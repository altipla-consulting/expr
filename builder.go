package expr

import (
	"fmt"
	"strings"
)

type BuildField func() string

func Builder(fields ...BuildField) string {
	return And(fields...)()
}

func And(fields ...BuildField) BuildField {
	return func() string {
		q := make([]string, len(fields))
		for i, field := range fields {
			q[i] = field()
		}
		return strings.Join(q, " ")
	}
}

type enumValue interface {
	String() string
	EnumDescriptor() ([]byte, []int)
}

func Eq(name string, value interface{}) BuildField {
	return func() string {
		switch v := value.(type) {
		case enumValue:
			return fmt.Sprintf("%v=%v", name, v.String())

		case int64, int32, int, bool:
			return fmt.Sprintf("%v=%v", name, value)

		default:
			panic(fmt.Sprintf("unsupported type in query builder for field %s: %T", name, value))
		}
	}
}

func Exists(name string) BuildField {
	return func() string {
		return name + ":*"
	}
}
