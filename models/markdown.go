// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
	bf "gopkg.in/russross/blackfriday.v2"
)

// Defines the extensions that are used
var exts = bf.NoIntraEmphasis | bf.Tables | bf.FencedCode | bf.Autolink |
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
func MarkdownToHTML(md string) string {
	md = strings.Replace(md, "\r\n", "\n", -1)
	return sanitize(string(bf.Run([]byte(md), bf.WithExtensions(exts))))
}

func sanitize(in string) string {
	return p.Sanitize(in)
}
