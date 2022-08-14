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
	"os"
	"strings"
	"testing"
)

func TestSimpleExpressionPrinting(t *testing.T) {

	input := `query {
  likeStory(storyID: 12345) {
    story {
      likeCount
    }
  }
}`

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

	if err := testPrettyPrinting(input, expectedOutput,
		input); err != nil {
		t.Error(err)
		return
	}

	/*
	   From the spec 2.9.4 Strings

	   Since block strings represent freeform text often used in indented positions,
	   the string value semantics of a block string excludes uniform indentation and
	   blank initial and trailing lines via BlockStringValue().

	   For example, the following operation containing a block string:

	   mutation {
	     sendEmail(message: """
	       Hello,
	         World!

	       Yours,
	         GraphQL.
	     """)
	   }

	   Is identical to the standard quoted string:

	   mutation {
	     sendEmail(message: "Hello,\n  World!\n\nYours,\n  GraphQL.")
	   }
	*/

	input = `{
  foo(bar: """
    Hello,
      World!

    Yours,
      GraphQL.
  """)                      # Block string value
}
`
	expectedOutput = `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Name: foo
          Arguments
            Argument
              Name: bar
              Value: Hello,
  World!

Yours,
  GraphQL.
`[1:]

	astres, err := ParseWithRuntime("mytest", input, &TestRuntimeProvider{})
	if err != nil || fmt.Sprint(astres) != expectedOutput {
		t.Error(fmt.Sprintf("Unexpected parser output:\n%v expected was:\n%v Error: %v", astres, expectedOutput, err))
		return
	}

	ppOutput := `{
  foo(bar: """Hello,
    World!

  Yours,
    GraphQL.""")
}`

	ppres, err := PrettyPrint(astres)
	if err != nil || ppres != ppOutput {
		fmt.Fprintf(os.Stderr, "#\n%v#", ppres)
		t.Error(fmt.Sprintf("Unexpected result:\n%v\nError: %v", ppres, err))
		return
	}

	val := astres.Children[0].Children[0].Children[0].Children[0].Children[1].Children[0].Children[1].Token.Val
	if val != "Hello,\n  World!\n\nYours,\n  GraphQL." {
		t.Error("Unexpected result:", val)
	}

	input = `{
  foo(bar: $Hello)        # Variable value
  foo(bar: 1)             # Int value
  foo(bar: 1.1)           # Float value
  foo(bar: "Hello")       # String value
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

	expectedOutput = `
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

	expectedPPResult := `{
  foo(bar: $Hello)
  foo(bar: 1)
  foo(bar: 1.1)
  foo(bar: "Hello")
  foo(bar: false)
  foo(bar: null)
  foo(bar: MOBILE_WEB)
  foo(bar: [1, 2, [A, "B"]])
  foo(bar: {foo : "bar", foo2 : [12], foo3 : {X : Y}})
}`

	if err := testPrettyPrinting(input, expectedOutput,
		expectedPPResult); err != nil {
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

	if err := testPrettyPrinting(input, expectedOutput,
		input); err != nil {
		t.Error(err)
		return
	}

	input = `query getBozoProfile ($devicePicSize: Int, $foo: bar=123) {
  user(id: 4) {
    id
    name
    profilePic(size: $devicePicSize)
  }
}`

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

	if err := testPrettyPrinting(input, expectedOutput,
		input); err != nil {
		t.Error(err)
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
}`[1:]

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

	if err := testPrettyPrinting(input, expectedOutput,
		input); err != nil {
		t.Error(err)
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
}`[1:]

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

	if err := testPrettyPrinting(input, expectedOutput,
		input); err != nil {
		t.Error(err)
		return
	}

	input = `
{
  my : field(size: 4) @include(if: true) @id() @foo(x: 1, y: "z")
}`[1:]

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

	if err := testPrettyPrinting(input, expectedOutput,
		input); err != nil {
		t.Error(err)
		return
	}
}

func TestErrorCases(t *testing.T) {

	astres, _ := ParseWithRuntime("mytest", `{ a }`, &TestRuntimeProvider{})
	astres.Children[0].Name = "foo"
	_, err := PrettyPrint(astres)

	if err == nil || err.Error() != "Could not find template for foo (tempkey: foo_1)" {
		t.Error("Unexpected result:", err)
		return
	}

	astres, err = ParseWithRuntime("mytest", `{ a(b:"""a"a""" x:""){ d} }`, &TestRuntimeProvider{})
	if err != nil {
		t.Error(err)
		return
	}

	pp, _ := PrettyPrint(astres)

	astres, err = ParseWithRuntime("mytest", pp, &TestRuntimeProvider{})
	if err != nil {
		t.Error(err)
		return
	}

	if astres.String() != `
Document
  ExecutableDefinition
    OperationDefinition
      SelectionSet
        Field
          Name: a
          Arguments
            Argument
              Name: b
              Value: a"a
            Argument
              Name: x
              Value: 
          SelectionSet
            Field
              Name: d
`[1:] {
		t.Error("Unexpected result:", astres)
		return
	}

	plainAST := astres.Plain()

	plainAST["children"].([]map[string]interface{})[0]["name"] = "Value"
	_, err = ASTFromPlain(plainAST)

	if err == nil || err.Error() != "Found plain ast value node without a value: Value" {
		t.Error("Unexpected result:", err)
		return
	}

	delete(plainAST["children"].([]map[string]interface{})[0], "name")
	_, err = ASTFromPlain(plainAST)

	if err == nil || !strings.HasPrefix(err.Error(), "Found plain ast node without a name") {
		t.Error("Unexpected result:", err)
		return
	}
}

func testPPOut(input string) (string, error) {
	var ppres string

	astres, err := ParseWithRuntime("mytest", input, &TestRuntimeProvider{})

	if err == nil {
		ppres, err = PrettyPrint(astres)
	}

	return ppres, err
}

func testPrettyPrinting(input, astOutput, ppOutput string) error {

	astres, err := ParseWithRuntime("mytest", input, &TestRuntimeProvider{})
	if err != nil || fmt.Sprint(astres) != astOutput {
		return fmt.Errorf("Unexpected parser output:\n%v expected was:\n%v Error: %v", astres, astOutput, err)
	}

	ppres, err := PrettyPrint(astres)
	if err != nil || ppres != ppOutput {
		fmt.Fprintf(os.Stderr, "#\n%v#", ppres)
		return fmt.Errorf("Unexpected result:\n%v\nError: %v", ppres, err)
	}

	// Make sure the pretty printed result is valid and gets the same parse tree

	astres2, err := ParseWithRuntime("mytest", ppres, &TestRuntimeProvider{})
	if err != nil || fmt.Sprint(astres2) != astOutput {
		return fmt.Errorf("Unexpected parser output from pretty print string:\n%v expected was:\n%v Error: %v", astres2, astOutput, err)
	}

	mAST, _ := json.Marshal(astres2.Plain())
	var plainAST interface{}
	json.Unmarshal(mAST, &plainAST)

	ast, err := ASTFromPlain(plainAST.(map[string]interface{}))
	if err != nil {
		return fmt.Errorf("Could not build AST from plain AST: %v", err)
	}

	ppres2, err := PrettyPrint(ast)
	if err != nil {
		return fmt.Errorf("Could not pretty print plain AST: %v", err)
	}

	if ppres != ppres2 {
		return fmt.Errorf("Unexpected pretty print - normal:\n%v\nFrom plain:\n%v", ppres, ppres2)
	}

	return nil
}
