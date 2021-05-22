/*
 * Public Domain Software
 *
 * I (Matthias Ladkau) am the author of the source code in this file.
 * I have placed the source code in this file in the public domain.
 *
 * For further information see: http://creativecommons.org/publicdomain/zero/1.0/
 */

package parser

/*
RuntimeProvider provides runtime components for a parse tree.
*/
type RuntimeProvider interface {

	/*
	   Runtime returns a runtime component for a given ASTNode.
	*/
	Runtime(node *ASTNode) Runtime
}

/*
Runtime provides the runtime for an ASTNode.
*/
type Runtime interface {

	/*
	   Validate this runtime component and all its child components.
	*/
	Validate() error

	/*
		Eval evaluate this runtime component.
	*/
	Eval() (map[string]interface{}, error)
}
