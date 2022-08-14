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

	"github.com/krotik/common/stringutil"
)

/*
ASTNode models a node in the AST
*/
type ASTNode struct {
	Name     string     // Name of the node
	Token    *LexToken  // Lexer token of this ASTNode
	Children []*ASTNode // Child nodes
	Runtime  Runtime    // Runtime component for this ASTNode

	binding        int                                                             // Binding power of this node
	nullDenotation func(p *parser, self *ASTNode) (*ASTNode, error)                // Configure token as beginning node
	leftDenotation func(p *parser, self *ASTNode, left *ASTNode) (*ASTNode, error) // Configure token as left node
}

/*
ASTFromPlain creates an AST from a plain AST.
A plain AST is a nested map structure like this:

	{
		name     : <name of node>
		value    : <value of node>
		children : [ <child nodes> ]
	}
*/
func ASTFromPlain(plainAST map[string]interface{}) (*ASTNode, error) {
	var astChildren []*ASTNode

	nameValue, ok := plainAST["name"]
	if !ok {
		return nil, fmt.Errorf("Found plain ast node without a name: %v", plainAST)
	}
	name := fmt.Sprint(nameValue)

	valueValue, ok := plainAST["value"]
	if stringutil.IndexOf(name, ValueNodes) != -1 && !ok {
		return nil, fmt.Errorf("Found plain ast value node without a value: %v", name)
	}
	value := fmt.Sprint(valueValue)

	// Create children

	if children, ok := plainAST["children"]; ok {

		if ic, ok := children.([]interface{}); ok {

			// Do a list conversion if necessary - this is necessary when we parse
			// JSON with map[string]interface{} this

			childrenList := make([]map[string]interface{}, len(ic))
			for i := range ic {
				childrenList[i] = ic[i].(map[string]interface{})
			}

			children = childrenList
		}

		for _, child := range children.([]map[string]interface{}) {

			astChild, err := ASTFromPlain(child)
			if err != nil {
				return nil, err
			}

			astChildren = append(astChildren, astChild)
		}
	}

	return &ASTNode{fmt.Sprint(name), &LexToken{TokenGeneral, 0,
		fmt.Sprint(value), 0, 0}, astChildren, nil, 0, nil, nil}, nil
}

/*
newAstNode creates an instance of this ASTNode which is connected to a concrete lexer token.
*/
func newAstNode(name string, p *parser, t *LexToken) *ASTNode {
	ret := &ASTNode{name, t, make([]*ASTNode, 0, 2), nil, 0, nil, nil}
	if p.rp != nil {
		ret.Runtime = p.rp.Runtime(ret)
	}
	return ret
}

/*
changeAstNode changes the name of a given ASTNode.
*/
func changeAstNode(node *ASTNode, newname string, p *parser) *ASTNode {
	node.Name = newname
	node.Runtime = nil
	if p.rp != nil {
		node.Runtime = p.rp.Runtime(node)
	}
	return node
}

/*
instane creates a new instance of this ASTNode which is connected to a concrete lexer token.
*/
func (n *ASTNode) instance(p *parser, t *LexToken) *ASTNode {
	ret := &ASTNode{n.Name, t, make([]*ASTNode, 0, 2), nil, n.binding, n.nullDenotation, n.leftDenotation}
	if p.rp != nil {
		ret.Runtime = p.rp.Runtime(ret)
	}
	return ret
}

/*
Plain returns this ASTNode and all its children as plain AST. A plain AST
only contains map objects, lists and primitive types which can be serialized
with JSON.
*/
func (n *ASTNode) Plain() map[string]interface{} {
	ret := make(map[string]interface{})

	ret["name"] = n.Name

	lenChildren := len(n.Children)

	if lenChildren > 0 {
		children := make([]map[string]interface{}, lenChildren)
		for i, child := range n.Children {
			children[i] = child.Plain()
		}

		ret["children"] = children
	}

	if stringutil.IndexOf(n.Name, ValueNodes) != -1 {
		ret["value"] = n.Token.Val
	}

	return ret
}

/*
String returns a string representation of this token.
*/
func (n *ASTNode) String() string {
	var buf bytes.Buffer
	n.levelString(0, &buf)
	return buf.String()
}

/*
levelString function to recursively print the tree.
*/
func (n *ASTNode) levelString(indent int, buf *bytes.Buffer) {

	// Print current level

	buf.WriteString(stringutil.GenerateRollingString(" ", indent*2))

	if stringutil.IndexOf(n.Name, ValueNodes) != -1 {
		buf.WriteString(fmt.Sprintf(n.Name+": %v", n.Token.Val))
	} else {
		buf.WriteString(n.Name)
	}

	buf.WriteString("\n")

	// Print children

	for _, child := range n.Children {
		child.levelString(indent+1, buf)
	}
}
