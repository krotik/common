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

	"github.com/krotik/common/errorutil"
)

// Parser Rules
// ============

/*
Maps of AST nodes corresponding to lexer tokens
*/
var astNodeMapValues map[string]*ASTNode
var astNodeMapTokens map[LexTokenID]*ASTNode
var astNodeMapIgnoredValues map[string]*ASTNode

func init() {
	astNodeMapValues = map[string]*ASTNode{
		"query":        {NodeOperationDefinition, nil, nil, nil, 0, ndOperationDefinition, nil},
		"mutation":     {NodeOperationDefinition, nil, nil, nil, 0, ndOperationDefinition, nil},
		"subscription": {NodeOperationDefinition, nil, nil, nil, 0, ndOperationDefinition, nil},
		"fragment":     {NodeFragmentDefinition, nil, nil, nil, 0, ndFragmentDefinition, nil},
		"{":            {NodeSelectionSet, nil, nil, nil, 0, ndSelectionSet, nil},
		"(":            {NodeArguments, nil, nil, nil, 0, ndArgsOrVarDef, nil},
		"@":            {NodeDirectives, nil, nil, nil, 0, ndDirectives, nil},
		"$":            {NodeVariable, nil, nil, nil, 0, ndVariable, nil},
		"...":          {NodeFragmentSpread, nil, nil, nil, 0, ndFragmentSpread, nil},
		"[":            {NodeListValue, nil, nil, nil, 0, ndListValue, nil},

		// Tokens which are not part of the AST (can be retrieved by next but not be inserted by run)

		"}": {"", nil, nil, nil, 0, nil, nil},
		":": {"", nil, nil, nil, 0, nil, nil},
		")": {"", nil, nil, nil, 0, nil, nil},
		"=": {"", nil, nil, nil, 0, nil, nil},
		"]": {"", nil, nil, nil, 0, nil, nil},
	}
	astNodeMapTokens = map[LexTokenID]*ASTNode{
		TokenName:        {NodeName, nil, nil, nil, 0, ndTerm, nil},
		TokenIntValue:    {NodeValue, nil, nil, nil, 0, ndTerm, nil},
		TokenStringValue: {NodeValue, nil, nil, nil, 0, ndTerm, nil},
		TokenFloatValue:  {NodeValue, nil, nil, nil, 0, ndTerm, nil},
		TokenEOF:         {NodeEOF, nil, nil, nil, 0, ndTerm, nil},
	}
}

// Parser
// ======

/*
Parser data structure
*/
type parser struct {
	name   string          // Name to identify the input
	node   *ASTNode        // Current ast node
	tokens chan LexToken   // Channel which contains lex tokens
	rp     RuntimeProvider // Runtime provider which creates runtime components

	// Flags

	isVarDef bool // The next Arguments block is parsed as a VariableDefinition
	isValue  bool // The next expression is parsed as a value
}

/*
Parse parses a given input string and returns an AST.
*/
func Parse(name string, input string) (*ASTNode, error) {
	return ParseWithRuntime(name, input, nil)
}

/*
ParseWithRuntime parses a given input string and returns an AST decorated with
runtime components.
*/
func ParseWithRuntime(name string, input string, rp RuntimeProvider) (*ASTNode, error) {
	p := &parser{name, nil, Lex(name, input), rp, false, false}

	node, err := p.next()

	if err != nil {
		return nil, err
	}

	p.node = node

	doc := newAstNode(NodeDocument, p, node.Token)

	for err == nil && p.node.Name != NodeEOF {

		if node, err = p.run(0); err == nil {

			if node != nil && node.Name == NodeSelectionSet {

				// Handle query shorthand

				if len(doc.Children) == 0 {
					ed := newAstNode(NodeExecutableDefinition, p, node.Token)
					doc.Children = append(doc.Children, ed)
					od := newAstNode(NodeOperationDefinition, p, node.Token)
					ed.Children = append(ed.Children, od)
					od.Children = append(od.Children, node)

				} else {

					return nil, p.newParserError(ErrMultipleShorthand,
						node.Token.String(), *node.Token)
				}
			} else {

				ed := newAstNode(NodeExecutableDefinition, p, node.Token)
				doc.Children = append(doc.Children, ed)
				ed.Children = append(ed.Children, node)
			}
		}
	}

	if err == nil {
		return doc, nil
	}

	return nil, err
}

