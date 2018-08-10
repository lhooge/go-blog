// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

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
