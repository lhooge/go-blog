// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"encoding/json"
	"net/http"

	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/models"
)

// JSONHandler marshals JSON and writes to the http response
type JSONHandler struct {
	AppCtx  *AppContext
	Handler JHandler
}

//JHandler enriches handler with the AppContext
type JHandler func(*AppContext, http.ResponseWriter, *http.Request) (*models.JSONData, error)

func (fn JSONHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	code := http.StatusOK

	rw.Header().Set("Content-Type", "application/json")

	data, err := fn.Handler(fn.AppCtx, rw, r)

	if err != nil {
		switch e := err.(type) {
		case *httperror.Error:
			code = e.HTTPStatus
		default:
			code = http.StatusInternalServerError
			logger.Log.Error(e)
		}

		logger.Log.Error(err)

		j, err := json.Marshal(err)
		if err != nil {
			logger.Log.Error(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(code)
		rw.Write(j)
		return
	}

	j, err := json.Marshal(data)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(code)
	rw.Write(j)
}
