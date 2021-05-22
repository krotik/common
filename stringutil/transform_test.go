/*
 * Public Domain Software
 *
 * I (Matthias Ladkau) am the author of the source code in this file.
 * I have placed the source code in this file in the public domain.
 *
 * For further information see: http://creativecommons.org/publicdomain/zero/1.0/
 */

package stringutil

import (
	"testing"
)

func TestStripCStyleComments(t *testing.T) {

	test := `
// Comment1
This is a test
/* A
comment
// Comment2
  */ bla
`

	if out := string(StripCStyleComments([]byte(test))); out != `
This is a test
 bla
` {
		t.Error("Unexpected return:", out)
		return
	}
}

func TestCreateDisplayString(t *testing.T) {
	testdata := []string{"this is a tEST", "_bla", "a_bla", "a__bla", "a__b_la", "",
		"a fool a to be to"}
	expected := []string{"This Is a Test", "Bla", "A Bla", "A Bla", "A B La", "",
		"A Fool a to Be To"}

	for i, str := range testdata {
		res := CreateDisplayString(str)
		if res != expected[i] {
			t.Error("Unexpected result for creating a display string from:", str,
				"result:", res, "expected:", expected[i])
		}
	}
}

func TestStripUniformIndentation(t *testing.T) {

	testdata := []string{`

    aaa
  aaa
      aaa

`, `
  bbb
    
    xx xx
  bbb
  bbb`, `
  ccc
ccc
    ccc
 `}

	expected := []string{`

  aaa
aaa
    aaa

`, `
bbb

  xx xx
bbb
bbb`, `
  ccc
ccc
    ccc
`}

	for i, str := range testdata {
		res := StripUniformIndentation(str)
		if res != expected[i] {
			t.Error("Unexpected result:", str,
				"result: '"+res+"' expected:", expected[i])
			return
		}
	}
}

func TestNewLineTransform(t *testing.T) {
	res := TrimBlankLines(ToUnixNewlines("\r\n  test123\r\ntest123\r\n"))
	if res != "  test123\ntest123" {
		t.Error("Unexpected result:", res)
		return
	}
}
