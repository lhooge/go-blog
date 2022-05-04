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
	"syscall"

	"git.hoogi.eu/snafu/go-blog/database"
	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/models"
	"golang.org/x/crypto/ssh/terminal"
)

type createUserFlag struct {
	username    string
	password    []byte
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

	username := flag.String("username", "", "Username for the admin user. (required)")
	email := flag.String("email", "", "Email for the created user. (required)")
	displayName := flag.String("displayname", "", "Display name for the admin user. (required)")
	isAdmin := flag.Bool("admin", false, "If set a new user with admin permissions will be created; otherwise a non-admin is created.")
	file := flag.String("sqlite", "", "Location to the sqlite3 database file. (required)")

	flag.Parse()

	if *username == "" {
		fmt.Println("the username (-username) must be specified")
		os.Exit(1)
	}
	if *email == "" {
		fmt.Println("the email (-email) must be specified")
		os.Exit(1)
	}
	if *displayName == "" {
		fmt.Println("the display name (-displayname) must be specified")
		os.Exit(1)
	}
	if *file == "" {
		fmt.Println("the argument -sqlite is empty. Please specify the location of the sqlite3 database file")
		os.Exit(1)
	}

	if flag.Parsed() {
		initUser := createUserFlag{
			username:    *username,
			email:       *email,
			displayName: *displayName,
			admin:       *isAdmin,
			sqlite:      *file,
		}

		fmt.Printf("Password: ")
		pw, err := terminal.ReadPassword(int(syscall.Stdin))

		fmt.Println("")

		if err != nil {
			fmt.Printf("could not read password %v\n", err)
			os.Exit(1)
		}

		initUser.password = pw

		if len(initUser.password) == 0 {
			fmt.Println("the password is empty")
			os.Exit(1)
		}

		err = initUser.CreateUser()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("The user '%s' was successfully created\n", *username)
	}
}

func (userFlags createUserFlag) CreateUser() error {
	dbConfig := database.SQLiteConfig{
		File: userFlags.sqlite,
	}

	db, err := dbConfig.Open()

	if err != nil {
		return err
	}

	userService := &models.UserService{
		Datasource: &models.SQLiteUserDatasource{
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
