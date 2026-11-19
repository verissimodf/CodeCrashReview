package arena

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/big"
	"regexp"
	"strings"
)

const syntax = "Use a, b, decimal integers, +, -, *, parentheses, abs(x), min(x,y) or max(x,y)."

var integer = regexp.MustCompile("^(0|[1-9][0-9]{0,8})$")
var alphabet = regexp.MustCompile("^[a-z0-9 +*(),-]+$")
var function = regexp.MustCompile("^def solve\\(a, *b\\):\\r?\\n[ \\t]+return[ \\t]+([^\\n]+)$")
var badExpression = errors.New(syntax)

func parse(code string) (ast.Expr, error) {
	code = strings.TrimSpace(code)
	if m := function.FindStringSubmatch(code); m != nil {
		code = m[1]
	}
	if len(code) > 256 || !alphabet.MatchString(code) || strings.Contains(code, "/*") {
		return nil, badExpression
	}
	expr, err := parser.ParseExpr(code)
	if err != nil {
		return nil, badExpression
	}
	return expr, nil
}

func evaluate(expr ast.Expr, a, b int64) (*big.Int, error) {
	switch n := expr.(type) {
	case *ast.Ident:
		if n.Name == "a" {
			return big.NewInt(a), nil
		}
		if n.Name == "b" {
			return big.NewInt(b), nil
		}
	case *ast.BasicLit:
		if integer.MatchString(n.Value) {
			v, _ := new(big.Int).SetString(n.Value, 10)
			return v, nil
		}
	case *ast.ParenExpr:
		return evaluate(n.X, a, b)
	case *ast.UnaryExpr:
		v, e := evaluate(n.X, a, b)
		if e != nil {
			return nil, e
		}
		if n.Op == token.SUB {
			return v.Neg(v), nil
		}
		if n.Op == token.ADD {
			return v, nil
		}
	case *ast.BinaryExpr:
		x, e := evaluate(n.X, a, b)
		if e != nil {
			return nil, e
		}
		y, e := evaluate(n.Y, a, b)
		if e != nil {
			return nil, e
		}
		switch n.Op {
		case token.ADD:
			return x.Add(x, y), nil
		case token.SUB:
			return x.Sub(x, y), nil
		case token.MUL:
			return x.Mul(x, y), nil
		}
	case *ast.CallExpr:
		f, ok := n.Fun.(*ast.Ident)
		if !ok || len(n.Args) < 1 || len(n.Args) > 2 {
			return nil, badExpression
		}
		x, e := evaluate(n.Args[0], a, b)
		if e != nil {
			return nil, e
		}
		if len(n.Args) == 1 && f.Name == "abs" {
			return x.Abs(x), nil
		}
		if len(n.Args) == 2 {
			y, e := evaluate(n.Args[1], a, b)
			if e != nil {
				return nil, e
			}
			switch f.Name {
			case "min":
				if x.Cmp(y) <= 0 {
					return x, nil
				}
				return y, nil
			case "max":
				if x.Cmp(y) >= 0 {
					return x, nil
				}
				return y, nil
			}
		}
	}
	return nil, badExpression
}

type result struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

func grade(code string, cases [][3]int64) (result, error) {
	expr, err := parse(code)
	if err != nil {
		return result{}, err
	}
	if len(cases) == 0 {
		return result{}, errors.New("Missing acceptance cases.")
	}
	passed := 0
	for _, c := range cases {
		v, e := evaluate(expr, c[0], c[1])
		if e != nil {
			return result{}, e
		}
		if v.Cmp(big.NewInt(c[2])) == 0 {
			passed++
		}
	}
	return result{100 * passed / len(cases), fmt.Sprintf("%d of %d acceptance cases passed.", passed, len(cases))}, nil
}
