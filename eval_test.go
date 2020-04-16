package expr

import (
	"testing"

	"github.com/stretchr/testify/require"

	pb "github.com/altipla-consulting/expr/testdata/foo"
)

func TestEvalSQL(t *testing.T) {
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
		root, filters, err := filters.parseQuery(test.query)
		require.NoError(t, err)
		cond, err := evalSQL(root, filters)
		require.NoError(t, err)

		require.Equal(t, cond.sql, test.expected, "test %v: [%v]", i, test.query)

		require.Len(t, cond.vals, len(test.vals))
		for j, val := range cond.vals {
			require.EqualValues(t, val, test.vals[j], "test %v, value %v: [%v]", i, j, test.query)
		}
	}
}

func TestMatcher(t *testing.T) {
	filters := Filters{
		IDParam("foo"),
	}

	query := `foo=3`
	matcher, err := filters.Matcher(query)
	require.NoError(t, err)

	data := map[string]interface{}{
		"foo": int64(3),
	}
	require.True(t, matcher(data))

	data = map[string]interface{}{
		"foo": int64(5),
	}
	require.False(t, matcher(data))
}

func TestMatcherEnum(t *testing.T) {
	filters := Filters{
		EnumParam("foo", pb.FooEnum_value),
	}

	query := `foo=FOOENUM_FIRST`
	matcher, err := filters.Matcher(query)
	require.NoError(t, err)

	data := map[string]interface{}{
		"foo": pb.FooEnum_FOOENUM_FIRST,
	}
	require.True(t, matcher(data))

	data = map[string]interface{}{
		"foo": pb.FooEnum_FOOENUM_SECOND,
	}
	require.False(t, matcher(data))
}
