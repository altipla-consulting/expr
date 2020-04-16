package parse

type Operator string

const (
	OpEqual    = Operator("=")
	OpNotEqual = Operator("!=")
	OpContains = Operator(":")
)

var allOperators = []Operator{
	OpEqual,
	OpNotEqual,
	OpContains,
}
