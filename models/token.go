package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"time"

	"git.hoogi.eu/snafu/go-blog/crypt"
	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/logger"
)

// TokenDatasourceService defines an interface for CRUD operations for tokens
type TokenDatasourceService interface {
	Create(t *Token) (int, error)
	Get(hash string, tt TokenType) (*Token, error)
	ListByUser(userID int, tt TokenType) ([]Token, error)
	Remove(hash string, tt TokenType) error
}

// Token represents a token
type Token struct {
	ID          int
	Hash        string
	Type        TokenType
	RequestedAt time.Time

	Author *User
}

const (
	// PasswordReset token generated for resetting passwords
	PasswordReset = iota
)

var types = [...]string{"password_reset"}

// TokenType specifies the type where token can be used
type TokenType int

// Scan implements the Scanner interface.
func (tt *TokenType) Scan(value interface{}) error {
	for k, t := range types {
		if t == (value.(string)) {
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

// TokenService containing the service to access tokens
type TokenService struct {
	Datasource TokenDatasourceService
}

// Create creates a new token
func (ts *TokenService) Create(t *Token) error {
	t.Hash = crypt.RandomHash(32)

	if _, err := ts.Datasource.Create(t); err != nil {
		return err
	}

	return nil
}

// Get token for a defined token type expires after a defined time
// Expired token will be removed
func (ts *TokenService) Get(hash string, tt TokenType, expireAfter time.Duration) (*Token, error) {
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

// RateLimit returns an error if a token is requested greater three times in a time span of 15 minutes
func (ts *TokenService) RateLimit(userID int, tt TokenType) error {
	tokens, err := ts.Datasource.ListByUser(userID, tt)

	if err != nil {
		return err
	}

	now := time.Now()

	var rate []Token
	for _, t := range tokens {
		if now.Sub(t.RequestedAt) < time.Minute*15 {
			rate = append(rate, t)
		}
	}

	if len(rate) > 3 {
		return fmt.Errorf("too many tokens of type %s were requested for user %d, not sending mail", tt, userID)
	}

	return nil
}

// Remove removes a token
func (ts *TokenService) Remove(hash string, tt TokenType) error {
	return ts.Datasource.Remove(hash, tt)
}
