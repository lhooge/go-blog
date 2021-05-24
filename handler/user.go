// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"net/http"

	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/middleware"
	"git.hoogi.eu/snafu/go-blog/models"
)

// AdminUsersHandler returns an overview of the created users
func AdminUsersHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	page := getPageParam(r)

	total, err := ctx.UserService.Count(models.All)

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

	users, err := ctx.UserService.List(p)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Err:    err,
			Active: "users",
		}
	}

	var userInvites []models.UserInvite

	if cu, _ := middleware.User(r); cu.IsAdmin {
		userInvites, err = ctx.UserInviteService.List()

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

// AdminUserNewHandler returns the form for adding new user
func AdminUserNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name:   tplAdminUserNew,
		Active: "users",
	}
}

// AdminUserNewPostHandler handles the creation of new users
func AdminUserNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u := &models.User{
		DisplayName:   r.FormValue("displayname"),
		Username:      r.FormValue("username"),
		Email:         r.FormValue("email"),
		PlainPassword: []byte(r.FormValue("password")),
		Active:        convertCheckbox(r, "active"),
		IsAdmin:       convertCheckbox(r, "admin"),
	}

	userID, err := ctx.UserService.Create(u)
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

// AdminUserEditHandler returns the form for editing an user
func AdminUserEditHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	userID, err := parseInt(getVar(r, "userID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminUsers,
			Err:  err,
		}
	}

	u, err := ctx.UserService.GetByID(userID)

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

// AdminUserEditPostHandler handles the updating of an user
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
		ID:            userID,
		Email:         r.FormValue("email"),
		DisplayName:   r.FormValue("displayname"),
		Username:      r.FormValue("username"),
		PlainPassword: []byte(r.FormValue("password")),
		Active:        convertCheckbox(r, "active"),
		IsAdmin:       convertCheckbox(r, "admin"),
	}

	changePassword := false

	if len(u.PlainPassword) > 0 {
		changePassword = true
	}

	if err := ctx.UserService.Update(u, changePassword); err != nil {
		return &middleware.Template{
			Name:   tplAdminUserEdit,
			Err:    err,
			Active: "users",
			Data: map[string]interface{}{
				"user": u,
			},
		}
	}

	if changePassword {
		session, err := ctx.SessionService.Get(w, r)

		if err != nil {
			return &middleware.Template{
				Name:   tplAdminUserEdit,
				Err:    err,
				Active: "users",
				Data: map[string]interface{}{
					"user": u,
				},
			}
		}

		sessions := ctx.SessionService.SessionProvider.FindByValue("userid", u.ID)

		for _, s := range sessions {
			if session.SessionID() != s.SessionID() {
				ctx.SessionService.SessionProvider.Remove(s.SessionID())
			}
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
		SuccessMsg:   "Successfully edited user " + u.Email,
	}
}

// AdminUserDeleteHandler returns the form for removing a user
func AdminUserDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	userID, err := parseInt(getVar(r, "userID"))

	user, err := ctx.UserService.GetByID(userID)

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
				"Could not remove administrator. No administrator would remain.",
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
		Active: "users",
		Data: map[string]interface{}{
			"action": remove,
		},
	}
}

// AdminUserDeletePostHandler handles removing of a user
func AdminUserDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	userID, err := parseInt(getVar(r, "userID"))

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Active: "users",
			Err:    err,
		}
	}

	user, err := ctx.UserService.GetByID(userID)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Active: "users",
			Err:    err,
		}
	}

	if err := ctx.UserService.Remove(user); err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Active: "users",
			Err:    err,
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
	}
}
