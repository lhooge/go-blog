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

	CreatedBy User
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
	Datasource UserInviteDatasourceService
}

func (ui UserInvite) validate() error {
	return nil
}

func (uis UserInviteService) ListUserInvites() ([]UserInvite, error) {
	return uis.Datasource.List()
}

func (uis UserInviteService) CreateUserInvite(ui *UserInvite) (int, error) {
	ui.Hash = utils.RandomHash(32)

	if err := ui.validate(); err != nil {
		return -1, err
	}

	return uis.Datasource.Create(ui)
}
