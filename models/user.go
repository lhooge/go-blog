// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/go-blog/utils"
	"golang.org/x/crypto/bcrypt"
)

//UserDatasourceService defines an interface for CRUD operations for users
type UserDatasourceService interface {
	Create(u *User) (int, error)
	List(p *Pagination) ([]User, error)
	Get(userID int) (*User, error)
	Update(u *User, changePassword bool) error
	Count(ac AdminCriteria) (int, error)
	GetByMail(mail string) (*User, error)
	GetByUsername(username string) (*User, error)
	Remove(userID int) error
}

//User represents a user
type User struct {
	ID           int
	Username     string
	Email        string
	DisplayName  string
	Password     []byte
	Salt         []byte
	LastModified time.Time
	Active       bool
	IsAdmin      bool
}

const (
	bcryptRounds = 12
)

//UserService containing the service to access users
type UserService struct {
	Datasource      UserDatasourceService
	Config          settings.User
	UserInterceptor UserInterceptor
}

//UserInterceptor will be executed before and after updating/creating users
//build your own interceptor as plugin
type UserInterceptor interface {
	PreCreate(user *User) error
	PostCreate(user *User) error
	PreUpdate(user *User) error
	PostUpdate(user *User) error
}

func (u *User) validate(us UserService, minPasswordLength int, changeMail, changeUserName, changePassword bool) error {
	u.DisplayName = strings.TrimSpace(u.DisplayName)
	u.Email = strings.TrimSpace(u.Email)
	u.Username = strings.TrimSpace(u.Username)

	if len(u.DisplayName) == 0 {
		return httperror.ValueRequired("display name")
	}

	if len([]rune(u.DisplayName)) > 191 {
		return httperror.ValueTooLong("display name", 191)
	}

	if len(u.Email) == 0 {
		return httperror.ValueRequired("email")
	}

	if len(u.Email) > 191 {
		return httperror.ValueTooLong("email", 191)
	}

	if len(u.Username) == 0 {
		return httperror.ValueRequired("username")
	}

	if len([]rune(u.Username)) > 60 {
		return httperror.ValueTooLong("username", 60)
	}

	if changePassword {
		if len(u.Password) < minPasswordLength && len(u.Password) >= 0 {
			return httperror.New(http.StatusUnprocessableEntity,
				fmt.Sprintf("The password is too short. It must be at least %d characters long.", minPasswordLength),
				fmt.Errorf("the password is too short, it must be at least %d characters long", minPasswordLength),
			)
		}
	}

	if changeMail {
		err := us.DuplicateMail(u.Email)

		if err != nil {
			return err
		}
	}

	if changeUserName {
		err := us.DuplicateUsername(u.Email)

		if err != nil {
			return err
		}
	}

	return nil
}

func (us UserService) DuplicateMail(mail string) error {
	user, err := us.Datasource.GetByMail(mail)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if user != nil {
		return httperror.New(http.StatusUnprocessableEntity,
			fmt.Sprintf("The mail %s already exists.", mail),
			fmt.Errorf("the mail %s already exits", mail))
	}

	return nil
}

func (us UserService) DuplicateUsername(username string) error {
	user, err := us.Datasource.GetByUsername(username)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}
	if user != nil {
		return httperror.New(http.StatusUnprocessableEntity,
			fmt.Sprintf("The username %s already exists.", username),
			fmt.Errorf("the username %s already exists", username))
	}

	return nil
}

//CountUsers returns the amount of users
func (us UserService) CountUsers(a AdminCriteria) (int, error) {
	return us.Datasource.Count(a)
}

//ListUsers returns a list of users. Limits the amount based on the defined pagination
func (us UserService) ListUsers(p *Pagination) ([]User, error) {
	return us.Datasource.List(p)
}

//GetUserByID gets the user based on the given id; will not contain the user password
func (us UserService) GetUserByID(userID int) (*User, error) {
	u, err := us.Datasource.Get(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("user", fmt.Errorf("the user with id %d was not found", userID))
		}
		return nil, err
	}

	return u, nil
}

//GetUserByUsername gets the user based on the given username; will contain the user password
func (us UserService) GetUserByUsername(username string) (*User, error) {
	u, err := us.Datasource.GetByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("user", fmt.Errorf("the user with username %s was not found", username))
		}
		return nil, err
	}

	return u, nil
}

//GetUserByMail gets the user based on the given mail; will contain the user password
func (us UserService) GetUserByMail(mail string) (*User, error) {
	u, err := us.Datasource.GetByMail(mail)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("user", fmt.Errorf("the user with mail %s was not found", mail))
		}
	}
	return u, nil
}

