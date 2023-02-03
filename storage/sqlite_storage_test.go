package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strib/forgefs/filter"
)

func TestFilterNodeToSQLConstraint(t *testing.T) {
	checkFilter := func(toParse, expectedSQL string) {
		n, err := filter.Parse(toParse)
		require.NoError(t, err)
		s, err := filterNodeToSQLConstraint(n)
		require.NoError(t, err)
		require.Equal(t, expectedSQL, s)
	}
	checkFilter("a=10", "a = 10")
	checkFilter("e=20.1", "e = 20.100000")
	checkFilter("f=10,d=2", "(f = 10 AND d = 2)")
	checkFilter("f=10^d=2", "(f = 10 OR d = 2)")
	checkFilter(
		"house=lOgOs",
		"(house1 = \"Logos\" OR house2 = \"Logos\" OR house3 = \"Logos\")")
	checkFilter("expansion=CotA", "expansion = \"CALL_OF_THE_ARCHONS\"")
	checkFilter("a=10:", "a >= 10")
	checkFilter("c=:10", "c <= 10")
	checkFilter("r=1.2:1.7", "(r >= 1.2 AND r <= 1.7)")
	// TODO(#15): The AND should take precedence here.
	checkFilter(
		"sas=80:85+aerc=50:^a=5",
		"((sas >= 80 AND sas <= 85) AND (aerc >= 50 OR a = 5))")
}
