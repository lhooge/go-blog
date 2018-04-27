// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sessions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.hoogi.eu/go-blog/models/sessions"
)

func sampleCookieStore() sessions.CookieStore {
	return sessions.CookieStore{
		Secure:          true,
		Path:            "/test",
		HTTPOnly:        true,
		Name:            "test-session",
		SessionProvider: sessions.NewInMemoryProvider(),
		IdleSessionTTL:  2,
	}
}

var cs sessions.CookieStore

func init() {
	cs = sampleCookieStore()
}

func TestSessionLifeCycle(t *testing.T) {
	mockID := 1

	createdSess, rcvCookie := createSession(t, mockID, cs)
	if createdSess.GetUserID() != mockID {
		t.Fatalf("got an unexpected user id. Expected %d, bot got %d", mockID, createdSess.GetUserID())
	}

	expectedCookie := &http.Cookie{
		Name:     cs.Name,
		HttpOnly: cs.HTTPOnly,
		Secure:   cs.Secure,
		Path:     "/test",
		Value:    createdSess.GetSessionID(),
	}

	checkCookie(t, rcvCookie, expectedCookie)

	time.Sleep(time.Duration(1 * 1e9))
	getSess, err := getSession(t, rcvCookie.Raw, cs)
	if err != nil {
		t.Fatal(err)
	}
	if getSess.GetLastTouchTime().Unix() <= createdSess.GetLastTouchTime().Unix() {
		t.Fatalf("Last touch time not updated or equals. Initial %v, after get %v", createdSess.GetLastTouchTime().Unix(), getSess.GetLastTouchTime().Unix())
	}

	if getSess.GetSessionID() != createdSess.GetSessionID() {
		t.Fatalf("Got an invalid session id. Initial %s, after get %s", createdSess.GetSessionID(), getSess.GetSessionID())
	}

	removeSession(t, rcvCookie.Raw, cs)

	_, err = getSession(t, rcvCookie.Raw, cs)
	if err == nil {
		t.Fatalf("The session should be removed, but is still there %v", err)
	}
}

func TestSessionGarbageCollection(t *testing.T) {
	mockID := 1
	cs := sampleCookieStore()

	ticker := time.NewTicker(time.Duration(1 * 1e9))
	cs.InitGC(ticker, time.Duration(2*time.Second))

	createdSess, rcvCookie := createSession(t, mockID, cs)
	if createdSess.GetUserID() != mockID {
		t.Fatalf("got an unexpected user id. Expected %d, bot got %d", mockID, createdSess.GetUserID())
	}

	expectedCookie := &http.Cookie{
		Name:     cs.Name,
		HttpOnly: cs.HTTPOnly,
		Secure:   cs.Secure,
		Path:     "/test",
		Value:    createdSess.GetSessionID(),
	}

	checkCookie(t, rcvCookie, expectedCookie)

	time.Sleep(time.Duration(3 * 1e9))

	getSess, err := getSession(t, rcvCookie.Raw, cs)
	if err == nil {
		t.Fatal(err)
	}

	if getSess != nil {
		t.Fatalf("Got a session which should be invalidated. Initial %s", createdSess.GetSessionID())
	}
}

func createSession(t *testing.T, userID int, cs sessions.CookieStore) (sessions.Session, *http.Cookie) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := httptest.NewRecorder()

	createdSession := cs.CreateForUser(rw, req, userID)
	cookies := rw.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookies received")
	}
	return createdSession, cookies[0]
}

func getSession(t *testing.T, rawCookieValue string, cs sessions.CookieStore) (*sessions.Session, error) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", rawCookieValue)
	rw := httptest.NewRecorder()
	getSession, err := cs.Get(rw, req)
	return getSession, err
}

func removeSession(t *testing.T, rawCookieValue string, cs sessions.CookieStore) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Cookie", rawCookieValue)
	rw := httptest.NewRecorder()
	if err := cs.Remove(rw, req); err != nil {
		t.Fatal(err)
	}
}

func checkCookie(t *testing.T, rcvCookie, expectedCookie *http.Cookie) {
	if rcvCookie.Name != expectedCookie.Name {
		t.Errorf("got an unexpected cookie name. Expected %s, bot got %s", expectedCookie.Name, rcvCookie.Name)
	}
	if rcvCookie.Path != expectedCookie.Path {
		t.Errorf("got an unexpected cookie path. Expected %s, bot got %s", expectedCookie.Path, rcvCookie.Path)
	}
	if rcvCookie.Value != expectedCookie.Value {
		t.Errorf("got an unexpected cookie value. Expected %s, bot got %s", expectedCookie.Value, rcvCookie.Value)
	}
	if rcvCookie.HttpOnly != expectedCookie.HttpOnly {
		t.Errorf("got an unexpected cookie http only f;ag. Expected %t, bot got %t", expectedCookie.HttpOnly, rcvCookie.HttpOnly)
	}
	if rcvCookie.Secure != expectedCookie.Secure {
		t.Errorf("got an unexpected cookie secure flag. Expected %t, bot got %t", expectedCookie.Secure, rcvCookie.Secure)
	}
}
