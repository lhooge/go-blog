// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"html/template"

	"git.hoogi.eu/go-blog/components/mail"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/models/sessions"
	"git.hoogi.eu/go-blog/settings"
)

//AppContext contains the services, session store, templates
type AppContext struct {
	SessionStore   *sessions.CookieStore
	ArticleService models.ArticleService
	UserService    models.UserService
	SiteService    models.SiteService
	FileService    models.FileService
	TokenService   models.TokenService
	MailService    mail.Service
	ConfigService  *settings.Settings
	Templates      *template.Template
}
