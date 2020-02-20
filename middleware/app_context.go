// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"html/template"

	"git.hoogi.eu/snafu/go-blog/models"
	"git.hoogi.eu/snafu/go-blog/settings"
	"git.hoogi.eu/snafu/session"
)

//AppContext contains the services, session store, templates, ...
type AppContext struct {
	SessionService    *session.SessionService
	ArticleService    models.ArticleService
	CategoryService   models.CategoryService
	UserService       models.UserService
	UserInviteService models.UserInviteService
	SiteService       models.SiteService
	FileService       models.FileService
	TokenService      models.TokenService
	Mailer            models.Mailer
	ConfigService     *settings.Settings
	Templates         *template.Template
}
