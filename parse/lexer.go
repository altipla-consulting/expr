package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type itemType int

const (
	itemError itemType = iota + 1
	itemEOF
	itemField
	itemOperator
	itemString
	itemNumber
	itemConstant
	itemAnd
	itemNot
)

const eof = -1

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemError:
		return i.val
	case itemEOF:
		return "EOF"
	case itemField:
		return fmt.Sprintf("field:%q", i.val)
	case itemOperator:
		return fmt.Sprintf("op:%q", i.val)
	case itemString:
		return fmt.Sprintf("string:%q", i.val)
	case itemNumber:
		return fmt.Sprintf("number:%q", i.val)
	case itemConstant:
		return fmt.Sprintf("const:%q", i.val)
	case itemAnd:
		return " AND "
	case itemNot:
		return "NOT "
	}
	panic(fmt.Sprintf("should not reach here: %v", i.typ))
}

type stateFn func(l *lexer) stateFn

type lexer struct {
	input      string
	start, pos int
	width      int
	items      chan item
}

func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 { // revive:disable-line:empty-block
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *lexer) nextItem() item {
	return <-l.items
}

func (l *lexer) drain() {
	for range l.items { // revive:disable-line:empty-block
	}
}

func (l *lexer) ignoreSpaces() {
	l.acceptRun(" ")
	l.ignore()
}

func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

func lexField(l *lexer) stateFn {
	l.ignoreSpaces()

	if r := l.peek(); r == '-' {
		l.next()
		l.ignore()
		l.emit(itemNot)
	}

	l.acceptRun("abcdefghijklmnopqrstuvwxyz1234567890-.")
	if l.start == l.pos {
		return l.errorf("field name: %q", l.input[l.start:])
	}

	l.emit(itemField)
	return lexOperator
}

func lexOperator(l *lexer) stateFn {
	l.ignoreSpaces()
	l.acceptRun(":<=!>*")

	if l.start == l.pos {
		return l.errorf("empty operator")
	}

	if l.input[l.start:l.pos] == ":*" {
		l.emit(itemOperator)
		return lexAnd
	}
	l.emit(itemOperator)

	return lexValue
}

func lexValue(l *lexer) stateFn {
	l.ignoreSpaces()

	switch r := l.peek(); {
	case r == '"':
		return lexString
	case isDigit(r):
		return lexNumber
	default:
		return lexConstant
	}
}

func lexString(l *lexer) stateFn {
	l.next()

Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof {
				break
			}
			fallthrough
		case eof:
			return l.errorf("unterminated quoted string: %q", l.input[l.start:])
		case '"':
			break Loop
		}
	}

	l.emit(itemString)
	return lexAnd
}

func lexNumber(l *lexer) stateFn {
	l.accept("+-")
	l.acceptRun("0123456789")

	if l.start == l.pos {
		return l.errorf("unknown number: %q", l.input[l.start:])
	}

	l.emit(itemNumber)
	return lexAnd
}

func lexConstant(l *lexer) stateFn {
	l.acceptRun("ABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789")

	if l.start == l.pos {
		return l.errorf("unknown constant: %q", l.input[l.start:])
	}

	l.emit(itemConstant)
	return lexAnd
}

func lexAnd(l *lexer) stateFn {
	switch r := l.next(); r {
	case ' ':
	case eof:
		l.emit(itemEOF)
		return nil
	default:
		return l.errorf("unknown character: %c", r)
	}
	l.ignore()

	l.ignoreSpaces()
	if l.peek() == eof {
		l.emit(itemEOF)
		return nil
	}

	return lexField
}

func lexStart(l *lexer) stateFn {
	l.emit(itemAnd)
	if r := l.peek(); r == eof {
		l.emit(itemEOF)
		return nil
	}
	return lexField
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}
