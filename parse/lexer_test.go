package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	filters := []struct {
		query    string
		expected []item
	}{
		{
			query: ``,
			expected: []item{
				{itemAnd, ""},
				{itemEOF, ""},
			},
		},
		{
			query: `foo=3`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, "="}, {itemNumber, "3"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo=MY_CONSTANT`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, "="}, {itemConstant, "MY_CONSTANT"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo:3 bar:"hola"`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, ":"}, {itemNumber, "3"},
				{itemField, "bar"}, {itemOperator, ":"}, {itemString, `"hola"`},
				{itemEOF, ""},
			},
		},
		{
			query: `foo.bar=3`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo.bar"}, {itemOperator, "="}, {itemNumber, "3"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo:*`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, ":"}, {itemHasContent, ""},
				{itemEOF, ""},
			},
		},
		// {
		//  query: `-foo:3`,
		//  expected: []item{
		//    {itemEOF, ""},
		//  },
		// },
		// {
		//  query: `NOT foo:3`,
		//  expected: []item{
		//    {itemEOF, ""},
		//  },
		// },
	}
	for i, f := range filters {
		l := lex(f.query)
		for _, expected := range f.expected {
			got := l.nextItem()
			if got.typ == 0 {
				require.Fail(t, "missing items in the lexer", "filter %v: [%v]: expected [%s]", i, f.query, expected)
			}
			require.Equal(t, expected.typ, got.typ, "filter %v: [%v]: got [%v], expected [%v]", i, f.query, got, expected)
			require.Equal(t, expected.val, got.val, "filter %v: [%v]: got [%v], expected [%v]", i, f.query, got, expected)
		}
		if got := l.nextItem(); got.typ != 0 {
			require.Fail(t, "should not have additional items in the lexer", "filter %v: [%v]: got [%s]", i, f.query, got)
		}
	}
}
