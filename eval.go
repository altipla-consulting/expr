package expr

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"libs.altipla.consulting/database"
	"libs.altipla.consulting/errors"

	"github.com/altipla-consulting/expr/parse"
)

type Filters []*Filter

func (fs Filters) ApplySQL(q *database.Collection, query string) (*database.Collection, error) {
	root, err := parse.Parse(query)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid filter expression: %s", err)
	}

	present := make(map[string]bool)
	for _, child := range root.Nodes {
		f := fs.getFilter(child.Field.Name)
		if f == nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown field in query: %v", child.Field.Name)
		}

		if !f.hasOperator(child.Op.Val) {
			return nil, status.Errorf(codes.InvalidArgument, "operator not allowed for field %v: %q", child.Field.Name, child.Op.Val)
		}

		val, err := f.eval(child.Val)
		if err != nil {
			return nil, errors.Trace(err)
		}
		q = q.Filter(child.Field.Name+" "+child.Op.Val.ToSQL(), val)

		present[child.Field.Name] = true
	}

	for _, f := range fs {
		if f.required && !present[f.name] {
			return nil, status.Errorf(codes.InvalidArgument, "required filter in query: %v", f.name)
		}
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
