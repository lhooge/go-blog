// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.hoogi.eu/snafu/cfg"
	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/models"
	"git.hoogi.eu/snafu/go-blog/settings"
)

// Template contains the information about the template to render.
// Active contains the current active navigation.
// Data the data which is injected into the templates.
// SuccessMsg is an optional variable which is displayed as a green message.
// WarnMsg is an optional variable which is displayed as an orange message.
// RedirectPath contains the path where the request should be redirected.
// Err will be shown as red message in templates. If it's a httperror, the display message will be shown,
// otherwise generich 'An internal error occurred' is shown.
type Template struct {
	Name         string
	Active       string
	Data         map[string]interface{}
	SuccessMsg   string
	WarnMsg      string
	RedirectPath string
	Err          error
}

// Templates defines the directory where the templates are located, the FuncMap are additional functions, which can
// be used in the templates.
type Templates struct {
	Directory string
	FuncMap   template.FuncMap
}

// NotFound returned if no route matches
func NotFound(ctx *AppContext, rw http.ResponseWriter, r *http.Request) *Template {
	//For deleting flash cookies
	getFlash(rw, r, "ErrorMsg")
	getFlash(rw, r, "SuccessMsg")

	session, _ := ctx.SessionService.Get(rw, r)

	if session != nil && strings.HasPrefix(r.URL.EscapedPath(), "/admin") {
		return &Template{
			Name: "admin/error",
			Err:  httperror.New(http.StatusNotFound, "Nothing was found at this location", fmt.Errorf("page %s not found", r.URL.EscapedPath())),
		}
	} else {
		return &Template{
			Name: "front/error",
			Err:  httperror.New(http.StatusNotFound, "Nothing was found at this location", fmt.Errorf("page %s not found", r.URL.EscapedPath())),
		}
	}
}

// FuncMap some function that can be used in templates
func FuncMap(ss models.SiteService, settings *settings.Settings) template.FuncMap {
	return template.FuncMap{
		"GetMetadata": func(data map[string]interface{}) template.HTML {
			var meta, desc string

			if len(settings.Description) > 0 {
				desc = settings.Description

				if len(desc) > 200 {
					desc = desc[0:200] + "..."
				}

				meta = fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", html.EscapeString(desc))
			}

			if value, ok := data["article"]; ok {
				if art, ok := value.(*models.Article); ok {
					desc = art.Teaser

					if len(desc) > 200 {
						desc = desc[0:200] + "..."
					}

					meta = fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", html.EscapeString(desc))
					meta += fmt.Sprintf("\t\t<meta name=\"author\" content=\"%s\">\n", html.EscapeString(art.Author.DisplayName))
				}
			}

			if value, ok := data["site"]; ok {
				if site, ok := value.(*models.Site); ok {
					desc = site.Content

					if len(desc) > 200 {
						desc = desc[0:200] + "..."
					}

					meta = fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", html.EscapeString(desc))
					meta += fmt.Sprintf("\t\t<meta name=\"author\" content=\"%s\">\n", html.EscapeString(site.Author.DisplayName))
				}
			}
			return template.HTML(meta)
		},
		"GetTitle": func(data map[string]interface{}) string {
			if value, ok := data["article"]; ok {
				if art, ok := value.(*models.Article); ok {
					return art.Headline
				}
			}

			if value, ok := data["site"]; ok {
				if site, ok := value.(*models.Site); ok {
					return site.Title
				}
			}
			return settings.Title
		},
		"Language": func() string {
			return settings.Language
		},
		"ApplicationURL": func() string {
			return settings.Application.Domain
		},
		"CustomCSS": func() string {
			return settings.Application.CustomCSS
		},
		"OverwriteCSS": func() bool {
			return settings.Application.OverwriteCSS
		},
		"KeepAliveInterval": func() int64 {
			return (settings.Session.TTL.Nanoseconds() / 1e9) - 5
		},
		"PageTitle": func() string {
			return settings.Title
		},
		"BuildVersion": func() string {
			return settings.BuildVersion
		},
		"BuildGitHash": func() string {
			return settings.BuildGitHash
		},
		"NilString": func(s sql.NullString) string {
			if !s.Valid {
				return ""
			}
			return s.String
		},
		"FormatNilDateTime": func(t models.NullTime) string {
			if !t.Valid {
				return ""
			}
			return t.Time.In(time.Local).Format("January 2, 2006 at 3:04 PM")
		},
		"HumanizeFilesize": func(size int64) string {
			return cfg.FileSize(size).HumanReadable()
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
			return template.HTML(models.MarkdownToHTML([]byte(s)))
		},
		"NToBr": func(in string) template.HTML {
			return template.HTML(models.NewlineToBr(models.EscapeHTML(in)))
		},
		"EscapeHTML": func(in string) string {
			return html.EscapeString(in)
		},
		"GetSites": func() []models.Site {
			sites, err := ss.List(models.OnlyPublished, nil)

			if err != nil {
				logger.Log.Error(err)
				return nil
			}

			return sites
		},
	}
}

// Load walks threw directory and parses templates ending with html
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

// RedirectURL builds a URL for redirecting
func (t Template) RedirectURL() string {
	if t.RedirectPath[0] == byte('/') {
		return t.RedirectPath
	}
	return "/" + t.RedirectPath
}
