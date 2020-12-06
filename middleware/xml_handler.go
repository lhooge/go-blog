package middleware

import (
	"bytes"
	"encoding/xml"
	"net/http"

	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/models"
)

// XMLHandler marshals XML and writes to the http response
type XMLHandler struct {
	AppCtx  *AppContext
	Handler XHandler
}

//XNLHandler enriches handler with the AppContext
type XHandler func(*AppContext, http.ResponseWriter, *http.Request) (*models.XMLData, error)

func (fn XMLHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/xml")

	h, err := fn.Handler(fn.AppCtx, rw, r)

	if err != nil {
		logger.Log.Error(err)

		xml, err := xml.Marshal(err)

		if err != nil {
			logger.Log.Error(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Write(xml)
		return
	}

	x, err := xml.MarshalIndent(h.Data, "", "\t")

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	x = []byte(xml.Header + string(x))

	if h.HexEncode {
		x = bytes.Replace(x, []byte("&amp;"), []byte("&#x26;"), -1) // &
		x = bytes.Replace(x, []byte("&#39;"), []byte("&#x27;"), -1) // '
		x = bytes.Replace(x, []byte("&#34;"), []byte("&#x22;"), -1) // "
		x = bytes.Replace(x, []byte("&lt;"), []byte("&#x3c;"), -1)  // <
		x = bytes.Replace(x, []byte("&gt;"), []byte("&#x3e;"), -1)  // >
	}

	rw.Write(x)
}
