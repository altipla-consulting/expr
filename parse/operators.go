package parse

type Operator string

func (op Operator) ToSQL() string {
	return string(op)
}

const (
	OpEqual    = Operator("=")
	OpNotEqual = Operator("!=")
)

var allOperators = []Operator{
	OpEqual,
	OpNotEqual,
}
