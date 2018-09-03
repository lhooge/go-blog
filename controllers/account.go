package controllers

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/components/mail"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/utils"
)

//AdminProfileHandler returns page for updating the profile of the currently logged-in user
func AdminProfileHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	return &middleware.Template{
		Name: tplAdminProfile,
		Data: map[string]interface{}{
			"user": user,
		},
		Active: "profile",
	}
}

//AdminProfilePostHandler handles the updating of the user profile
func AdminProfilePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	ctxUser, _ := middleware.User(r)

	u := &models.User{
		ID:          ctxUser.ID,
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		DisplayName: r.FormValue("displayname"),
		Active:      true,
		IsAdmin:     ctxUser.IsAdmin,
	}

	if _, err := ctx.UserService.Authenticate(ctxUser, ctx.ConfigService.LoginMethod, []byte(r.PostFormValue("current_password"))); err != nil {
		return &middleware.Template{
			Name:   tplAdminProfile,
			Err:    httperror.New(http.StatusUnauthorized, "Your password is invalid.", err),
			Active: "profile",
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	changePassword := false

	if len(r.FormValue("password")) > 0 {
		changePassword = true
		// Password change
		u.Password = []byte(r.FormValue("password"))

		if !bytes.Equal(u.Password, []byte(r.FormValue("retyped_password"))) {
			return &middleware.Template{
				Name:   tplAdminProfile,
				Active: "profile",
				Err: httperror.New(http.StatusUnprocessableEntity,
					"Please check your retyped password",
					errors.New("the password did not match the retyped one")),
				Data: map[string]interface{}{
					"user": u,
				},
			}
		}
	}

	if err := ctx.UserService.UpdateUser(u, changePassword); err != nil {
		return &middleware.Template{
			Name:   tplAdminProfile,
			Active: "profile",
			Err:    err,
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/user/profile",
		Active:       "profile",
		SuccessMsg:   "Your profile update was successful",
		Data: map[string]interface{}{
			"user": u,
		},
	}
}

func ActivateAccountHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	hash := getVar(r, "hash")

	_, err := ctx.UserInviteService.GetByHash(hash)

	if err != nil {
		if err == sql.ErrNoRows {
			return &middleware.Template{
				Name: tplAdminLogin,
				Err:  httperror.New(http.StatusNotFound, "Could not find an invitation", errors.New("the invitation was not found")),
			}
		}

		return &middleware.Template{
			Name: tplAdminLogin,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name: tplAdminActivateAccount,
		Data: map[string]interface{}{
			"hash": hash,
		},
	}
}

func ActivateAccountPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	password := r.FormValue("password")
	repassword := r.FormValue("password_repeat")
	hash := getVar(r, "hash")

	if !bytes.Equal([]byte(password), []byte(repassword)) {
		return &middleware.Template{
			Name: tplAdminActivateAccount,
			Err: httperror.New(http.StatusUnprocessableEntity,
				"Please check your retyped password",
				errors.New("the password did not match the retyped one")),
			Data: map[string]interface{}{
				"hash": hash,
			},
		}
	}

	ui, err := ctx.UserInviteService.GetByHash(hash)

	if err != nil {
		if err == sql.ErrNoRows {
			return &middleware.Template{
				Name: tplAdminLogin,
				Err:  httperror.New(http.StatusNotFound, "Could not find the invitation", errors.New("the invitation was not found")),
			}
		}

		return &middleware.Template{
			Name: tplAdminActivateAccount,
			Err:  err,
		}
	}

	user := ui.Copy()

	user.Password = []byte(password)
	user.Active = true

	_, err = ctx.UserService.CreateUser(user)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminActivateAccount,
			Err:  err,
		}
	}

	return &middleware.Template{
		RedirectPath: "admin",
		SuccessMsg:   "The account was successfully activated. You can now log in.",
	}
}

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

	err = ctx.TokenService.CreateToken(t)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  err,
		}
	}

	resetLink := utils.AppendString(ctx.ConfigService.Blog.Domain, "/admin/reset-password/", t.Hash)

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
