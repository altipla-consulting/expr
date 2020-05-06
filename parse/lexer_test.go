package parse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	tests := []struct {
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
				{itemField, "foo"}, {itemOperator, ":*"},
				{itemEOF, ""},
			},
		},
		{
			query: `-foo=3`,
			expected: []item{
				{itemAnd, ""},
				{itemNot, ""},
				{itemField, "foo"}, {itemOperator, "="}, {itemNumber, "3"},
				{itemEOF, ""},
			},
		},
		{
			query: `-foo:*`,
			expected: []item{
				{itemAnd, ""},
				{itemNot, ""},
				{itemField, "foo"}, {itemOperator, ":*"},
				{itemEOF, ""},
			},
		},
		{
			query: `fooBar:*`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "fooBar"}, {itemOperator, ":*"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo=true`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, "="}, {itemConstant, "true"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo>3`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, ">"}, {itemNumber, "3"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo>=3`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, ">="}, {itemNumber, "3"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo<3`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, "<"}, {itemNumber, "3"},
				{itemEOF, ""},
			},
		},
		{
			query: `foo<=3`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "foo"}, {itemOperator, "<="}, {itemNumber, "3"},
				{itemEOF, ""},
			},
		},
		{
			query: `ts>"2019-03-02"`,
			expected: []item{
				{itemAnd, ""},
				{itemField, "ts"}, {itemOperator, ">"}, {itemString, `"2019-03-02"`},
				{itemEOF, ""},
			},
		},
		// {
		//  query: `NOT foo:3`,
		//  expected: []item{
		//    {itemEOF, ""},
		//  },
		// },
	}
	for i, test := range tests {
		l := lex(test.query)
		for _, expected := range test.expected {
			got := l.nextItem()
			if got.typ == 0 {
				require.Fail(t, "missing items in the lexer", "test %v: [%v]: expected [%s]", i, test.query, expected)
			}
			require.Equal(t, expected.typ, got.typ, "test %v: [%v]: got [%v], expected [%v]", i, test.query, got, expected)
			require.Equal(t, expected.val, got.val, "test %v: [%v]: got [%v], expected [%v]", i, test.query, got, expected)
		}
		if got := l.nextItem(); got.typ != 0 {
			require.Fail(t, "should not have additional items in the lexer", "test %v: [%v]: got [%s]", i, test.query, got)
		}
	}
}
