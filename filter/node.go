package filter

type Node struct {
	Op    BooleanOperator
	Left  *Node
	Right *Node

	Constraint *Constraint
}

func (n *Node) String() string {
	if n.Constraint != nil {
		return n.Constraint.String()
	}

	return "(" + n.Left.String() + " " + n.Op.String() + " " +
		n.Right.String() + ")"
}
