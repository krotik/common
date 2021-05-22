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
	"encoding/json"
	"fmt"
	"testing"
)

/*
Test RuntimeProvider provides runtime components for a parse tree.
*/
type TestRuntimeProvider struct {
}

/*
Runtime returns a runtime component for a given ASTNode.
*/
func (trp *TestRuntimeProvider) Runtime(node *ASTNode) Runtime {
	return &TestRuntime{}
}

/*
Test Runtime provides the runtime for an ASTNode.
*/
type TestRuntime struct {
}

/*
Validate this runtime component and all its child components.
*/
func (tr *TestRuntime) Validate() error {
	return nil
}

/*
Eval evaluate this runtime component.
*/
func (tr *TestRuntime) Eval() (map[string]interface{}, error) {
	return nil, nil
}

func TestInputValueParsing(t *testing.T) {

	input := `{
  foo(bar: $Hello)        # Variable value
  foo(bar: 1)             # Int value
  foo(bar: 1.1)           # Float value
  foo(bar: "Hello")       # String value
  foo(bar: """
               Hello
               test123
""")                      # Block string value
  foo(bar: false)         # Boolean value
  foo(bar: null)          # Null value
  foo(bar: MOBILE_WEB)    # Enum value
  foo(bar: [1,2,[A,"B"]]) # List value
  foo(bar: {foo:"bar" 
    foo2 : [12],
    foo3 : { X:Y }
    })         # Object value
}
`
	expectedOutput := `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Variable: Hello
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Value: 1
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Value: 1.1
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Value: Hello
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Value: Hello
test123
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Value: false
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Value: null
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              EnumValue: MOBILE_WEB
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              ListValue
                Value: 1
                Value: 2
                ListValue
                  EnumValue: A
                  Value: B
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              ObjectValue
                ObjectField: foo
                  Value: bar
                ObjectField: foo2
                  ListValue
                    Value: 12
                ObjectField: foo3
                  ObjectValue
                    ObjectField: X
                      EnumValue: Y
`[1:]

	if res, err := ParseWithRuntime("mytest", input, &TestRuntimeProvider{}); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}
}

