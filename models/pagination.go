// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"html/template"
	"math"
	"strings"
)

// Pagination type is used to provide a page selector
type Pagination struct {
	Total       int
	Limit       int
	CurrentPage int
	RelURL      string
}

// Offset returns the offset where to start
func (p *Pagination) Offset() int {
	return (p.CurrentPage - 1) * p.Limit
}

// url returns the absolute url
func (p *Pagination) url() string {
	if p.RelURL[0] == '/' {
		return p.RelURL
	}
	return "/" + p.RelURL
}

// pages returns the amount of pages
func (p *Pagination) pages() int {
	return int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}

// hasNext returns true if a next page is available
func (p *Pagination) hasNext() bool {
	if p.CurrentPage*p.Limit >= p.Total {
		return false
	}
	return true
}

// hasMoreThanOnePage returns true if the bar has more than one page
func (p *Pagination) hasMoreThanOnePage() bool {
	return p.Limit < p.Total
}

// hasPrevious returns true if a previous page is available
func (p *Pagination) hasPrevious() bool {
	return !(p.CurrentPage == 1)
}

// nextPage returns the next page
func (p *Pagination) nextPage() int {
	if !p.hasNext() {
		return p.CurrentPage
	}
	return p.CurrentPage + 1
}

// previousPage returns the previous page
func (p *Pagination) previousPage() int {
	if !p.hasPrevious() {
		return p.CurrentPage
	}
	return p.CurrentPage - 1
}

// PaginationBar returns the HTML for the pagination bar which can be embedded
func (p *Pagination) PaginationBar() template.HTML {
	var sb strings.Builder

	if p.pages() > 1 {
		sb.WriteString(`<div id="pagination">`)

		if !p.hasPrevious() {
			sb.WriteString(`<a class="button button-inactive" href="#">&laquo; Backward</a>`)
		} else {
			sb.WriteString(fmt.Sprintf(`<a class="button button-active" href="%s/%d">&laquo; Backward</a>`, p.url(), p.previousPage()))
		}

		for i := 1; i <= p.pages(); i++ {
			if p.CurrentPage == i {
				sb.WriteString(fmt.Sprintf(`<a class="button button-inactive" href="#">%d</a>`, i))
			} else {
				sb.WriteString(fmt.Sprintf(`<a class="button button-active" href="%s/%d">%d</a>`, p.url(), i, i))
			}
		}

		if !p.hasNext() {
			sb.WriteString(`<a class="button button-inactive" href="#">Forward &raquo;</a>`)
		} else {
			sb.WriteString(fmt.Sprintf(`<a class="button button-active" href="%s/%d">Forward &raquo;</a>`, p.url(), p.nextPage()))
		}

		sb.WriteString(`</div>`)
	}
	return template.HTML(sb.String())
}
