package arena

import (
	"codecrashreview/backend/pkg/freellm"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strings"
)

var bot = uuid.MustParse("ec000000-0000-4000-8000-0000000000ff")

func (a api) generate(ctx context.Context, role, description string, cases [][3]int64) (string, result, error) {
	if a.model == "" {
		return "", result{}, refuse(503, "Configure a model to generate a solution.")
	}
	prompt := "You are " + role + ", competing in a Python arithmetic challenge. " + roles[role] + " " + syntax + " Return only the expression for the body of solve(a,b). No prose or Markdown."
	resp, e := a.llm.ChatCompletion(ctx, &freellm.ChatRequest{Model: a.model, Temperature: 0.2, MaxTokens: 512, Messages: []freellm.ChatMessage{{Role: "system", Content: prompt}, {Role: "user", Content: description}}})
	if e != nil || len(resp.Choices) == 0 {
		return "", result{}, refuse(502, "The configured model did not return a solution. Please retry.")
	}
	code := strings.TrimSpace(resp.Choices[0].Message.Content)
	if strings.HasPrefix(code, "```") {
		if i := strings.IndexByte(code, '\n'); i >= 0 {
			code = strings.TrimSpace(strings.TrimSuffix(code[i+1:], "```"))
		}
	}
	code = strings.TrimSpace(strings.TrimPrefix(code, "return "))
	res, e := grade(code, cases)
	if e != nil {
		return "", result{}, refuse(502, "The model returned an unsupported expression. Please retry.")
	}
	if !strings.HasPrefix(code, "def ") {
		code = "def solve(a, b):\n    return " + code
	}
	return code, res, nil
}
func (a api) suggest(c call) (any, error) {
	var role, description string
	var cases [][3]int64
	e := a.db.QueryRow(c.ctx, `select p.agent_persona,ch.description,ch.cases from matches m join arena_challenges ch on ch.id=m.challenge_id join match_participants p on p.match_id=m.id where m.id=$1 and p.user_id=$2 and m.status in ('pending','active') and p.status!='completed'`, c.Param("id"), c.user).Scan(&role, &description, &cases)
	if e != nil {
		return nil, e
	}
	code, _, e := a.generate(c.ctx, role, description, cases)
	return gin.H{"code": code}, e
}
func (a api) duel(c call) (any, error) {
	r, e := c.input()
	if e != nil {
		return nil, e
	}
	if e = persona(r.Persona); e != nil {
		return nil, e
	}
	if e = persona(r.Opponent); e != nil {
		return nil, e
	}
	if out, e := lookup(c, a.db, r, "practice"); e != nil || out != nil {
		return out, e
	}
	var desc string
	var cases [][3]int64
	e = a.db.QueryRow(c.ctx, `select description,cases from arena_challenges where id=$1`, r.ChallengeID).Scan(&desc, &cases)
	if e != nil {
		return nil, e
	}
	roles := []string{r.Persona, r.Opponent}
	users := []uuid.UUID{c.user, bot}
	var codes [2]string
	var grades [2]result
	for i, role := range roles {
		codes[i], grades[i], e = a.generate(c.ctx, role, desc, cases)
		if e != nil {
			return nil, e
		}
	}
	return a.transaction(c, func(tx pgx.Tx) (any, error) {
		if out, e := acquireRequest(c, tx, r, "practice"); out != nil || e != nil {
			return out, e
		}
		id, e := newMatch(c.ctx, tx, r.ChallengeID, "practice")
		if e != nil {
			return nil, e
		}
		_, e = tx.Exec(c.ctx, `update matches set started_at=now() where id=$1`, id)
		if e != nil {
			return nil, e
		}
		for i, role := range roles {
			if e = addPlayer(c.ctx, tx, id, users[i], role); e != nil {
				return nil, e
			}
			if e = saveSolution(c.ctx, tx, id, users[i], codes[i], grades[i]); e != nil {
				return nil, e
			}
		}
		if e = finish(c.ctx, tx, id, false); e != nil {
			return nil, e
		}
		return rememberRequest(c, tx, r, "practice", id)
	})
}
