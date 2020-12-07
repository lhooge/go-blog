// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	bf "github.com/russross/blackfriday/v2"
)

// Defines the extensions that are used
var ext = bf.NoIntraEmphasis | bf.Tables | bf.FencedCode | bf.Autolink |
	bf.Strikethrough | bf.SpaceHeadings | bf.BackslashLineBreak |
	bf.DefinitionLists | bf.Footnotes | bf.HardLineBreak

// Defines the HTML rendering flags that are used
var flags = bf.UseXHTML | bf.Smartypants | bf.SmartypantsFractions |
	bf.SmartypantsDashes | bf.SmartypantsLatexDashes | bf.TOC

var p *bluemonday.Policy

func init() {
	p = bluemonday.UGCPolicy()
	p.AllowAttrs("style").OnElements("pre")
	p.AllowAttrs("style").OnElements("span")
}

//MarkdownToHTML sanitizes and parses markdown to HTML
func MarkdownToHTML(md []byte) []byte {
	md = bytes.Replace(md, []byte("\r\n"), []byte("\n"), -1)
	unsafe := bf.Run((md), bf.WithExtensions(ext))

	return sanitize(unsafe)
}

func sanitize(in []byte) []byte {
	return p.SanitizeBytes(in)
}

func EscapeHTML(in string) string {
	return template.HTMLEscapeString(in)
}

func NewlineToBr(in string) string {
	out := strings.Replace(in, "\r\n", "<br>", -1)
	out = strings.Replace(out, "\n", "<br>", -1)
	out = strings.Replace(out, "\r", "<br>", -1)
	return out
}
