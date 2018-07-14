package middleware

import (
	"encoding/xml"
	"net/http"

	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/models"
)

// XMLHandler marshals XML and writes to the http response
type XMLHandler struct {
	AppCtx  *AppContext
	Handler XHandler
}

//XNLHandler enriches handler with the AppContext
type XHandler func(*AppContext, http.ResponseWriter, *http.Request) (*models.Data, error)

func (fn XMLHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/xml")

	h, err := fn.Handler(fn.AppCtx, rw, r)

	if err != nil {
		logger.Log.Error(err)

		x, err2 := xml.Marshal(err)

		if err2 != nil {
			logger.Log.Error(err2)
			http.Error(rw, err2.Error(), http.StatusInternalServerError)
			return
		}

		rw.Write(x)
		return
	}

	x, err2 := xml.Marshal(h.Data)

	if err2 != nil {
		http.Error(rw, err2.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(x)
}
