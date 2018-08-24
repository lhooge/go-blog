// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

//AdminProfileHandler returns page for updating the profile of the currently logged-in user
func AdminProfileHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	return &middleware.Template{
		Name: tplAdminProfile,
		Data: map[string]interface{}{
			"user": user,
		},
		Active: "profile",
	}
}

//AdminProfilePostHandler handles the updating of the user profile
func AdminProfilePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	ctxUser, _ := middleware.User(r)

	u := &models.User{
		ID:          ctxUser.ID,
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		DisplayName: r.FormValue("displayname"),
		Active:      true,
		IsAdmin:     ctxUser.IsAdmin,
	}

	if _, err := ctx.UserService.Authenticate(ctxUser, ctx.ConfigService.LoginMethod, []byte(r.PostFormValue("current_password"))); err != nil {
		return &middleware.Template{
			Name:   tplAdminProfile,
			Err:    httperror.New(http.StatusUnauthorized, "Your password is invalid.", err),
			Active: "profile",
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	changePassword := false

	if len(r.FormValue("password")) > 0 {
		changePassword = true
		// Password change
		u.Password = []byte(r.FormValue("password"))

		if !bytes.Equal(u.Password, []byte(r.FormValue("retyped_password"))) {
			return &middleware.Template{
				Name:   tplAdminProfile,
				Active: "profile",
				Err: httperror.New(http.StatusUnprocessableEntity,
					"Please check your retyped password",
					errors.New("the password did not match the retyped one")),
				Data: map[string]interface{}{
					"user": u,
				},
			}
		}
	}

	if err := ctx.UserService.UpdateUser(u, changePassword); err != nil {
		return &middleware.Template{
			Name:   tplAdminProfile,
			Active: "profile",
			Err:    err,
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/user/profile",
		Active:       "profile",
		SuccessMsg:   "Your profile update was successful",
		Data: map[string]interface{}{
			"user": u,
		},
	}
}

//AdminUsersHandler returns an overview of the created users  (admin only action)
func AdminUsersHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	page := getPageParam(r)

	total, err := ctx.UserService.CountUsers(models.All)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Err:    err,
			Active: "users",
		}
	}

	p := &models.Pagination{
		Total:       total,
		Limit:       20,
		CurrentPage: page,
		RelURL:      "admin/users/page",
	}

	users, err := ctx.UserService.ListUsers(p)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Err:    err,
			Active: "users",
		}
	}

	userInvites, err := ctx.UserInviteService.ListUserInvites()
	fmt.Println("user_invites", userInvites)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Err:    err,
			Active: "users",
			Data: map[string]interface{}{
				"users":      users,
				"pagination": p,
			},
		}
	}

	return &middleware.Template{
		Name:   tplAdminUsers,
		Active: "users",
		Data: map[string]interface{}{
			"users":        users,
			"user_invites": userInvites,
			"pagination":   p,
		},
	}
}

//AdminUserNewHandler returns the form for adding new user (admin only action)
func AdminUserNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name:   tplAdminUserNew,
		Active: "users",
	}
}

//AdminUserNewPostHandler handles the creation of new users (admin only action)
func AdminUserNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u := &models.User{
		DisplayName: r.FormValue("displayname"),
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		Password:    []byte(r.FormValue("password")),
		Active:      convertCheckbox(r, "active"),
		IsAdmin:     convertCheckbox(r, "admin"),
	}

	userID, err := ctx.UserService.CreateUser(u)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUserNew,
			Err:    err,
			Active: "users",
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
		SuccessMsg:   "Successfully added user " + u.Email,
		Data: map[string]interface{}{
			"userID": userID,
		},
	}
}

//AdminUserEditHandler returns the form for editing an user (admin only action)
func AdminUserEditHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	userID, err := parseInt(getVar(r, "userID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminUsers,
			Err:  err,
		}
	}

	u, err := ctx.UserService.GetUserByID(userID)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminUsers,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminUserEdit,
		Active: "users",
		Data: map[string]interface{}{
			"user": u,
		},
	}
}

//AdminUserEditPostHandler handles the updating of an user (admin only action)
func AdminUserEditPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	userID, err := parseInt(getVar(r, "userID"))

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/users",
			Active:       "users",
			Err:          err,
		}
	}

	u := &models.User{
		ID:          userID,
		Email:       r.FormValue("email"),
		DisplayName: r.FormValue("displayname"),
		Username:    r.FormValue("username"),
		Password:    []byte(r.FormValue("password")),
		Active:      convertCheckbox(r, "active"),
		IsAdmin:     convertCheckbox(r, "admin"),
	}

	changePassword := false

	if len(u.Password) > 0 {
		changePassword = true
	}

	if err := ctx.UserService.UpdateUser(u, changePassword); err != nil {
		return &middleware.Template{
			Name:   tplAdminUserEdit,
			Err:    err,
			Active: "users",
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
		SuccessMsg:   "Successfully edited user " + u.Email,
	}
}

//AdminUserDeleteHandler returns the form for removing user (admin only action)
func AdminUserDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	userID, err := parseInt(getVar(r, "userID"))

	user, err := ctx.UserService.GetUserByID(userID)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/users",
			Active:       "users",
			Err:          err,
		}
	}

	oneAdmin, err := ctx.UserService.OneAdmin()

	if oneAdmin && user.IsAdmin {
		return &middleware.Template{
			RedirectPath: "admin/users",
			Active:       "users",
			Err: httperror.New(http.StatusUnprocessableEntity,
				"Could not remove administrator. No Administrator would remain.",
				fmt.Errorf("could not remove administrator %s no administrator would remain", user.Username)),
		}
	}

	remove := models.Action{
		ID:          "removeUser",
		ActionURL:   fmt.Sprintf("/admin/user/delete/%d", user.ID),
		BackLinkURL: "/admin/users",
		Description: fmt.Sprintf("Please confirm removing of user %s?", user.Username),
		WarnMsg:     "All articles, sites and files belonging to this user will be deleted!",
		Title:       "Confirm removing of user",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "sites",
		Data: map[string]interface{}{
			"action": remove,
		},
	}
}

//AdminUserDeletePostHandler handles removing of a user (admin only action)
func AdminUserDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	userID, err := parseInt(getVar(r, "userID"))

	user, err := ctx.UserService.GetUserByID(userID)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUserDelete,
			Active: "users",
			Err:    err,
		}
	}

	if err := ctx.UserService.RemoveUser(user); err != nil {
		return &middleware.Template{
			Name:   tplAdminUserDelete,
			Active: "users",
			Err:    err,
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
	}
}

//AdminUserNewHandler returns the form for adding new user (admin only action)
func AdminUserInviteNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name:   tplAdminUserInviteNew,
		Active: "users",
	}
}

//AdminUserNewPostHandler handles the creation of new users (admin only action)
func AdminUserInviteNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	ui := &models.UserInvite{
		DisplayName: r.FormValue("displayname"),
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		IsAdmin:     convertCheckbox(r, "admin"),
		CreatedBy:   user,
	}

	inviteID, err := ctx.UserInviteService.CreateUserInvite(ui)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUserInviteNew,
			Err:    err,
			Active: "users",
			Data: map[string]interface{}{
				"user_invite": ui,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
		SuccessMsg:   "Successfully invited user " + ui.Email,
		Data: map[string]interface{}{
			"userID": inviteID,
		},
	}
}
