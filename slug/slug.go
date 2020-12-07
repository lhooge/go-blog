package slug

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var multipleDashes = regexp.MustCompile(`[-]{2,}`)

//CreateURLSafeSlug creates a URL safe slug to use in URL
func CreateURLSafeSlug(input string, suffix int) string {
	input = strings.Replace(input, "-", "", -1)

	input = strings.Map(func(r rune) rune {
		switch {
		case r == ' ':
			return '-'
		case unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		default:
			return -1
		}
	}, strings.ToLower(strings.TrimSpace(input)))

	input = strings.Trim(input, "-")

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
