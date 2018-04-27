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

	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/models/sessions"
)

var ctx *middleware.AppContext

func init() {
	articleService := models.ArticleService{Datasource: &inMemoryArticle{articles: make(map[int]*models.Article)}}
	userService := models.UserService{Datasource: &inMemoryUser{users: make(map[int]*models.User)}}

	sessionStore := sessions.CookieStore{
		Path:            "/admin",
		Name:            "test-session",
		HTTPOnly:        true,
		Secure:          true,
		SessionProvider: sessions.NewInMemoryProvider(),
		IdleSessionTTL:  10,
	}

	ctx = &middleware.AppContext{
		UserService:    userService,
		ArticleService: articleService,
		SessionStore:   &sessionStore,
	}
}

func setHeader(r *http.Request, key, value string) {
	r.Header.Set("X-Unit-Testing-Value-"+key, value)
}

func setValues(m url.Values, key, value string) {
	m.Add(key, value)
}

func dummyAdminUser() *models.User {
	return &models.User{ID: 1, Email: "test@example.com", Lastname: "Simpson", Firstname: "Homer", Active: true, IsAdmin: true}
}

func dummyUser() *models.User {
	return &models.User{ID: 1, Email: "test@example.com", Lastname: "Simpson", Firstname: "Marge", Active: true, IsAdmin: false}
}

func dummyInactiveUser() models.User {
	return models.User{ID: 1, Email: "test@example.com", Lastname: "Simpson", Firstname: "Bart", Active: false, IsAdmin: false}
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
