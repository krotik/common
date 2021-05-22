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
	"bufio"
	"bytes"
	"io"
	"math"
	"regexp"
	"strings"
	"unicode"
)

var cSyleCommentsRegexp = regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/")

/*
StripCStyleComments strips out C-Style comments from a given string.
*/
func StripCStyleComments(text []byte) []byte {
	return cSyleCommentsRegexp.ReplaceAll(text, nil)
}

/*
CreateDisplayString changes all "_" characters into spaces and properly capitalizes
the resulting string.
*/
func CreateDisplayString(str string) string {
	if len(str) == 0 {
		return ""
	}

	return ProperTitle(strings.Replace(str, "_", " ", -1))
}

// The following words should not be capitalized
//
var notCapitalize = map[string]string{
	"a":    "",
	"an":   "",
	"and":  "",
	"at":   "",
	"but":  "",
	"by":   "",
	"for":  "",
	"from": "",
	"in":   "",
	"nor":  "",
	"on":   "",
	"of":   "",
	"or":   "",
	"the":  "",
	"to":   "",
	"with": "",
}

/*
ProperTitle will properly capitalize a title string by capitalizing the first, last
and any important words. Not capitalized are articles: a, an, the; coordinating
conjunctions: and, but, or, for, nor; prepositions (fewer than five
letters): on, at, to, from, by.
*/
func ProperTitle(input string) string {
	words := strings.Fields(strings.ToLower(input))
	size := len(words)

	for index, word := range words {
		if _, ok := notCapitalize[word]; !ok || index == 0 || index == size-1 {
			words[index] = strings.Title(word)
		}
	}

	return strings.Join(words, " ")
}

/*
ToUnixNewlines converts all newlines in a given string to unix newlines.
*/
func ToUnixNewlines(s string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	return strings.Replace(s, "\r", "\n", -1)
}

/*
TrimBlankLines removes blank initial and trailing lines.
*/
func TrimBlankLines(s string) string {
	return strings.Trim(s, "\r\n")
}

/*
StripUniformIndentation removes uniform indentation from a string.
*/
func StripUniformIndentation(s string) string {
	leadingWhitespace := func(line string) int {
		var count int

		// Count leading whitespaces in a string

		for _, r := range line {
			if unicode.IsSpace(r) || unicode.IsControl(r) {
				count++
			} else {
				return count
			}
		}

		return -1 // Special case line is full of whitespace
	}

	// Count the minimum number of leading whitespace excluding
	// empty lines

	minCount := math.MaxInt16
	reader := strings.NewReader(s)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		if lw := leadingWhitespace(scanner.Text()); lw != -1 {
			if lw < minCount {
				minCount = lw
			}
		}
	}

	// Go through the string again and build up the output

	var buf bytes.Buffer

	reader.Seek(0, io.SeekStart)
	scanner = bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) != "" {
			for i, r := range line {
				if i >= minCount {
					buf.WriteRune(r)
				}
			}
		}

		buf.WriteString("\n")
	}

	// Prepare output string

	ret := buf.String()

	if !strings.HasSuffix(s, "\n") {
		ret = ret[:len(ret)-1]
	}

	return ret
}
