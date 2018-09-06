package controllers

import (
	"fmt"
	"net/http"

	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

func AdminUserInviteNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name:   tplAdminUserInviteNew,
		Active: "users",
	}
}

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

	err = ctx.Mailer.SendActivationLink(ui)

	if err != nil {
		logger.Log.Error(err)
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

func AdminUserInviteDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	inviteID, err := parseInt(getVar(r, "inviteID"))

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/user-invite",
			Active:       "users",
			Err:          err,
		}
	}

	invite, err := ctx.UserInviteService.GetInvite(inviteID)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/user-invite",
			Active:       "users",
			Err:          err,
		}
	}

	remove := models.Action{
		ID:          "removeUserInvite",
		ActionURL:   fmt.Sprintf("/admin/user-invite/delete/%d", invite.ID),
		BackLinkURL: "/admin/users",
		Description: fmt.Sprintf("Please confirm removing of user invitation %s?", invite.Username),
		Title:       "Confirm removing of user invitation",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "users",
		Data: map[string]interface{}{
			"action": remove,
		},
	}
}

func AdminUserInviteDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	inviteID, err := parseInt(getVar(r, "inviteID"))

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUserDelete,
			Active: "users",
			Err:    err,
		}
	}

	if err := ctx.UserInviteService.RemoveInvite(inviteID); err != nil {
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
