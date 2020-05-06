package expr

import (
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/altipla-consulting/expr/parse"
)

type ParamOption func(f *Filter)

func Required() ParamOption {
	return func(f *Filter) {
		f.required = true
	}
}

func IDParam(name string, opts ...ParamOption) *Filter {
	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpEqual, parse.OpNotEqual},
		eval: func(value parse.Node) (interface{}, error) {
			switch v := value.(type) {
			case *parse.NumberNode:
				if v.Val < 0 {
					return nil, status.Errorf(codes.InvalidArgument, "id field cannot be negative: %v: %d", name, v.Val)
				}
				return v.Val, nil

			default:
				return nil, status.Errorf(codes.InvalidArgument, "id fields require numeric filters: %v: %v", name, value)
			}
		},
	}
}

func EnumParam(name string, values map[string]int32, opts ...ParamOption) *Filter {
	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpEqual, parse.OpNotEqual},
		eval: func(value parse.Node) (interface{}, error) {
			switch v := value.(type) {
			case *parse.ConstantNode:
				if strings.HasSuffix(v.Name, "_UNKNOWN") {
					return nil, status.Errorf(codes.InvalidArgument, "enum fields cannot be filtered by the unknown value: %v: %s", name, v.Name)
				}
				if _, ok := values[v.Name]; !ok {
					return nil, status.Errorf(codes.InvalidArgument, "unknown enum field value: %v: %s", name, v.Name)
				}
				return v.Name, nil

			default:
				return nil, status.Errorf(codes.InvalidArgument, "enum fields require constants filters: %v: %v", name, value)
			}
		},
	}
}

func BoolParam(name string, opts ...ParamOption) *Filter {
	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpEqual, parse.OpNotEqual},
		eval: func(value parse.Node) (interface{}, error) {
			switch v := value.(type) {
			case *parse.ConstantNode:
				switch strings.ToLower(v.Name) {
				case "true":
					return true, nil
				case "false":
					return false, nil
				}
				return nil, status.Errorf(codes.InvalidArgument, "boolean fields should be either true or nil: %v: %s", name, v.Name)

			default:
				return nil, status.Errorf(codes.InvalidArgument, "boolean fields require boolean filters: %v: %v", name, value)
			}
		},
	}
}

func TimestampParam(name string, opts ...ParamOption) *Filter {
	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpExists, parse.OpGreaterThan, parse.OpGreaterOrEqualThan, parse.OpLessThan, parse.OpLessOrEqualThan},
		eval: func(value parse.Node) (interface{}, error) {
			switch v := value.(type) {
			case *parse.StringNode:
				if len(v.Unquoted()) == len("2006-01-02") {
					t, err := time.Parse("2006-01-02", v.Unquoted())
					if err != nil {
						return nil, status.Errorf(codes.InvalidArgument, "invalid date: %v: %v", v.Unquoted(), err)
					}
					return t, err
				}

				t, err := time.Parse(time.RFC3339, v.Unquoted())
				if err != nil {
					return nil, status.Errorf(codes.InvalidArgument, "invalid rfc3339 timestamp: %v: %v", v.Unquoted(), err)
				}
				return t, err

			default:
				return nil, status.Errorf(codes.InvalidArgument, "timestamp fields require string filters: %v: %v", name, value)
			}
		},
	}
}

func StringParam(name string, opts ...ParamOption) *Filter {
	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpEqual, parse.OpNotEqual, parse.OpContains},
		eval: func(value parse.Node) (interface{}, error) {
			switch v := value.(type) {
			case *parse.ConstantNode:
				return v.Name, nil

			case *parse.StringNode:
				return v.Unquoted(), nil

			default:
				return nil, status.Errorf(codes.InvalidArgument, "string fields require string filters: %v: %v", name, value)
			}
		},
	}
}