//CreateUser creates the user
//If an UserInterceptor is available the action PreCreate is executed before creating and PostCreate after creating the user
func (us UserService) CreateUser(u *User) (int, error) {
	if us.UserInterceptor != nil {
		errUserInterceptor := us.UserInterceptor.PreCreate(u)
		return -1, httperror.InternalServerError(fmt.Errorf("error while executing user interceptor 'PreCreate' error %v", errUserInterceptor))
	}

	if err := u.validate(us, us.Config.MinPasswordLength, true, true, true); err != nil {
		return -1, err
	}

	salt := utils.GenerateSalt()
	saltedPassword := utils.AppendBytes(u.Password, salt)
	password, err := utils.CryptPassword([]byte(saltedPassword), bcryptRounds)

	if err != nil {
		return -1, err
	}

	u.Salt = salt
	u.Password = password

	userID, err := us.Datasource.Create(u)

	if err != nil {
		return -1, err
	}

	if us.UserInterceptor != nil {
		errUserInterceptor := us.UserInterceptor.PostCreate(u)
		logger.Log.Errorf("error while executing PostUpdate user interceptor method %v", errUserInterceptor)
	}

	salt = nil
	saltedPassword = nil

	return userID, nil
}

//UpdateUser updates the user
//If an UserInterceptor is available the action PreUpdate is executed before updating and PostUpdate after updating the user
func (us UserService) UpdateUser(u *User, changePassword bool) error {
	oldUser, err := us.Datasource.Get(u.ID)

	if err != nil {
		return err
	}

	if !oldUser.IsAdmin {
		if oldUser.ID != u.ID {
			return httperror.PermissionDenied("update", "user", fmt.Errorf("permission denied user %d is not granted to update user %d", oldUser.ID, u.ID))
		}
	}

	if us.UserInterceptor != nil {
		errUserInterceptor := us.UserInterceptor.PreUpdate(u)
		return httperror.InternalServerError(fmt.Errorf("error while executing user interceptor 'PreUpdate' error %v", errUserInterceptor))
	}

	var changedMail = !(u.Email == oldUser.Email)
	var changedUserName = !(u.Username == oldUser.Username)

	if err = u.validate(us, us.Config.MinPasswordLength, changedMail, changedUserName, changePassword); err != nil {
		return err
	}

	oneAdmin, err := us.OneAdmin()

	if err != nil {
		return err
	}

	if oneAdmin {
		if (oldUser.IsAdmin && !u.IsAdmin) || (oldUser.IsAdmin && !u.Active) {
			return httperror.New(http.StatusUnprocessableEntity,
				"Could not update user, because no administrator would remain",
				fmt.Errorf("could not update user %s action, because no administrator would remain", oldUser.Username))
		}
	}

	if changePassword {
		salt := utils.GenerateSalt()
		saltedPassword := utils.AppendBytes(u.Password, salt)
		password, err := utils.CryptPassword([]byte(saltedPassword), bcryptRounds)

		if err != nil {
			return err
		}

		u.Password = password
		u.Salt = salt
	}

	err = us.Datasource.Update(u, changePassword)

	u.Password = nil

	if us.UserInterceptor != nil {
		errUserInterceptor := us.UserInterceptor.PostUpdate(u)
		logger.Log.Errorf("error while executing PostUpdate user interceptor method %v", errUserInterceptor)
	}

	return err
}

//Authenticate tries to authenticates the user
// if the user was found;; but the password is wrong the found user and an error will be returned
func (us UserService) Authenticate(u *User, loginMethod settings.LoginMethod, password []byte) (*User, error) {
	var err error

	if loginMethod == settings.EMail {
		u, err = us.GetUserByMail(u.Email)
	} else {
		u, err = us.GetUserByUsername(u.Username)
	}

	if err != nil {
		return nil, err
	}

	if err := u.comparePassword(password); err != nil {
		return u, err
	}
	return u, nil
}

//RemoveUser removes the user; returns an error if no dministrator would remain
func (us UserService) RemoveUser(u *User) error {
	oneAdmin, err := us.OneAdmin()

	if err != nil {
		return err
	}

	if oneAdmin {
		if u.IsAdmin {
			return httperror.New(http.StatusUnprocessableEntity,
				"Could not remove administrator. No Administrator would remain.",
				fmt.Errorf("could not remove administrator %s no administrator would remain", u.Username))
		}
	}
	return us.Datasource.Remove(u.ID)
}

//OneAdmin returns true if there is only one admin
func (us UserService) OneAdmin() (bool, error) {
	c, err := us.Datasource.Count(OnlyAdmins)

	if err != nil {
		return true, err
	}

	if c == 1 {
		return true, nil
	}

	return false, nil
}

func (u User) comparePassword(password []byte) error {
	return bcrypt.CompareHashAndPassword(u.Password, utils.AppendBytes(password, u.Salt))
}
