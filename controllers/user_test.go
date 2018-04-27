// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

func TestCreateGetEditUser(t *testing.T) {
	defer ctx.UserService.Datasource.(*inMemoryUser).Flush()

	expectedUser := &models.User{
		Firstname: "Homer",
		Lastname:  "Simpson",
		Email:     "homer@example.com",
		Username:  "homer",
		Password:  []byte("1234567890"),
		Active:    true,
		IsAdmin:   false,
	}

	userID, err := doCreateUserRequest(expectedUser)
	if err != nil {
		t.Fatal(err)
	}

	expectedUser = &models.User{
		ID:        userID,
		Firstname: "Homer12",
		Lastname:  "Simpson",
		Email:     "homer@example.com",
		Username:  "homer",
		Password:  []byte("1234567890"),
		Active:    true,
		IsAdmin:   false,
	}

	err = doEditUsersRequest(expectedUser)
	if err != nil {
		t.Fatal(err)
	}

	user, err := doGetUserRequest(userID)
	if err != nil {
		t.Fatal(err)
	}

	err = checkUser(user, expectedUser)
	if err != nil {
		t.Fatal(err)
	}
}

func checkUser(user, expectedUser *models.User) error {
	if user.Firstname != expectedUser.Firstname {
		return fmt.Errorf("got an unexpected firstname. expected: %s, actual: %s", expectedUser.Firstname, user.Firstname)
	}
	if user.Lastname != expectedUser.Lastname {
		return fmt.Errorf("got an unexpected lastname. expected: %s, actual: %s", expectedUser.Lastname, user.Lastname)
	}
	if user.Email != expectedUser.Email {
		return fmt.Errorf("got an unexpected email. expected: %s, actual: %s", expectedUser.Email, user.Email)
	}
	if user.ID != expectedUser.ID {
		return fmt.Errorf("got an unexpected id. expected: %d, actual: %d", expectedUser.ID, user.ID)
	}
	if user.IsAdmin != expectedUser.IsAdmin {
		return fmt.Errorf("got an unexpected isAdmin. expected: %t, actual: %t", expectedUser.IsAdmin, user.IsAdmin)
	}
	if user.Active != expectedUser.Active {
		return fmt.Errorf("got an unexpected active. expected: %t, actual: %t", expectedUser.Active, user.Active)
	}
	return nil
}

func doGetUserRequest(userid int) (*models.User, error) {
	req, err := postRequest("/admin/user/", nil)
	if err != nil {
		return nil, err
	}

	setHeader(req, "userID", strconv.Itoa(userid))
	reqCtx := context.WithValue(req.Context(), middleware.UserContextKey, dummyAdminUser())
	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserEditHandler(ctx, rw, req.WithContext(reqCtx))

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	if w, ok := tpl.Data["user"].(*models.User); ok {
		return w, nil
	}

	return nil, fmt.Errorf("no user were returned %v", tpl.Err)
}

func doEditUsersRequest(user *models.User) error {
	values := url.Values{}

	setValues(values, "firstname", user.Firstname)
	setValues(values, "lastname", user.Lastname)
	setValues(values, "username", user.Username)
	setValues(values, "email", user.Email)
	s := "on"
	if user.Active == false {
		s = "off"
	}
	setValues(values, "active", s)

	req, err := postRequest("/admin/user/edit", values)
	if err != nil {
		return err
	}
	setHeader(req, "userID", strconv.Itoa(user.ID))
	reqCtx := context.WithValue(req.Context(), middleware.UserContextKey, dummyAdminUser())
	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserEditPostHandler(ctx, rw, req.WithContext(reqCtx))

	if tpl.Err != nil {
		return tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return fmt.Errorf("there is no success message returned")
	}
	return nil
}

func doListUsersRequest() ([]models.User, error) {
	req, err := postRequest("/admin/users", nil)
	if err != nil {
		return nil, err
	}

	reqCtx := context.WithValue(req.Context(), middleware.UserContextKey, dummyAdminUser())
	rw := httptest.NewRecorder()
	tpl := controllers.AdminUsersHandler(ctx, rw, req.WithContext(reqCtx))

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["users"].([]models.User), nil
}

func doCreateUserRequest(user *models.User) (int, error) {
	values := url.Values{}
	setValues(values, "firstname", user.Firstname)
	setValues(values, "lastname", user.Lastname)
	setValues(values, "username", user.Username)
	setValues(values, "email", user.Email)
	setValues(values, "password", string(user.Password))

	req, err := postRequest("/admin/user/new", values)
	if err != nil {
		return 0, err
	}

	reqCtx := context.WithValue(req.Context(), middleware.UserContextKey, dummyAdminUser())
	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserNewPostHandler(ctx, rw, req.WithContext(reqCtx))

	if tpl.Err != nil {
		return 0, tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return -1, fmt.Errorf("there is no success message returned")
	}

	return tpl.Data["userID"].(int), nil
}

type inMemoryUser struct {
	users map[int]*models.User
	mutex sync.RWMutex
}

func (imu *inMemoryUser) Create(u *models.User) (int, error) {
	userID := len(imu.users) + 1
	u.ID = userID
	imu.users[userID] = u
	return userID, nil
}

func (imu *inMemoryUser) List(*models.Pagination) ([]models.User, error) {
	var u []models.User

	for _, v := range imu.users {
		u = append(u, *v)
	}

	return u, nil
}

func (imu *inMemoryUser) Get(userID int) (*models.User, error) {
	if _, ok := imu.users[userID]; ok {
		return imu.users[userID], nil
	}
	return nil, sql.ErrNoRows

}

func (imu *inMemoryUser) Update(u *models.User, changePassword bool) error {
	if _, ok := imu.users[u.ID]; ok {

		imu.users[u.ID] = u
		return nil
	}
	return sql.ErrNoRows
}

func (imu *inMemoryUser) Count(ac models.AdminCriteria) (int, error) {
	//TODO() evaluate admin criteria
	return len(imu.users), nil
}

func (imu *inMemoryUser) Flush() {
	for k := range imu.users {
		delete(imu.users, k)
	}
}

func (imu *inMemoryUser) UpdateLoginDate(userID int) error {
	if k, ok := imu.users[userID]; ok {
		k.LastLogin = models.NullTime{Time: time.Now(), Valid: true}
		return nil
	}
	return sql.ErrNoRows
}

func (imu *inMemoryUser) GetByMail(mail string) (*models.User, error) {
	for _, v := range imu.users {
		if v.Email == mail {
			return v, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (imu *inMemoryUser) GetByUsername(username string) (*models.User, error) {
	for _, v := range imu.users {
		if v.Username == username {
			return v, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (imu *inMemoryUser) Remove(userID int) error {
	delete(imu.users, userID)
	return nil
}
