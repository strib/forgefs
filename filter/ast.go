package filter

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
)

type Var interface {
	String() string
}

type AmberControl struct {
	Value string `@"a"`
}

func (a AmberControl) String() string {
	return "a"
}

type ExpectedAmber struct {
	Value string `@"e"`
}

func (a ExpectedAmber) String() string {
	return "e"
}

type ArtifactControl struct {
	Value string `@"r"`
}

func (a ArtifactControl) String() string {
	return "r"
}

type CreatureControl struct {
	Value string `@"c"`
}

func (c CreatureControl) String() string {
	return "c"
}

type Efficiency struct {
	Value string `@"f"`
}

func (e Efficiency) String() string {
	return "f"
}

type Disruption struct {
	Value string `@"d"`
}

func (e Disruption) String() string {
	return "d"
}

type SAS struct {
	Value string `@"sas"`
}

func (s SAS) String() string {
	return "sas"
}

type Expansion struct {
	Value string `@"expansion" | @"set"`
}

func (e Expansion) String() string {
	return "expansion"
}

type House struct {
	Value string `@"house"`
}

func (h House) String() string {
	return "house"
}

type AERC struct {
	Value string `@"aerc"`
}

func (a AERC) String() string {
	return "aerc"
}

type Op struct {
	Equal bool `@"="?`
}

type Value struct {
	Range  []string `@(Float|Int)* @":" @(Float|Int)*`
	Float  *float64 `| @Float`
	Int    *int     `| @Int`
	String *string  `| @Ident`
}

func (v *Value) MinString() string {
	if len(v.Range) > 1 && v.Range[0] != ":" {
		return v.Range[0]
	}
	return ""
}

func (v *Value) MaxString() string {
	last := len(v.Range) - 1
	if last >= 1 && v.Range[last] != ":" {
		return v.Range[last]
	}
	return ""
}

type Constraint struct {
	Var   Var    `@@`
	Op    *Op    `@@`
	Value *Value `@@`
}

func (c *Constraint) String() string {
	if len(c.Value.Range) != 0 {
		return fmt.Sprintf("[%s = %s:%s]",
			c.Var, c.Value.MinString(), c.Value.MaxString())
	} else if c.Value.Float != nil {
		return fmt.Sprintf("[%s = %f]", c.Var, *c.Value.Float)
	} else if c.Value.Int != nil {
		return fmt.Sprintf("[%s = %d]", c.Var, *c.Value.Int)
	}
	return fmt.Sprintf("[%s = %s]", c.Var, *c.Value.String)
}

type BooleanOperator interface {
	Eval(s *Statement) string
	String() string
}

type And struct {
	Value string `@"," | @"+"`
}

func (a And) String() string {
	return "AND"
}

func (a And) Eval(s *Statement) string {
	return a.String() + " " + s.String()
}

type Or struct {
	Value string `@"^"`
}

func (o Or) String() string {
	return "OR"
}

func (o Or) Eval(s *Statement) string {
	return o.String() + " " + s.String()
}

type OpRight struct {
	Op    BooleanOperator `@@`
	Right *Statement      `@@`
}

func (or *OpRight) String() string {
	return or.Op.Eval(or.Right)
}

type Expression struct {
	Constraint   *Constraint `@@`
	Substatement *Statement  `| "(" @@ ")"`
}

func (e *Expression) String() string {
	if e.Constraint != nil {
		return e.Constraint.String()
	}
	return "(" + e.Substatement.String() + ")"
}

func (e *Expression) MakeTree() *Node {
	if e.Constraint != nil {
		return &Node{
			Constraint: e.Constraint,
		}
	}

	return e.Substatement.MakeTree()
}

type ExprChain struct {
	Left    *Expression `@@`
	OpRight []*OpRight  `@@*`
}

func (ec *ExprChain) String() string {
	s := make([]string, 1+len(ec.OpRight))
	s[0] = ec.Left.String()
	for i, r := range ec.OpRight {
		s[i+1] = r.String()
	}
	return strings.Join(s, " ")
}

func (ex *ExprChain) MakeTree() *Node {
	if len(ex.OpRight) == 0 {
		return ex.Left.MakeTree()
	}

	// AND has a higher precedence than OR.  Which means, in a chain,
	// we need to make sure to output the first OR node at the top of
	// the tree.
	splitAt := 0
	for i, r := range ex.OpRight {
		if _, ok := r.Op.(Or); ok {
			splitAt = i
			break
		}
	}

	r := ex.OpRight[splitAt]
	return &Node{
		Op: r.Op,
		Left: (&ExprChain{
			Left:    ex.Left,
			OpRight: ex.OpRight[:splitAt],
		}).MakeTree(),
		Right: r.Right.Expr.MakeTree(),
	}
}

type Statement struct {
	Expr *ExprChain `@@`
}

func (s *Statement) String() string {
	return s.Expr.String()
}

func (s *Statement) MakeTree() *Node {
	return s.Expr.MakeTree()
}

var parser = participle.MustBuild[Statement](
	participle.Union[Var](
		AmberControl{},
		ExpectedAmber{},
		ArtifactControl{},
		CreatureControl{},
		Efficiency{},
		Disruption{},
		SAS{},
		AERC{},
		Expansion{},
		House{},
	),
	participle.Union[BooleanOperator](
		And{},
		Or{},
	),
)

func Parse(s string) (*Node, error) {
	stmt, err := parser.ParseString("", s)
	if err != nil {
		return nil, err
	}

	return stmt.MakeTree(), nil
}
