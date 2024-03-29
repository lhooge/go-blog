package models

import (
	"database/sql"
	"git.hoogi.eu/snafu/go-blog/logger"
	"strings"
	"time"
)

// SQLiteUserDatasource providing an implementation of UserDatasourceService using SQLite
type SQLiteUserDatasource struct {
	SQLConn *sql.DB
}

// List returns a list of users
func (rdb *SQLiteUserDatasource) List(p *Pagination) ([]User, error) {
	var stmt strings.Builder
	var args []interface{}
	var users []User
	var u User

	stmt.WriteString("SELECT id, username, email, display_name, last_modified, active, is_admin FROM user ORDER BY username ASC ")

	if p != nil {
		stmt.WriteString("LIMIT ? OFFSET ? ")
		args = append(args, p.Limit, p.Offset())
	}

	rows, err := rdb.SQLConn.Query(stmt.String(), args...)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logger.Log.Error(err)
		}
	}()

	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.LastModified, &u.Active, &u.IsAdmin); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Get gets a user by his userID
func (rdb *SQLiteUserDatasource) Get(userID int) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT u.id, u.username, u.email, u.display_name, u.last_modified, u.active, u.is_admin,  u.salt "+
		"FROM user as u "+
		"WHERE u.id=? ", userID).
		Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.LastModified, &u.Active, &u.IsAdmin, &u.Salt); err != nil {
		return nil, err
	}

	return &u, nil
}

// GetByMail gets a user by his mail, includes the password and salt
func (rdb *SQLiteUserDatasource) GetByMail(mail string) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT id, is_admin, active, display_name, username, email, salt, password FROM user WHERE email=? ", mail).
		Scan(&u.ID, &u.IsAdmin, &u.Active, &u.DisplayName, &u.Username, &u.Email, &u.Salt, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByUsername gets a user by his username, includes the password and salt
func (rdb *SQLiteUserDatasource) GetByUsername(username string) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT id, is_admin, active, display_name, username, email, salt, password FROM user WHERE username=? ", username).
		Scan(&u.ID, &u.IsAdmin, &u.Active, &u.DisplayName, &u.Username, &u.Email, &u.Salt, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}

// Create creates a new user
func (rdb *SQLiteUserDatasource) Create(u *User) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO user (salt, password, username, email, display_name, last_modified, active, is_admin) VALUES(?, ?, ?, ?, ?, ?, ?, ?);",
		u.Salt, u.Password, u.Username, u.Email, u.DisplayName, time.Now(), u.Active, u.IsAdmin)

	if err != nil {
		return -1, err
	}

	i, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return int(i), nil
}

// Update updates an user
func (rdb *SQLiteUserDatasource) Update(u *User, changePassword bool) error {
	var stmt strings.Builder
	var args []interface{}

	stmt.WriteString("UPDATE user SET display_name=?, username=?, email=?, last_modified=?, active=?, is_admin=? ")
	args = append(args, u.DisplayName, u.Username, u.Email, time.Now(), u.Active, u.IsAdmin)

	if changePassword {
		stmt.WriteString(", salt=?, password=? ")
		args = append(args, u.Salt, u.Password)
	}

	stmt.WriteString("WHERE id=?;")

	args = append(args, u.ID)

	if _, err := rdb.SQLConn.Exec(stmt.String(), args...); err != nil {
		return err
	}

	return nil
}

// Count returns the amount of users matches the AdminCriteria
func (rdb *SQLiteUserDatasource) Count(ac AdminCriteria) (int, error) {
	var stmt strings.Builder
	stmt.WriteString("SELECT count(id) FROM user ")

	if ac == OnlyAdmins {
		stmt.WriteString("WHERE is_admin = '1'")
	} else if ac == NoAdmins {
		stmt.WriteString("WHERE is_admin = '0'")
	}

	var total int

	if err := rdb.SQLConn.QueryRow(stmt.String()).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

// Remove removes an user
func (rdb *SQLiteUserDatasource) Remove(userID int) error {
	if _, err := rdb.SQLConn.Exec("DELETE FROM user WHERE id=?", userID); err != nil {
		return err
	}

	return nil
}
