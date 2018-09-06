package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/utils"
)

//TokenDatasourceService defines an interface for CRUD operations for tokens
type TokenDatasourceService interface {
	Create(t *Token) (int, error)
	Get(hash string, tt TokenType) (*Token, error)
	Remove(hash string, tt TokenType) error
}

//Token represents a token
type Token struct {
	Hash        string
	Type        TokenType
	RequestedAt time.Time

	Author *User
}

const (
	//PasswordReset token generated for resetting passwords
	PasswordReset = iota
)

var types = [...]string{"password_reset"}

//TokenType specifies the type where token can be used
type TokenType int

// Scan implements the Scanner interface.
func (tt *TokenType) Scan(value interface{}) error {
	for k, t := range types {
		if t == string(value.([]byte)) {
			tts := TokenType(k)
			tt = &tts
			return nil
		}
	}
	return fmt.Errorf("no valid token type found")
}

// Value implements the driver Valuer interface.
func (tt TokenType) Value() (driver.Value, error) {
	return tt.String(), nil
}

func (tt TokenType) String() string {
	return types[tt]
}

//TokenService containing the service to access tokens
type TokenService struct {
	Datasource TokenDatasourceService
}

//CreateToken creates a new token
func (ts TokenService) CreateToken(t *Token) error {
	t.Hash = utils.RandomHash(32)

	_, err := ts.Datasource.Create(t)

	return err
}

//GetToken gets all token for a defined token type expires after a defined time
//Expired tokens will be removed
func (ts TokenService) GetToken(hash string, tt TokenType, expireAfter time.Duration) (*Token, error) {
	token, err := ts.Datasource.Get(hash, tt)

	if err != nil {
		return nil, err
	}

	now := time.Now()

	if now.After(token.RequestedAt.Add(expireAfter)) {
		err = ts.Datasource.Remove(token.Hash, tt)
		logger.Log.Errorf("could not remove expired token, err %v", err)

		return nil, httperror.New(http.StatusNotFound, "The token is already expired. Fill out the form to receive a new token", errors.New("the token was expired"))
	}

	return token, nil
}

//RemoveToken removes a token
func (ts TokenService) RemoveToken(hash string, tt TokenType) error {
	return ts.Datasource.Remove(hash, tt)
}
