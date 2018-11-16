// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/session"
)

var ctx *middleware.AppContext

func init() {
	articleService := models.ArticleService{Datasource: &inMemoryArticle{articles: make(map[int]*models.Article)}}
	userService := models.UserService{Datasource: &inMemoryUser{users: make(map[int]*models.User)}}

	//TODO proper test config
	cfg, err := settings.LoadConfig("../go-blog.conf")

	if err != nil {
		logger.Log.Error(err)
		panic(1)
	}

	s := session.SessionService{
		Path:            "/admin",
		Name:            "test-session",
		HTTPOnly:        true,
		Secure:          true,
		SessionProvider: session.NewInMemoryProvider(),
		IdleSessionTTL:  10,
	}

	ctx = &middleware.AppContext{
		UserService:    userService,
		ArticleService: articleService,
		SessionService: &s,
		ConfigService:  cfg,
	}
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

func dummyAdminUser() *models.User {
	return &models.User{ID: 1, Email: "test@example.com", Username: "homer", DisplayName: "Homer Simpson", Active: true}
}

func dummyUser() *models.User {
	return &models.User{ID: 1, Email: "test-marge@example.com", Username: "marge", DisplayName: "Marge Simpson", Active: true}
}

func dummyInactiveUser() models.User {
	return models.User{ID: 1, Email: "test-bart@example.com", Username: "bart", DisplayName: "Bart Simpson", Active: false}
}

func postRequest(path string, values url.Values) (*http.Request, error) {
	var b bytes.Buffer
	b.WriteString(values.Encode())
	req, err := http.NewRequest("POST", "/admin/article/new", &b)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

func getRequest(path string, values url.Values) (*http.Request, error) {
	var b bytes.Buffer
	b.WriteString(values.Encode())
	req, err := http.NewRequest("GET", "/admin/article/new", &b)
	return req, err
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
