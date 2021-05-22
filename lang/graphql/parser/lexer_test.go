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
	"testing"
)

func TestNextAndPeek(t *testing.T) {
	l := &lexer{"", "Test", 0, 0, 0, 0, 0, make(chan LexToken)}

	if res := fmt.Sprintf("%c", l.next(0)); res != "T" {
		t.Error("Unexpected result:", res)
	}

	if res := fmt.Sprintf("%c", l.next(1)); res != "e" {
		t.Error("Unexpected result:", res)
	}

	if res := fmt.Sprintf("%c", l.next(2)); res != "s" {
		t.Error("Unexpected result:", res)
	}

	if res := fmt.Sprintf("%c", l.next(3)); res != "t" {
		t.Error("Unexpected result:", res)
	}

	if l.pos != 0 {
		t.Error("Lexer moved forward when it shouldn't: ", l.pos)
		return
	}

	if res := fmt.Sprintf("%c", l.next(-1)); res != "T" {
		t.Error("Unexpected result:", res)
	}

	if l.pos != 1 {
		t.Error("Lexer moved forward when it shouldn't: ", l.pos)
		return
	}

	if res := fmt.Sprintf("%c", l.next(-1)); res != "e" {
		t.Error("Unexpected result:", res)
	}

	if l.pos != 2 {
		t.Error("Lexer moved forward when it shouldn't: ", l.pos)
		return
	}
}

func TestSimpleLexing(t *testing.T) {

	if res := fmt.Sprint(LexToList("test", "\ufeff1!23")); res != `[int(1) ! int(23) EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(LexToList("test", "1!23.4e+11 3E-5 11.1 .4$")); res !=
		`[int(1) ! flt(23.4e+11) flt(3e-5) flt(11.1) flt(.4) $ EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(LexToList("test", "12!foo...bar99")); res !=
		`[int(12) ! <foo> ... <bar99> EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(LexToList("test", "-0 0 1230 0123")); res !=
		`[int(-0) int(0) int(1230) Error: 0123 (Line 1, Pos 11) EOF]` {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestLexingErrors(t *testing.T) {

	if res := fmt.Sprint(LexToList("test", `"te`)); res != `[Error: EOF inside quotes (Line 1, Pos 1) EOF]` {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestMultilineLexing(t *testing.T) {

	if res := fmt.Sprint(LexToList("test", `1!23#...4e+11
123
true
`)); res != `[int(1) ! int(23) int(123) <true> EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(LexToList("test", `"""
123
"""
"["
[
"123"
"123\u2318"
"""123\u2318"""
"""
  bla
"""
`)); res != `["123" "[" [ "123" "123⌘" "123\u2318" "bla" EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(LexToList("test", `"""
    Hello,
      World!

    Yours,
      GraphQL.
  """
`)); res != `["Hello,
  World!

Yours,
  GraphQL." EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	if res := fmt.Sprint(LexToList("test", `"Hello,\n  World!\n\nYours,\n  GraphQL."
`)); res != `["Hello,
  World!

Yours,
  GraphQL." EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	ll := LexToList("test", `"Hello,\n  World!\n\nYours,\n  GraphQL."
`)
	if res := ll[len(ll)-1].PosString(); res != "Line 2, Pos 1" {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestIgnoredLexing(t *testing.T) {

	res := fmt.Sprint(LexToList("test", "1,2,3...abc\t\r\n#123\n"))

	if res != `[int(1) int(2) int(3) ... <abc> EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	res = fmt.Sprint(LexToList("test", "1,2,3 .. x abc\r\n#123\n"))

	if res != `[int(1) int(2) int(3) Error: .. (Line 1, Pos 7) <x> <abc> EOF]` {
		t.Error("Unexpected result:", res)
		return
	}

	res = fmt.Sprint(LexToList("test", "1,2,3 .. x abc\r\n#123"))

	if res != `[int(1) int(2) int(3) Error: .. (Line 1, Pos 7) <x> <abc> EOF]` {
		t.Error("Unexpected result:", res)
		return
	}
}

func TestSampleQueries(t *testing.T) {

	sampleQueries := [][]string{{`
query StudentsNormal {
  allStudents(pagination: {offset: 0, limit: 10}, sort: {fields: [{field: "studentNumber", order: ASC}]}, 
                           filter: {fields: [{op: NIN, value: "[Harry]", field: "name"}]}) {
    result {
      ...studentFields
      subjects {
        name
        classroom
      }
    }
    pagination {
      offset
      limit
      total
    }
  }
}
`, `[<query> <StudentsNormal> { <allStudents> ( <pagination> : { <offset> : int(0) <limit> : int(10) } <sort> : { <fields> : [ { <field> : "studentNumber" <order> : <ASC> } ] } <filter> : { <fields> : [ { <op> : <NIN> <value> : "[Harry]" <field> : "name" } ] } ) { <result> { ... <studentFields> <subjects> { <name> <classroom> } } <pagination> { <offset> <limit> <total> } } } EOF]`},

		{`
query StudentsJPA {
  allStudentsJPA(pagination: {offset: 0, limit: 10}, sort: {fields: [{field: "studentNumber", order: ASC}]}, filter: {fields: [{op: NIN, value: "[Harry]", field: "name"}]}) {
    ... on PaginationWrapper_Student {
      result {
        name
      }
    }
    result {
      ...studentFields
      ... on Student {
        enrolled
      }
      subjects {
        name
        classroom
      }
    }
    pagination {
      offset
      limit
      total
    }
  }
}
`, `[<query> <StudentsJPA> { <allStudentsJPA> ( <pagination> : { <offset> : int(0) <limit> : int(10) } <sort> : { <fields> : [ { <field> : "studentNumber" <order> : <ASC> } ] } <filter> : { <fields> : [ { <op> : <NIN> <value> : "[Harry]" <field> : "name" } ] } ) { ... <on> <PaginationWrapper_Student> { <result> { <name> } } <result> { ... <studentFields> ... <on> <Student> { <enrolled> } <subjects> { <name> <classroom> } } <pagination> { <offset> <limit> <total> } } } EOF]`},

		{`

# query variables
{
  "st": {
    "studentNumber": 63170004,
    "studentLoan": 631700.04,
    "name": "Latest",
    "surname": "Greatest"
  }
`, `[{ "st" : { "studentNumber" : int(63170004) "studentLoan" : flt(631700.04) "name" : "Latest" "surname" : "Greatest" } EOF]`}}

	for _, sampleQuery := range sampleQueries {

		if res := fmt.Sprint(LexToList("test", sampleQuery[0])); res != sampleQuery[1] {
			t.Error("Unexpected result\nGiven:\n", sampleQuery[0], "\nGot:\n", res, "\nExpected:\n", sampleQuery[1])
			return
		}
	}
}
