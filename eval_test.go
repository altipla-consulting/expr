package expr

import (
	"testing"

	"github.com/stretchr/testify/require"

	pb "github.com/altipla-consulting/expr/testdata/foo"
)

func TestToSQL(t *testing.T) {
	filters := Filters{
		IDParam("foo"),
		EnumParam("bar", pb.FooEnum_value),
	}

	tests := []struct {
		query    string
		expected string
		vals     []interface{}
	}{
		{
			query:    ``,
			expected: ``,
			vals:     nil,
		},
		{
			query:    `foo=3`,
			expected: `(foo = ?)`,
			vals:     []interface{}{3},
		},
		{
			query:    `bar=FOOENUM_FIRST`,
			expected: `(bar = ?)`,
			vals:     []interface{}{"FOOENUM_FIRST"},
		},
		{
			query:    `foo!=4 bar=FOOENUM_FIRST`,
			expected: `(foo != ?) AND (bar = ?)`,
			vals:     []interface{}{4, "FOOENUM_FIRST"},
		},
		{
			query:    `-foo=3`,
			expected: `(NOT foo = ?)`,
			vals:     []interface{}{3},
		},
	}
	for i, test := range tests {
		sql, vals, err := filters.ToSQL(test.query)
		require.NoError(t, err)

		require.Equal(t, sql, test.expected, "test %v: [%v]", i, test.query)

		require.Len(t, vals, len(test.vals))
		for j, val := range vals {
			require.EqualValues(t, val, test.vals[j], "test %v, value %v: [%v]", i, j, test.query)
		}
	}
}
