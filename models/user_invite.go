package models

import (
	"time"

	"git.hoogi.eu/snafu/go-blog/crypt"
	"git.hoogi.eu/snafu/go-blog/mail"
)

// TODO: refactor
// UserInvite represents a new invited user
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

// UserInviteDatasourceService defines an interface for CRUD operations for users
type UserInviteDatasourceService interface {
	List() ([]UserInvite, error)
	Get(inviteID int) (*UserInvite, error)
	GetByHash(hash string) (*UserInvite, error)
	Create(ui *UserInvite) (int, error)
	Update(ui *UserInvite) error
	Remove(inviteID int) error
}

// UserInviteService
type UserInviteService struct {
	Datasource  UserInviteDatasourceService
	UserService UserService
	MailService mail.Service
}

// validate A user invitation must conform the user validations except the password checks
func (ui UserInvite) validate(uis UserInviteService) error {
	user := ui.Copy()

	return user.validate(uis.UserService, -1, VDupEmail|VDupUsername)
}

func (uis UserInviteService) List() ([]UserInvite, error) {
	return uis.Datasource.List()
}

func (uis UserInviteService) Update(ui *UserInvite) error {
	ui.Hash = crypt.RandomHash(32)

	if err := ui.validate(uis); err != nil {
		return err
	}

	return uis.Datasource.Update(ui)
}

func (uis UserInviteService) Create(ui *UserInvite) (int, error) {
	ui.Hash = crypt.RandomHash(32)

	if err := ui.validate(uis); err != nil {
		return -1, err
	}

	return uis.Datasource.Create(ui)
}

func (uis UserInviteService) Get(inviteID int) (*UserInvite, error) {
	return uis.Datasource.Get(inviteID)
}

func (uis UserInviteService) GetByHash(hash string) (*UserInvite, error) {
	return uis.Datasource.GetByHash(hash)
}

func (uis UserInviteService) Remove(inviteID int) error {
	return uis.Datasource.Remove(inviteID)
}
