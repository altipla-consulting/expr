package expr

import (
	"fmt"
	"strings"
	"unicode"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"libs.altipla.consulting/database"
	"libs.altipla.consulting/errors"

	"github.com/altipla-consulting/expr/parse"
)

type Filter struct {
	name      string
	required  bool
	operators []parse.Operator
	eval      func(value parse.Node) (interface{}, error)
}

func (f *Filter) hasOperator(op parse.Operator) bool {
	for _, o := range f.operators {
		if o == op {
			return true
		}
	}
	return false
}

type Filters []*Filter

func (fs Filters) parseQuery(query string) (*parse.AndNode, map[string]*Filter, error) {
	root, err := parse.Parse(query)
	if err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid filter expression: %s", err)
	}

	filters := make(map[string]*Filter)
	for _, f := range fs {
		filters[f.name] = f
	}

	present := make(map[string]bool)
	for _, expr := range root.Nodes {
		f := filters[expr.Field.Name]
		if f == nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "unknown field in query: %v", expr.Field.Name)
		}

		if !f.hasOperator(expr.Op.Val) {
			return nil, nil, status.Errorf(codes.InvalidArgument, "operator not allowed for field %v: %q", expr.Field.Name, expr.Op.Val)
		}

		if _, err := f.eval(expr.Val); err != nil {
			return nil, nil, errors.Trace(err)
		}

		present[expr.Field.Name] = true
	}

	for _, f := range fs {
		if f.required && !present[f.name] {
			return nil, nil, status.Errorf(codes.InvalidArgument, "required filter in query: %v", f.name)
		}
	}

	return root, filters, nil
}

func (fs Filters) ApplySQL(q *database.Collection, query string) (*database.Collection, error) {
	root, filters, err := fs.parseQuery(query)
	if err != nil {
		return nil, errors.Trace(err)
	}
	cond, err := evalSQL(root, filters)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if cond != nil {
		q = q.FilterCond(cond)
	}

	return q, nil
}

type sqlCondition struct {
	sql  string
	vals []interface{}
}

func (cond *sqlCondition) SQL() string           { return cond.sql }
func (cond *sqlCondition) Values() []interface{} { return cond.vals }

func evalSQL(root *parse.AndNode, filters map[string]*Filter) (*sqlCondition, error) {
	var conds []string
	var vals []interface{}
	for _, expr := range root.Nodes {
		switch expr.Op.Val {
		case parse.OpExists:
			if expr.Negative {
				conds = append(conds, fmt.Sprintf("(%s IS NULL)", sqlizeName(expr.Field.Name)))
			} else {
				conds = append(conds, fmt.Sprintf("(%s IS NOT NULL)", sqlizeName(expr.Field.Name)))
			}

		case parse.OpEqual, parse.OpNotEqual:
			val, err := filters[expr.Field.Name].eval(expr.Val)
			if err != nil {
				return nil, errors.Trace(err)
			}

			var not string
			if expr.Negative {
				not = "NOT "
			}
			conds = append(conds, fmt.Sprintf("(%s%s %s ?)", not, sqlizeName(expr.Field.Name), expr.Op.Val))
			vals = append(vals, val)

		default:
			return nil, errors.Errorf("cannot use operator in SQL queries: %v", expr.Op.Val)
		}
	}

	if len(conds) == 0 {
		return nil, nil
	}

	return &sqlCondition{
		sql:  strings.Join(conds, " AND "),
		vals: vals,
	}, nil
}

func sqlizeName(s string) string {
	var result []rune
	for _, r := range s {
		if unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

type Matcher func(value map[string]interface{}) bool

func (fs Filters) Matcher(query string) (Matcher, error) {
	root, filters, err := fs.parseQuery(query)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return func(value map[string]interface{}) bool {
		for _, expr := range root.Nodes {
			// Podemos ignorar el error porque ya se comprueban antes al parsear la query.
			want, _ := filters[expr.Field.Name].eval(expr.Val)

			got, exists := value[expr.Field.Name]

			// Las enumeraciones se comparan como strings, así que las convertimos.
			if enumv, ok := got.(enumValue); ok {
				got = enumv.String()
			}

			var result bool
			switch expr.Op.Val {
			case parse.OpEqual:
				result = (want == got)
			case parse.OpNotEqual:
				result = (want != got)
			case parse.OpContains:
				result = strings.Contains(got.(string), want.(string))
			case parse.OpExists:
				result = exists
			}

			// Si no encontramos lo que necesitamos podemos parar de comprobar
			// condiciones y salirnos ya.
			if expr.Negative == result {
				return false
			}
		}

		return true
	}, nil
}
