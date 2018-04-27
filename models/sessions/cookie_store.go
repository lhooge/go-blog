// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sessions

import (
	"net/http"
	"time"
)

//CookieStore contains settings for the session
type CookieStore struct {
	Path            string
	HTTPOnly        bool
	Name            string
	Secure          bool
	IdleSessionTTL  int64
	SessionProvider SessionProvider
}

//CreateForUser creates a session for a user
func (cs CookieStore) CreateForUser(rw http.ResponseWriter, r *http.Request, userID int) Session {
	sid := cs.SessionProvider.Create(userID)
	cookie := &http.Cookie{
		Path:     cs.Path,
		HttpOnly: cs.HTTPOnly,
		Name:     cs.Name,
		Secure:   cs.Secure,
		Value:    sid.GetSessionID(),
	}

	http.SetCookie(rw, cookie)

	return sid
}

//Get receives the session from the cookie
func (cs CookieStore) Get(rw http.ResponseWriter, r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(cs.Name)

	if err != nil {
		return nil, err
	}
	sess, err := cs.SessionProvider.Get(cookie.Value)

	if err != nil {
		//Try to remove client cookie as it is not valid anymore
		dc := &http.Cookie{
			Name:    cs.Name,
			MaxAge:  -1,
			Expires: time.Unix(1, 0),
			Path:    cs.Path,
		}

		http.SetCookie(rw, dc)

		return nil, err
	}

	return &sess, nil
}

//Remove removes the session from the session map and the cookie
func (cs CookieStore) Remove(rw http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(cs.Name)

	if err != nil {
		return err
	}

	cs.SessionProvider.Remove(cookie.Value)

	dc := &http.Cookie{
		Name:    cs.Name,
		MaxAge:  -1,
		Expires: time.Unix(1, 0),
		Path:    cs.Path,
	}

	http.SetCookie(rw, dc)
	return nil
}

//InitGC initialized the garbage collection for removing the session after the TTL has reached
func (cs CookieStore) InitGC(ticker *time.Ticker, timeoutAfter time.Duration) {
	cs.SessionProvider.Clean(ticker, timeoutAfter)
}
