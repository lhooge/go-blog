// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"errors"
	"net/http"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/models"
	"github.com/gorilla/csrf"
)

type contextKey string

var (
	UserContextKey = contextKey("user")
)

//AuthHandler checks if the user is authenticated; if not next handler in chain is not called
func (ctx AppContext) AuthHandler(handler http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		session, err := ctx.SessionService.Get(rw, r)

		if err != nil {
			logger.Log.Error(err)
			rw.WriteHeader(http.StatusUnauthorized)
			ctx.Templates.ExecuteTemplate(rw, "admin/login", map[string]interface{}{
				"ErrorMsg":       "Please provide login credentials.",
				"state":          r.URL.EscapedPath(),
				csrf.TemplateTag: csrf.TemplateField(r),
			})
			return
		}

		userid, ok := session.GetValue("userid").(int)
		if !ok {
			logger.Log.Error("userid is not an integer %v", userid)

			rw.WriteHeader(http.StatusUnauthorized)
			ctx.Templates.ExecuteTemplate(rw, "admin/login", map[string]interface{}{
				"ErrorMsg":       "Please provide login credentials.",
				"state":          r.URL.EscapedPath(),
				csrf.TemplateTag: csrf.TemplateField(r),
			})
			return
		}

		u, err := ctx.UserService.GetUserByID(userid)
		if err != nil {
			logger.Log.Error(err)
			rw.WriteHeader(http.StatusUnauthorized)
			ctx.Templates.ExecuteTemplate(rw, "admin/login", map[string]interface{}{
				"ErrorMsg":       "Please provide login credentials.",
				"state":          r.URL.EscapedPath(),
				csrf.TemplateTag: csrf.TemplateField(r),
			})
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, u)
		handler.ServeHTTP(rw, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

//RequireAdmin ensures that the user is an admin; if not next handler in chain is not called
func (ctx AppContext) RequireAdmin(handler http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		u, err := User(r)

		if err != nil {
			ctx.Templates.ExecuteTemplate(rw, "admin/error", map[string]interface{}{
				"ErrorMsg": "An internal server error occured",
			})
			return
		}

		if u.IsAdmin == false {
			ctx.Templates.ExecuteTemplate(rw, "admin/error", map[string]interface{}{
				"ErrorMsg":    "You have not the permissions to execute this action",
				"currentUser": u,
			})
			return
		}

		handler.ServeHTTP(rw, r)
	}
	return http.HandlerFunc(fn)
}

//User gets the user from the request context
func User(r *http.Request) (*models.User, error) {
	v := r.Context().Value(UserContextKey)
	if v == nil {
		return nil, httperror.InternalServerError(errors.New("user is not available in context. is the authentication handler in chain?"))
	}

	return v.(*models.User), nil
}
