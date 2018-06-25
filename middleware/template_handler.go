// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
)

//TemplateHandler enriches handlers with a application context containing 'services'
type TemplateHandler struct {
	AppCtx  *AppContext
	Handler Handler
}

//Handler enriches handler with the AppContext
type Handler func(*AppContext, http.ResponseWriter, *http.Request) *Template

func (fn TemplateHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var errorMsg, successMsg string
	statusCode := 200

	t := fn.Handler(fn.AppCtx, rw, r)

	if t.Data == nil {
		t.Data = make(map[string]interface{})
	}

	user, err := User(r)

	if err == nil {
		t.Data["currentUser"] = user
	}

	successMsg = t.SuccessMsg

	if t.Err != nil {
		switch e := t.Err.(type) {
		case *httperror.Error:
			statusCode = e.HTTPStatus
			ip := getIP(r)
			en := logger.Log.WithField("ip", ip)
			en.Error(e)
			errorMsg = e.DisplayMsg
		default:
			logger.Log.Error(e)
			errorMsg = "Sorry, an internal server error occured"
		}

		t.Data["ErrorMsg"] = errorMsg
	}

	if len(t.RedirectPath) == 0 {
		fl, err := getFlash(rw, r, "SuccessMsg")

		if err != nil {
			logger.Log.Error(err)
		} else if len(fl) > 0 {
			t.Data["SuccessMsg"] = fl
		}

		fl, err = getFlash(rw, r, "ErrorMsg")

		if err != nil {
			logger.Log.Error(err)
		} else if len(fl) > 0 {
			t.Data["ErrorMsg"] = fl
		}

		t.Data[csrf.TemplateTag] = csrf.TemplateField(r)
		t.Data["active"] = t.Active

		rw.WriteHeader(statusCode)

		if err := fn.AppCtx.Templates.ExecuteTemplate(rw, t.Name, t.Data); err != nil {
			logger.Log.Error(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	} else {
		statusCode = http.StatusFound
		if len(errorMsg) > 0 {
			setCookie(rw, "ErrorMsg", "/", errorMsg)
		} else if len(successMsg) > 0 {
			setCookie(rw, "SuccessMsg", "/", successMsg)
		}
		http.Redirect(rw, r, t.RedirectURL(), statusCode)
	}
}
