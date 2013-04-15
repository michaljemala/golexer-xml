package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type stateFn func(*lexer) stateFn

type itemType int

type item struct {
	typ itemType
	val string
}
type lexer struct {
	init  stateFn
	input string
	start int
	pos   int
	width int
	items chan item
}

const (
	eof rune = iota
)

const (
	tokenError itemType = iota
	tokenEndOfFile
	tokenTagBegin           // <
	tokenTagEnd             // >
	tokenTagBeginDash       // </
	tokenTagEndDash         // />
	tokenTagName            // tag name
	tokenAttrName           // attribute name
	tokenEquals             // =
	tokenDoubleQuotedString // "lorem ipsum"
	tokenSingleQuotedString // 'lorem ipsum'
	tokenText               // text
)

func NewLexer(input string) (*lexer, chan item) {
	l := &lexer{
		init:  lexInit,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) run() {
	for state := l.init; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		typ: tokenError,
		val: fmt.Sprintf(format, args),
	}
	return nil
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) emit(typ itemType) {
	l.items <- item{
		typ: typ,
		val: l.input[l.start:l.pos],
	}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func lexInit(l *lexer) stateFn {
	if r := l.peek(); r == eof {
		return l.errorf("Unexpected end of file")
	} else {
		return lexCommon
	}
	panic("Unreachable")
}

func lexCommon(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return nil
		case r == '<':
			switch s := l.next(); {
			case s == eof:
				return l.errorf("Unexpected end of file")
			case s == '/':
				l.emit(tokenTagBeginDash)
				return lexTagName
			case s == '?':
				return l.errorf("XML Declaration not supported yet") // TODO parse XML declaration
			case s == '!':
				return l.errorf("Comments not supported yet") // TODO parse XML comment
			case isValidFirstChar(s):
				l.backup()
				l.emit(tokenTagBegin)
				return lexTagName
			default:
				return l.errorf("Invalid character: Expected '<first-character-of-tagname>', found '%v'", string(s))
			}
		case r == '>':
		case r == '/':
			// />
		case r == '=':
		case r == '"':
		case r == '\'':
		}
	}
	panic("Unreachable")
}

func lexTagName(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("Unexpected end of file")
		case r == ' ' || r == '/':
			l.backup()
			l.emit(tokenTagName)
			return lexTagInside
		case !isValidChar(r):
			return l.errorf("Invalid character: Expected '<first-character-of-tagname>', found '%v'", string(r))
		}
	}
	panic("Unreachable")
}

func lexTagInside(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			return l.errorf("Unexpected end of file")
		case r == ' ':
			l.ignore()
		case r == '/':
			if s := l.next(); s != '>' {
				return l.errorf("Invalid character: Expected '>', found '%v'", string(s))
			}
			l.emit(tokenTagEndDash)
			return nil // TODO where to go?
		case r == '>':
			l.emit(tokenTagEnd)
			return nil // TODO where to go?
		}
	}
	panic("Unreachable")
}

// Helpers
func isValidFirstChar(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == ':'
}

func isValidChar(r rune) bool {
	return isValidFirstChar(r) || unicode.IsDigit(r) || r == '-' || r == '.'
}
