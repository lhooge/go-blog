package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/controllers"
)

func TestLogin(t *testing.T) {
	resp, err := doLoginRequest(rGuest, "alice", "123456789012")

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
	resp, err := doLoginRequest(rGuest, "alice", "test2")

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

func doLoginRequest(user reqUser, login, password string) (responseWrapper, error) {
	values := url.Values{}
	addValue(values, "username", login)
	addValue(values, "password", password)

	r := request{
		url:    "/admin/login",
		method: "POST",
		user:   user,
		values: values,
	}

	rr := httptest.NewRecorder()
	tpl := controllers.LoginPostHandler(ctx, rr, r.buildRequest())

	if tpl.Err != nil {
		return responseWrapper{response: rr, template: tpl}, tpl.Err
	}

	return responseWrapper{response: rr, template: tpl}, nil
}
