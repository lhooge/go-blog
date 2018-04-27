package models

import (
	"bytes"
	"database/sql"
	"time"
)

//MySQLFileDatasource providing an implementation of FileDatasourceService using MariaDB
type MySQLFileDatasource struct {
	SQLConn *sql.DB
}

//GetByFilename returns the file based on the filename; it the user is given and it is a non admin
//only file specific to this user is returned
func (rdb MySQLFileDatasource) GetByFilename(filename string, u *User) (*File, error) {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT f.id, f.filename, f.content_type, f.size, f.last_modified, f.user_id, ")
	stmt.WriteString("u.display_name, u.username, u.email, u.is_admin ")
	stmt.WriteString("FROM file as f ")
	stmt.WriteString("INNER JOIN user as u ")
	stmt.WriteString("ON u.id = f.user_id ")
	stmt.WriteString("WHERE f.filename=? ")

	args = append(args, filename)

	if u != nil {
		if !u.IsAdmin {
			stmt.WriteString("AND f.user_id=? ")
			args = append(args, u.ID)
		}
	}

	var f File
	var ru User

	if err := rdb.SQLConn.QueryRow(stmt.String(), args...).Scan(&f.ID, &f.Filename, &f.ContentType, &f.Size, &f.LastModified, &ru.ID,
		&ru.DisplayName, &ru.Username, &ru.Email, &ru.IsAdmin); err != nil {
		return nil, err
	}

	f.Author = &ru

	return &f, nil
}

//Get returns the file based on the filename; it the user is given and it is a non admin
//only file specific to this user is returned
func (rdb MySQLFileDatasource) Get(fileID int, u *User) (*File, error) {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT f.id, f.filename, f.content_type, f.size, f.last_modified, f.user_id, ")
	stmt.WriteString("u.display_name, u.username, u.email, u.is_admin ")
	stmt.WriteString("FROM file as f ")
	stmt.WriteString("INNER JOIN user as u ")
	stmt.WriteString("ON u.id = f.user_id ")
	stmt.WriteString("WHERE f.id=? ")

	args = append(args, fileID)

	if u != nil {
		if !u.IsAdmin {
			stmt.WriteString("AND f.user_id=? ")
			args = append(args, u.ID)
		}
	}

	var f File
	var ru User

	if err := rdb.SQLConn.QueryRow(stmt.String(), args...).Scan(&f.ID, &f.Filename, &f.ContentType, &f.Size, &f.LastModified, &ru.ID,
		&ru.DisplayName, &ru.Username, &ru.Email, &ru.IsAdmin); err != nil {
		return nil, err
	}

	f.Author = &ru

	return &f, nil
}

//Create inserts some file meta information into the database
func (rdb MySQLFileDatasource) Create(f *File) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO file (id, filename, content_type, size, last_modified, user_id) VALUES(?, ?, ?, ?, ?, ?)",
		f.ID, f.Filename, f.ContentType, f.Size, time.Now(), f.Author.ID)

	if err != nil {
		return -1, err
	}

	i, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return int(i), nil
}

//List returns a list of files based on the filename; it the user is given and it is a non admin
//only files specific to this user are returned
func (rdb MySQLFileDatasource) List(u *User, p *Pagination) ([]File, error) {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT f.id, f.filename, f.content_type, f.size, f.last_modified, ")
	stmt.WriteString("u.id, u.display_name, u.username, u.email, u.is_admin ")
	stmt.WriteString("FROM file as f ")
	stmt.WriteString("INNER JOIN user as u ")
	stmt.WriteString("ON f.user_id = u.id ")

	if u != nil {
		if !u.IsAdmin {
			stmt.WriteString("WHERE f.user_id=? ")
			args = append(args, u.ID)
		}
	}

	stmt.WriteString("ORDER BY f.last_modified DESC ")

	if p != nil {
		stmt.WriteString("LIMIT ? OFFSET ? ")
		args = append(args, p.Limit, p.Offset())
	}

	rows, err := rdb.SQLConn.Query(stmt.String(), args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	files := []File{}

	var f File
	var us User

	for rows.Next() {
		if err = rows.Scan(&f.ID, &f.Filename, &f.ContentType, &f.Size, &f.LastModified, &us.ID, &us.DisplayName,
			&us.Username, &us.Email, &u.IsAdmin); err != nil {
			return nil, err
		}

		f.Author = &us

		files = append(files, f)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return files, nil

}

//Count returns a number of files based on the filename; it the user is given and it is a non admin
//only files specific to this user are counted
func (rdb MySQLFileDatasource) Count(u *User) (int, error) {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT count(id) FROM file ")

	if u != nil {
		if !u.IsAdmin {
			stmt.WriteString("WHERE user_id = ?")
			args = append(args, u.ID)
		}
	}

	var total int

	if err := rdb.SQLConn.QueryRow(stmt.String(), args...).Scan(&total); err != nil {
		return -1, err
	}

	return total, nil
}

//Delete deletes a file based on fileID; users which are not the owner are not allowed to remove files; except admins
func (rdb MySQLFileDatasource) Delete(fileID int) error {
	if _, err := rdb.SQLConn.Exec("DELETE FROM file WHERE id=?", fileID); err != nil {
		return err
	}
	return nil

}
