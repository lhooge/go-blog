// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.hoogi.eu/snafu/go-blog/components/httperror"
	"git.hoogi.eu/snafu/go-blog/components/logger"
	"git.hoogi.eu/snafu/go-blog/settings"
	"git.hoogi.eu/snafu/go-blog/utils"
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
	ID            int
	Username      string
	Email         string
	DisplayName   string
	Password      []byte
	PlainPassword []byte
	Salt          []byte
	LastModified  time.Time
	Active        bool
	IsAdmin       bool
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
type UserInterceptor interface {
	PreCreate(user *User) error
	PostCreate(user *User) error
	PreUpdate(oldUser *User, user *User) error
	PostUpdate(oldUser *User, user *User) error
	PreRemove(user *User) error
	PostRemove(user *User) error
}

type Validations int

const (
	VDupEmail = 1 << iota
	VDupUsername
	VPassword
)

func (u *User) validate(us UserService, minPasswordLength int, v Validations) error {
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

	if (v & VPassword) != 0 {
		if len(u.PlainPassword) < minPasswordLength && len(u.PlainPassword) > 0 {
			return httperror.New(http.StatusUnprocessableEntity,
				fmt.Sprintf("The password is too short. It must be at least %d characters long.", minPasswordLength),
				fmt.Errorf("the password is too short, it must be at least %d characters long", minPasswordLength),
			)
		}
	}

	if (v & VDupEmail) != 0 {
		err := us.duplicateMail(u.Email)

		if err != nil {
			return err
		}
	}

	if (v & VDupUsername) != 0 {
		err := us.duplicateUsername(u.Username)

		if err != nil {
			return err
		}
	}

	return nil
}

func (us UserService) duplicateMail(mail string) error {
	user, err := us.Datasource.GetByMail(mail)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if user != nil {
		return httperror.New(http.StatusUnprocessableEntity, fmt.Sprintf("The mail %s already exists.", mail), fmt.Errorf("the mail %s already exits", mail))
	}

	return nil
}

func (us UserService) duplicateUsername(username string) error {
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

//Count returns the amount of users
func (us UserService) Count(a AdminCriteria) (int, error) {
	return us.Datasource.Count(a)
}

//List returns a list of users. Limits the amount based on the defined pagination
func (us UserService) List(p *Pagination) ([]User, error) {
	return us.Datasource.List(p)
}

//GetByID gets the user based on the given id; will not contain the user password
func (us UserService) GetByID(userID int) (*User, error) {
	u, err := us.Datasource.Get(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("user", fmt.Errorf("the user with id %d was not found", userID))
		}
		return nil, err
	}

	return u, nil
}

//GetByUsername gets the user based on the given username; will contain the user password
func (us UserService) GetByUsername(username string) (*User, error) {
	u, err := us.Datasource.GetByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("user", err)
		}
		return nil, err
	}

	return u, nil
}

//GetByMail gets the user based on the given mail; will contain the user password
func (us UserService) GetByMail(mail string) (*User, error) {
	u, err := us.Datasource.GetByMail(mail)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("user", err)
		}

		return nil, err
	}
	return u, nil
}

//Create creates the user
//If an UserInterceptor is available the action PreCreate is executed before creating and PostCreate after creating the user
func (us UserService) Create(u *User) (int, error) {
	if us.UserInterceptor != nil {
		if err := us.UserInterceptor.PreCreate(u); err != nil {
			return -1, httperror.InternalServerError(fmt.Errorf("error while executing user interceptor 'PreCreate' error %v", err))
		}
	}

	if err := u.validate(us, us.Config.MinPasswordLength, VDupUsername|VDupEmail|VPassword); err != nil {
		return -1, err
	}

	salt := utils.GenerateSalt()
	saltedPassword := utils.AppendBytes(u.PlainPassword, salt)
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
		logger.Log.Errorf("error while executing PostCreate user interceptor method %v", errUserInterceptor)
	}

	salt = nil
	saltedPassword = nil
	u.PlainPassword = nil

	return userID, nil
}

//Update updates the user
//If an UserInterceptor is available the action PreUpdate is executed before updating and PostUpdate after updating the user
func (us UserService) Update(u *User, changePassword bool) error {
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

		if err := us.UserInterceptor.PreUpdate(oldUser, u); err != nil {
			return httperror.InternalServerError(fmt.Errorf("error while executing user interceptor 'PreUpdate' error %v", err))
		}
	}

	var v Validations

	if u.Email != oldUser.Email {
		v |= VDupEmail
	}

	if u.Username != oldUser.Username {
		v |= VDupUsername
	}

	if changePassword {
		v |= VPassword
	}

	if err = u.validate(us, us.Config.MinPasswordLength, v); err != nil {
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
		saltedPassword := utils.AppendBytes(u.PlainPassword, salt)
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
		if err := us.UserInterceptor.PostUpdate(oldUser, u); err != nil {
			logger.Log.Errorf("error while executing PostUpdate user interceptor method %v", err)
		}
	}
	u.PlainPassword = nil

	return err
}

// Authenticate authenticates the user by the given login method (email or username)
// if the user was found but the password is wrong the found user and an error will be returned
func (us UserService) Authenticate(u *User, loginMethod settings.LoginMethod) (*User, error) {
	var err error

	if len(u.Username) == 0 || len(u.PlainPassword) == 0 {
		return nil, httperror.New(http.StatusUnauthorized, "Your username or password is invalid.", errors.New("no username or password were given"))
	}

	var password = u.PlainPassword

	if loginMethod == settings.EMail {
		u, err = us.Datasource.GetByMail(u.Email)
	} else {
		u, err = us.Datasource.GetByUsername(u.Username)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			//Do some extra work
			bcrypt.CompareHashAndPassword([]byte("$2a$12$bQlRnXTNZMp6kCyoAlnf3uZW5vtmSj9CHP7pYplRUVK2n0C5xBHBa"), password)
			return nil, httperror.New(http.StatusUnauthorized, "Your username or password is invalid.", err)
		}
		return nil, err
	}

	u.PlainPassword = password

	if err := u.comparePassword(); err != nil {
		return u, httperror.New(http.StatusUnauthorized, "Your username or password is invalid.", err)
	}

	if !u.Active {
		return nil, httperror.New(http.StatusUnprocessableEntity,
			"Your account is deactivated.",
			fmt.Errorf("the user with id %d tried to logged in but the account is deactivated", u.ID))
	}

	u.PlainPassword = nil
	u.Password = nil
	u.Salt = nil

	return u, nil
}

// Remove removes the user returns an error if no administrator would remain
func (us UserService) Remove(u *User) error {
	if us.UserInterceptor != nil {
		if err := us.UserInterceptor.PreRemove(u); err != nil {
			return httperror.InternalServerError(fmt.Errorf("error while executing user interceptor 'PreRemove' error %v", err))
		}
	}

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

	err = us.Datasource.Remove(u.ID)

	if us.UserInterceptor != nil {
		if err := us.UserInterceptor.PostRemove(u); err != nil {
			logger.Log.Errorf("error while executing PostRemove user interceptor method %v", err)
		}
	}

	return err
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

func (u User) comparePassword() error {
	return bcrypt.CompareHashAndPassword(u.Password, utils.AppendBytes(u.PlainPassword, u.Salt))
}
