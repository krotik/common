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
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/krotik/common/errorutil"
	"github.com/krotik/common/stringutil"
)

/*
IndentationLevel is the level of indentation which the pretty printer should use
*/
const IndentationLevel = 2

/*
Map of pretty printer templates for AST nodes

There is special treatment for NodeVALUE.
*/
var prettyPrinterMap = map[string]*template.Template{
	NodeArgument + "_2": template.Must(template.New(NodeArgument).Parse("{{.c1}}: {{.c2}}")),

	NodeOperationDefinition + "_1": template.Must(template.New(NodeArgument).Parse("{{.c1}}")),
	NodeOperationDefinition + "_2": template.Must(template.New(NodeArgument).Parse("{{.c1}} {{.c2}}")),
	NodeOperationDefinition + "_3": template.Must(template.New(NodeArgument).Parse("{{.c1}} {{.c2}} {{.c3}}")),
	NodeOperationDefinition + "_4": template.Must(template.New(NodeArgument).Parse("{{.c1}} {{.c2}} {{.c3}} {{.c4}}")),
	NodeOperationDefinition + "_5": template.Must(template.New(NodeArgument).Parse("{{.c1}} {{.c2}} {{.c3}} {{.c4}} {{.c5}}")),

	NodeFragmentDefinition + "_3": template.Must(template.New(NodeArgument).Parse("fragment {{.c1}} {{.c2}} {{.c3}}")),
	NodeFragmentDefinition + "_4": template.Must(template.New(NodeArgument).Parse("fragment {{.c1}} {{.c2}} {{.c3}} {{.c4}}")),

	NodeInlineFragment + "_1": template.Must(template.New(NodeArgument).Parse("... {{.c1}}\n")),
	NodeInlineFragment + "_2": template.Must(template.New(NodeArgument).Parse("... {{.c1}} {{.c2}}\n")),
	NodeInlineFragment + "_3": template.Must(template.New(NodeArgument).Parse("... {{.c1}} {{.c2}} {{.c3}}\n")),

	NodeExecutableDefinition + "_1": template.Must(template.New(NodeArgument).Parse("{{.c1}}")),

	NodeVariableDefinition + "_2": template.Must(template.New(NodeArgument).Parse("{{.c1}}: {{.c2}}")),
	NodeVariableDefinition + "_3": template.Must(template.New(NodeArgument).Parse("{{.c1}}: {{.c2}}{{.c3}}")),

	NodeDirective + "_1": template.Must(template.New(NodeArgument).Parse("@{{.c1}}")),
	NodeDirective + "_2": template.Must(template.New(NodeArgument).Parse("@{{.c1}}{{.c2}}")),
}

