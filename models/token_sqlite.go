package models

import (
	"database/sql"
	"time"
)

//SQLiteTokenDatasource providing an implementation of TokenDatasourceService using MariaDB
type SQLiteTokenDatasource struct {
	SQLConn *sql.DB
}

//Create creates a new token
func (rdb SQLiteTokenDatasource) Create(t *Token) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO token (hash, requested_at, token_type, user_id) VALUES(?, ?, ?, ?)",
		t.Hash, time.Now(), t.Type, t.Author.ID)

	if err != nil {
		return -1, err
	}

	i, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return int(i), nil
}

//Get gets a token based on the hash and the token type
func (rdb SQLiteTokenDatasource) Get(hash string, tt TokenType) (*Token, error) {
	var t Token
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT t.hash, t.requested_at, t.token_type, t.user_id FROM token as t WHERE hash=? AND token_type=? ", hash, tt.String()).
		Scan(&t.Hash, &t.RequestedAt, &t.Type, &u.ID); err != nil {
		return nil, err
	}

	t.Author = &u

	return &t, nil
}

//Remove removes a token based on the hash
func (rdb SQLiteTokenDatasource) Remove(hash string, tt TokenType) error {
	if _, err := rdb.SQLConn.Exec("DELETE FROM token WHERE hash=? AND token_type=? ", hash, tt.String()); err != nil {
		return err
	}
	return nil
}
