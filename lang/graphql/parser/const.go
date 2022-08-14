/*
 * Public Domain Software
 *
 * I (Matthias Ladkau) am the author of the source code in this file.
 * I have placed the source code in this file in the public domain.
 *
 * For further information see: http://creativecommons.org/publicdomain/zero/1.0/
 */

/*
Package parser contains a GraphQL parser. Based on GraphQL spec June 2018.

Lexer for Source Text - @spec 2.1

Lex() is a lexer function to convert a given search query into a list of tokens.

Based on a talk by Rob Pike: Lexical Scanning in Go

https://www.youtube.com/watch?v=HxaD_trXwRE

The lexer's output is pushed into a channel which is consumed by the parser.
This design enables the concurrent processing of the input text by lexer and
parser.

Parser

Parse() is a parser which produces a parse tree from a given set of lexer tokens.

Based on an article by Douglas Crockford: Top Down Operator Precedence

http://crockford.com/javascript/tdop/tdop.html

which is based on the ideas of Vaughan Pratt and his paper: Top Down Operator Precedence

http://portal.acm.org/citation.cfm?id=512931
https://tdop.github.io/

ParseWithRuntime() parses a given input and decorates the resulting parse tree
with runtime components which can be used to interpret the parsed query.
*/
package parser

/*
LexTokenID represents a unique lexer token ID
*/
type LexTokenID int

/*
Available lexer token types
*/
const (
	TokenError LexTokenID = iota // Lexing error token with a message as val
	TokenEOF                     // End-of-file token

	// Punctuators - @spec 2.1.8

	// GraphQL documents include punctuation in order to describe structure.
	// GraphQL is a data description language and not a programming language,
	// therefore GraphQL lacks the punctuation often used to describe mathematical expressions.

	TokenPunctuator

	// Names - @spec 2.1.9

	// GraphQL Documents are full of named things: operations, fields, arguments, types,
	// directives, fragments, and variables. All names must follow the same grammatical
	// form. Names in GraphQL are case‐sensitive. That is to say name, Name, and NAME
	// all refer to different names. Underscores are significant, which means
	// other_name and othername are two different names. Names in GraphQL are limited
	// to this ASCII subset of possible characters to support interoperation with as
	// many other systems as possible.

	TokenName

	// Integer value - @spec 2.9.1

	// An Integer number is specified without a decimal point or exponent (ex. 1).

	TokenIntValue

	// Float value - @spec 2.9.2

	// A Float number includes either a decimal point (ex. 1.0) or an exponent
	// (ex. 1e50) or both (ex. 6.0221413e23).

	TokenFloatValue

	// String Value - @spec 2.9.4

	// Strings are sequences of characters wrapped in double‐quotes (").
	// (ex. "Hello World"). White space and other otherwise‐ignored characters are
	// significant within a string value. Unicode characters are allowed within String
	// value literals, however SourceCharacter must not contain some ASCII control
	// characters so escape sequences must be used to represent these characters.

	TokenStringValue

	// General token used for plain ASTs

	TokenGeneral
)

/*
Available parser AST node types
*/
const (
	NodeAlias                = "Alias"
	NodeArgument             = "Argument"
	NodeArguments            = "Arguments"
	NodeDefaultValue         = "DefaultValue"
	NodeDirective            = "Directive"
	NodeDirectives           = "Directives"
	NodeDocument             = "Document"
	NodeEnumValue            = "EnumValue"
	NodeEOF                  = "EOF"
	NodeExecutableDefinition = "ExecutableDefinition"
	NodeField                = "Field"
	NodeFragmentDefinition   = "FragmentDefinition"
	NodeFragmentName         = "FragmentName"
	NodeFragmentSpread       = "FragmentSpread"
	NodeInlineFragment       = "InlineFragment"
	NodeListValue            = "ListValue"
	NodeName                 = "Name"
	NodeObjectField          = "ObjectField"
	NodeObjectValue          = "ObjectValue"
	NodeOperationDefinition  = "OperationDefinition"
	NodeOperationType        = "OperationType"
	NodeSelectionSet         = "SelectionSet"
	NodeType                 = "Type"
	NodeTypeCondition        = "TypeCondition"
	NodeValue                = "Value"
	NodeVariable             = "Variable"
	NodeVariableDefinition   = "VariableDefinition"
	NodeVariableDefinitions  = "VariableDefinitions"
)

/*
ValueNodes are AST nodes which contain a significant value
*/
var ValueNodes = []string{
	NodeAlias,
	NodeDefaultValue,
	NodeEnumValue,
	NodeFragmentName,
	NodeFragmentSpread,
	NodeName,
	NodeObjectField,
	NodeOperationType,
	NodeType,
	NodeTypeCondition,
	NodeValue,
	NodeVariable,
}
