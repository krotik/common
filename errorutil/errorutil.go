/*
 * Public Domain Software
 *
 * I (Matthias Ladkau) am the author of the source code in this file.
 * I have placed the source code in this file in the public domain.
 *
 * For further information see: http://creativecommons.org/publicdomain/zero/1.0/
 */

/*
Package errorutil contains common error objects and functions.
*/
package errorutil

import (
	"bytes"
)

/*
AssertOk will panic on any non-nil error parameter.
*/
func AssertOk(err error) {
	if err != nil {
		panic(err.Error())
	}
}

/*
AssertTrue will panic if the given condition is negative.
*/
func AssertTrue(condition bool, errString string) {
	if !condition {
		panic(errString)
	}
}

/*
CompositeError can collect multiple errors in a single error object.
*/
type CompositeError struct {
	Errors []error
}

/*
NewCompositeError creates a new composite error object.
*/
func NewCompositeError() *CompositeError {
	return &CompositeError{make([]error, 0)}
}

/*
Add adds an error.
*/
func (ce *CompositeError) Add(e error) {
	ce.Errors = append(ce.Errors, e)
}

/*
HasErrors returns true if any error have been collected.
*/
func (ce *CompositeError) HasErrors() bool {
	return len(ce.Errors) > 0
}

/*
Error returns all collected errors as a string.
*/
func (ce *CompositeError) Error() string {
	var buf bytes.Buffer
	for i, e := range ce.Errors {
		buf.WriteString(e.Error())
		if i < len(ce.Errors)-1 {
			buf.WriteString("; ")
		}
	}
	return buf.String()
}
