// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

const (
	tplAdminLogin = "admin/login"
)

// LoginHandler shows the login form;
// if the user is already logged in the user will be redirected to the administration page of aricles
func LoginHandler(ctx *middleware.AppContext, rw http.ResponseWriter, r *http.Request) *middleware.Template {
	_, err := ctx.SessionService.Get(rw, r)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminLogin,
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
	}

}

// LoginPostHandler receives the login information from the form; checks the login and
// starts a session for the user. The sesion will be stored in a cookie
func LoginPostHandler(ctx *middleware.AppContext, rw http.ResponseWriter, r *http.Request) *middleware.Template {
	if err := r.ParseForm(); err != nil {
		return &middleware.Template{
			Name: tplAdminLogin,
			Err:  err,
		}
	}

	username := r.PostFormValue("username")
	password := []byte(r.PostFormValue("password"))
	redirectTo := r.PostFormValue("state")

	if len(redirectTo) == 0 {
		redirectTo = "admin/articles"
	}

	user := &models.User{
		Username: username,
		Email:    username,
	}

	user, err := ctx.UserService.Authenticate(user, ctx.ConfigService.LoginMethod, password)

	if err != nil {
		//Do some extra work
		if user == nil {
			bcrypt.CompareHashAndPassword([]byte("$2a$12$bQlRnXTNZMp6kCyoAlnf3uZW5vtmSj9CHP7pYplRUVK2n0C5xBHBa"), password)
		}

		hErr, ok := err.(*httperror.Error)

		if ok {
			return &middleware.Template{
				Name: tplAdminLogin,
				Err:  httperror.New(http.StatusUnauthorized, "Your username or password is invalid.", hErr.Err),
				Data: map[string]interface{}{
					"user": models.User{
						Username: username,
					},
				},
			}
		}
		return &middleware.Template{
			Name: tplAdminLogin,
			Err:  httperror.New(http.StatusUnauthorized, "Your username or password is invalid.", err),
			Data: map[string]interface{}{
				"user": models.User{
					Username: username,
				},
			},
		}
	}

	session := ctx.SessionService.Create(rw, r)
	session.SetValue("userid", user.ID)

	return &middleware.Template{
		RedirectPath: redirectTo,
	}
}

// LogoutHandler logs the user out by removing the cookie and removing the session from the session store
func LogoutHandler(ctx *middleware.AppContext, rw http.ResponseWriter, r *http.Request) *middleware.Template {
	ctx.SessionService.Remove(rw, r)

	return &middleware.Template{
		RedirectPath: "admin",
		SuccessMsg:   "Successfully logged out",
	}
}

// KeepAliveSessionHandler keeps a session alive.
func KeepAliveSessionHandler(ctx *middleware.AppContext, rw http.ResponseWriter, r *http.Request) (*models.JSONData, error) {
	_, err := ctx.SessionService.Get(rw, r)

	if err != nil {
		return nil, err
	}

	data := &models.JSONData{
		Data: map[string]bool{
			"acknowledge": true,
		},
	}

	return data, nil
}
