// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"git.hoogi.eu/go-blog/components/database"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/go-blog/utils"
	"git.hoogi.eu/session"

	_ "github.com/mattn/go-sqlite3"
)

var ctx *middleware.AppContext

func setup(t *testing.T) {
	logger.InitLogger(ioutil.Discard, "Debug")

	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		t.Fatal(err)
	}

	err = database.InitTables(db)

	if err != nil {
		t.Fatal(err)
	}

	err = fillSeeds(db)

	if err != nil {
		t.Fatal(err)
	}

	cfg, err := settings.LoadConfig("../go-blog.conf")

	if err != nil {
		t.Fatal(err)
	}

	userService := models.UserService{
		Datasource: models.SQLiteUserDatasource{
			SQLConn: db,
		},
		Config: cfg.User,
	}

	userInviteService := models.UserInviteService{
		Datasource: models.SQLiteUserInviteDatasource{
			SQLConn: db,
		},
		UserService: userService,
	}

	articleService := models.ArticleService{
		AppConfig: cfg.Application,
		Datasource: models.SQLiteArticleDatasource{
			SQLConn: db,
		},
	}

	siteService := models.SiteService{
		Datasource: models.SQLiteSiteDatasource{
			SQLConn: db,
		},
	}

	fileService := models.FileService{
		Datasource: models.SQLiteFileDatasource{
			SQLConn: db,
		},
	}

	categoryService := models.CategoryService{
		Datasource: models.SQLiteCategoryDatasource{
			SQLConn: db,
		},
	}

	tokenService := models.TokenService{
		Datasource: models.SQLiteTokenDatasource{
			SQLConn: db,
		},
	}

	sessionService := session.SessionService{
		Path:            "/admin",
		Name:            "test-session",
		HTTPOnly:        true,
		Secure:          true,
		SessionProvider: session.NewInMemoryProvider(),
		IdleSessionTTL:  10,
	}

	ctx = &middleware.AppContext{
		UserService:       userService,
		UserInviteService: userInviteService,
		ArticleService:    articleService,
		CategoryService:   categoryService,
		SiteService:       siteService,
		FileService:       fileService,
		TokenService:      tokenService,
		SessionService:    &sessionService,
		ConfigService:     cfg,
	}
}

func fillSeeds(db *sql.DB) error {
	salt := utils.GenerateSalt()
	saltedPassword := utils.AppendBytes([]byte("123456789012"), salt)
	password, err := utils.CryptPassword([]byte(saltedPassword), 12)

	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user (id, username, email, display_name, salt, password, active, is_admin, last_modified) VALUES (1, 'alice', 'alice@example.org', 'Alice Schneier', ?, ?, 1, 1, date('now'))", string(salt), password)

	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user (id, username, email, display_name, salt, password, active, is_admin, last_modified) VALUES (2, 'bob', 'bob@example.org', 'Bob Stallman', ?, ?, 1, 0, date('now'))", string(salt), string(password))

	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user (id, username, email, display_name, salt, password, active, is_admin, last_modified) VALUES (3, 'mallory', 'mallory@example.org', 'Mallory Pike', ?, ?, 0, 1, date('now'))", string(salt), string(password))

	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO user (id, username, email, display_name, salt, password, active, is_admin, last_modified) VALUES (4, 'eve', 'eve@example.org', 'Mallory Pike', ?, ?, 0, 0, date('now'))", string(salt), string(password))

	if err != nil {
		return err
	}

	return nil

}

func dummyAdminUser() *models.User {
	u, _ := ctx.UserService.GetByID(1)
	return u
}

func dummyUser() *models.User {
	u, _ := ctx.UserService.GetByID(2)
	return u
}

func setHeader(r *http.Request, key, value string) {
	r.Header.Set("X-Unit-Testing-Value-"+key, value)
}

func addValue(m url.Values, key, value string) {
	m.Add(key, value)
}

func addCheckboxValue(m url.Values, key string, value bool) {
	if value {
		m.Add(key, "on")
	}
	m.Add(key, "off")
}

func post(path string, values url.Values) (*http.Request, error) {
	var b bytes.Buffer
	b.WriteString(values.Encode())
	req, err := http.NewRequest("POST", path, &b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

func get(path string, values url.Values) (*http.Request, error) {
	var b bytes.Buffer
	b.WriteString(values.Encode())
	req, err := http.NewRequest("GET", path, &b)
	return req, err
}

//reqUser the user which should be added to the context
type reqUser int

const (
	rGuest = iota
	rAdminUser
	rUser
	rInactiveAdminUser
	rInactiveUser
)

type request struct {
	url     string
	user    reqUser
	method  string
	values  url.Values
	pathVar []pathVar
}

type pathVar struct {
	key   string
	value string
}

func (r request) buildRequest() *http.Request {
	var req *http.Request

	if r.method == http.MethodPost {
		req, _ = post(r.url, r.values)
	} else {
		req, _ = get(r.url, r.values)
	}

	if r.pathVar != nil {
		for _, v := range r.pathVar {
			setHeader(req, v.key, v.value)
		}
	}

	var user *models.User

	if r.user == rAdminUser {
		user, _ = ctx.UserService.GetByID(1)
	} else if r.user == rUser {
		user, _ = ctx.UserService.GetByID(2)
	} else if r.user == rInactiveAdminUser {
		user, _ = ctx.UserService.GetByID(3)
	} else if r.user == rInactiveUser {
		user, _ = ctx.UserService.GetByID(4)
	} else {
		return req
	}

	reqCtx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(reqCtx)

	return req
}

type responseWrapper struct {
	template *middleware.Template
	response *httptest.ResponseRecorder
}

func (r responseWrapper) getTemplateError() error {
	return r.template.Err
}

func (r responseWrapper) isCodeSuccess() bool {
	return r.response.Result().StatusCode == http.StatusOK
}

func (r responseWrapper) getStatus() int {
	return r.response.Result().StatusCode
}

func (r responseWrapper) getCookie(name string) (*http.Cookie, error) {
	for _, c := range r.response.Result().Cookies() {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("cookie %s not found", name)
}
