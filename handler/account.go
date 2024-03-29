package handler

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/middleware"
	"git.hoogi.eu/snafu/go-blog/models"
)

// AdminProfileHandler returns the page for updating the profile of the currently logged-in user
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

// AdminProfilePostHandler handles the updating of the user profile which is currently logged in
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

		u.PlainPassword = []byte(r.FormValue("password"))

		if !bytes.Equal(u.PlainPassword, []byte(r.FormValue("retyped_password"))) {
			return &middleware.Template{
				Name:   tplAdminProfile,
				Active: "profile",
				Err:    httperror.New(http.StatusUnprocessableEntity, "The passwords entered do not match.", errors.New("the password entered did not match")),
				Data: map[string]interface{}{
					"user": u,
				},
			}
		}
	}

	if changePassword {
		session, err := ctx.SessionService.Renew(w, r)

		if err != nil {
			logger.Log.Error(err)
		}

		session.SetValue("userid", u.ID)

		sessions := ctx.SessionService.SessionProvider.FindByValue("userid", u.ID)

		for _, sid := range sessions {
			if sid.SessionID() != session.SessionID() {
				ctx.SessionService.SessionProvider.Remove(sid.SessionID())
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
		SuccessMsg:   "Your profile was successfully updated.",
		Data: map[string]interface{}{
			"user": u,
		},
	}
}

// ActivateAccountHandler shows the form to activate an account
func ActivateAccountHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	hash := getVar(r, "hash")

	_, err := ctx.UserInviteService.GetByHash(hash)

	if err != nil {
		if err == sql.ErrNoRows {
			return &middleware.Template{
				Name: tplAdminLogin,
				Err:  httperror.New(http.StatusNotFound, "Can't find an invitation.", fmt.Errorf("the invitation with hash %s was not found", hash)),
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

// ActivateAccountPostHandler activates an user account
func ActivateAccountPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	// Delete cookie if the user is logged in
	ctx.SessionService.Remove(w, r)

	password := r.FormValue("password")
	repassword := r.FormValue("password_repeat")
	hash := getVar(r, "hash")

	if !bytes.Equal([]byte(password), []byte(repassword)) {
		return &middleware.Template{
			Name: tplAdminActivateAccount,
			Err:  httperror.New(http.StatusUnprocessableEntity, "The passwords entered do not match.", errors.New("the password entered did not match")),
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
				Err:  httperror.New(http.StatusNotFound, "Can't find an invitation.", fmt.Errorf("the invitation with hash %s was not found", hash)),
			}
		}

		return &middleware.Template{
			Name: tplAdminActivateAccount,
			Err:  err,
		}
	}

	user := ui.Copy()

	user.PlainPassword = []byte(password)
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
		SuccessMsg:   "The account was successfully activated. You can now login.",
	}
}

// ResetPasswordHandler returns the form to reset the password
func ResetPasswordHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	hash := getVar(r, "hash")

	t, err := ctx.TokenService.Get(hash, models.PasswordReset, time.Duration(1)*time.Hour)

	if err != nil {
		if err == sql.ErrNoRows {
			return &middleware.Template{
				Name: tplAdminForgotPassword,
				Err:  httperror.New(http.StatusNotFound, "The token was not found. Please fill out the form to receive another token.", errors.New("the token to reset the password was not found")),
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

// ResetPasswordPostHandler handles a password reset
func ResetPasswordPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	hash := getVar(r, "hash")
	password := r.FormValue("password")
	password2 := r.FormValue("password_repeat")

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

	if password != password2 {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  httperror.New(http.StatusUnprocessableEntity, "The passwords entered do not match.", errors.New("the password entered did not match")),
		}
	}

	u.PlainPassword = []byte(password)

	err = ctx.UserService.Update(u, true)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminResetPassword,
			Err:  err,
		}
	}

	err = ctx.TokenService.Remove(hash, models.PasswordReset)

	if err != nil {
		logger.Log.Errorf("could not remove token %s error %v", hash, err)
	}

	ctx.Mailer.SendPasswordChangeConfirmation(u)

	if !u.Active {
		logger.Log.Warnf("password reset for user '%s' was successful, but user is deactivated", u.Email)

		return &middleware.Template{
			Name:    tplAdminResetPassword,
			WarnMsg: "Your password reset was successful, but your account is deactivated.",
		}
	}

	return &middleware.Template{
		RedirectPath: "admin",
		SuccessMsg:   "Your password reset was successful.",
	}
}

// ForgotPasswordHandler returns the form for the password reset
func ForgotPasswordHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name: tplAdminForgotPassword,
	}
}

// ForgotPasswordPostHandler handles the processing for the password reset
func ForgotPasswordPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	email := r.FormValue("email")

	u, err := ctx.UserService.GetByMail(email)

	if err != nil {
		var e *httperror.Error
		if errors.As(err, &e) {
			if errors.Is(e.Err, sql.ErrNoRows) {
				logger.Log.Error(err)
				return &middleware.Template{
					Name:       tplAdminForgotPassword,
					SuccessMsg: fmt.Sprintf("An email to '%s' with password reset instructions is on the way.", email),
				}
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

	t := &models.Token{
		Author: u,
		Type:   models.PasswordReset,
	}

	if err = ctx.TokenService.RateLimit(u.ID, models.PasswordReset); err != nil {
		logger.Log.Error(err)
	}

	if err = ctx.TokenService.Create(t); err != nil {
		return &middleware.Template{
			Name: tplAdminForgotPassword,
			Err:  err,
		}
	}

	ctx.Mailer.SendPasswordResetLink(u, t)

	return &middleware.Template{
		Name:       tplAdminForgotPassword,
		SuccessMsg: fmt.Sprintf("An email to '%s' with password reset instructions is on the way.", email),
	}
}
