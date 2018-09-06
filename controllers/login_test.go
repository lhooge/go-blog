package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/models"
)

func TestLogin(t *testing.T) {
	defer ctx.UserService.Datasource.(*inMemoryUser).Flush()

	expectedUser := &models.User{
		DisplayName: "Homer",
		Email:       "homer@example.com",
		Username:    "homer",
		Password:    []byte("1234567890"),
		Active:      true,
		IsAdmin:     false,
	}

	_, err := doCreateUserRequest(expectedUser)

	if err != nil {
		t.Fatal(err)
	}

	resp, err := doLoginRequest()

	if err != nil {
		t.Fatal(err)
	}

	if resp.getTemplateError() != nil {
		t.Fatalf("an error is returned %v", resp.getTemplateError().Error())
	}

	if !resp.isCodeSuccess() {
		t.Fatalf("got an invalid http response code %d", resp.getStatus())
	}

	c, err := resp.getCookie("test-session")

	if err != nil {
		t.Fatal(err)
	}

	if c.HttpOnly == false {
		t.Error("cookie with session id is missing http only flag")
	}
	if c.Secure == false {
		t.Error("cookie with session id is missing secure flag")
	}
}

func TestFailLogin(t *testing.T) {
	defer ctx.UserService.Datasource.(*inMemoryUser).Flush()

	expectedUser := &models.User{
		DisplayName: "Homer",
		Email:       "homer@example.com",
		Username:    "homer",
		Password:    []byte("12345678123"),
		Active:      true,
		IsAdmin:     false,
	}

	_, err := doCreateUserRequest(expectedUser)

	if err != nil {
		t.Fatal(err)
	}

	resp, err := doLoginRequest()

	if err == nil {
		t.Fatalf("Expected an error when credentials are wrong. But error is nil %v", resp.template)
	}

	if resp.getTemplateError().(*httperror.Error).HTTPStatus != http.StatusUnauthorized {
		t.Errorf("Got an invalid status code. Should be %d, but was %d", http.StatusUnauthorized, resp.getStatus())
	}

	_, err = resp.getCookie("test-session")

	if err == nil {
		t.Fatal("the cookie test-session should not be set but is available")
	}
}

func doLoginRequest() (responseWrapper, error) {
	values := url.Values{}
	addValue(values, "username", "homer")
	addValue(values, "password", "1234567890")

	req, err := postRequest("/admin/login", values)
	if err != nil {
		return responseWrapper{}, err
	}

	rr := httptest.NewRecorder()
	tpl := controllers.LoginPostHandler(ctx, rr, req)

	if tpl.Err != nil {
		return responseWrapper{response: rr, template: tpl}, tpl.Err
	}

	return responseWrapper{response: rr, template: tpl}, nil
}
