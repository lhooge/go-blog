package models

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

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
