// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"errors"
	"net/http"
	"path"

	"github.com/gorilla/csrf"

	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/models"
)

type contextKey string

var (
	UserContextKey = contextKey("user")
)

// TemplateHandler enriches handlers with a application context containing 'services'
type TemplateHandler struct {
	AppCtx  *AppContext
	Handler Handler
}

// Handler enriches handler with the AppContext
type Handler func(*AppContext, http.ResponseWriter, *http.Request) *Template

func (fn TemplateHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	t := fn.Handler(fn.AppCtx, rw, r)

	if t.Data == nil {
		t.Data = make(map[string]interface{})
	}

	var errorMsg, warnMsg, successMsg string
	successMsg = t.SuccessMsg
	warnMsg = t.WarnMsg
	code := http.StatusOK

	logWithIP := logger.Log.WithField("ip", getIP(r))

	t.Data["CSRFToken"] = csrf.Token(r)

	if t.Err != nil {
		switch e := t.Err.(type) {
		case *httperror.Error:
			code = e.HTTPStatus
			logWithIP.Error(e)
			errorMsg = e.DisplayMsg
		default:
			logWithIP.Error(e)
			errorMsg = "Sorry, an internal server error occurred"
		}

		t.Data["ErrorMsg"] = errorMsg
	}

	if user, err := User(r); err == nil {
		t.Data["currentUser"] = user
	}

	if len(t.RedirectPath) == 0 {
		t.Data["SuccessMsg"] = successMsg
		t.Data["WarnsMsg"] = warnMsg

		fl, err := getFlash(rw, r, "SuccessMsg")

		if err != nil {
			logWithIP.Error(err)
		} else if len(fl) > 0 {
			t.Data["SuccessMsg"] = fl
		}

		fl, err = getFlash(rw, r, "ErrorMsg")

		if err != nil {
			logWithIP.Error(err)
		} else if len(fl) > 0 {
			t.Data["ErrorMsg"] = fl
		}

		fl, err = getFlash(rw, r, "WarnMsg")

		if err != nil {
			logWithIP.Error(err)
		} else if len(fl) > 0 {
			t.Data["WarnMsg"] = fl
		}

		t.Data[csrf.TemplateTag] = csrf.TemplateField(r)
		t.Data["active"] = t.Active

		rw.WriteHeader(code)

		if err := fn.AppCtx.Templates.ExecuteTemplate(rw, t.Name, t.Data); err != nil {
			logWithIP.Error(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else {
		code = http.StatusFound
		if len(errorMsg) > 0 {
			setCookie(rw, "ErrorMsg", "/", errorMsg)
		} else if len(warnMsg) > 0 {
			setCookie(rw, "WarnMsg", "/", warnMsg)
		} else if len(successMsg) > 0 {
			setCookie(rw, "SuccessMsg", "/", successMsg)
		}
		http.Redirect(rw, r, path.Clean(t.RedirectURL()), code)
	}
}

// AuthHandler checks if the user is authenticated; if not next handler in chain is not called
func (ctx AppContext) AuthHandler(handler http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		session, err := ctx.SessionService.Get(rw, r)

		if err != nil {
			logger.Log.Error(err)
			rw.WriteHeader(http.StatusUnauthorized)
			if err := ctx.Templates.ExecuteTemplate(rw, "admin/login", map[string]interface{}{
				"ErrorMsg":       "Please provide login credentials.",
				"state":          r.URL.EscapedPath(),
				csrf.TemplateTag: csrf.TemplateField(r),
			}); err != nil {
				logger.Log.Errorf("error while executing the template %v", err)
				return
			}
			return
		}

		userid, ok := session.GetValue("userid").(int)

		if !ok {
			logger.Log.Errorf("userid is not an integer %v", userid)

			rw.WriteHeader(http.StatusUnauthorized)
			if err := ctx.Templates.ExecuteTemplate(rw, "admin/login", map[string]interface{}{
				"ErrorMsg":       "Please provide login credentials.",
				"state":          r.URL.EscapedPath(),
				csrf.TemplateTag: csrf.TemplateField(r),
			}); err != nil {
				logger.Log.Errorf("error while executing the template %v", err)
				return
			}

			return
		}

		u, err := ctx.UserService.GetByID(userid)

		if err != nil {
			logger.Log.Error(err)
			rw.WriteHeader(http.StatusUnauthorized)
			if err := ctx.Templates.ExecuteTemplate(rw, "admin/login", map[string]interface{}{
				"ErrorMsg":       "Please provide login credentials.",
				"state":          r.URL.EscapedPath(),
				csrf.TemplateTag: csrf.TemplateField(r),
			}); err != nil {
				logger.Log.Errorf("error while executing the template %v", err)
				return
			}
			return
		}

		handler.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), UserContextKey, u)))
	}
	return http.HandlerFunc(fn)
}

// RequireAdmin ensures that the user is an admin; if not next handler in chain is not called.
func (ctx AppContext) RequireAdmin(handler http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		u, err := User(r)

		if err != nil {
			logger.Log.Error(err)
			if err := ctx.Templates.ExecuteTemplate(rw, "admin/error", map[string]interface{}{
				"ErrorMsg": "An internal server error occurred",
			}); err != nil {
				logger.Log.Errorf("error while executing the template %v", err)
				return
			}
			return
		}

		if u.IsAdmin == false {
			if err := ctx.Templates.ExecuteTemplate(rw, "admin/error", map[string]interface{}{
				"ErrorMsg":    "You have not the permissions to execute this action",
				"currentUser": u,
			}); err != nil {
				logger.Log.Errorf("error while executing the template %v", err)
				return
			}
			return
		}

		handler.ServeHTTP(rw, r)
	}
	return http.HandlerFunc(fn)
}

// User gets the user from the request context
func User(r *http.Request) (*models.User, error) {
	v := r.Context().Value(UserContextKey)
	if v == nil {
		return nil, httperror.InternalServerError(errors.New("user is not available in context"))
	}

	return v.(*models.User), nil
}
