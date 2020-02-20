package controllers_test

import (
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"git.hoogi.eu/snafu/go-blog/controllers"
	"git.hoogi.eu/snafu/go-blog/models"
)

func TestUserInviteWorkflow(t *testing.T) {
	setup(t)

	defer teardown()

	ui := &models.UserInvite{
		DisplayName: "Homer Simpson",
		Email:       "homer@example.com",
		Username:    "homer",
		IsAdmin:     false,
	}

	inviteID, hash, err := doAdminCreateUserInviteRequest(rAdminUser, ui)

	if err != nil {
		t.Fatal(err)
	}

	if len(hash) == 0 {
		t.Error("no hash received")
	}

	inviteID, hash, err = doAdminResendUserInviteRequest(rAdminUser, inviteID)

	if err != nil {
		t.Fatal(err)
	}

	if len(hash) == 0 {
		t.Error("no hash received")
	}

	err = doAdminRemoveUserInviteRequest(rAdminUser, inviteID)

	if err != nil {
		t.Fatal(err)
	}

	_, _, err = doAdminResendUserInviteRequest(rAdminUser, inviteID)

	if err == nil {
		t.Errorf("removed user invitation is still there, inviteID %d", inviteID)
	}
}

func doAdminCreateUserInviteRequest(user reqUser, ui *models.UserInvite) (int, string, error) {
	values := url.Values{}
	addValue(values, "displayname", ui.DisplayName)
	addValue(values, "username", ui.Username)
	addValue(values, "email", ui.Email)
	addCheckboxValue(values, "admin", ui.IsAdmin)

	r := request{
		url:    "/admin/user-invite/new",
		user:   user,
		method: "POST",
		values: values,
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserInviteNewPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return -1, "", tpl.Err
	}

	return tpl.Data["inviteID"].(int), tpl.Data["hash"].(string), nil
}

func doAdminResendUserInviteRequest(user reqUser, inviteID int) (int, string, error) {
	r := request{
		url:    "/admin/user-invite/resend/" + strconv.Itoa(inviteID),
		user:   user,
		method: "POST",
		pathVar: []pathVar{
			pathVar{
				key:   "inviteID",
				value: strconv.Itoa(inviteID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserInviteResendPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return -1, "", tpl.Err
	}

	return tpl.Data["inviteID"].(int), tpl.Data["hash"].(string), nil
}

func doAdminRemoveUserInviteRequest(user reqUser, inviteID int) error {
	r := request{
		url:    "/admin/user-invite/delete/" + strconv.Itoa(inviteID),
		user:   user,
		method: "POST",
		pathVar: []pathVar{
			pathVar{
				key:   "inviteID",
				value: strconv.Itoa(inviteID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminUserInviteDeletePostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}
