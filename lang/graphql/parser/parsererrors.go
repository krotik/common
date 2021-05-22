/*
 * Public Domain Software
 *
 * I (Matthias Ladkau) am the author of the source code in this file.
 * I have placed the source code in this file in the public domain.
 *
 * For further information see: http://creativecommons.org/publicdomain/zero/1.0/
 */

package parser

import (
	"errors"
	"fmt"
)

/*
newParserError creates a new ParserError object.
*/
func (p *parser) newParserError(t error, d string, token LexToken) error {
	return &Error{p.name, t, d, token.Lline, token.Lpos}
}

/*
Error models a parser related error
*/
type Error struct {
	Source string // Name of the source which was given to the parser
	Type   error  // Error type (to be used for equal checks)
	Detail string // Details of this error
	Line   int    // Line of the error
	Pos    int    // Position of the error
}

/*
Error returns a human-readable string representation of this error.
*/
func (pe *Error) Error() string {
	var ret string

	if pe.Detail != "" {
		ret = fmt.Sprintf("Parse error in %s: %v (%v)", pe.Source, pe.Type, pe.Detail)
	} else {
		ret = fmt.Sprintf("Parse error in %s: %v", pe.Source, pe.Type)
	}

	return fmt.Sprintf("%s (Line:%d Pos:%d)", ret, pe.Line, pe.Pos)
}

/*
Parser related error types
*/
var (
	ErrImpossibleLeftDenotation = errors.New("Term can only start an expression")
	ErrImpossibleNullDenotation = errors.New("Term cannot start an expression")
	ErrLexicalError             = errors.New("Lexical error")
	ErrNameExpected             = errors.New("Name expected")
	ErrOnExpected               = errors.New("Type condition starting with 'on' expected")
	ErrSelectionSetExpected     = errors.New("Selection Set expected")
	ErrMultipleShorthand        = errors.New("Query shorthand only allowed for one query operation")
	ErrUnexpectedEnd            = errors.New("Unexpected end")
	ErrUnexpectedToken          = errors.New("Unexpected term")
	ErrUnknownToken             = errors.New("Unknown term")
	ErrValueOrVariableExpected  = errors.New("Value or variable expected")
	ErrVariableExpected         = errors.New("Variable expected")
)