func TestOperationDefinitionParsing(t *testing.T) {

	input := `query {
  likeStory(storyID: 12345) {
    story {
      likeCount
    }
  }
}
`
	expectedOutput := `
Document
  ExecutableDefinition
    OperationDefinition
      OperationType: query
      SelectionSet
        Field
          Name: likeStory
          Arguments
            Argument
              Name: storyID
              Value: 12345
          SelectionSet
            Field
              Name: story
              SelectionSet
                Field
                  Name: likeCount
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	expectedOutput = `
{
  "children": [
    {
      "children": [
        {
          "children": [
            {
              "name": "OperationType",
              "value": "query"
            },
            {
              "children": [
                {
                  "children": [
                    {
                      "name": "Name",
                      "value": "likeStory"
                    },
                    {
                      "children": [
                        {
                          "children": [
                            {
                              "name": "Name",
                              "value": "storyID"
                            },
                            {
                              "name": "Value",
                              "value": "12345"
                            }
                          ],
                          "name": "Argument"
                        }
                      ],
                      "name": "Arguments"
                    },
                    {
                      "children": [
                        {
                          "children": [
                            {
                              "name": "Name",
                              "value": "story"
                            },
                            {
                              "children": [
                                {
                                  "children": [
                                    {
                                      "name": "Name",
                                      "value": "likeCount"
                                    }
                                  ],
                                  "name": "Field"
                                }
                              ],
                              "name": "SelectionSet"
                            }
                          ],
                          "name": "Field"
                        }
                      ],
                      "name": "SelectionSet"
                    }
                  ],
                  "name": "Field"
                }
              ],
              "name": "SelectionSet"
            }
          ],
          "name": "OperationDefinition"
        }
      ],
      "name": "ExecutableDefinition"
    }
  ],
  "name": "Document"
}`[1:]

	res, err := Parse("mytest", input)
	plain, _ := json.MarshalIndent(res.Plain(), "", "  ")
	plainString := string(plain)

	if err != nil || plainString != expectedOutput {
		t.Error("Unexpected parser output:\n", plainString, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `query hans {
  likeStory(storyID: 12345) {
    story {
      likeCount
    }
  }
}
`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      OperationType: query
      Name: hans
      SelectionSet
        Field
          Name: likeStory
          Arguments
            Argument
              Name: storyID
              Value: 12345
          SelectionSet
            Field
              Name: story
              SelectionSet
                Field
                  Name: likeCount
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `subscription hans @foo(a:1) @bar(b:2) {
      likeCount
}
`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      OperationType: subscription
      Name: hans
      Directives
        Directive
          Name: foo
          Arguments
            Argument
              Name: a
              Value: 1
        Directive
          Name: bar
          Arguments
            Argument
              Name: b
              Value: 2
      SelectionSet
        Field
          Name: likeCount
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
query getBozoProfile($devicePicSize: Int, $foo: bar=123) {
  user(id: 4) {
    id
    name
    profilePic(size: $devicePicSize)
  }
}
`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      OperationType: query
      Name: getBozoProfile
      VariableDefinitions
        VariableDefinition
          Variable: devicePicSize
          Type: Int
        VariableDefinition
          Variable: foo
          Type: bar
          DefaultValue: 123
      SelectionSet
        Field
          Name: user
          Arguments
            Argument
              Name: id
              Value: 4
          SelectionSet
            Field
              Name: id
            Field
              Name: name
            Field
              Name: profilePic
              Arguments
                Argument
                  Name: size
                  Variable: devicePicSize
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `query (foo:bar) {}`
	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		"Parse error in mytest: Variable expected (<foo>) (Line:1 Pos:8)" {
		t.Error("Unexpected output:", out, err)
		return
	}

	input = `query @foo() ()  {}`
	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Selection Set expected (@) (Line:1 Pos:7)` {
		t.Error("Unexpected output:", out, err)
		return
	}
}

func TestParserErrors(t *testing.T) {

	input := `"bl\*a`
	if _, err := Parse("mytest", input); err.Error() !=
		"Parse error in mytest: Lexical error (EOF inside quotes) (Line:1 Pos:1)" {
		t.Error(err)
		return
	}

	input = `"bl\*a"`
	if _, err := Parse("mytest", input); err.Error() !=
		"Parse error in mytest: Lexical error (Could not interpret escape sequence: invalid syntax) (Line:1 Pos:1)" {
		t.Error(err)
		return
	}

	input = `{ "bla"`
	if _, err := Parse("mytest", input); err.Error() !=
		"Parse error in mytest: Unexpected end (Line:1 Pos:7)" {
		t.Error(err)
		return
	}

	input = `{ bla : "bla" }`
	if _, err := Parse("mytest", input); err.Error() !=
		`Parse error in mytest: Name expected ("bla") (Line:1 Pos:9)` {
		t.Error(err)
		return
	}

	tokens := make(chan LexToken, 1)
	close(tokens)
	p := &parser{"test", nil, tokens, nil, false, false}

	if _, err := p.next(); err == nil || err.Error() != "Parse error in test: Unexpected end (Line:0 Pos:0)" {
		t.Error(err)
		return
	}

	tokens = make(chan LexToken, 1)
	tokens <- LexToken{-1, 0, "foo", 0, 0}
	close(tokens)
	p = &parser{"test", nil, tokens, nil, false, false}

	if _, err := p.next(); err == nil || err.Error() != `Parse error in test: Unknown term (id:-1 (foo)) (Line:0 Pos:0)` {
		t.Error(err)
		return
	}

	// Test incomplete expression

	input = `{ a `
	if _, err := Parse("mytest", input); err.Error() !=
		"Parse error in mytest: Unexpected end (Line:1 Pos:4)" {
		t.Error(err)
		return
	}

	input = `[ 11, "tes`
	if _, err := Parse("mytest", input); err == nil || err.Error() !=
		"Parse error in mytest: Lexical error (EOF inside quotes) (Line:1 Pos:7)" {
		t.Error(err)
		return
	}

	input = `[ { "a"`
	if _, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Name expected ("a") (Line:1 Pos:5)` {
		t.Error(err)
		return
	}

	input = `{
  foo(bar: {)
}`
	if _, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Term cannot start an expression (}) (Line:3 Pos:2)` {
		t.Error(err)
		return
	}

	input = `{
  foo(bar: { a b })
}`
	if _, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Unexpected term (b) (Line:2 Pos:17)` {
		t.Error(err)
		return
	}

	input = `@1`
	if _, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Unexpected term (1) (Line:1 Pos:2)` {
		t.Error(err)
		return
	}
}

func TestQueryShorthandParsing(t *testing.T) {

	// Test shorthand operation

	input := `{ field }`
	expectedOutput := `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Name: field
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{ field }{ field }`

	if _, err := Parse("mytest", input); err.Error() !=
		`Parse error in mytest: Query shorthand only allowed for one query operation ({) (Line:1 Pos:10)` {
		t.Error(err)
		return
	}

	input = `{ 
	my : field
}`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Alias: my
          Name: field
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{ 
	my : field
	foo,
	bar
}`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Alias: my
          Name: field
        Field
          Name: foo
        Field
          Name: bar
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{ 
	my : field(size : 4)
}`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Alias: my
          Name: field
          Arguments
            Argument
              Name: size
              Value: 4
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{ my : field(4 : 4) }`
	if _, err := Parse("mytest", input); err == nil || err.Error() !=
		"Parse error in mytest: Name expected (int(4)) (Line:1 Pos:14)" {
		t.Error("Unexpected result:", err)
		return
	}

	input = `{ 
	my : field(size : 4, fred : "boo"),
	test(x:12.5e-50)
	foo
	bar(x:"[")
}`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Alias: my
          Name: field
          Arguments
            Argument
              Name: size
              Value: 4
            Argument
              Name: fred
              Value: boo
        Field
          Name: test
          Arguments
            Argument
              Name: x
              Value: 12.5e-50
        Field
          Name: foo
        Field
          Name: bar
          Arguments
            Argument
              Name: x
              Value: [
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{ 
	my : field(size : 4) @include(if: true)
}`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Alias: my
          Name: field
          Arguments
            Argument
              Name: size
              Value: 4
          Directives
            Directive
              Name: include
              Arguments
                Argument
                  Name: if
                  Value: true
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{ 
	my : field(size : 4) @include(if: true) @id() @foo(x:1 y:"z")
}`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Alias: my
          Name: field
          Arguments
            Argument
              Name: size
              Value: 4
          Directives
            Directive
              Name: include
              Arguments
                Argument
                  Name: if
                  Value: true
            Directive
              Name: id
              Arguments
            Directive
              Name: foo
              Arguments
                Argument
                  Name: x
                  Value: 1
                Argument
                  Name: y
                  Value: z
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `# "user" represents one of many users in a graph of data, referred to by a
# unique identifier.
{
  user(id: 4) {
    name
  }
}`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Name: user
          Arguments
            Argument
              Name: id
              Value: 4
          SelectionSet
            Field
              Name: name
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{
  me {
    id
    firstName
    lastName
    birthday {
      month
      day
    }
    friends {
      name
    }
  }
}
`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Name: me
          SelectionSet
            Field
              Name: id
            Field
              Name: firstName
            Field
              Name: lastName
            Field
              Name: birthday
              SelectionSet
                Field
                  Name: month
                Field
                  Name: day
            Field
              Name: friends
              SelectionSet
                Field
                  Name: name
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `{ 
	my : field(size : 4) @x(a:"b") @y(a:$b,c:$d) {
		foo : field(size : 4) @x(a:"b") @y(a:"b",c:"d")
		bar : test,
	}
}`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Alias: my
          Name: field
          Arguments
            Argument
              Name: size
              Value: 4
          Directives
            Directive
              Name: x
              Arguments
                Argument
                  Name: a
                  Value: b
            Directive
              Name: y
              Arguments
                Argument
                  Name: a
                  Variable: b
                Argument
                  Name: c
                  Variable: d
          SelectionSet
            Field
              Alias: foo
              Name: field
              Arguments
                Argument
                  Name: size
                  Value: 4
              Directives
                Directive
                  Name: x
                  Arguments
                    Argument
                      Name: a
                      Value: b
                Directive
                  Name: y
                  Arguments
                    Argument
                      Name: a
                      Value: b
                    Argument
                      Name: c
                      Value: d
            Field
              Alias: bar
              Name: test
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

}

func TestFragmentParsing(t *testing.T) {

	input := `
fragment friendFields on User {
  id
  name
  profilePic(size: 50)
}
`
	expectedOutput := `
Document
  ExecutableDefinition
    FragmentDefinition
      FragmentName: friendFields
      TypeCondition: User
      SelectionSet
        Field
          Name: id
        Field
          Name: name
        Field
          Name: profilePic
          Arguments
            Argument
              Name: size
              Value: 50
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
fragment friendFields on User @foo() {
  id
  name
  profilePic(size: 50)
}
`
	expectedOutput = `
Document
  ExecutableDefinition
    FragmentDefinition
      FragmentName: friendFields
      TypeCondition: User
      Directives
        Directive
          Name: foo
          Arguments
      SelectionSet
        Field
          Name: id
        Field
          Name: name
        Field
          Name: profilePic
          Arguments
            Argument
              Name: size
              Value: 50
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
query withNestedFragments {
  user(id: 4) {
    friends(first: 10) {
      ...friendFields
    }
    mutualFriends(first: 10) {
      ...friendFields
    }
  }
}

fragment friendFields on User {
  id
  name
  ...standardProfilePic
}

fragment standardProfilePic on User {
  profilePic(size: 50)
}
`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      OperationType: query
      Name: withNestedFragments
      SelectionSet
        Field
          Name: user
          Arguments
            Argument
              Name: id
              Value: 4
          SelectionSet
            Field
              Name: friends
              Arguments
                Argument
                  Name: first
                  Value: 10
              SelectionSet
                FragmentSpread: friendFields
            Field
              Name: mutualFriends
              Arguments
                Argument
                  Name: first
                  Value: 10
              SelectionSet
                FragmentSpread: friendFields
  ExecutableDefinition
    FragmentDefinition
      FragmentName: friendFields
      TypeCondition: User
      SelectionSet
        Field
          Name: id
        Field
          Name: name
        FragmentSpread: standardProfilePic
  ExecutableDefinition
    FragmentDefinition
      FragmentName: standardProfilePic
      TypeCondition: User
      SelectionSet
        Field
          Name: profilePic
          Arguments
            Argument
              Name: size
              Value: 50
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
query inlineFragmentTyping {
  profiles(handles: ["zuck", "cocacola"]) {
    handle
    ... on User {
      friends {
        count
      }
    }
    ... on Page {
      likers {
        count
      }
    }
  }
}
`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      OperationType: query
      Name: inlineFragmentTyping
      SelectionSet
        Field
          Name: profiles
          Arguments
            Argument
              Name: handles
              ListValue
                Value: zuck
                Value: cocacola
          SelectionSet
            Field
              Name: handle
            InlineFragment
              TypeCondition: User
              SelectionSet
                Field
                  Name: friends
                  SelectionSet
                    Field
                      Name: count
            InlineFragment
              TypeCondition: Page
              SelectionSet
                Field
                  Name: likers
                  SelectionSet
                    Field
                      Name: count
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
query inlineFragmentNoType($expandedInfo: Boolean) {
  user(handle: "zuck") {
    id
    name
    ... @include(if: $expandedInfo) {
      firstName
      lastName
      birthday
    }
  }
}
`

	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      OperationType: query
      Name: inlineFragmentNoType
      VariableDefinitions
        VariableDefinition
          Variable: expandedInfo
          Type: Boolean
      SelectionSet
        Field
          Name: user
          Arguments
            Argument
              Name: handle
              Value: zuck
          SelectionSet
            Field
              Name: id
            Field
              Name: name
            InlineFragment
              Directives
                Directive
                  Name: include
                  Arguments
                    Argument
                      Name: if
                      Variable: expandedInfo
              SelectionSet
                Field
                  Name: firstName
                Field
                  Name: lastName
                Field
                  Name: birthday
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `fragment foo {}`
	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Type condition starting with 'on' expected ({) (Line:1 Pos:14)` {
		t.Error("Unexpected output:", out, err)
		return
	}

	input = `fragment on foo {}`
	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Type condition starting with 'on' expected (<foo>) (Line:1 Pos:13)` {
		t.Error("Unexpected output:", out, err)
		return
	}

	input = `
{
  user(n:1) {
    ...on
  }
}
fragment on on User {
  id
}
`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Name: user
          Arguments
            Argument
              Name: n
              Value: 1
          SelectionSet
            FragmentSpread: on
  ExecutableDefinition
    FragmentDefinition
      FragmentName: on
      TypeCondition: User
      SelectionSet
        Field
          Name: id
`[1:]

	if res, err := Parse("mytest", input); err != nil || fmt.Sprint(res) != expectedOutput {
		t.Error("Unexpected parser output:\n", res, "expected was:\n", expectedOutput, "Error:", err)
		return
	}

	input = `
{
    ...1
}
`[1:]

	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Name expected (int(1)) (Line:2 Pos:9)` {
		t.Error("Unexpected output:", out, err)
		return
	}

	input = `
fragment {
    field
}
`[1:]

	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Name expected ({) (Line:1 Pos:10)` {
		t.Error("Unexpected output:", out, err)
		return
	}

	input = `
fragment foo on {
    field
}
`[1:]

	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Name expected ({) (Line:1 Pos:17)` {
		t.Error("Unexpected output:", out, err)
		return
	}

	input = `
fragment foo on bar
`[1:]

	if out, err := Parse("mytest", input); err == nil || err.Error() !=
		`Parse error in mytest: Selection Set expected (EOF) (Line:2 Pos:1)` {
		t.Error("Unexpected output:", out, err)
		return
	}
}
