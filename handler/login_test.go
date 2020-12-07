package handler_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"git.hoogi.eu/snafu/go-blog/handler"
)

func TestLogin(t *testing.T) {
	setup(t)

	defer teardown()

	err := login("alice", "123456789012")

	if err != nil {
		t.Error(err)
	}

	err = login("alice", "test2")

	if err == nil {
		t.Error("expected a failed login, but error is nil")
	}
}

func login(username, password string) error {
	resp, err := doLoginRequest(rGuest, username, password)

	if err != nil {
		return err
	}

	if resp.getTemplateError() != nil {
		return fmt.Errorf("an error is returned %v", resp.getTemplateError().Error())
	}

	if !resp.isCodeSuccess() {
		return fmt.Errorf("got an invalid http response code %d", resp.getStatus())
	}

	c, err := resp.getCookie("test-session")

	if err != nil {
		return err
	}

	if c.HttpOnly == false {
		return fmt.Errorf("cookie with session id is missing http only flag")
	}
	if c.Secure == false {
		return fmt.Errorf("cookie with session id is missing secure flag")
	}

	return nil
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
	tpl := handler.LoginPostHandler(ctx, rr, r.buildRequest())

	if tpl.Err != nil {
		return responseWrapper{response: rr, template: tpl}, tpl.Err
	}

	return responseWrapper{response: rr, template: tpl}, nil
}
