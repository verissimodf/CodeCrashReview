package arena

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"codecrashreview/backend/pkg/freellm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type api struct {
	db    *pgxpool.Pool
	model string
	llm   *freellm.Client
}
type call struct {
	*gin.Context
	user uuid.UUID
	ctx  context.Context
}
type problem struct {
	status  int
	message string
}

func (p problem) Error() string           { return p.message }
func refuse(status int, msg string) error { return problem{status, msg} }

func Register(protected *gin.RouterGroup, db *pgxpool.Pool) {
	a := api{db, os.Getenv("DEFAULT_MODEL"), freellm.NewClient(strings.TrimRight(os.Getenv("FREELLM_BASE_URL"), "/"), os.Getenv("FREELLM_API_KEY"))}
	g := protected.Group("/arena")
	g.GET("/challenges", route(a.challenges))
	g.GET("/matches", route(a.list))
	g.POST("/matches", route(a.create))
	g.GET("/matches/:id", route(a.detail))
	g.POST("/matches/:id/join", route(a.join))
	g.POST("/matches/:id/submit", route(a.submit))
	g.POST("/matches/:id/suggest", route(a.suggest))
	g.POST("/duels", route(a.duel))
}
func route(f func(call) (any, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, _ := c.Get("user_id")
		user, ok := value.(uuid.UUID)
		if !ok || user == uuid.Nil {
			c.JSON(401, gin.H{"error": "Sign in to play."})
			return
		}
		if id := c.Param("id"); id != "" {
			if _, err := uuid.Parse(id); err != nil {
				c.JSON(400, gin.H{"error": "Invalid match ID."})
				return
			}
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
		data, err := f(call{c, user, c.Request.Context()})
		if err == nil {
			c.JSON(200, data)
			return
		}
		p := problem{500, "Could not complete this operation."}
		if errors.Is(err, pgx.ErrNoRows) {
			p = problem{404, "Challenge or match unavailable."}
		} else if !errors.As(err, &p) {
			log.Printf("arena operation %s failed: %v", c.FullPath(), err)
		}
		c.JSON(p.status, gin.H{"error": p.message})
	}
}

type request struct {
	ChallengeID uuid.UUID `json:"challenge_id"`
	Persona     string    `json:"persona"`
	Opponent    string    `json:"opponent_persona"`
	Code        string    `json:"code"`
}

func (c call) input() (request, error) {
	var r request
	if e := c.ShouldBindJSON(&r); e != nil {
		return r, refuse(400, "Invalid request body.")
	}
	return r, nil
}

var roles = map[string]string{"sentinel": "Check sign and boundary handling.", "optimizer": "Prefer a short efficient expression.", "refactor": "Prefer readable expressions.", "guardian": "Check edge cases.", "docubot": "Favor clarity.", "polyglot": "Use idiomatic Python arithmetic."}

func persona(p string) error {
	if _, ok := roles[p]; !ok {
		return refuse(400, "Select a valid persona.")
	}
	return nil
}
func (a api) challenges(c call) (any, error) {
	var data json.RawMessage
	err := a.db.QueryRow(c.ctx, `select coalesce(jsonb_agg(to_jsonb(k) order by id),'[]') from (select id,title,description || ' ' || $1::text as description,starter_code from arena_challenges) k`, syntax).Scan(&data)
	return data, err
}

const summary = `jsonb_build_object('id',m.id,'status',m.status,'mode',m.mode,'challenge_id',m.challenge_id,'winner_id',m.winner_id,
 'joined',exists(select from match_participants p where p.match_id=m.id and p.user_id=$1),
 'joinable',m.mode='ranked' and m.status='pending' and not exists(select from match_participants p where p.match_id=m.id and p.user_id=$1))`

func (a api) list(c call) (any, error) {
	var data json.RawMessage
	err := a.db.QueryRow(c.ctx, `select coalesce(jsonb_agg(item order by created_at desc),'[]') from (select `+summary+` item,m.created_at from matches m join arena_challenges ch on ch.id=m.challenge_id
 where m.mode='ranked' and m.status='pending' or exists(select from match_participants p where p.match_id=m.id and p.user_id=$1) order by m.created_at desc limit 100) k`, c.user).Scan(&data)
	return data, err
}
func (a api) detail(c call) (any, error) {
	var data json.RawMessage
	err := a.db.QueryRow(c.ctx, `select `+summary+` || jsonb_build_object(
 'participants',(select jsonb_agg(jsonb_build_object('user_id',u.id,'username',u.username,'persona',p.agent_persona,'score',p.score,'status',p.status) order by p.created_at,p.id) from match_participants p join users u on u.id=p.user_id where p.match_id=m.id),
 'submissions',(select coalesce(jsonb_agg(jsonb_build_object('username',u.username,'score',s.score,'feedback',s.feedback,'files',s.files) order by s.submitted_at,s.id),'[]') from submissions s join match_participants p on p.id=s.participant_id join users u on u.id=p.user_id where s.match_id=m.id and (m.status='completed' or p.user_id=$1)))
 from matches m join arena_challenges ch on ch.id=m.challenge_id where m.id=$2 and (m.mode='ranked' and m.status='pending' or exists(select from match_participants p where p.match_id=m.id and p.user_id=$1))`, c.user, c.Param("id")).Scan(&data)
	return data, err
}
func (a api) transaction(c call, f func(pgx.Tx) (any, error)) (any, error) {
	tx, err := a.db.Begin(c.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	out, err := f(tx)
	if err != nil {
		return nil, err
	}
	return out, tx.Commit(c.ctx)
}
func addPlayer(ctx context.Context, tx pgx.Tx, id, user uuid.UUID, role string) error {
	_, err := tx.Exec(ctx, `insert into match_participants(match_id,user_id,agent_persona,rating_before) select $1,id,$3,rating from users where id=$2`, id, user, role)
	return err
}
func newMatch(ctx context.Context, tx pgx.Tx, ch uuid.UUID, mode string) (uuid.UUID, error) {
	var id uuid.UUID
	e := tx.QueryRow(ctx, `insert into matches(mode,challenge_id,max_turns) select $2,id,1 from arena_challenges where id=$1 returning id`, ch, mode).Scan(&id)
	return id, e
}
func (a api) create(c call) (any, error) {
	r, e := c.input()
	if e != nil {
		return nil, e
	}
	if e = persona(r.Persona); e != nil {
		return nil, e
	}
	return a.transaction(c, func(tx pgx.Tx) (any, error) {
		if out, e := acquireRequest(c, tx, r, "ranked"); out != nil || e != nil {
			return out, e
		}
		id, e := newMatch(c.ctx, tx, r.ChallengeID, "ranked")
		if e != nil {
			return nil, e
		}
		if e = addPlayer(c.ctx, tx, id, c.user, r.Persona); e != nil {
			return nil, e
		}
		return rememberRequest(c, tx, r, "ranked", id)
	})
}
func lockMatch(c call, tx pgx.Tx) (uuid.UUID, string, [][3]int64, error) {
	var id uuid.UUID
	var status string
	var cases [][3]int64
	err := tx.QueryRow(c.ctx, `select m.id,m.status,ch.cases from matches m join arena_challenges ch on ch.id=m.challenge_id where m.id=$1 and mode='ranked' for update of m`, c.Param("id")).Scan(&id, &status, &cases)
	return id, status, cases, err
}
func (a api) join(c call) (any, error) {
	r, e := c.input()
	if e != nil {
		return nil, e
	}
	if e = persona(r.Persona); e != nil {
		return nil, e
	}
	return a.transaction(c, func(tx pgx.Tx) (any, error) {
		id, status, _, e := lockMatch(c, tx)
		if e != nil {
			return nil, e
		}
		var joined bool
		e = tx.QueryRow(c.ctx, `select exists(select from match_participants where user_id=$1 and match_id=$2)`, c.user, id).Scan(&joined)
		if e != nil {
			return nil, e
		}
		out := gin.H{"id": id}
		if joined {
			return out, nil
		}
		if status != "pending" {
			return nil, refuse(409, "This match already has two players.")
		}
		if e = addPlayer(c.ctx, tx, id, c.user, r.Persona); e != nil {
			return nil, e
		}
		_, e = tx.Exec(c.ctx, `update matches set status='active',started_at=now() where id=$1`, id)
		return out, e
	})
}
