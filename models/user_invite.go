package models

import (
	"time"

	"git.hoogi.eu/go-blog/utils"
)

//User represents a user
type UserInvite struct {
	ID          int
	Hash        string
	Username    string
	Email       string
	DisplayName string
	CreatedAt   time.Time
	IsAdmin     bool

	CreatedBy *User
}

func (ui UserInvite) Copy() *User {
	return &User{
		Username:    ui.Username,
		Email:       ui.Email,
		DisplayName: ui.DisplayName,
		IsAdmin:     ui.IsAdmin,
	}
}

//UserInviteDatasourceService defines an interface for CRUD operations for users
type UserInviteDatasourceService interface {
	List() ([]UserInvite, error)
	Get(inviteID int) (*UserInvite, error)
	GetByHash(hash string) (*UserInvite, error)
	Create(ui *UserInvite) (int, error)
	Remove(inviteID int) error
}

//UserInviteService
type UserInviteService struct {
	Datasource  UserInviteDatasourceService
	UserService UserService
}

// validate A user invitation must conform the user validations except the password checks
func (ui UserInvite) validate(uis UserInviteService) error {
	user := ui.Copy()

	err := user.validate(uis.UserService, -1, VDupEmail|VDupUsername)

	if err != nil {
		return err
	}

	return nil
}

func (uis UserInviteService) ListUserInvites() ([]UserInvite, error) {
	return uis.Datasource.List()
}

func (uis UserInviteService) CreateUserInvite(ui *UserInvite) (int, error) {
	ui.Hash = utils.RandomHash(32)

	if err := ui.validate(uis); err != nil {
		return -1, err
	}

	return uis.Datasource.Create(ui)
}

func (uis UserInviteService) GetInvite(inviteID int) (*UserInvite, error) {
	return uis.Datasource.Get(inviteID)
}

func (uis UserInviteService) GetByHash(hash string) (*UserInvite, error) {
	return uis.Datasource.GetByHash(hash)
}

func (uis UserInviteService) RemoveInvite(inviteID int) error {
	return uis.Datasource.Remove(inviteID)
}
