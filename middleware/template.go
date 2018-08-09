// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/go-blog/utils"
)

// Template contains the information about the template to render.
// The DisplayMsg in models.Error will be the ErrorMsg in the flash bubble,
// the SuccessMsg is an optional variable which is also displayed as a green flash bubble,
// both are appended to the data map with keys 'ErrorMsg' or 'SuccessMsg' in the AppHandler
type Template struct {
	Name         string
	Active       string
	Data         map[string]interface{}
	SuccessMsg   string
	RedirectPath string
	Err          error
}

//Templates defines in which directory should be looked for template
type Templates struct {
	Directory string
	FuncMap   template.FuncMap
}

//NotFound returned if no route matches
func NotFound(ctx *AppContext, rw http.ResponseWriter, r *http.Request) *Template {
	//For deleting flash cookies
	getFlash(rw, r, "ErrorMsg")
	getFlash(rw, r, "SuccessMsg")

	session, _ := ctx.SessionService.Get(rw, r)
	if session != nil && strings.HasPrefix(r.URL.EscapedPath(), "/admin") {
		return &Template{
			Name: "admin/error",
			Err:  httperror.New(http.StatusNotFound, "Nothing was found at this location", errors.New("page not found")),
		}
	} else {
		return &Template{
			Name: "front/error",
			Err:  httperror.New(http.StatusNotFound, "Nothing was found at this location", errors.New("page not found")),
		}
	}
}

//FuncMap some functions for use in templates
func FuncMap(ss models.SiteService, cfg *settings.Settings) template.FuncMap {
	return template.FuncMap{
		"GetMetadata": func(data map[string]interface{}) template.HTML {
			var meta string

			if len(cfg.Description) > 0 {
				meta = fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", cfg.Description)
			}

			if value, ok := data["article"]; ok {
				if art, ok := value.(*models.Article); ok {
					meta = fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", art.Teaser)
					meta += fmt.Sprintf("\t\t<meta name=\"author\" content=\"%s\">\n", art.Author.DisplayName)
				}
			} else if value, ok := data["site"]; ok {
				if site, ok := value.(*models.Site); ok {
					meta = fmt.Sprintf("\t\t<meta name=\"author\" content=\"%s\">\n", site.Author.DisplayName)
				}
			}
			return template.HTML(meta)
		},
		"Language": func() string {
			return cfg.Language
		},
		"KeepAliveInterval": func() int64 {
			return (cfg.Session.TTL.Nanoseconds() / 1e9) - 5
		},
		"PageTitle": func() string {
			return cfg.Title
		},
		"BuildVersion": func() string {
			return cfg.BuildVersion
		},
		"BuildDate": func() string {
			return cfg.BuildDate
		},
		"FormatNilDateTime": func(t models.NullTime) string {
			if !t.Valid {
				return ""
			}
			return t.Time.In(time.Local).Format("January 2, 2006 at 3:04 PM")
		},
		"FormatDateTime": func(t time.Time) string {
			return t.In(time.Local).Format("January 2, 2006 at 3:04 PM")
		},
		"FormatNilDate": func(t models.NullTime) string {
			if !t.Valid {
				return ""
			}
			return t.Time.In(time.Local).Format("January 2, 2006")
		},
		"FormatDate": func(t time.Time) string {
			return t.In(time.Local).Format("January 2, 2006")
		},
		"BoolToIcon": func(b bool) template.HTML {
			if b {
				return template.HTML(`<img alt="circle-checked" src="../assets/svg/circle-check.svg">`)
			}
			return template.HTML("")
		},
		"PaginationBar": func(p models.Pagination) template.HTML {
			return p.PaginationBar()
		},
		"ParseMarkdown": func(s string) template.HTML {
			return template.HTML(models.MarkdownToHTML(s))
		},
		"NToBr": func(input string) template.HTML {
			out := template.HTMLEscapeString(input)
			out = strings.Replace(out, "\r\n", "<br>", -1)
			out = strings.Replace(out, "\n", "<br>", -1)
			out = strings.Replace(out, "\r", "<br>", -1)
			return template.HTML(out)
		},
		"GetSites": func() []models.Site {
			sites, err := ss.ListSites(models.OnlyPublished, nil)

			if err != nil {
				logger.Log.Error(err)
				return nil
			}

			return sites
		},
	}
}

//Load walks threw directory and parses templates ending with html
func (ts Templates) Load() (*template.Template, error) {
	tpl := template.New("").Funcs(ts.FuncMap)

	err := filepath.Walk(ts.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(path, ".html") {
			tpls, err := tpl.ParseFiles(path)

			if err != nil {
				return err
			}
			template.Must(tpls, err)

		}

		return nil
	})

	return tpl, err
}

//RedirectURL builds a URL for redirecting
func (t Template) RedirectURL() string {
	if t.RedirectPath[0] == byte('/') {
		return t.RedirectPath
	}
	return utils.AppendString("/", t.RedirectPath)
}
