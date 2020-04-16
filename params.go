package expr

import (
	"strings"

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
				return nil, status.Errorf(codes.InvalidArgument, "id fields can only be filtered with numbers: %v: %s", name, value)
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
					return nil, status.Errorf(codes.InvalidArgument, "enum fields cannot be filtered by the unknown value: %v: %s", name, value)
				}
				if _, ok := values[v.Name]; !ok {
					return nil, status.Errorf(codes.InvalidArgument, "unknown enum field value: %v: %s", name, value)
				}
				return v.Name, nil

			default:
				return nil, status.Errorf(codes.InvalidArgument, "enum fields can only be filtered with constants: %v: %s", name, value)
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
				return nil, status.Errorf(codes.InvalidArgument, "boolean fields should be true or false: %v: %s", name, value)

			default:
				return nil, status.Errorf(codes.InvalidArgument, "boolean fields can only be filtered with booleans: %v: %s", name, value)
			}
		},
	}
}
