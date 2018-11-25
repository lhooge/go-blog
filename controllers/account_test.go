package controllers_test

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/models"
)

func TestProfileUpdate(t *testing.T) {
	setup(t)

	defer teardown()

	user := &models.User{
		DisplayName:   "Homer Simpson",
		Email:         "homer@example.com",
		Username:      "homer",
		PlainPassword: []byte("123456789012"),
		Active:        true,
		IsAdmin:       false,
	}

	userID, err := doAdminCreateUserRequest(rAdminUser, user)
	if err != nil {
		t.Fatal(err)
	}

	user.Username = "marge"
	user.PlainPassword = []byte("2109876543210")
	user.DisplayName = "Marge Simpson"
	user.IsAdmin = true
	user.Email = "marge@example.com"

	err = doAdminProfileRequest(reqUser(userID), user, "123456789012")
	if err != nil {
		t.Fatal(err)
	}

	err = login(user.Username, string(user.PlainPassword))
	if err != nil {
		t.Fatal(err)
	}
}

func doAdminProfileRequest(user reqUser, u *models.User, currentPassword string) error {
	values := url.Values{}
	addValue(values, "username", u.Username)
	addValue(values, "email", u.Email)
	addValue(values, "displayname", u.DisplayName)
	addValue(values, "password", string(u.PlainPassword))
	addValue(values, "retyped_password", string(u.PlainPassword))
	addValue(values, "current_password", string(currentPassword))
	r := request{
		url:    "/admin/profile",
		user:   user,
		method: "POST",
		values: values,
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminProfilePostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}

func doActivateAccountRequest(user reqUser, password, passwordRepeat, hash string) error {
	values := url.Values{}
	addValue(values, "password", password)
	addValue(values, "password_repeat", passwordRepeat)

	r := request{
		url:    "/admin/activate-account/" + hash,
		user:   user,
		method: "POST",
		values: values,
		pathVar: []pathVar{
			pathVar{
				key:   "hash",
				value: hash,
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSiteEditPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}

func doResetPasswordRequest(user reqUser, password, passwordRepeat, hash string) error {
	values := url.Values{}
	addValue(values, "password", password)
	addValue(values, "password_repeat", passwordRepeat)

	r := request{
		url:    "/admin/reset-password/" + hash,
		user:   user,
		method: "POST",
		values: values,
		pathVar: []pathVar{
			pathVar{
				key:   "hash",
				value: hash,
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSiteEditPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}
