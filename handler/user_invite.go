package handler

import (
	"fmt"
	"net/http"

	"git.hoogi.eu/snafu/go-blog/middleware"
	"git.hoogi.eu/snafu/go-blog/models"
)

// AdminUserInviteNewHandler shows the form to invite an user
func AdminUserInviteNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name:   tplAdminUserInviteNew,
		Active: "users",
	}
}

// AdminUserInviteNewHandler handles the invitation, sends an activation mail to the invited user
func AdminUserInviteNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	ui := &models.UserInvite{
		DisplayName: r.FormValue("displayname"),
		Username:    r.FormValue("username"),
		Email:       r.FormValue("email"),
		IsAdmin:     convertCheckbox(r, "admin"),
		CreatedBy:   user,
	}

	inviteID, err := ctx.UserInviteService.Create(ui)

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

	ctx.Mailer.SendActivationLink(ui)

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
		SuccessMsg:   fmt.Sprintf("%s %s. ", "Successfully invited user ", ui.Email),
		Data: map[string]interface{}{
			"inviteID": inviteID,
			"hash":     ui.Hash,
		},
	}
}

// AdminUserInviteResendPostHandler resends the activation link to the user
func AdminUserInviteResendPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	inviteID, err := parseInt(getVar(r, "inviteID"))

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/user-invite",
			Active:       "users",
			Err:          err,
		}
	}

	ui, err := ctx.UserInviteService.Get(inviteID)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Err:    err,
			Active: "users",
		}
	}

	ctx.Mailer.SendActivationLink(ui)

	return &middleware.Template{
		RedirectPath: "admin/users",
		Active:       "users",
		SuccessMsg:   fmt.Sprintf("%s %s. ", "Successfully invited user ", ui.Email),
		Data: map[string]interface{}{
			"inviteID": inviteID,
			"hash":     ui.Hash,
		},
	}
}

// AdminUserInviteDeleteHandler shows the form to remove an user invitation
func AdminUserInviteDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	inviteID, err := parseInt(getVar(r, "inviteID"))

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/user-invite",
			Active:       "users",
			Err:          err,
		}
	}

	invite, err := ctx.UserInviteService.Get(inviteID)

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

// AdminUserInviteDeletePostHandler handles the removing of an user invitation
func AdminUserInviteDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	inviteID, err := parseInt(getVar(r, "inviteID"))

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminUsers,
			Active: "users",
			Err:    err,
		}
	}

	if err := ctx.UserInviteService.Remove(inviteID); err != nil {
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
