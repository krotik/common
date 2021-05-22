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
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/krotik/common/stringutil"
)

/*
LexToken represents a token which is returned by the lexer.
*/
type LexToken struct {
	ID    LexTokenID // Token kind
	Pos   int        // Starting position (in runes)
	Val   string     // Token value
	Lline int        // Line in the input this token appears
	Lpos  int        // Position in the input line this token appears
}

/*
PosString returns the position of this token in the origianl input as a string.
*/
func (t LexToken) PosString() string {
	return fmt.Sprintf("Line %v, Pos %v", t.Lline, t.Lpos)
}

/*
String returns a string representation of a token.
*/
func (t LexToken) String() string {

	switch {

	case t.ID == TokenEOF:
		return "EOF"

	case t.ID == TokenError:
		return fmt.Sprintf("Error: %s (%s)", t.Val, t.PosString())

	case t.ID == TokenName:
		return fmt.Sprintf("<%s>", t.Val)

	case t.ID == TokenStringValue:
		return fmt.Sprintf("\"%s\"", t.Val)

	case t.ID == TokenIntValue:
		return fmt.Sprintf("int(%s)", t.Val)

	case t.ID == TokenFloatValue:
		return fmt.Sprintf("flt(%s)", t.Val)
	}

	return fmt.Sprintf("%s", t.Val)
}

/*
SymbolMap is a map of special symbols
*/
var SymbolMap = map[string]LexTokenID{
	"!": TokenPunctuator,
	"$": TokenPunctuator,
	"(": TokenPunctuator,
	")": TokenPunctuator,
	":": TokenPunctuator,
	"=": TokenPunctuator,
	"@": TokenPunctuator,
	"[": TokenPunctuator,
	"]": TokenPunctuator,
	"{": TokenPunctuator,
	"|": TokenPunctuator,
	"}": TokenPunctuator,
	// "..." Is checked as a special case
}

// Lexer
// =====

/*
RuneEOF is a special rune which represents the end of the input
*/
const RuneEOF = -1

/*
RuneComma is the rune for a comma
*/
const RuneComma = ','

/*
Function which represents the current state of the lexer and returns the next state
*/
type lexFunc func() lexFunc

/*
Lexer data structure
*/
type lexer struct {
	name   string        // Name to identify the input
	input  string        // Input string of the lexer
	pos    int           // Current rune pointer
	line   int           // Current line pointer
	lastnl int           // Last newline position
	width  int           // Width of last rune
	start  int           // Start position of the current red token
	tokens chan LexToken // Channel for lexer output
}

/*
Lex lexes a given input. Returns a channel which contains tokens.
*/
func Lex(name string, input string) chan LexToken {

	l := &lexer{name, input, 0, 0, 0, 0, 0, make(chan LexToken)}
	go l.run()

	return l.tokens
}

/*
LexToList lexes a given input. Returns a list of tokens.
*/
func LexToList(name string, input string) []LexToken {
	var tokens []LexToken

	for t := range Lex(name, input) {
		tokens = append(tokens, t)
	}

	return tokens
}

/*
run is the main loop of the lexer.
*/
func (l *lexer) run() {

	if l.skipWhiteSpace() {
		for state := l.lexToken; state != nil; {
			state = state()

			if !l.skipWhiteSpace() {
				break
			}
		}
	}

	close(l.tokens)
}

/*
next returns the next rune in the input and advances the current rune pointer if the
peek value is -1 or smaller. If the peek value is 0 or greater then the nth token from the current
position is returned without advancing the current rune pointer.
*/
func (l *lexer) next(peek int) rune {
	var r rune
	var w, peekw int

	// Check if we reached the end

	if int(l.pos) >= len(l.input) {
		return RuneEOF
	}

	// Decode the next rune

	peeklen := 1 + peek
	if peeklen < 1 {
		peeklen = 1
	}

	for i := 0; i < peeklen; i++ {
		r, w = utf8.DecodeRuneInString(l.input[l.pos+peekw:])
		peekw += w
	}

	if peek == -1 {
		l.width = w
		l.pos += l.width
	}

	return r
}

/*
hasSequence checks if the next characters are of the following sequence.
*/
func (l *lexer) hasSequence(s string) bool {
	runes := stringutil.StringToRuneSlice(s)
	for i := 0; i < len(runes); i++ {
		if l.next(i) != runes[i] {
			return false
		}
	}
	return true
}

/*
startNew starts a new token.
*/
func (l *lexer) startNew() {
	l.start = l.pos
}

/*
emitTokenAndValue passes a token with a given value back to the client.
*/
func (l *lexer) emitToken(i LexTokenID, val string) {
	if l.tokens != nil {
		l.tokens <- LexToken{i, l.start, val, l.line + 1, l.start - l.lastnl + 1}
	}
}

// State functions
// ===============

