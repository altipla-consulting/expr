package expr

import (
	"strings"

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
	sqlValue  func(value parse.Node) (interface{}, error)
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

type sqlCondition struct {
	sql  string
	vals []interface{}
}

func (cond *sqlCondition) SQL() string           { return cond.sql }
func (cond *sqlCondition) Values() []interface{} { return cond.vals }

func (fs Filters) ApplySQL(q *database.Collection, query string) (*database.Collection, error) {
	sql, vals, err := fs.ToSQL(query)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if sql != "" {
		q = q.FilterCond(&sqlCondition{sql, vals})
	}

	return q, nil
}

func (fs Filters) getFilter(name string) *Filter {
	for _, f := range fs {
		if f.name == name {
			return f
		}
	}
	return nil
}

func (fs Filters) ToSQL(query string) (string, []interface{}, error) {
	root, err := parse.Parse(query)
	if err != nil {
		return "", nil, status.Errorf(codes.InvalidArgument, "invalid filter expression: %s", err)
	}

	var conds []string
	var vals []interface{}

	present := make(map[string]bool)
	for _, expr := range root.Nodes {
		f := fs.getFilter(expr.Field.Name)
		if f == nil {
			return "", nil, status.Errorf(codes.InvalidArgument, "unknown field in query: %v", expr.Field.Name)
		}

		if !f.hasOperator(expr.Op.Val) {
			return "", nil, status.Errorf(codes.InvalidArgument, "operator not allowed for field %v: %q", expr.Field.Name, expr.Op.Val)
		}

		val, err := f.sqlValue(expr.Val)
		if err != nil {
			return "", nil, errors.Trace(err)
		}

		var parts []string
		if expr.Negative {
			parts = append(parts, "NOT")
		}
		parts = append(parts, expr.Field.Name)
		switch expr.Op.Val {
		case parse.OpEqual, parse.OpNotEqual:
			parts = append(parts, string(expr.Op.Val))
		default:
			return "", nil, errors.Errorf("cannot use operator in SQL queries: %v", expr.Op.Val)
		}
		parts = append(parts, "?")

		conds = append(conds, "("+strings.Join(parts, " ")+")")
		vals = append(vals, val)

		present[expr.Field.Name] = true
	}

	for _, f := range fs {
		if f.required && !present[f.name] {
			return "", nil, status.Errorf(codes.InvalidArgument, "required filter in query: %v", f.name)
		}
	}

	return strings.Join(conds, " AND "), vals, nil
}
