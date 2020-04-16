package parse

import (
	"strconv"
	"strings"
)

type Node interface {
	Type() NodeType
	String() string
}

type NodeType int

func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeField NodeType = iota
	NodeOperator
	NodeString
	NodeNumber
	NodeConstant
	NodeAnd
	NodeExpr
	NodeHasContent
)

type FieldNode struct {
	NodeType
	Name string
}

func (f *FieldNode) String() string {
	return f.Name
}

type OperatorNode struct {
	NodeType
	Val Operator
}

func (f *OperatorNode) String() string {
	return string(f.Val)
}

type StringNode struct {
	NodeType
	Quoted string
}

func (s *StringNode) String() string {
	return s.Quoted
}

type NumberNode struct {
	NodeType
	Val int64
}

func (n *NumberNode) String() string {
	return strconv.FormatInt(n.Val, 10)
}

type ConstantNode struct {
	NodeType
	Name string
}

func (c *ConstantNode) String() string {
	return c.Name
}

type AndNode struct {
	NodeType
	Nodes []*ExprNode
}

func (a *AndNode) String() string {
	s := make([]string, len(a.Nodes))
	for i, node := range a.Nodes {
		s[i] = node.String()
	}
	return strings.Join(s, " ")
}

type HasContentNode struct {
	NodeType
}

func (c *HasContentNode) String() string {
	return "*"
}

type ExprNode struct {
	NodeType
	Field *FieldNode
	Op    *OperatorNode
	Val   Node
}

func (e ExprNode) String() string {
	return e.Field.String() + e.Op.String() + e.Val.String()
}
