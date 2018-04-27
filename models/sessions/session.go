// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sessions

import (
	"time"

	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/models"
)

//Session contains the session
type Session struct {
	sid         string
	userID      int
	lastTouched time.Time

	values map[interface{}]interface{}
}

// GetUserID gets the userID from the session; which was set by createForUser
func (s *Session) GetUserID() int {
	return s.userID
}

// GetUser returns the user from the session
func (s *Session) GetUser(userService models.UserService) *models.User {
	user, err := userService.GetUserByID(s.GetUserID())
	if err != nil {
		logger.Log.Errorf("could not return user id %d from session %v", s.GetUserID(), err)
	}
	return user
}

//GetLastTouchTime recveives the date when the session was touched
func (s *Session) GetLastTouchTime() time.Time {
	return s.lastTouched
}

//GetSessionID gets the sessionID
func (s *Session) GetSessionID() string {
	return s.sid
}

//SetValue sets a value into the session
func (s *Session) SetValue(key interface{}, value interface{}) {
	s.values[key] = value
}

//GetValue receives a value from the session
func (s *Session) GetValue(key interface{}) interface{} {
	return s.values[key]
}

//RemoveValue removes a previously set value from the session
func (s *Session) RemoveValue(key interface{}) {
	delete(s.values, key)
}
