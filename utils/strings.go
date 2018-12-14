// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package utils

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

//AppendString uses byte buffer to append multiple strings
func AppendString(s ...string) string {
	var buffer bytes.Buffer
	for _, value := range s {
		buffer.WriteString(value)
	}
	return buffer.String()
}

//AppendBytes uses byte buffer to append multiple byte arrays
func AppendBytes(s ...[]byte) []byte {
	var buffer bytes.Buffer
	for _, value := range s {
		buffer.Write(value)
	}
	return buffer.Bytes()
}

//TrimmedStringIsEmpty trims spaces returns true if length is 0
func TrimmedStringIsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

//IsOneOfStringsEmpty checks if one of the given strings is empty
func IsOneOfStringsEmpty(s ...string) bool {
	for _, value := range s {
		if len(value) == 0 {
			return true
		}
	}
	return true
}

var filenameSubs = map[rune]string{
	'/':  "",
	'\\': "",
	':':  "",
	'*':  "",
	'?':  "",
	'"':  "",
	'<':  "",
	'>':  "",
	'|':  "",
	' ':  "",
}

func isDot(r rune) bool {
	return '.' == r
}

//SanitizeFilename sanitizes a filename for safe use when serving file
func SanitizeFilename(s string) string {
	s = strings.TrimFunc(s, unicode.IsSpace)
	s = removeControlCharacters(s)
	s = substitute(s, filenameSubs)
	s = strings.TrimFunc(s, isDot)
	return s
}

var slugSubs = map[rune]string{
	'&':  "",
	'$':  "",
	'+':  "",
	',':  "",
	'/':  "",
	':':  "",
	';':  "",
	'=':  "",
	'?':  "",
	'@':  "",
	'#':  "",
	'!':  "",
	'\'': "",
	'(':  "",
	')':  "",
	'*':  "",
	'%':  "",
}

var multipleDashes = regexp.MustCompile(`[-]{2,}`)

//CreateURLSafeSlug creates a url safe slug to use in urls
func CreateURLSafeSlug(input string, suffix int) string {
	input = removeControlCharacters(input)
	input = substitute(input, slugSubs)
	input = strings.TrimSpace(input)

	input = strings.Replace(input, " ", "-", -1)

	input = strings.ToLower(input)

	input = multipleDashes.ReplaceAllString(input, "-")

	if suffix > 0 {
		input += strconv.Itoa(suffix)
	}

	return input
}

func substitute(input string, subs map[rune]string) string {
	var b bytes.Buffer

	for _, c := range input {
		if _, ok := subs[c]; ok {
			b.WriteString(subs[c])
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func removeControlCharacters(input string) string {
	var b bytes.Buffer
	for _, c := range input {
		if c > 31 {
			b.WriteRune(c)
		}
	}
	return b.String()
}
