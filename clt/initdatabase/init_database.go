// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Inititializes the database schema
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"git.hoogi.eu/go-blog/components/database"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/settings"
)

type initDatabaseFlags struct {
	engine       string
	databaseName string
	host         string
	port         int
	user         string
	password     string
	file         string
}

var (
	BuildVersion = "develop"
	GitHash      = ""
)

func main() {
	logger.InitLogger(ioutil.Discard, "Error")

	fmt.Printf("init_database version %s\n", BuildVersion)

	databaseName := flag.String("database", "", "The name of the database")
	host := flag.String("host", "127.0.0.1", "The address of the database")
	port := flag.Int("port", 3306, "The port of the database")
	user := flag.String("user", "root", "The name of the database user")
	password := flag.String("password", "", "The password of the database user")
	file := flag.String("file", "", "The database file to use needed if sqlite is used")
	config := flag.String("config", "", "Config to the blog configuration file. This will get the mysql connection parameters from the config")

	flag.Parse()

	if flag.Parsed() {
		initDB := initDatabaseFlags{
			databaseName: *databaseName,
			host:         *host,
			port:         *port,
			user:         *user,
			password:     *password,
			file:         *file,
		}

		if err := initDB.initSQLite(*config); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func (dbFlags initDatabaseFlags) initSQLite(config string) error {
	fmt.Print(">> Do you want to create the tables now? (y|N): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	lInput := strings.ToLower(input)

	if strings.ToLower(lInput) != "y\n" {
		fmt.Println("Aborted. Tables were not created.")
		os.Exit(0)
	}

	dbConfig := database.SQLiteConfig{
		File: dbFlags.file,
	}

	if len(config) > 0 {
		c, err := settings.LoadConfig(config)

		if err != nil {
			return err
		}

		dbConfig = database.SQLiteConfig{
			File: c.Database.File,
		}
	}

	db, err := dbConfig.Open()

	if err != nil {
		return err
	}

	defer func() {
		db.Close()
	}()

	if _, err := db.Exec("CREATE TABLE user " +
		"(" +
		"username VARCHAR(60) NOT NULL, " +
		"email VARCHAR(191) NOT NULL, " +
		"display_name VARCHAR(191) NOT NULL, " +
		"password CHAR(60) NOT NULL, " +
		"salt CHAR(32) NOT NULL, " +
		"is_admin boolean NOT NULL DEFAULT false, " +
		"active boolean NOT NULL DEFAULT true, " +
		"last_modified datetime NOT NULL," +
		"CONSTRAINT user_email_key UNIQUE (username), " +
		"CONSTRAINT user_email_key UNIQUE (email) " +
		");"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE TABLE user_invite " +
		"(" +
		"hash VARCHAR(191) NOT NULL, " +
		"username VARCHAR(60) NOT NULL, " +
		"email VARCHAR(191) NOT NULL, " +
		"display_name VARCHAR(191) NOT NULL, " +
		"is_admin boolean NOT NULL DEFAULT false, " +
		"active boolean NOT NULL DEFAULT true, " +
		"created_at datetime NOT NULL," +
		"created_by INT NOT NULL, " +
		"FOREIGN KEY (created_by) REFERENCES user(id), " +
		"CONSTRAINT userinvite_hash_key UNIQUE (hash), " +
		"CONSTRAINT userinvite_username_key UNIQUE (username), " +
		"CONSTRAINT userinvite_email_key UNIQUE (email) " +
		");"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE TABLE article " +
		"(" +
		"headline VARCHAR(100) NOT NULL, " +
		"slug VARCHAR(191) NOT NULL, " +
		"teaser text NOT NULL, " +
		"content text NOT NULL, " +
		"published boolean NOT NULL DEFAULT false, " +
		"published_on datetime, " +
		"last_modified datetime NOT NULL, " +
		"user_id INT NOT NULL, " +
		"CONSTRAINT blog_slug_key UNIQUE (slug), " +
		"CONSTRAINT `fk_article_user` " +
		"FOREIGN KEY (user_id) REFERENCES user(id) " +
		"ON DELETE CASCADE " +
		");"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE TABLE site " +
		"(" +
		"title VARCHAR(100) NOT NULL, " +
		"link VARCHAR(100) NOT NULL, " +
		"content text NOT NULL, " +
		"published boolean NOT NULL DEFAULT false, " +
		"published_on datetime, " +
		"last_modified datetime NOT NULL, " +
		"order_no INT NOT NULL, " +
		"user_id INT NOT NULL, " +
		"CONSTRAINT site_link_key UNIQUE (link), " +
		"FOREIGN KEY (user_id) REFERENCES user(id) " +
		"ON DELETE CASCADE " +
		");"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE TABLE file " +
		"(" +
		"filename VARCHAR(191) NOT NULL, " +
		"size BIGINT NOT NULL, " +
		"content_type VARCHAR(150) NOT NULL, " +
		"last_modified datetime NOT NULL, " +
		"user_id INT NOT NULL, " +
		"CONSTRAINT `fk_file_user` " +
		"FOREIGN KEY (user_id) REFERENCES user(id) " +
		"ON DELETE CASCADE, " +
		"CONSTRAINT file_filename_key UNIQUE (filename) " +
		");"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE TABLE token " +
		"(" +
		"hash VARCHAR(191) NOT NULL, " +
		"requested_at datetime NOT NULL, " +
		"token_type VARCHAR(100) NOT NULL, " +
		"user_id INT NOT NULL, " +
		"CONSTRAINT `fk_token_user` " +
		"FOREIGN KEY (user_id) REFERENCES user(id) " +
		"ON DELETE CASCADE, " +
		"CONSTRAINT token_pkey PRIMARY KEY (hash) " +
		");"); err != nil {
		return err
	}

	fmt.Println("The tables were created successfully")

	return nil
}
