package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAST(t *testing.T) {
	n, err := Parse("a=10")
	require.NoError(t, err)
	require.NotNil(t, n.Constraint)
	require.IsType(t, AmberControl{}, n.Constraint.Var)
	require.Equal(t, 10, *n.Constraint.Value.Int)

	n, err = Parse("expansion=MM")
	require.NoError(t, err)
	require.NotNil(t, n.Constraint)
	require.IsType(t, Expansion{}, n.Constraint.Var)
	require.Equal(t, "MM", *n.Constraint.Value.String)

	n, err = Parse("e=10.5")
	require.NoError(t, err)
	require.NotNil(t, n.Constraint)
	require.IsType(t, ExpectedAmber{}, n.Constraint.Var)
	require.Equal(t, 10.5, *n.Constraint.Value.Float)

	n, err = Parse("a=10,(e=10.5^a=3)")
	require.NoError(t, err)
	require.Nil(t, n.Constraint)
	require.NotNil(t, n.Left.Constraint)
	require.IsType(t, AmberControl{}, n.Left.Constraint.Var)
	require.Equal(t, 10, *n.Left.Constraint.Value.Int)
	require.IsType(t, And{}, n.Op)
	require.Nil(t, n.Right.Constraint)
	require.NotNil(t, n.Right.Left.Constraint)
	require.IsType(t, ExpectedAmber{}, n.Right.Left.Constraint.Var)
	require.Equal(t, 10.5, *n.Right.Left.Constraint.Value.Float)
	require.IsType(t, Or{}, n.Right.Op)
	require.NotNil(t, n.Right.Right.Constraint)
	require.IsType(t, AmberControl{}, n.Right.Right.Constraint.Var)
	require.Equal(t, 3, *n.Right.Right.Constraint.Value.Int)

	n, err = Parse("a=10^e=10.5+a=3^e=11.5")
	require.NoError(t, err)
	require.Nil(t, n.Constraint)
	require.NotNil(t, n.Left.Constraint)
	require.IsType(t, AmberControl{}, n.Left.Constraint.Var)
	require.Equal(t, 10, *n.Left.Constraint.Value.Int)
	require.IsType(t, Or{}, n.Op)
	require.Nil(t, n.Right.Constraint)
	require.NotNil(t, n.Right.Left.Constraint)
	// TODO(#15): This doesn't seem right, since the AND should be at
	// the lowest level of the tree since it has precedence. Hmm.
	require.IsType(t, ExpectedAmber{}, n.Right.Left.Constraint.Var)
	require.Equal(t, 10.5, *n.Right.Left.Constraint.Value.Float)
	require.IsType(t, And{}, n.Right.Op)
	require.Nil(t, n.Right.Right.Constraint)
	require.IsType(t, AmberControl{}, n.Right.Right.Left.Constraint.Var)
	require.Equal(t, 3, *n.Right.Right.Left.Constraint.Value.Int)
	require.IsType(t, Or{}, n.Right.Right.Op)
	require.IsType(t, ExpectedAmber{}, n.Right.Right.Right.Constraint.Var)
	require.Equal(t, 11.5, *n.Right.Right.Right.Constraint.Value.Float)

	n, err = Parse("sas=10:15+a=5.0:7.0")
	require.NoError(t, err)
	require.Nil(t, n.Constraint)
	require.NotNil(t, n.Left.Constraint)
	require.IsType(t, SAS{}, n.Left.Constraint.Var)
	require.Equal(t, "10", n.Left.Constraint.Value.MinString())
	require.Equal(t, "15", n.Left.Constraint.Value.MaxString())
	require.IsType(t, And{}, n.Op)
	require.IsType(t, AmberControl{}, n.Right.Constraint.Var)
	require.Equal(t, "5.0", n.Right.Constraint.Value.MinString())
	require.Equal(t, "7.0", n.Right.Constraint.Value.MaxString())
}
