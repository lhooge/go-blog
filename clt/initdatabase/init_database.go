// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

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
)

var (
	BuildVersion = "develop"
	GitHash      = ""
)

func main() {
	logger.InitLogger(ioutil.Discard, "Error")

	fmt.Printf("init_database version %s\n", BuildVersion)

	file := flag.String("sqlite", "", "Location for the sqlite3 database")

	flag.Parse()

	if flag.Parsed() {
		if err := initSQLite(*file); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("The tables were created")
	}
}

func initSQLite(sqlitefile string) error {
	if len(sqlitefile) == 0 {
		return fmt.Errorf("the argument -sqlite is empty. Please specify the location of the sqlite3 database file")
	}

	fmt.Print(">> Do you want to create the tables now? (y|N): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	lInput := strings.ToLower(input)

	if strings.ToLower(lInput) != "y\n" {
		fmt.Println("Aborted. Tables were not created.")
		os.Exit(0)
	}

	dbConfig := database.SQLiteConfig{
		File: sqlitefile,
	}

	db, err := dbConfig.Open()

	if err != nil {
		return err
	}

	defer func() {
		db.Close()
	}()

	return database.InitTables(db)
}
