// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers

import (
	"net/http"
	"strconv"

	"git.hoogi.eu/go-blog/components/logger"
	"github.com/gorilla/mux"
)

const (
	tplAdminAction = "admin/action"
	tplAdminError  = "admin/error"
)

func getVar(r *http.Request, key string) string {
	v := mux.Vars(r)[key]
	if len(v) > 0 {
		return v
	}
	return r.Header.Get("X-Unit-Testing-Value-" + key)
}

func parseInt(v string) (int, error) {
	i, err := strconv.Atoi(v)
	if err != nil {
		return -1, err
	}
	return i, nil
}

func getPageParam(r *http.Request) int {
	var page = 1

	if len(getVar(r, "page")) > 0 {
		i, err := strconv.Atoi(getVar(r, "page"))

		if err != nil {
			logger.Log.Errorf("could not parse page number %v\n", err)
			return 1
		}

		page = i
	}

	return page
}

func convertCheckbox(r *http.Request, name string) bool {
	return r.FormValue(name) == "on"
}
