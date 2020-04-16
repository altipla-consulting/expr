package parse

type Operator string

const (
	OpEqual    = Operator("=")
	OpNotEqual = Operator("!=")
	OpContains = Operator(":")
	OpExists   = Operator(":*")
)

var allOperators = []Operator{
	OpEqual,
	OpNotEqual,
	OpContains,
	OpExists,
}
