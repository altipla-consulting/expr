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
			return fmt.Sprintf("%s=%s", name, v.String())

		case int64, int32, int:
			return fmt.Sprintf("%s=%d", name, value)

		default:
			panic(fmt.Sprintf("unsupported type in query builder for field %s: %T", name, value))
		}
	}
}
