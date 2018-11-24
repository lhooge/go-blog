// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/models"
)

func TestUserWorklfow(t *testing.T) {
	expectedUser := &models.User{
		DisplayName: "Homer Simpson",
		Email:       "homer@example.com",
		Username:    "homer",
		Password:    []byte("123456789012"),
		Active:      true,
	}

	userID, err := doAdminCreateUserRequest(rAdminUser, expectedUser)
	if err != nil {
		t.Fatal(err)
	}

	expectedUser = &models.User{
		ID:          userID,
		DisplayName: "Homer12 Simpson",
		Email:       "homer@example.com",
		Username:    "homer",
		Password:    []byte("12345678901234"),
		Active:      true,
	}

	err = doAdminEditUsersRequest(rAdminUser, expectedUser)
	if err != nil {
		t.Fatal(err)
	}

	user, err := doAdminGetUserRequest(rAdminUser, userID)
	if err != nil {
		t.Fatal(err)
	}

	err = checkUser(user, expectedUser)
	if err != nil {
		t.Fatal(err)
	}
}

func checkUser(user, expectedUser *models.User) error {
	if user.DisplayName != expectedUser.DisplayName {
		return fmt.Errorf("got an unexpected displayname. expected: %s, actual: %s", expectedUser.DisplayName, user.DisplayName)
	}
	if user.Email != expectedUser.Email {
		return fmt.Errorf("got an unexpected email. expected: %s, actual: %s", expectedUser.Email, user.Email)
	}
	if user.ID != expectedUser.ID {
		return fmt.Errorf("got an unexpected id. expected: %d, actual: %d", expectedUser.ID, user.ID)
	}
	if user.Active != expectedUser.Active {
		return fmt.Errorf("got an unexpected active. expected: %t, actual: %t", expectedUser.Active, user.Active)
	}
	return nil
}

func doAdminGetUserRequest(user reqUser, userID int) (*models.User, error) {
	r := request{
		url:    "/admin/user/" + strconv.Itoa(userID),
		method: "GET",
		user:   user,
		pathVar: []pathVar{
			pathVar{
				key:   "userID",
				value: strconv.Itoa(userID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserEditHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	if w, ok := tpl.Data["user"].(*models.User); ok {
		return w, nil
	}

	return nil, fmt.Errorf("no user were returned %v", tpl.Err)
}

func doAdminEditUsersRequest(user reqUser, u *models.User) error {
	values := url.Values{}
	addValue(values, "displayname", u.DisplayName)
	addValue(values, "username", u.Username)
	addValue(values, "email", u.Email)

	s := "on"

	if u.Active == false {
		s = "off"
	}
	addValue(values, "active", s)

	r := request{
		url:    "/admin/user/edit" + strconv.Itoa(u.ID),
		method: "POST",
		user:   user,
		values: values,
		pathVar: []pathVar{
			pathVar{
				key:   "userID",
				value: strconv.Itoa(u.ID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserEditPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return fmt.Errorf("there is no success message returned")
	}
	return nil
}

func doAdminListUsersRequest(user reqUser) ([]models.User, error) {
	r := request{
		url:    "/admin/users",
		method: "GET",
		user:   user,
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminUsersHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["users"].([]models.User), nil
}

func doAdminCreateUserRequest(user reqUser, u *models.User) (int, error) {
	values := url.Values{}
	addValue(values, "displayname", u.DisplayName)
	addValue(values, "username", u.Username)
	addValue(values, "email", u.Email)
	addValue(values, "password", string(u.Password))
	addCheckboxValue(values, "active", u.Active)

	r := request{
		url:    "/admin/user/edit",
		method: "POST",
		user:   user,
		values: values,
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserNewPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return 0, tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return -1, fmt.Errorf("there is no success message returned")
	}

	return tpl.Data["userID"].(int), nil
}
