// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sessions

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/utils"
)

//SessionProvider an interface for storing and accessing sessions
type SessionProvider interface {
	Create(userID int) Session
	Get(sid string) (Session, error)
	Remove(sid string)
	Clean(ticker *time.Ticker, timeoutAfter time.Duration)
}

//InMemoryProvider implements a in memory storage for sessions
type InMemoryProvider struct {
	mutex    sync.RWMutex
	sessions map[string]Session
}

//NewInMemoryProvider creates a new in memory provider
func NewInMemoryProvider() *InMemoryProvider {
	return &InMemoryProvider{
		sessions: make(map[string]Session),
	}
}

//Create stores a session in the map
func (imp *InMemoryProvider) Create(userID int) Session {
	imp.mutex.Lock()
	defer imp.mutex.Unlock()
	sid := base64.URLEncoding.EncodeToString(utils.RandomSecureKey(64))
	imp.sessions[sid] = Session{sid: sid, userID: userID, lastTouched: time.Now(), values: make(map[interface{}]interface{})}

	return imp.sessions[sid]
}

//Get receives a session from the map
func (imp *InMemoryProvider) Get(sid string) (Session, error) {
	imp.mutex.RLock()
	defer imp.mutex.RUnlock()
	if sess, ok := imp.sessions[sid]; ok {
		sess.lastTouched = time.Now()
		imp.sessions[sid] = sess
		return sess, nil
	}

	return Session{}, fmt.Errorf("no session with id %s found", sid)
}

//Remove removes a session from the map
func (imp *InMemoryProvider) Remove(sid string) {
	imp.mutex.Lock()
	defer imp.mutex.Unlock()
	delete(imp.sessions, sid)
}

//Clean clean sessions after a specified timeout
//Checks every x durations defined in config
func (imp *InMemoryProvider) Clean(ticker *time.Ticker, timeoutAfter time.Duration) {
	go func() {
		for range ticker.C {
			imp.mutex.Lock()
			for key, value := range imp.sessions {
				if time.Now().After(value.lastTouched.Add(timeoutAfter)) {
					logger.Log.Debugf("deleting session user id: %d, last touched: %s", value.userID, value.lastTouched.String())
					delete(imp.sessions, key)
				}
			}
			imp.mutex.Unlock()
		}
	}()
}