/*
run is the main parser function.
*/
func (p *parser) run(rightBinding int) (*ASTNode, error) {
	var err error
	var left *ASTNode

	n := p.node

	// Get the next ASTNode

	if p.node, err = p.next(); err == nil {

		// All nodes have a null denotation

		if n.nullDenotation == nil {
			return nil, p.newParserError(ErrImpossibleNullDenotation, p.node.Token.Val, *p.node.Token)
		}

		left, err = n.nullDenotation(p, n)
	}

	if err != nil {
		return nil, err
	}

	// At this point we would normally collect left denotations but this
	// parser has only null denotations

	errorutil.AssertTrue(rightBinding == p.node.binding, "Unexpected right binding")

	return left, nil
}

/*
next retrieves the next lexer token and return it as ASTNode.
*/
func (p *parser) next() (*ASTNode, error) {

	token, more := <-p.tokens

	if !more {

		// Unexpected end of input - the associated token is an empty error token

		return nil, p.newParserError(ErrUnexpectedEnd, "", token)

	} else if token.ID == TokenError {

		// There was a lexer error wrap it in a parser error

		return nil, p.newParserError(ErrLexicalError, token.Val, token)

	} else if node, ok := astNodeMapValues[token.Val]; ok &&
		(!p.isValue || token.ID == TokenPunctuator) && token.ID != TokenStringValue {

		// Parse complex expressions unless we parse a value (then just deal with punctuators)

		return node.instance(p, &token), nil

	} else if node, ok := astNodeMapTokens[token.ID]; ok {

		return node.instance(p, &token), nil
	}

	return nil, p.newParserError(ErrUnknownToken, fmt.Sprintf("id:%v (%v)", token.ID, token), token)
}

// Null denotation functions
// =========================

/*
ndTerm is used for terminals.
*/
func ndTerm(p *parser, self *ASTNode) (*ASTNode, error) {
	return self, nil
}

/*
ndVariable is used for variables.  (@spec 2.10)
*/
func ndVariable(p *parser, self *ASTNode) (*ASTNode, error) {
	var err error

	if p.node.Token.ID == TokenName {

		// Append the query name and move on

		self.Token = p.node.Token
		p.node, err = p.next()
	}

	return self, err
}

/*
ndListValue parses a list value. (@spec 2.9.7)
*/
func ndListValue(p *parser, self *ASTNode) (*ASTNode, error) {
	var current *ASTNode
	var err error

	for p.node.Token.ID != TokenEOF && p.node.Token.Val != "]" {

		// Parse list values

		if current, err = parseValue(p); err == nil {
			self.Children = append(self.Children, current)
		} else {
			return nil, err
		}
	}

	return self, skipToken(p, "]")
}

