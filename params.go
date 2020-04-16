package expr

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"libs.altipla.consulting/errors"

	"github.com/altipla-consulting/expr/parse"
)

type ParamOption func(f *Filter)

func Required() ParamOption {
	return func(f *Filter) {
		f.required = true
	}
}

func IDParam(name string, opts ...ParamOption) *Filter {
	readValue := func(value parse.Node) (int64, error) {
		switch v := value.(type) {
		case *parse.NumberNode:
			if v.Val < 0 {
				return 0, status.Errorf(codes.InvalidArgument, "id field cannot be negative: %v: %d", name, v.Val)
			}
			return v.Val, nil

		default:
			return 0, status.Errorf(codes.InvalidArgument, "id fields can only be filtered with numbers: %v: %s", name, value)
		}
	}

	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpEqual, parse.OpNotEqual},
		sqlValue:  func(value parse.Node) (interface{}, error) { return readValue(value) },
		match: func(value parse.Node, op parse.Operator, other interface{}) (bool, error) {
			v, err := readValue(value)
			if err != nil {
				return false, errors.Trace(err)
			}
			switch other := other.(int64); op {
			case parse.OpEqual:
				return v == other, nil
			case parse.OpNotEqual:
				return v != other, nil
			default:
				panic("should not reach here")
			}
		},
	}
}

func EnumParam(name string, values map[string]int32, opts ...ParamOption) *Filter {
	readValue := func(value parse.Node) (string, error) {
		switch v := value.(type) {
		case *parse.ConstantNode:
			if strings.HasSuffix(v.Name, "_UNKNOWN") {
				return "", status.Errorf(codes.InvalidArgument, "enum fields cannot be filtered by the unknown value: %v: %s", name, value)
			}
			if _, ok := values[v.Name]; !ok {
				return "", status.Errorf(codes.InvalidArgument, "unknown enum field value: %v: %s", name, value)
			}
			return v.Name, nil

		default:
			return "", status.Errorf(codes.InvalidArgument, "enum fields can only be filtered with constants: %v: %s", name, value)
		}
	}

	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpEqual, parse.OpNotEqual},
		sqlValue:  func(value parse.Node) (interface{}, error) { return readValue(value) },
		match: func(value parse.Node, op parse.Operator, other interface{}) (bool, error) {
			v, err := readValue(value)
			if err != nil {
				return false, errors.Trace(err)
			}
			switch other := other.(int32); op {
			case parse.OpEqual:
				return values[v] == other, nil
			case parse.OpNotEqual:
				return values[v] != other, nil
			default:
				panic("should not reach here")
			}
		},
	}
}

func BoolParam(name string, opts ...ParamOption) *Filter {
	readValue := func(value parse.Node) (bool, error) {
		switch v := value.(type) {
		case *parse.ConstantNode:
			switch strings.ToLower(v.Name) {
			case "true":
				return true, nil
			case "false":
				return false, nil
			}
			return false, status.Errorf(codes.InvalidArgument, "boolean fields should be true or false: %v: %s", name, value)

		default:
			return false, status.Errorf(codes.InvalidArgument, "boolean fields can only be filtered with booleans: %v: %s", name, value)
		}
	}

	return &Filter{
		name:      name,
		operators: []parse.Operator{parse.OpEqual, parse.OpNotEqual},
		sqlValue:  func(value parse.Node) (interface{}, error) { return readValue(value) },
		match: func(value parse.Node, op parse.Operator, other interface{}) (bool, error) {
			v, err := readValue(value)
			if err != nil {
				return false, errors.Trace(err)
			}
			switch other := other.(bool); op {
			case parse.OpEqual:
				return v == other, nil
			case parse.OpNotEqual:
				return v != other, nil
			default:
				panic("should not reach here")
			}
		},
	}
}
