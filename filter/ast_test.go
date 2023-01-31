package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAST(t *testing.T) {
	n, err := Parse("a=10")
	require.NoError(t, err)
	t.Log(n)
	n, err = Parse("expansion=MM")
	require.NoError(t, err)
	t.Log(n)
	n, err = Parse("e=10.5")
	require.NoError(t, err)
	t.Log(n)
	n, err = Parse("a=10,(e=10.5^a=3)")
	require.NoError(t, err)
	t.Log(n)
	n, err = Parse("a=10^e=10.5+a=3^e=11.5")
	require.NoError(t, err)
	t.Log(n)
	n, err = Parse("sas=10:15+a=5.0:7.0")
	require.NoError(t, err)
	t.Log(n)
}
