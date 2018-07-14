// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"encoding/json"
	"net/http"

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
type JHandler func(*AppContext, http.ResponseWriter, *http.Request) (*models.Data, error)

func (fn JSONHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	retJSON, err := fn.Handler(fn.AppCtx, rw, r)
	if err != nil {
		logger.Log.Error(retJSON)

		js, err2 := json.Marshal(err)
		if err2 != nil {
			logger.Log.Error(err2)
			http.Error(rw, err2.Error(), http.StatusInternalServerError)
			return
		}

		rw.Write(js)
		return
	}

	js, err2 := json.Marshal(retJSON)

	if err2 != nil {
		http.Error(rw, err2.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(js)
}
