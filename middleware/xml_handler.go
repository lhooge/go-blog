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

// XHandler enriches handler with the AppContext
type XHandler func(*AppContext, http.ResponseWriter, *http.Request) (*models.XMLData, error)

func (fn XMLHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	logWithIP := logger.Log.WithField("ip", getIP(r))

	rw.Header().Set("Content-Type", "application/xml")

	h, err := fn.Handler(fn.AppCtx, rw, r)

	if err != nil {
		logWithIP.Error(err)

		x, err := xml.Marshal(err)

		if err != nil {
			logWithIP.Error(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = rw.Write(x); err != nil {
			logWithIP.Error(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	x, err2 := xml.MarshalIndent(h.Data, "", "\t")

	if err2 != nil {
		logWithIP.Error(err)
		http.Error(rw, err2.Error(), http.StatusInternalServerError)
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

	if _, err := rw.Write(x); err != nil {
		logger.Log.Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
