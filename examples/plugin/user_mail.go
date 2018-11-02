package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"

	"git.hoogi.eu/go-blog/models"
	_ "github.com/mattn/go-sqlite3"
)

type userMailserverInterceptor struct {
	db *sql.DB
}

func GetUserInterceptor() (ui models.UserInterceptor, err error) {
	db, err := sql.Open("sqlite3", "mailserver.sqlite")

	if err != nil {
		return nil, err
	}

	ui = userMailserverInterceptor{
		db: db,
	}

	return
}

func (um userMailserverInterceptor) PreCreate(user *models.User) error {
	return nil
}

func (um userMailserverInterceptor) PostCreate(user *models.User) error {
	pw, err := hashPassword(user.PlainPassword)

	if err != nil {
		return err
	}

	_, err = um.db.Exec("INSERT INTO user (email, password) VALUES (?, ?)", user.Email, pw)

	return err
}

func (um userMailserverInterceptor) PreUpdate(oldUser *models.User, user *models.User) error {
	return nil
}

func (um userMailserverInterceptor) PostUpdate(oldUser *models.User, user *models.User) error {
	pw, err := hashPassword(user.PlainPassword)

	if err != nil {
		return err
	}

	fmt.Println(oldUser.Email)

	_, err = um.db.Exec("UPDATE user SET email=?, password=? WHERE email = ?", user.Email, pw, oldUser.Email)

	return err
}

func (um userMailserverInterceptor) PreRemove(user *models.User) error {
	return nil
}

func (um userMailserverInterceptor) PostRemove(user *models.User) error {
	fmt.Println(user)
	_, err := um.db.Exec("DELETE FROM user WHERE email = ?", user.Email)
	return err
}

func hashPassword(pass []byte) (string, error) {
	args := []string{"pw", "-s", "argon2id", "-p", string(pass)}

	cmd := exec.Command("doveadm", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}