/*
ndInputObject parses an input object literal. (@spec 2.9.8)
*/
func ndInputObject(p *parser, self *ASTNode) (*ASTNode, error) {
	var current *ASTNode
	var err error

	changeAstNode(self, NodeObjectValue, p)

	for p.node.Token.ID != TokenEOF && p.node.Token.Val != "}" {

		if current, err = p.run(0); err == nil {

			if current.Name != NodeName {

				err = p.newParserError(ErrNameExpected,
					current.Token.String(), *current.Token)

			} else {

				of := newAstNode(NodeObjectField, p, current.Token)
				self.Children = append(self.Children, of)

				if err = skipToken(p, ":"); err == nil {

					// Parse object value

					if current, err = parseValue(p); err == nil {
						of.Children = append(of.Children, current)
					}
				}
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return self, skipToken(p, "}")
}

/*
ndFragmentSpread is used for fragment spreads and inline fragments.
(@spec 2.8, 2.8.2)
*/
func ndFragmentSpread(p *parser, self *ASTNode) (*ASTNode, error) {
	var current, expectedNameNode *ASTNode
	var err error

	if p.node.Token.Val == "on" {

		// We might have an inline fragment

		onToken := p.node.Token
		p.node, err = p.next()

		if err == nil && p.node.Name == NodeName {

			// Append the fragment name

			changeAstNode(p.node, NodeTypeCondition, p)
			self.Children = append(self.Children, p.node)
			p.node, err = p.next()

		} else {

			self.Token = onToken
		}

	} else if p.node.Token.ID == TokenName {

		// Append the query name and move on

		self.Token = p.node.Token
		p.node, err = p.next()

	} else {

		expectedNameNode = p.node
	}

	if err == nil && p.node.Token.Val == "@" {

		// Parse directives

		if current, err = p.run(0); err == nil {
			self.Children = append(self.Children, current)
		}
	}

	if err == nil && p.node.Token.Val == "{" {

		// Parse selection set

		if current, err = p.run(0); err == nil {
			self.Children = append(self.Children, current)

			// If there is a selection set we must have an inline fragment

			changeAstNode(self, NodeInlineFragment, p)
		}

	} else if err == nil && expectedNameNode != nil {

		// Using the fragment spread operatior without specifying a name nor
		// a selection set is an error

		err = p.newParserError(ErrNameExpected,
			expectedNameNode.Token.String(), *expectedNameNode.Token)
	}

	return self, err
}

/*
ndOperationDefinition parses an operation definition. Each operation is
represented by an optional operation name and a selection set. (@spec 2.3)
*/
func ndOperationDefinition(p *parser, self *ASTNode) (*ASTNode, error) {
	var err error
	var current = p.node

	ot := newAstNode(NodeOperationType, p, self.Token)
	self.Children = append(self.Children, ot)

	if p.node.Token.ID == TokenName {

		// Append the query name and move on

		self.Children = append(self.Children, p.node)
		p.node, err = p.next()
	}

	if err == nil && p.node.Token.Val == "(" {

		// Parse variable definition

		p.isVarDef = true

		if current, err = p.run(0); err == nil {
			self.Children = append(self.Children, current)
		}

		p.isVarDef = false
	}

	if err == nil && p.node.Token.Val == "@" {

		// Parse directive

		if current, err = p.run(0); err == nil {
			self.Children = append(self.Children, current)
		}
	}

	if err == nil && p.node.Token.Val == "{" {

		// Parse selection set

		if current, err = p.run(0); err == nil {
			self.Children = append(self.Children, current)
		}

	} else if err == nil {

		// Selection Set is mandatory

		err = p.newParserError(ErrSelectionSetExpected,
			current.Token.String(), *current.Token)
	}

	return self, err
}

/*
ndFragmentDefinition parses a fragment definition. Each fragment is
represented by an optional fragment name, a type condition and a selection set.
(@spec 2.8)
*/
func ndFragmentDefinition(p *parser, self *ASTNode) (*ASTNode, error) {
	var err error
	var current *ASTNode

	if p.node.Token.ID == TokenName {

		// Append the fragment name and move on

		changeAstNode(p.node, NodeFragmentName, p)
		self.Children = append(self.Children, p.node)

		p.node, err = p.next()

	} else {

		err = p.newParserError(ErrNameExpected,
			p.node.Token.String(), *p.node.Token)
	}

	if err == nil {

		if p.node.Token.Val != "on" {

			// Type conditions must start with on

			err = p.newParserError(ErrOnExpected,
				p.node.Token.String(), *p.node.Token)

		} else {
			p.node, err = p.next()

			if err == nil {
				if p.node.Token.ID == TokenName {

					// Append the fragment name

					changeAstNode(p.node, NodeTypeCondition, p)
					self.Children = append(self.Children, p.node)
					p.node, err = p.next()

				} else {

					err = p.newParserError(ErrNameExpected,
						p.node.Token.String(), *p.node.Token)
				}
			}
		}
	}

	if err == nil && p.node.Token.Val == "@" {

		// Parse directive

		if current, err = p.run(0); err == nil {
			self.Children = append(self.Children, current)
		}
	}

	if err == nil && p.node.Token.Val == "{" {

		// Parse selection set

		if current, err = p.run(0); err == nil {
			self.Children = append(self.Children, current)
		}

	} else if err == nil {

		// Selection Set is mandatory

		err = p.newParserError(ErrSelectionSetExpected,
			p.node.Token.String(), *p.node.Token)
	}

	return self, err
}

/*
ndSelectionSet parses a selection set. An operation selects the set of
information it needs. (@spec 2.4)
*/
func ndSelectionSet(p *parser, self *ASTNode) (*ASTNode, error) {
	var current *ASTNode
	var err error

	// Special case if we are parsing an input object literal (@spec 2.9.8)

	if p.isValue {
		return ndInputObject(p, self)
	}

	for p.node.Token.ID != TokenEOF && p.node.Token.Val != "}" {

		if p.node.Token.Val == "..." {

			// Add a simple fragment spread

			if current, err = p.run(0); err == nil {
				self.Children = append(self.Children, current)
			}

		} else {

			err = acceptFieldExpression(p, self)
		}

		if err != nil {
			return nil, err
		}
	}

	return self, skipToken(p, "}")
}

/*
acceptFieldExpression parses a field expression. (@spec 2.5, 2.6, 2.7)
*/
func acceptFieldExpression(p *parser, self *ASTNode) error {
	var err error

	// Field node gets the first token in the field expression

	fe := newAstNode(NodeField, p, p.node.Token)
	self.Children = append(self.Children, fe)

	current := p.node

	if p.node, err = p.next(); err == nil && p.node.Name != NodeEOF {

		if p.node.Token.Val == ":" {

			// Last node was an Alias not a name

			changeAstNode(current, NodeAlias, p)

			// Append Alias to Field children and move on

			fe.Children = append(fe.Children, current)

			if p.node, err = p.next(); err == nil && p.node.Name != NodeEOF {
				current = p.node
				p.node, err = p.next()
			}
		}

		if err == nil && p.node.Name != NodeEOF {

			// Next node must be a Name

			if current.Name == NodeName {

				// Append Name to Field children and move on

				fe.Children = append(fe.Children, current)

			} else {

				err = p.newParserError(ErrNameExpected,
					current.Token.String(), *current.Token)
			}

			if err == nil && p.node.Token.Val == "(" {

				// Parse arguments

				if current, err = p.run(0); err == nil {
					fe.Children = append(fe.Children, current)
				}
			}

			if err == nil && p.node.Token.Val == "@" {

				// Parse directives

				if current, err = p.run(0); err == nil {
					fe.Children = append(fe.Children, current)
				}
			}

			if err == nil && p.node.Token.Val == "{" {

				// Parse nested selection set

				if current, err = p.run(0); err == nil {
					fe.Children = append(fe.Children, current)
				}
			}
		}
	}

	return err
}

/*
ndArgsOrVarDef parses an argument or variable definition expression. (@spec 2.6, 2.10)
*/
func ndArgsOrVarDef(p *parser, self *ASTNode) (*ASTNode, error) {
	var err error
	var args, arg, current *ASTNode

	// Create a list token

	if p.isVarDef {
		args = newAstNode(NodeVariableDefinitions, p, p.node.Token)
	} else {
		args = newAstNode(NodeArguments, p, p.node.Token)
	}

	for err == nil && p.node.Token.ID != TokenEOF && p.node.Token.Val != ")" {

		if p.isVarDef {
			arg = newAstNode(NodeVariableDefinition, p, p.node.Token)
		} else {
			arg = newAstNode(NodeArgument, p, p.node.Token)
		}

		args.Children = append(args.Children, arg)

		if current, err = p.run(0); err == nil {

			if !p.isVarDef && current.Name != NodeName {
				err = p.newParserError(ErrNameExpected,
					current.Token.String(), *current.Token)

			} else if p.isVarDef && current.Name != NodeVariable {
				err = p.newParserError(ErrVariableExpected,
					current.Token.String(), *current.Token)

			} else {

				// Add name

				arg.Children = append(arg.Children, current)

				if err = skipToken(p, ":"); err == nil {

					// Add value

					if p.isVarDef {
						if current, err = p.run(0); err == nil {
							changeAstNode(current, NodeType, p)
							arg.Children = append(arg.Children, current)
						}
					} else {
						if current, err = parseValue(p); err == nil {
							arg.Children = append(arg.Children, current)
						}
					}

					if err == nil && p.isVarDef && p.node.Token.Val == "=" {

						skipToken(p, "=")

						// Parse default value

						if current, err = parseValue(p); err == nil {
							changeAstNode(current, NodeDefaultValue, p)
							arg.Children = append(arg.Children, current)
						}
					}

				}
			}
		}
	}

	// Must have a closing bracket

	if err == nil {
		return args, skipToken(p, ")")
	}

	return nil, err
}

/*
parseValue parses a value and returns the result. (@spec 2.9)
*/
func parseValue(p *parser) (*ASTNode, error) {
	p.isValue = true
	current, err := p.run(0)
	p.isValue = false

	if err == nil {

		if current.Token.Val == "true" ||
			current.Token.Val == "false" ||
			current.Token.Val == "null" ||
			current.Token.ID == TokenIntValue ||
			current.Token.ID == TokenFloatValue ||
			current.Token.ID == TokenStringValue {

			// Simple constant values

			changeAstNode(current, NodeValue, p)

		} else if current.Name == NodeName {

			// Enum values

			changeAstNode(current, NodeEnumValue, p)

		} else {

			// Everything else must be a variable or a complex data type

			errorutil.AssertTrue(current.Name == NodeVariable ||
				current.Name == NodeListValue ||
				current.Name == NodeObjectValue, fmt.Sprint("Unexpected value node:", current))
		}

		if err == nil {
			return current, err
		}
	}

	return nil, err
}

/*
ndDirectives parses a directive expression. (@spec 2.12)
*/
func ndDirectives(p *parser, self *ASTNode) (*ASTNode, error) {
	var err error
	var current *ASTNode

	dir := newAstNode(NodeDirective, p, p.node.Token)

	if err = acceptChild(p, dir, TokenName); err == nil {

		if current, err = p.run(0); err == nil {

			dir.Children = append(dir.Children, current)
			self.Children = append(self.Children, dir)

			if p.node.Token.Val == "@" {

				if p.node, err = p.next(); err == nil {
					return ndDirectives(p, self)
				}
			}
		}
	}

	return self, err
}

// Helper functions
// ================

/*
skipToken skips over a token if it has one of the given valid values.
*/
func skipToken(p *parser, validValues ...string) error {
	var err error

	canSkip := func(val string) bool {
		for _, i := range validValues {
			if i == val {
				return true
			}
		}
		return false
	}

	if !canSkip(p.node.Token.Val) {

		if p.node.Token.ID == TokenEOF {
			return p.newParserError(ErrUnexpectedEnd, "", *p.node.Token)
		}

		return p.newParserError(ErrUnexpectedToken, p.node.Token.Val, *p.node.Token)
	}

	// This should never return an error unless we skip over EOF or complex tokens
	// like values

	p.node, err = p.next()

	return err
}

/*
acceptChild accepts the current token as a child.
*/
func acceptChild(p *parser, self *ASTNode, id LexTokenID) error {
	var err error

	current := p.node

	if p.node, err = p.next(); err == nil {

		if current.Token.ID == id {
			self.Children = append(self.Children, current)
		} else {
			err = p.newParserError(ErrUnexpectedToken, current.Token.Val, *current.Token)
		}
	}

	return err
}
