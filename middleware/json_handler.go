// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"encoding/json"
	"net/http"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/models"
)

// JSONHandler marshals JSON and writes to the http response
// Currently just used for keeping the session alive (if writing or editing an article or site)
// see controllers/json/session.go
type JSONHandler struct {
	AppCtx  *AppContext
	Handler JHandler
}

//JHandler enriches handler with the AppContext
type JHandler func(*AppContext, http.ResponseWriter, *http.Request) (*models.JSONData, error)

func (fn JSONHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	statusCode := 200

	rw.Header().Set("Content-Type", "application/json")

	data, err := fn.Handler(fn.AppCtx, rw, r)

	if err != nil {
		switch e := err.(type) {
		case *httperror.Error:
			statusCode = e.HTTPStatus
		default:
			statusCode = 500
			logger.Log.Error(e)
		}

		logger.Log.Error(err)

		mjson, err2 := json.Marshal(err)
		if err2 != nil {
			logger.Log.Error(err2)
			http.Error(rw, err2.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(statusCode)
		rw.Write(mjson)
		return
	}

	mjson, err2 := json.Marshal(data)

	if err2 != nil {
		http.Error(rw, err2.Error(), http.StatusInternalServerError)
		rw.WriteHeader(500)
		return
	}

	rw.WriteHeader(statusCode)
	rw.Write(mjson)
}
