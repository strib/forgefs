package filter

// Node represents a node in the filter tree.  It can either have a
// non-nil constraint (making it a leaf in the filter tree), or left
// and a right non-nil nodes combined with a boolean operator.
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
