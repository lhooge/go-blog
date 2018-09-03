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

	stmt.WriteString("SELECT rowid, username, email, display_name, last_modified, active, is_admin FROM user ORDER BY username ASC ")

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

//Get gets an user by his userID
func (rdb SQLiteUserDatasource) Get(userID int) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT u.rowid, u.username, u.email, u.display_name, u.last_modified, u.active, u.is_admin, u.salt "+
		"FROM user as u "+
		"WHERE u.rowid=? ", userID).
		Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.LastModified, &u.Active, &u.IsAdmin, &u.Salt); err != nil {
		return nil, err
	}

	return &u, nil
}

//GetByMail gets an user by his mail, includes the password and salt
func (rdb SQLiteUserDatasource) GetByMail(mail string) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT rowid, is_admin, active, display_name, username, email, salt, password FROM user WHERE email=? AND active='1' ", mail).
		Scan(&u.ID, &u.IsAdmin, &u.Active, &u.DisplayName, &u.Username, &u.Email, &u.Salt, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}

//GetByUsername gets an user by his username, includes the password and salt
func (rdb SQLiteUserDatasource) GetByUsername(username string) (*User, error) {
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT rowid, is_admin, active, display_name, username, email, salt, password FROM user WHERE username=? AND active='1' ", username).
		Scan(&u.ID, &u.IsAdmin, &u.Active, &u.DisplayName, &u.Username, &u.Email, &u.Salt, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}

//Create creates an new user
func (rdb SQLiteUserDatasource) Create(u *User) (int, error) {
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

//Update updates an user
func (rdb SQLiteUserDatasource) Update(u *User, changePassword bool) error {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("UPDATE user SET display_name=?, username=?, email=?, last_modified=?, active=?, is_admin=? ")
	args = append(args, u.DisplayName, u.Username, u.Email, time.Now(), u.Active, u.IsAdmin)

	if changePassword {
		stmt.WriteString(", salt=?, password=? ")

		args = append(args, u.Salt, u.Password)
	}
	stmt.WriteString("WHERE rowid=?;")

	args = append(args, u.ID)

	if _, err := rdb.SQLConn.Exec(stmt.String(), args...); err != nil {
		return err
	}

	return nil
}

//Count retuns the amount of users matches the AdminCriteria
func (rdb SQLiteUserDatasource) Count(ac AdminCriteria) (int, error) {
	var stmt bytes.Buffer

	stmt.WriteString("SELECT count(rowid) FROM user ")

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

//Removes an user
func (rdb SQLiteUserDatasource) Remove(userID int) error {
	var stmt bytes.Buffer

	stmt.WriteString("DELETE FROM user WHERE rowid=?")

	if _, err := rdb.SQLConn.Exec(stmt.String(), userID); err != nil {
		return err
	}

	return nil
}
