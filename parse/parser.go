package parse

import (
	"runtime"
	"strconv"

	"libs.altipla.consulting/errors"
)

type parser struct {
	lex       *lexer
	token     [1]item
	peekCount int
}

func Parse(query string) (root *AndNode, err error) {
	p := &parser{lex: lex(query)}
	defer p.recover(&err)
	return p.parseAnd(), nil
}

func (p *parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

func (p *parser) backup() {
	p.peekCount++
}

func (p *parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

func (p *parser) errorf(format string, args ...interface{}) {
	panic(errors.Errorf(format, args...))
}

func (p *parser) unexpected(token item, context string) {
	p.errorf("unexpected %s in %s", token, context)
}

func (p *parser) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if p != nil {
			p.lex.drain()
		}
		*errp = e.(error)
	}
}

func (p *parser) parseAnd() *AndNode {
	n := &AndNode{
		NodeType: NodeAnd,
	}

	switch next := p.next(); next.typ {
	case itemAnd:
	case itemEOF:
		return n
	default:
		p.unexpected(next, "AND")
	}

	for p.peek().typ != itemEOF {
		n.Nodes = append(n.Nodes, p.parseExpr())
	}

	return n
}

func (p *parser) parseOperator() *OperatorNode {
	tok := p.next()
	if tok.typ != itemOperator {
		p.unexpected(tok, "expression operator")
	}

	for _, op := range allOperators {
		if string(op) == tok.val {
			return &OperatorNode{
				NodeType: NodeOperator,
				Val:      op,
			}
		}
	}

	p.errorf("unknown operator: %v", tok.val)

	panic("should not reach here")
}

func (p *parser) parseExpr() *ExprNode {
	// Lee el nombre del campo, posiblemente con una negativa delante.
	var negative bool
	tok := p.next()
	switch tok.typ {
	case itemField:
	case itemNot:
		negative = true
		tok = p.next()
	default:
		p.unexpected(tok, "expression field")
	}

	expr := &ExprNode{
		NodeType: NodeExpr,
		Field: &FieldNode{
			NodeType: NodeField,
			Name:     tok.val,
		},
		Op:       p.parseOperator(),
		Negative: negative,
	}

	// Operadores que no tienen argumentos adicionales.
	if !expr.Op.Val.HasArg() {
		return expr
	}

	// Operadores con argumentos de varios posibles tipos. Aquí no se comprueba
	// el tipo, solamente se lee lo que haya y se guarda en la expresión.
	switch tok := p.next(); tok.typ {
	case itemNumber:
		val, err := strconv.ParseInt(tok.val, 10, 64)
		if err != nil {
			p.errorf("cannot parse number: %v: %s", tok.val, err)
		}
		expr.Val = &NumberNode{
			NodeType: NodeNumber,
			Val:      val,
		}

	case itemString:
		expr.Val = &StringNode{
			NodeType: NodeString,
			Quoted:   tok.val,
		}

	case itemConstant:
		expr.Val = &ConstantNode{
			NodeType: NodeConstant,
			Name:     tok.val,
		}

	default:
		p.unexpected(tok, "expression value")
	}

	return expr
}
