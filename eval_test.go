package expr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pb "github.com/altipla-consulting/expr/testdata/foo"
)

func TestEvalSQL(t *testing.T) {
	filters := Filters{
		IDParam("id"),
		EnumParam("enum", pb.FooEnum_value),
		TimestampParam("ts"),
		BoolParam("boolUppercase"),
	}

	tests := []struct {
		query    string
		expected string
		vals     []interface{}
	}{
		{
			query:    `id=3`,
			expected: `(id = ?)`,
			vals:     []interface{}{3},
		},
		{
			query:    `enum=FOOENUM_FIRST`,
			expected: `(enum = ?)`,
			vals:     []interface{}{"FOOENUM_FIRST"},
		},
		{
			query:    `id!=4 enum=FOOENUM_FIRST`,
			expected: `(id != ?) AND (enum = ?)`,
			vals:     []interface{}{4, "FOOENUM_FIRST"},
		},
		{
			query:    `-id=3`,
			expected: `(NOT id = ?)`,
			vals:     []interface{}{3},
		},
		{
			query:    `ts:*`,
			expected: `(ts IS NOT NULL)`,
			vals:     []interface{}{},
		},
		{
			query:    `-ts:*`,
			expected: `(ts IS NULL)`,
			vals:     []interface{}{},
		},
		{
			query:    `boolUppercase=TRUE`,
			expected: `(bool_uppercase = ?)`,
			vals:     []interface{}{true},
		},
		{
			query:    `boolUppercase=true`,
			expected: `(bool_uppercase = ?)`,
			vals:     []interface{}{true},
		},
		{
			query:    `ts>"2019-03-02"`,
			expected: `(ts > ?)`,
			vals:     []interface{}{time.Date(2019, time.March, 2, 0, 0, 0, 0, time.UTC)},
		},
		{
			query:    `ts>"2019-03-02T14:15:16Z"`,
			expected: `(ts > ?)`,
			vals:     []interface{}{time.Date(2019, time.March, 2, 14, 15, 16, 0, time.UTC)},
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

	matcher, err := filters.Matcher(`foo=3`)
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

	matcher, err := filters.Matcher(`foo=FOOENUM_FIRST`)
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

func TestMatcherTimestampExists(t *testing.T) {
	filters := Filters{
		TimestampParam("foo"),
	}

	matcher, err := filters.Matcher(`foo:*`)
	require.NoError(t, err)

	data := map[string]interface{}{
		"foo": time.Now(),
	}
	require.True(t, matcher(data))

	data = make(map[string]interface{})
	require.False(t, matcher(data))
}
