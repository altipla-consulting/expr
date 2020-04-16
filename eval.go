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
	match     func(value parse.Node, op parse.Operator, other interface{}) (bool, error)
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
	evaler, err := fs.Evaler(query)
	if err != nil {
		return nil, errors.Trace(err)
	}
	sql, vals, err := evaler.ToSQL()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if sql != "" {
		q = q.FilterCond(&sqlCondition{sql, vals})
	}

	return q, nil
}

func (fs Filters) Evaler(query string) (*Evaler, error) {
	root, err := parse.Parse(query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid filter expression: %s", err)
	}

	evaler := &Evaler{
		root:    root,
		filters: make(map[string]*Filter),
	}
	for _, f := range fs {
		evaler.filters[f.name] = f
	}

	present := make(map[string]bool)
	for _, expr := range root.Nodes {
		f := evaler.filters[expr.Field.Name]
		if f == nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown field in query: %v", expr.Field.Name)
		}

		if !f.hasOperator(expr.Op.Val) {
			return nil, status.Errorf(codes.InvalidArgument, "operator not allowed for field %v: %q", expr.Field.Name, expr.Op.Val)
		}

		present[expr.Field.Name] = true
	}

	for _, f := range fs {
		if f.required && !present[f.name] {
			return nil, status.Errorf(codes.InvalidArgument, "required filter in query: %v", f.name)
		}
	}

	return evaler, nil
}

type Evaler struct {
	root    *parse.AndNode
	filters map[string]*Filter
}

func (evaler *Evaler) Match(fields map[string]interface{}) (bool, error) {
	for _, expr := range evaler.root.Nodes {
		result, err := evaler.filters[expr.Field.Name].match(expr.Val, expr.Op.Val, fields[expr.Field.Name])
		if err != nil {
			return false, errors.Trace(err)
		}
		if !result {
			return false, nil
		}
	}

	return true, nil
}

func (evaler *Evaler) ToSQL() (string, []interface{}, error) {
	var conds []string
	var vals []interface{}

	for _, expr := range evaler.root.Nodes {
		val, err := evaler.filters[expr.Field.Name].sqlValue(expr.Val)
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
	}

	return strings.Join(conds, " AND "), vals, nil
}
