// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"git.hoogi.eu/go-blog/components/database"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/components/mail"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/go-blog/utils"
	"git.hoogi.eu/session"

	_ "github.com/mattn/go-sqlite3"
)

var ctx *middleware.AppContext
var db *sql.DB

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

	cfg.File.Location = os.TempDir()

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
		Config: cfg.File,
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

	mailer := models.Mailer{
		Sender:    MockSMTP{},
		AppConfig: &cfg.Application,
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
		Mailer:            mailer,
		ConfigService:     cfg,
	}
}

func teardown() {
	if db != nil {
		db.Close()
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

type MockSMTP struct{}

func (sm MockSMTP) Send(m mail.Mail) error {
	return nil
}

func (sm MockSMTP) SendAsync(m mail.Mail) error {
	return nil
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

func postMultipart(path string, mp []multipartRequest) (*http.Request, error) {
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)

	defer mw.Close()

	for _, v := range mp {
		fh, err := os.Open(v.file)

		if err != nil {
			return nil, err
		}

		defer fh.Close()

		fw, err := mw.CreateFormFile(v.key, filepath.Base(fh.Name()))

		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fw, fh)

		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("POST", path, buf)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	return req, nil
}

func post(path string, values url.Values) (*http.Request, error) {
	var b bytes.Buffer
	b.WriteString(values.Encode())
	req, err := http.NewRequest("POST", path, &b)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func get(path string, values url.Values) (*http.Request, error) {
	var b bytes.Buffer
	b.WriteString(values.Encode())
	req, err := http.NewRequest("GET", path, &b)

	if err != nil {
		return nil, err
	}

	return req, nil
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

//request used to build an http.Request with specified values
//url will not really considered as the requests are not send, the *http.Request is just passed directly to the controllers
//pathvar is an array of key/value pairs used as dynamic query parameters such as /article/{id}
type request struct {
	url          string
	user         reqUser
	method       string
	values       url.Values
	pathVar      []pathVar
	multipartReq []multipartRequest
}

type multipartRequest struct {
	key  string
	file string
}

type pathVar struct {
	key   string
	value string
}

func (r request) buildRequest() *http.Request {
	var req *http.Request
	var err error

	if len(r.multipartReq) > 0 {
		req, err = postMultipart(r.url, r.multipartReq)
	} else if r.method == http.MethodPost {
		req, err = post(r.url, r.values)
	} else {
		req, err = get(r.url, r.values)
	}

	if err != nil {
		log.Print(err)
	}

	if r.pathVar != nil {
		for _, v := range r.pathVar {
			setHeader(req, v.key, v.value)
		}
	}

	var user *models.User

	if r.user == rGuest {
		return req
	} else {
		user, _ = ctx.UserService.GetByID(int(r.user))

		recorder := httptest.NewRecorder()
		session := ctx.SessionService.Create(recorder, req)
		session.SetValue("userid", user.ID)

		cookie := recorder.Result().Cookies()[0]
		req.AddCookie(cookie)
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
