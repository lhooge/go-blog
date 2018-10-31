package models

import (
	"bytes"
	"database/sql"
	"time"
)

//SQLiteUserInviteDatasource
type SQLiteUserInviteDatasource struct {
	SQLConn *sql.DB
}

func (rdb SQLiteUserInviteDatasource) List() ([]UserInvite, error) {
	var stmt bytes.Buffer
	var args []interface{}

	stmt.WriteString("SELECT ui.rowid, ui.username, ui.email, ui.display_name, ui.created_at, ui.is_admin, ")
	stmt.WriteString("u.rowid, u.username, u.email, u.display_name ")
	stmt.WriteString("FROM user_invite as ui ")
	stmt.WriteString("INNER JOIN user as u ")
	stmt.WriteString("ON u.rowid = ui.created_by ")
	stmt.WriteString("ORDER BY ui.username ASC ")

	rows, err := rdb.SQLConn.Query(stmt.String(), args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	invites := []UserInvite{}

	var ui UserInvite
	var u User

	for rows.Next() {
		if err = rows.Scan(&ui.ID, &ui.Username, &ui.Email, &ui.DisplayName, &ui.CreatedAt, &ui.IsAdmin, &u.ID, &u.Username, &u.Email, &u.DisplayName); err != nil {
			return nil, err
		}
		ui.CreatedBy = &u
		invites = append(invites, ui)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return invites, nil
}

func (rdb SQLiteUserInviteDatasource) Get(inviteID int) (*UserInvite, error) {
	var u User
	var ui UserInvite

	if err := rdb.SQLConn.QueryRow("SELECT ui.rowid, ui.hash, ui.username, ui.email, ui.display_name, ui.created_at, ui.is_admin, "+
		"u.rowid, u.username, u.email, u.display_name "+
		"FROM user_invite as ui "+
		"INNER JOIN user as u "+
		"ON u.rowid = ui.created_by "+
		"WHERE ui.rowid=? ", inviteID).
		Scan(&ui.ID, &ui.Hash, &ui.Username, &ui.Email, &ui.DisplayName, &ui.CreatedAt, &ui.IsAdmin, &u.ID, &u.Username, &u.Email, &u.DisplayName); err != nil {
		return nil, err
	}

	ui.CreatedBy = &u

	return &ui, nil
}

func (rdb SQLiteUserInviteDatasource) GetByHash(hash string) (*UserInvite, error) {
	var ui UserInvite
	var u User

	if err := rdb.SQLConn.QueryRow("SELECT ui.rowid, ui.hash, ui.username, ui.email, ui.display_name, ui.created_at, ui.is_admin, "+
		"u.rowid, u.username, u.email, u.display_name "+
		"FROM user_invite as ui "+
		"INNER JOIN user as u "+
		"ON u.rowid = ui.created_by "+
		"WHERE ui.hash=? ", hash).
		Scan(&ui.ID, &ui.Hash, &ui.Username, &ui.Email, &ui.DisplayName, &ui.CreatedAt, &ui.IsAdmin, &u.ID, &u.Username, &u.Email, &u.DisplayName); err != nil {
		return nil, err
	}

	ui.CreatedBy = &u

	return &ui, nil
}

func (rdb SQLiteUserInviteDatasource) Update(ui *UserInvite) error {
	if _, err := rdb.SQLConn.Exec("UPDATE user_invite SET hash=?, username=?, email=?, display_name=?, is_admin=?, created_at=?, created_by=? "+
		"WHERE rowid=? ", ui.Hash, ui.Username, ui.Email, ui.DisplayName, ui.IsAdmin, ui.CreatedBy.ID, ui.ID); err != nil {
		return err
	}

	return nil
}

//Create creates an new user invitation
func (rdb SQLiteUserInviteDatasource) Create(ui *UserInvite) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO user_invite (hash, username, email, display_name, is_admin, created_at, created_by) VALUES(?, ?, ?, ?, ?, ?, ?);",
		ui.Hash, ui.Username, ui.Email, ui.DisplayName, ui.IsAdmin, time.Now(), ui.CreatedBy.ID)

	if err != nil {
		return -1, err
	}

	i, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}
	return int(i), nil
}

//Count retuns the amount of users invitations
func (rdb SQLiteUserInviteDatasource) Count() (int, error) {
	var total int

	if err := rdb.SQLConn.QueryRow("SELECT count(rowid) FROM user_invite").Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

//Removes an user invitation
func (rdb SQLiteUserInviteDatasource) Remove(inviteID int) error {
	var stmt bytes.Buffer

	stmt.WriteString("DELETE FROM user_invite WHERE rowid=?")

	if _, err := rdb.SQLConn.Exec(stmt.String(), inviteID); err != nil {
		return err
	}

	return nil
}
