// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package json

import (
	"net/http"

	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

// KeepAliveSessionHandler - Keeps a session alive.
func KeepAliveSessionHandler(ctx *middleware.AppContext, rw http.ResponseWriter, r *http.Request) (*models.JSONData, error) {
	_, err := ctx.SessionStore.Get(rw, r)

	if err != nil {
		return nil, err
	}

	data := &models.JSONData{
		Data: map[string]bool{"acknowledge": true},
	}

	return data, nil
}
