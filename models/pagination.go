// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"fmt"
	"html/template"
	"math"

	"git.hoogi.eu/snafu/go-blog/utils"
)

//Pagination type is used to provide a page selector
type Pagination struct {
	Total       int
	Limit       int
	CurrentPage int
	RelURL      string
}

//Offset returns the offset where to start
func (p Pagination) Offset() int {
	return (p.CurrentPage - 1) * p.Limit
}

//url returns the absolute url
func (p Pagination) url() string {
	if p.RelURL[0] == '/' {
		return utils.AppendString(p.RelURL)
	}
	return utils.AppendString("/", p.RelURL)
}

//pages returns the amount of pages
func (p Pagination) pages() int {
	return int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}

//hasNext returns true if a next page is available
func (p Pagination) hasNext() bool {
	if p.CurrentPage*p.Limit >= p.Total {
		return false
	}
	return true
}

//hasMoreThanOnePage returns true if the bar has more than one page
func (p Pagination) hasMoreThanOnePage() bool {
	return p.Limit < p.Total
}

//hasPrevious returns true if a previous page is available
func (p Pagination) hasPrevious() bool {
	return !(p.CurrentPage == 1)
}

//nextPage returns the next page
func (p Pagination) nextPage() int {
	if !p.hasNext() {
		return p.CurrentPage
	}
	return p.CurrentPage + 1
}

//previousPage returns the previous page
func (p Pagination) previousPage() int {
	if !p.hasPrevious() {
		return p.CurrentPage
	}
	return p.CurrentPage - 1
}

//PaginationBar returns the HTML for the pagination bar which can be embedded
func (p Pagination) PaginationBar() template.HTML {
	var buffer bytes.Buffer

	if p.pages() > 1 {
		buffer.WriteString(`<div id="pagination">`)

		if !p.hasPrevious() {
			buffer.WriteString(`<a class="button button-inactive" href="#">&laquo; Backward</a>`)
		} else {
			buffer.WriteString(fmt.Sprintf(`<a class="button button-active" href="%s/%d">&laquo; Backward</a>`, p.url(), p.previousPage()))
		}

		for i := 1; i <= p.pages(); i++ {
			if p.CurrentPage == i {
				buffer.WriteString(fmt.Sprintf(`<a class="button button-inactive" href="#">%d</a></li>`, i))
			} else {
				buffer.WriteString(fmt.Sprintf(`<a class="button button-active" href="%s/%d">%d</a></li>`, p.url(), i, i))
			}
		}

		if !p.hasNext() {
			buffer.WriteString(`<a class="button button-inactive" href="#">Forward &raquo;</a>`)
		} else {
			buffer.WriteString(fmt.Sprintf(`<a class="button button-active" href="%s/%d">Forward &raquo;</a>`, p.url(), p.nextPage()))
		}

		buffer.WriteString(`</div>`)
	}
	return template.HTML(buffer.String())
}
