package controllers

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
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
	ctxUser.PlainPassword = []byte(r.FormValue("current_password"))

	u := &models.User{
		ID:            ctxUser.ID,
		Username:      r.FormValue("username"),
		Email:         r.FormValue("email"),
		DisplayName:   r.FormValue("displayname"),
		Active:        true,
		IsAdmin:       ctxUser.IsAdmin,
		PlainPassword: []byte(r.FormValue("password")),
	}

	if _, err := ctx.UserService.Authenticate(ctxUser, ctx.ConfigService.LoginMethod); err != nil {
		return &middleware.Template{
			Name:   tplAdminProfile,
			Err:    httperror.New(http.StatusUnauthorized, "Your current password is invalid.", err),
			Active: "profile",
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	changePassword := false

	if len(u.PlainPassword) > 0 {
		changePassword = true
		// Password change
		u.PlainPassword = []byte(r.FormValue("password"))

		if !bytes.Equal(u.PlainPassword, []byte(r.FormValue("retyped_password"))) {
			return &middleware.Template{
				Name:   tplAdminProfile,
				Active: "profile",
				Err:    httperror.New(http.StatusUnprocessableEntity, "Please check your retyped password", errors.New("the password did not match the retyped one")),
				Data: map[string]interface{}{
					"user": u,
				},
			}
		}
	}

	if err := ctx.UserService.Update(u, changePassword); err != nil {
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
	// Delete any cookies if user is logged in
	ctx.SessionService.Remove(w, r)

	password := r.FormValue("password")
	repassword := r.FormValue("password_repeat")
	hash := getVar(r, "hash")

	if !bytes.Equal([]byte(password), []byte(repassword)) {
		return &middleware.Template{
			Name: tplAdminActivateAccount,
			Err:  httperror.New(http.StatusUnprocessableEntity, "Please check your retyped password", errors.New("the password did not match the retyped one")),
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

	if _, err := ctx.UserService.Create(user); err != nil {
		return &middleware.Template{
			Name: tplAdminActivateAccount,
			Err:  err,
		}
	}

	if err := ctx.UserInviteService.Remove(ui.ID); err != nil {
		return &middleware.Template{
			Name:   tplAdminLogin,
			Active: "users",
			Err:    err,
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

	t, err := ctx.TokenService.Get(hash, models.PasswordReset, time.Duration(1)*time.Hour)

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

	t, err := ctx.TokenService.Get(hash, models.PasswordReset, time.Duration(1)*time.Hour)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  err,
		}
	}

	u, err := ctx.UserService.GetByID(t.Author.ID)

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

	err = ctx.UserService.Update(u, true)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  err,
		}
	}

	go func(hash string) {
		err = ctx.TokenService.Remove(hash, models.PasswordReset)

		if err != nil {
			logger.Log.Errorf("could not remove token %s error %v", hash, err)
		}

		err = ctx.Mailer.SendPasswordChangeConfirmation(u)

		if err != nil {
			logger.Log.Errorf("could not send password changed mail %v", err)
		}
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

	u, err := ctx.UserService.GetByMail(email)

	if err != nil {
		if httperror.Equals(err, sql.ErrNoRows) {
			return &middleware.Template{
				RedirectPath: "admin",
				SuccessMsg:   "An email with password reset instructions is on the way.",
			}
		} else {
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

	err = ctx.TokenService.Create(t)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  err,
		}
	}

	err = ctx.Mailer.SendPasswordResetLink(u, t)

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
