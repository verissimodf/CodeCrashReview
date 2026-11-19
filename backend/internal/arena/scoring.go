package arena

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"math"
)

func saveSolution(ctx context.Context, tx pgx.Tx, id, user uuid.UUID, code string, res result) error {
	files := []map[string]string{{"path": "solution.py", "content": code}}
	tag, e := tx.Exec(ctx, `insert into submissions(match_id,participant_id,agent_persona,turn,files,score,status,evaluated_at,feedback)
 select match_id,id,agent_persona,1,$3,$4::double precision,case when $4::double precision=100 then 'passed' else 'failed' end,now(),$5 from match_participants where match_id=$1 and user_id=$2`, id, user, files, res.Score, res.Feedback)
	if e != nil {
		return e
	}
	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	_, e = tx.Exec(ctx, `update match_participants set score=$3,status='completed' where match_id=$1 and user_id=$2`, id, user, res.Score)
	return e
}

type player struct {
	User            uuid.UUID
	Persona, Status string
	Rating, Score   int
	Deviation       float64
}

func finish(ctx context.Context, tx pgx.Tx, id uuid.UUID, ranked bool) error {
	// Lock both user records in a consistent order. Concurrent matches settle against current ratings.
	rows, e := tx.Query(ctx, `select u.id,p.agent_persona,p.status,u.rating,p.score::int,u.rating_deviation from match_participants p join users u on u.id=p.user_id where match_id=$1 order by u.id for update of u`, id)
	if e != nil {
		return e
	}
	players, e := pgx.CollectRows(rows, pgx.RowToStructByPos[player])
	if e != nil {
		return e
	}
	if len(players) != 2 {
		return errors.New("arena requires two players")
	}
	for _, p := range players {
		if p.Status != "completed" {
			return nil
		}
	}
	var winner *uuid.UUID
	var role *string
	for i, p := range players {
		other := players[1-i]
		actual := 0.5
		win, loss, draw := 0, 0, 0
		switch {
		case p.Score > other.Score:
			actual = 1
			win = 1
			winner = &players[i].User
			role = &players[i].Persona
		case p.Score < other.Score:
			actual = 0
			loss = 1
		default:
			draw = 1
		}
		if !ranked {
			continue
		}
		expected := 1 / (1 + math.Pow(10, float64(other.Rating-p.Rating)/400))
		delta := int(math.Round(32 * (actual - expected)))
		after := p.Rating + delta
		_, e = tx.Exec(ctx, `update users set rating=$2,peak_rating=greatest(peak_rating,$2),total_matches=total_matches+1,wins=wins+$3,losses=losses+$4,draws=draws+$5 where id=$1`, p.User, after, win, loss, draw)
		if e != nil {
			return e
		}
		_, e = tx.Exec(ctx, `update match_participants set rating_before=$3,rating_after=$4,rating_change=$5 where match_id=$1 and user_id=$2`, id, p.User, p.Rating, after, delta)
		if e != nil {
			return e
		}
		_, e = tx.Exec(ctx, `insert into rating_history(user_id,match_id,rating_before,rating_after,rating_change,deviation_before,deviation_after) values($1,$2,$3,$4,$5,$6,$6)`, p.User, id, p.Rating, after, delta, p.Deviation)
		if e != nil {
			return e
		}
	}
	_, e = tx.Exec(ctx, `update matches set status='completed',winner_id=$2,winner_persona=$3,ended_at=now(),current_turn=1 where id=$1`, id, winner, role)
	return e
}
func (a api) submit(c call) (any, error) {
	r, e := c.input()
	if e != nil {
		return nil, e
	}
	return a.transaction(c, func(tx pgx.Tx) (any, error) {
		id, status, cases, e := lockMatch(c, tx)
		if e != nil {
			return nil, e
		}
		var previous *string
		var old result
		e = tx.QueryRow(c.ctx, `select s.files->0->>'content',coalesce(s.score,0)::int,coalesce(s.feedback,'') from match_participants p left join submissions s on s.participant_id=p.id where p.match_id=$1 and p.user_id=$2`, id, c.user).Scan(&previous, &old.Score, &old.Feedback)
		if e != nil {
			return nil, e
		}
		if previous != nil {
			if *previous == r.Code {
				return old, nil
			}
			return nil, refuse(409, "Your final submission has already been recorded.")
		}
		if status != "active" {
			return nil, refuse(409, "Both players must join before submitting.")
		}
		res, e := grade(r.Code, cases)
		if e != nil {
			return nil, refuse(422, e.Error())
		}
		if e = saveSolution(c.ctx, tx, id, c.user, r.Code, res); e != nil {
			return nil, e
		}
		return res, finish(c.ctx, tx, id, true)
	})
}
func fingerprint(r request, mode string) string {
	return mode + "/" + r.ChallengeID.String() + "/" + r.Persona + "/" + r.Opponent
}
func lookup(c call, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, r request, mode string) (any, error) {
	key := c.GetHeader("Idempotency-Key")
	if key == "" {
		return nil, nil
	}
	if len(key) > 100 {
		return nil, refuse(400, "Idempotency key exceeds 100 bytes.")
	}
	var id uuid.UUID
	var stored string
	e := q.QueryRow(c.ctx, `select match_id,fingerprint from arena_requests where user_id=$1 and request_key=$2`, c.user, key).Scan(&id, &stored)
	if errors.Is(e, pgx.ErrNoRows) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	if stored != fingerprint(r, mode) {
		return nil, refuse(409, "Use a new key for a different match.")
	}
	return gin.H{"id": id}, nil
}
func acquireRequest(c call, tx pgx.Tx, r request, mode string) (any, error) {
	key := c.GetHeader("Idempotency-Key")
	if key == "" {
		return nil, nil
	}
	if len(key) > 100 {
		return nil, refuse(400, "Idempotency key exceeds 100 bytes.")
	}
	if _, e := tx.Exec(c.ctx, `select pg_advisory_xact_lock(hashtextextended($1, 0))`, c.user.String()+key); e != nil {
		return nil, e
	}
	return lookup(c, tx, r, mode)
}
func rememberRequest(c call, tx pgx.Tx, r request, mode string, id uuid.UUID) (any, error) {
	if key := c.GetHeader("Idempotency-Key"); key != "" {
		if _, e := tx.Exec(c.ctx, `insert into arena_requests(user_id,request_key,fingerprint,match_id) values($1,$2,$3,$4)`, c.user, key, fingerprint(r, mode), id); e != nil {
			return nil, e
		}
	}
	return gin.H{"id": id}, nil
}
