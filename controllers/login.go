// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/components/mail"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/utils"
)

const (
	tplAdminLogin          = "admin/login"
	tplAdminForgotPassword = "admin/forgot_password"
	tplAdminResetPassword  = "admin/reset_password"
)

//ResetPasswordHandler returns the form for the resetting the password
func ResetPasswordHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	hash := getVar(r, "hash")

	t, err := ctx.TokenService.GetToken(hash, models.PasswordReset, time.Duration(1)*time.Hour)

	if err != nil {
		if err == sql.ErrNoRows {
			return &middleware.Template{
				Name: tplAdminForgotPassword,
				Err:  httperror.New(http.StatusNotFound, "The token was not found. Fill out the form to receive another token", errors.New("the token was not found")),
			}
		}
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name: tplAdminResetPassword,
		Data: map[string]interface{}{
			"hash": t.Hash,
		},
	}
}

//ResetPasswordPostHandler handles the resetting of the password
func ResetPasswordPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	password := r.FormValue("password")
	repassword := r.FormValue("password_repeat")
	hash := getVar(r, "hash")

	t, err := ctx.TokenService.GetToken(hash, models.PasswordReset, time.Duration(1)*time.Hour)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  err,
		}
	}

	u, err := ctx.UserService.GetUserByID(t.Author.ID)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  err,
		}
	}

	if password != repassword {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  httperror.New(http.StatusUnprocessableEntity, "The passwords do not match", errors.New("the passwords did not match")),
		}
	}

	u.Password = []byte(password)

	err = ctx.UserService.UpdateUser(u, true)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  err,
		}
	}

	go func(hash string) {
		err = ctx.TokenService.RemoveToken(hash, models.PasswordReset)
		logger.Log.Errorf("could not remove token %s error %v", hash, err)

		m := mail.Mail{
			To:      u.Email,
			Subject: "Password change",
			Body:    fmt.Sprintf("Hi %s, \n\n your password change was sucessfully.", u.DisplayName),
		}

		err = ctx.MailService.Send(m)
		logger.Log.Errorf("could not send password changed mail %v", err)
	}(hash)

	return &middleware.Template{
		RedirectPath: "admin",
		SuccessMsg:   "Your password reset was successful. Please enter your login information.",
	}
}

//ForgotPasswordHandler returns the form for the reset password form
func ForgotPasswordHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name: tplAdminForgotPassword,
	}
}

//ForgotPasswordPostHandler handles the processing of the reset password function
func ForgotPasswordPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	email := r.FormValue("email")

	u, err := ctx.UserService.GetUserByMail(email)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  err,
			Data: map[string]interface{}{
				"user": models.User{
					Email: email,
				},
			},
		}
	}

	if !u.Active {
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  httperror.New(http.StatusUnauthorized, "Your account is deactivated.", err),
			Data: map[string]interface{}{
				"user": models.User{
					Email: email,
				},
			},
		}
	}

	t := &models.Token{
		Author: u,
		Type:   models.PasswordReset,
	}

	err = ctx.TokenService.AddToken(t)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  err,
		}
	}

	resetLink := utils.AppendString(ctx.ConfigService.Blog.Domain, "/reset-password/", t.Hash)

	m := mail.Mail{
		To:      u.Email,
		Subject: "Changing password instructions",
		Body:    fmt.Sprintf("Hi %s, \n\n use the following link to reset your password: \n\n. %s", u.DisplayName, resetLink),
	}

	err = ctx.MailService.Send(m)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  err,
		}
	}

	return &middleware.Template{
		RedirectPath: "admin",
		SuccessMsg:   "An email with password reset instructions is on the way.",
	}
}

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
func KeepAliveSessionHandler(ctx *middleware.AppContext, rw http.ResponseWriter, r *http.Request) (*models.Data, error) {
	_, err := ctx.SessionService.Get(rw, r)

	if err != nil {
		return nil, err
	}

	data := &models.Data{
		Data: map[string]bool{
			"acknowledge": true,
		},
	}

	return data, nil
}
