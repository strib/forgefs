package filter

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
)

// This file describes a simple grammar for specifying deck-filtering
// rules in a string. Each rule is a simple constraint like
// "var=value" (to specify an exact match) or "var=[min]:[max]" to
// specify a half-range or a full-range.  These constraints can be
// combined with AND logic (using `,` or `+` between constraints) or
// OR logic (using `^`).  Parenthesis can also be used to force
// precedence.

// Var represents a variable type that is being constrained.
type Var interface {
	String() string
}

// AmberControl represents the amber control variable type.
type AmberControl struct {
	Value string `@"a"`
}

func (a AmberControl) String() string {
	return "a"
}

// ExpectedAmber represents the expected amber variable type.
type ExpectedAmber struct {
	Value string `@"e"`
}

func (a ExpectedAmber) String() string {
	return "e"
}

// ArtifactControl represents the artifact control variable type.
type ArtifactControl struct {
	Value string `@"r"`
}

func (a ArtifactControl) String() string {
	return "r"
}

// CreatureControl represents the creature control variable type.
type CreatureControl struct {
	Value string `@"c"`
}

func (c CreatureControl) String() string {
	return "c"
}

// Efficiency represents the efficiency variable type.
type Efficiency struct {
	Value string `@"f"`
}

func (e Efficiency) String() string {
	return "f"
}

// Disruption represents the disruption variable type.
type Disruption struct {
	Value string `@"d"`
}

func (e Disruption) String() string {
	return "d"
}

// SAS represents the SAS variable type.
type SAS struct {
	Value string `@"sas"`
}

func (s SAS) String() string {
	return "sas"
}

// Expansion represents the expansion (or set) variable type.
type Expansion struct {
	Value string `@"expansion" | @"set"`
}

func (e Expansion) String() string {
	return "expansion"
}

// House represents the house variable type.
type House struct {
	Value string `@"house"`
}

func (h House) String() string {
	return "house"
}

// AERC represents the AERC variable type.
type AERC struct {
	Value string `@"aerc"`
}

func (a AERC) String() string {
	return "aerc"
}

// Op represent the operation specified by a constraint.  Currently
// can only be equals.
type Op struct {
	Equal bool `@"="?`
}

// Value represents the value of a constraint.  Currently could be a
// range, float, int, or string.
type Value struct {
	Range  []string `@(Float|Int)* @":" @(Float|Int)*`
	Float  *float64 `| @Float`
	Int    *int     `| @Int`
	String *string  `| @Ident`
}

// MinString returns the minimum value for the range, if this value is
// a range.
func (v *Value) MinString() string {
	if len(v.Range) > 1 && v.Range[0] != ":" {
		return v.Range[0]
	}
	return ""
}

// MaxString returns the maximum value for the range, if this value is
// a range.
func (v *Value) MaxString() string {
	last := len(v.Range) - 1
	if last >= 1 && v.Range[last] != ":" {
		return v.Range[last]
	}
	return ""
}

// Constraint represents a single constraint.
type Constraint struct {
	Var   Var    `@@`
	Op    *Op    `@@`
	Value *Value `@@`
}

func (c *Constraint) String() string {
	if len(c.Value.Range) != 0 {
		return fmt.Sprintf("[%s = %s:%s]",
			c.Var, c.Value.MinString(), c.Value.MaxString())
	}
	if c.Value.Float != nil {
		return fmt.Sprintf("[%s = %f]", c.Var, *c.Value.Float)
	}
	if c.Value.Int != nil {
		return fmt.Sprintf("[%s = %d]", c.Var, *c.Value.Int)
	}
	return fmt.Sprintf("[%s = %s]", c.Var, *c.Value.String)
}

// BooleanOperator represents the operator that combines two
// expressions.  Currently can be "and" or "or".
type BooleanOperator interface {
	Eval(s *Statement) string
	String() string
}

// And represents the "and" operator between two expressions.
type And struct {
	Value string `@"," | @"+"`
}

func (a And) String() string {
	return "AND"
}

// Eval helps build up a string for the full statement.
func (a And) Eval(s *Statement) string {
	return a.String() + " " + s.String()
}

// Or represents the "or" operator between two expressions.
type Or struct {
	Value string `@"^"`
}

func (o Or) String() string {
	return "OR"
}

// Eval helps build up a string for the full statement.
func (o Or) Eval(s *Statement) string {
	return o.String() + " " + s.String()
}

// OpRight represents the right side of an statement, including the
// operator.
type OpRight struct {
	Op    BooleanOperator `@@`
	Right *Statement      `@@`
}

func (or *OpRight) String() string {
	return or.Op.Eval(or.Right)
}

// Expression is either a single constraint or a parenthetical
// statement.
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

// MakeTree builds up a filter tree for the expression.
func (e *Expression) MakeTree() *Node {
	if e.Constraint != nil {
		return &Node{
			Constraint: e.Constraint,
		}
	}

	return e.Substatement.MakeTree()
}

// ExprChain represents a full statement, which consists of at least
// one expression followed by zero or more other expressions combined
// with boolean operators.
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

// MakeTree builds up a filter tree for the expression chain.
func (ec *ExprChain) MakeTree() *Node {
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

// Statement represents a statement specifying a full filter rule.
type Statement struct {
	Expr *ExprChain `@@`
}

func (s *Statement) String() string {
	return s.Expr.String()
}

// MakeTree builds a filter tree for the statement.
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

// Parse turns a string matching the above grammar into a filter tree.
func Parse(s string) (*Node, error) {
	stmt, err := parser.ParseString("", s)
	if err != nil {
		return nil, err
	}

	return stmt.MakeTree(), nil
}
