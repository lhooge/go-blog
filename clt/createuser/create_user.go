// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Provides a small CLT for creating an (administrator) user
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"git.hoogi.eu/go-blog/components/database"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/utils"
)

type createUserFlag struct {
	username    string
	password    string
	email       string
	displayName string
	admin       bool
	sqlite      string
}

var (
	BuildVersion = "develop"
	GitHash      = ""
)

func main() {
	logger.InitLogger(ioutil.Discard, "Error")

	fmt.Printf("create_user version %s\n", BuildVersion)

	username := flag.String("username", "", "Username for the admin user ")
	password := flag.String("password", "", "Password for the admin user ")
	email := flag.String("email", "", "Email for the created user ")
	displayName := flag.String("displayname", "", "Display name for the admin user ")
	isAdmin := flag.Bool("admin", false, "If set a new administrator will be created; otherwise a non-admin is created")
	file := flag.String("sqlite", "", "Location to the sqlite3 database file")

	flag.Parse()

	if flag.Parsed() {
		initUser := createUserFlag{
			username:    *username,
			password:    *password,
			email:       *email,
			displayName: *displayName,
			admin:       *isAdmin,
			sqlite:      *file,
		}

		err := initUser.CreateUser()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("The user '%s' was successfully created\n", *username)
	}
}

func (userFlags createUserFlag) CreateUser() error {
	if utils.TrimmedStringIsEmpty(userFlags.username) {
		return fmt.Errorf("the username (-username) must be specified")
	}
	if utils.TrimmedStringIsEmpty(userFlags.password) {
		return fmt.Errorf("the password (-password) must be specified")
	}
	if utils.TrimmedStringIsEmpty(userFlags.email) {
		return fmt.Errorf("the email (-email) must be specified")
	}
	if utils.TrimmedStringIsEmpty(userFlags.displayName) {
		return fmt.Errorf("the display name (-displayname) must be specified")
	}
	if utils.TrimmedStringIsEmpty(userFlags.sqlite) {
		return fmt.Errorf("the argument -sqlite is empty. Please specify the location of the sqlite3 database file")
	}

	var userService models.UserService

	dbConfig := database.SQLiteConfig{
		File: userFlags.sqlite,
	}

	db, err := dbConfig.Open()

	if err != nil {
		return err
	}

	userService = models.UserService{
		Datasource: models.SQLiteUserDatasource{
			SQLConn: db,
		},
	}

	user := &models.User{
		Username:      userFlags.username,
		DisplayName:   userFlags.displayName,
		Email:         userFlags.email,
		PlainPassword: []byte(userFlags.password),
		IsAdmin:       userFlags.admin,
		Active:        true,
	}

	if _, err := userService.Create(user); err != nil {
		return err
	}
	return nil
}
