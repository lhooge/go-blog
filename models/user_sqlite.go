package models

import (
	"bytes"
	"database/sql"
	"time"
)

//SQLiteUserDatasource providing an implementation of UserDatasourceService using SQLite
type SQLiteUserDatasource struct {
	SQLConn *sql.DB
}

//List returns a list of user
func (rdb SQLiteUserDatasource) List(p *Pagination) ([]User, error) {
	var stmt bytes.Buffer
	var args []interface{}

	stmt.WriteString("SELECT id, username, email, display_name, last_modified, active FROM user ORDER BY username ASC ")

	if p != nil {
		stmt.WriteString("LIMIT ? OFFSET ? ")
		args = append(args, p.Limit, p.Offset())
	}

	rows, err := rdb.SQLConn.Query(stmt.String(), args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []User{}

	var u User

	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.LastModified, &u.Active); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

//Get gets an user by his userID
func (rdb SQLiteUserDatasource) Get(userID int) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT u.id, u.username, u.email, u.display_name, u.last_modified, u.active,  u.salt "+
		"FROM user as u "+
		"WHERE u.id=? ", userID).
		Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.LastModified, &u.Active, &u.Salt); err != nil {
		return nil, err
	}

	return &u, nil
}

//GetByMail gets an user by his mail, includes the password and salt
func (rdb SQLiteUserDatasource) GetByMail(mail string) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT id, active, display_name, username, email, salt, password FROM user WHERE email=? ", mail).
		Scan(&u.ID, &u.Active, &u.DisplayName, &u.Username, &u.Email, &u.Salt, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}

//GetByUsername gets an user by his username, includes the password and salt
func (rdb SQLiteUserDatasource) GetByUsername(username string) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT id, active, display_name, username, email, salt, password FROM user WHERE username=? ", username).
		Scan(&u.ID, &u.Active, &u.DisplayName, &u.Username, &u.Email, &u.Salt, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}

//Create creates an new user
func (rdb SQLiteUserDatasource) Create(u *User) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO user (salt, password, username, email, display_name, last_modified, active) VALUES(?, ?, ?, ?, ?, ?, ?);",
		u.Salt, u.Password, u.Username, u.Email, u.DisplayName, time.Now(), u.Active)

	if err != nil {
		return -1, err
	}

	i, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}
	return int(i), nil
}

//Update updates an user
func (rdb SQLiteUserDatasource) Update(u *User, changePassword bool) error {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("UPDATE user SET display_name=?, username=?, email=?, last_modified=?, active=? ")
	args = append(args, u.DisplayName, u.Username, u.Email, time.Now(), u.Active)

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

//Count retuns the amount of users matches the AdminCriteria
func (rdb SQLiteUserDatasource) Count() (int, error) {
	var total int

	if err := rdb.SQLConn.QueryRow("SELECT count(id) FROM user ").Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

//Removes an user
func (rdb SQLiteUserDatasource) Remove(userID int) error {
	var stmt bytes.Buffer

	stmt.WriteString("DELETE FROM user WHERE id=?")

	if _, err := rdb.SQLConn.Exec(stmt.String(), userID); err != nil {
		return err
	}

	return nil
}