/*
lexToken is the main entry function for the lexer.
*/
func (l *lexer) lexToken() lexFunc {

	l.startNew()
	l.lexTextBlock()

	token := l.input[l.start:l.pos]

	// Check for Comment - @spec 2.1.4, 2.1.7

	if token == "#" {
		return l.skipRestOfLine()
	}

	// Lexical tokens - @spec 2.1.6

	// Check for String

	if token == "\"" {
		return l.lexStringValue()
	}

	// Check for Punctuator - @spec 2.1.8

	if _, ok := SymbolMap[token]; ok || token == "..." {
		l.emitToken(TokenPunctuator, token)
		return l.lexToken
	}

	// Check for Name - @spec 2.1.9

	isName, _ := regexp.MatchString("^[_A-Za-z][_0-9A-Za-z]*$", token)
	if isName {
		l.emitToken(TokenName, token)
		return l.lexToken
	}

	// Check for IntValue - @spec 2.9.1

	isZero, _ := regexp.MatchString("^-?0$", token)
	isInt, _ := regexp.MatchString("^-?[1-9][0-9]*$", token)
	if isZero || isInt {
		l.emitToken(TokenIntValue, token)
		return l.lexToken
	}

	// Check for FloatValue - @spec 2.9.2

	isFloat1, _ := regexp.MatchString("^[0-9]*\\.[0-9]*$", token)
	isFloat2, _ := regexp.MatchString("^[0-9][eE][+-]?[0-9]*$", token)
	isFloat3, _ := regexp.MatchString("^[0-9]*\\.[0-9][eE][+-]?[0-9]*$", token)

	if isFloat1 || isFloat2 || isFloat3 {
		l.emitToken(TokenFloatValue, strings.ToLower(token))
		return l.lexToken
	}

	// Everything else is an error

	l.emitToken(TokenError, token)

	return l.lexToken
}

/*
lexTextBlock lexes a block of text without whitespaces. Interprets
optionally all one or two letter tokens.
*/
func (l *lexer) lexTextBlock() {

	r := l.next(0)

	// Check if we start with a known symbol

	if _, ok := SymbolMap[strings.ToLower(string(r))]; ok || r == '#' || r == '"' {
		l.next(-1)
		return
	} else if r == '.' && l.hasSequence("...") {
		l.next(-1)
		l.next(-1)
		l.next(-1)
		return
	}

	for !l.isIgnoredRune(r) {
		l.next(-1)

		r = l.next(0)

		// Check if we find a token in the block

		if _, ok := SymbolMap[strings.ToLower(string(r))]; ok || r == '#' || r == '"' {
			return
		} else if r == '.' && l.hasSequence("...") {
			return
		}
	}
}

/*
lexStringValue lexes a string value either as a simple string or a block string.

Values can be declared in different ways:

" ... " A normal string (escape sequences are interpreted)

""" ... """ A multi-line string (escape sequences are not interpreted)
*/
func (l *lexer) lexStringValue() lexFunc {
	var isEnd func(rune) bool

	// String value lexing - @spec 2.9.4

	// Lookahead 2 tokens

	r1 := l.next(0)
	r2 := l.next(1)

	isBlockString := r1 == '"' && r2 == '"'

	if isBlockString {

		// Consume the initial quotes for blockstrings

		l.next(-1)
		l.next(-1)

		isEnd = func(r rune) bool {
			r1 := l.next(0)
			r2 := l.next(1)
			return r == '"' && r1 == '"' && r2 == '"'
		}

	} else {

		isEnd = func(r rune) bool {
			return r == '"'
		}
	}

	r := l.next(-1)
	lLine := l.line
	lLastnl := l.lastnl

	for !isEnd(r) {

		if r == '\n' {
			lLine++
			lLastnl = l.pos
		}

		r = l.next(-1)

		if r == RuneEOF {
			l.emitToken(TokenError, "EOF inside quotes")
			return nil
		}
	}

	if !isBlockString {
		val := l.input[l.start+1 : l.pos-1]

		s, err := strconv.Unquote("\"" + val + "\"")
		if err != nil {
			l.emitToken(TokenError, "Could not interpret escape sequence: "+err.Error())
			return nil
		}

		l.emitToken(TokenStringValue, s)

	} else {

		// Consume the final quotes for blockstrings

		l.next(-1)
		l.next(-1)

		token := l.input[l.start+3 : l.pos-3]

		// Since block strings represent freeform text often used in indented
		// positions, the string value semantics of a block string excludes uniform
		// indentation and blank initial and trailing lines
		// (from spec about 'Block Strings')

		token = stringutil.StripUniformIndentation(token)
		token = stringutil.TrimBlankLines(token)

		l.emitToken(TokenStringValue, token)
	}

	//  Set newline

	l.line = lLine
	l.lastnl = lLastnl

	return l.lexToken
}

/*
isIgnoredRune checks if a given rune should be ignored.
*/
func (l *lexer) isIgnoredRune(r rune) bool {

	// Ignored tokens - @spec 2.1.1, 2.1.2, 2.1.3, 2.1.3, 2.1.5, 2.1.7

	return unicode.IsSpace(r) || unicode.IsControl(r) || r == RuneEOF ||
		r == RuneComma || r == '\ufeff'
}

/*
skipWhiteSpace skips any number of whitespace characters. Returns false if the parser
reaches EOF while skipping whitespaces.
*/
func (l *lexer) skipWhiteSpace() bool {
	r := l.next(0)

	for l.isIgnoredRune(r) {
		if r == '\n' {
			l.line++
			l.lastnl = l.pos
		}

		l.next(-1)

		if r == RuneEOF {
			l.startNew()
			l.start--
			l.emitToken(TokenEOF, "")
			return false
		}

		r = l.next(0)
	}

	return true
}

/*
skipRestOfLine skips all characters until the next newline character.
*/
func (l *lexer) skipRestOfLine() lexFunc {
	r := l.next(-1)

	for r != '\n' && r != RuneEOF {
		r = l.next(-1)
	}

	if r == RuneEOF {
		return nil
	}

	l.line++
	l.lastnl = l.pos - 1

	return l.lexToken
}
