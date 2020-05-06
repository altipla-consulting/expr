package parse

type Operator string

func (op Operator) HasArg() bool {
	return op != OpExists
}

const (
	OpEqual              = Operator("=")
	OpNotEqual           = Operator("!=")
	OpContains           = Operator(":")
	OpExists             = Operator(":*")
	OpGreaterThan        = Operator(">")
	OpGreaterOrEqualThan = Operator(">=")
	OpLessThan           = Operator("<")
	OpLessOrEqualThan    = Operator("<=")
)

var allOperators = []Operator{
	OpEqual,
	OpNotEqual,
	OpContains,
	OpExists,
	OpGreaterThan,
	OpGreaterOrEqualThan,
	OpLessThan,
	OpLessOrEqualThan,
}