/*
PrettyPrint produces a pretty printed EQL query from a given AST.
*/
func PrettyPrint(ast *ASTNode) (string, error) {
	var visit func(ast *ASTNode, path []*ASTNode) (string, error)

	quoteValue := func(val string, allowNonQuotation bool) string {

		if val == "" {
			return `""`
		}

		isNumber, _ := regexp.MatchString("^[0-9][0-9\\.e-+]*$", val)
		isInlineString, _ := regexp.MatchString("^[a-zA-Z0-9_:.]*$", val)

		if allowNonQuotation && (isNumber || isInlineString) {
			return val
		} else if strings.ContainsRune(val, '"') {
			val = strings.Replace(val, "\"", "\\\"", -1)
		}
		if strings.Contains(val, "\n") {
			return fmt.Sprintf("\"\"\"%v\"\"\"", val)
		}
		return fmt.Sprintf("\"%v\"", val)
	}

	visit = func(ast *ASTNode, path []*ASTNode) (string, error) {

		// Handle special cases which don't have children but values

		if ast.Name == NodeValue {
			v := ast.Token.Val

			_, err := strconv.ParseFloat(v, 32)
			isNum := err == nil

			isConst := stringutil.IndexOf(v, []string{
				"true", "false", "null",
			}) != -1

			return quoteValue(ast.Token.Val, isConst || isNum), nil

		} else if ast.Name == NodeVariable {
			return fmt.Sprintf("$%v", ast.Token.Val), nil
		} else if ast.Name == NodeAlias {
			return fmt.Sprintf("%v :", ast.Token.Val), nil
		} else if ast.Name == NodeFragmentSpread {
			return ppPostProcessing(ast, path, fmt.Sprintf("...%v\n", ast.Token.Val)), nil
		} else if ast.Name == NodeTypeCondition {
			return fmt.Sprintf("on %v", ast.Token.Val), nil
		} else if ast.Name == NodeDefaultValue {
			return fmt.Sprintf("=%v", ast.Token.Val), nil
		}

		var children map[string]string
		var tempKey = ast.Name
		var buf bytes.Buffer

		// First pretty print children

		if len(ast.Children) > 0 {
			children = make(map[string]string)
			for i, child := range ast.Children {
				res, err := visit(child, append(path, child))
				if err != nil {
					return "", err
				}

				children[fmt.Sprint("c", i+1)] = res
			}

			tempKey += fmt.Sprint("_", len(children))
		}

		// Handle special cases requiring children

		if ast.Name == NodeDocument {
			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])

					if ast.Children[i].Name != NodeArguments {
						buf.WriteString("\n\n")
					}
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
			}

			return ppPostProcessing(ast, path, buf.String()), nil

		} else if ast.Name == NodeOperationType || ast.Name == NodeName ||
			ast.Name == NodeFragmentName || ast.Name == NodeType || ast.Name == NodeEnumValue {

			return ast.Token.Val, nil

		} else if ast.Name == NodeArguments {

			buf.WriteString("(")

			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])
					buf.WriteString(", ")
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
			}
			buf.WriteString(")")

			return ppPostProcessing(ast, path, buf.String()), nil

		} else if ast.Name == NodeListValue {
			buf.WriteString("[")
			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])
					buf.WriteString(", ")
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
			}
			buf.WriteString("]")

			return ppPostProcessing(ast, path, buf.String()), nil

		} else if ast.Name == NodeVariableDefinitions {
			buf.WriteString("(")
			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])
					buf.WriteString(", ")
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
			}
			buf.WriteString(")")

			return ppPostProcessing(ast, path, buf.String()), nil

		} else if ast.Name == NodeSelectionSet {
			buf.WriteString("{\n")
			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
			}
			buf.WriteString("}")

			return ppPostProcessing(ast, path, buf.String()), nil

		} else if ast.Name == NodeObjectValue {

			buf.WriteString("{")

			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])
					buf.WriteString(", ")
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
			}
			buf.WriteString("}")

			return ppPostProcessing(ast, path, buf.String()), nil

		} else if ast.Name == NodeObjectField {

			buf.WriteString(ast.Token.Val)
			buf.WriteString(" : ")
			buf.WriteString(children["c1"])

			return buf.String(), nil

		} else if ast.Name == NodeField {

			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])

					if ast.Children[i].Name != NodeArguments {
						buf.WriteString(" ")
					}
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
				buf.WriteString("\n")
			}

			return ppPostProcessing(ast, path, buf.String()), nil
		} else if ast.Name == NodeDirectives {

			if children != nil {
				i := 1
				for ; i < len(children); i++ {
					buf.WriteString(children[fmt.Sprint("c", i)])

					if ast.Children[i].Name != NodeArguments {
						buf.WriteString(" ")
					}
				}
				buf.WriteString(children[fmt.Sprint("c", i)])
			}

			return ppPostProcessing(ast, path, buf.String()), nil
		}

		// Retrieve the template

		temp, ok := prettyPrinterMap[tempKey]
		if !ok {
			return "", fmt.Errorf("Could not find template for %v (tempkey: %v)",
				ast.Name, tempKey)
		}

		// Use the children as parameters for template

		errorutil.AssertOk(temp.Execute(&buf, children))

		return ppPostProcessing(ast, path, buf.String()), nil
	}

	res, err := visit(ast, []*ASTNode{ast})

	return strings.TrimSpace(res), err
}

/*
ppPostProcessing applies post processing rules.
*/
func ppPostProcessing(ast *ASTNode, path []*ASTNode, ppString string) string {
	ret := ppString

	// Apply indentation

	if len(path) > 1 {
		if stringutil.IndexOf(ast.Name, []string{
			NodeField,
			NodeFragmentSpread,
			NodeInlineFragment,
		}) != -1 {

			parent := path[len(path)-3]

			indentSpaces := stringutil.GenerateRollingString(" ", IndentationLevel)
			ret = strings.ReplaceAll(ret, "\n", "\n"+indentSpaces)
			ret = fmt.Sprintf("%v%v", indentSpaces, ret)

			// Remove indentation from last line unless we have a special case

			if stringutil.IndexOf(parent.Name, []string{
				NodeField,
				NodeOperationDefinition,
			}) == -1 {

				if idx := strings.LastIndex(ret, "\n"); idx != -1 {
					ret = ret[:idx+1] + ret[idx+IndentationLevel+1:]
				}
			}

		}
	}

	// Remove all trailing spaces

	newlineSplit := strings.Split(ret, "\n")

	for i, s := range newlineSplit {
		newlineSplit[i] = strings.TrimRightFunc(s, unicode.IsSpace)
	}

	return strings.Join(newlineSplit, "\n")
}
