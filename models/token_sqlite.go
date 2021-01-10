package models

import (
	"database/sql"
	"time"
)

// SQLiteTokenDatasource providing an implementation of TokenDatasourceService using MariaDB
type SQLiteTokenDatasource struct {
	SQLConn *sql.DB
}

// Create creates a new token
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

// Get gets a token based on the hash and the token type
func (rdb SQLiteTokenDatasource) Get(hash string, tt TokenType) (*Token, error) {
	var t Token
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT t.id, t.hash, t.requested_at, t.token_type, t.user_id FROM token as t WHERE t.hash=? AND t.token_type=? ", hash, tt.String()).
		Scan(&t.ID, &t.Hash, &t.RequestedAt, &t.Type, &u.ID); err != nil {
		return nil, err
	}

	t.Author = &u

	return &t, nil
}

// ListByUser receives all tokens based on the user id and the token type ordered by requested
func (rdb SQLiteTokenDatasource) ListByUser(userID int, tt TokenType) ([]Token, error) {
	rows, err := rdb.SQLConn.Query("SELECT t.id, t.hash, t.requested_at, t.token_type, t.user_id FROM token as t WHERE t.user_id=? AND t.token_type=? ", userID, tt.String())

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tokens := []Token{}

	for rows.Next() {
		var u User
		var t Token
		if err = rows.Scan(&t.ID, &t.Hash, &t.RequestedAt, &t.Type, &u.ID); err != nil {
			return nil, err
		}

		t.Author = &u

		tokens = append(tokens, t)
	}

	return tokens, nil
}

// Remove removes a token based on the hash
func (rdb SQLiteTokenDatasource) Remove(hash string, tt TokenType) error {
	if _, err := rdb.SQLConn.Exec("DELETE FROM token WHERE hash=? AND token_type=? ", hash, tt.String()); err != nil {
		return err
	}
	return nil
}
