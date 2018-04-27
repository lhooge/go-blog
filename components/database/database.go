// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

//MySQLConfig represents mysql configuration type
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

//SQLiteConfig represents sqlite configuration type
type SQLiteConfig struct {
	File string
}

//Open receives handle for sqlite database; validates if opening the connection
func (d SQLiteConfig) Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", d.File)

	if err != nil {
		return nil, err
	}

	return db, nil
}

//Open receives handle for mysql database; validates if opening the connection
func (d MySQLConfig) Open() (*sql.DB, error) {
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4", d.User, d.Password, d.Host, d.Port, d.Database)
	db, err := sql.Open("mysql", url)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
