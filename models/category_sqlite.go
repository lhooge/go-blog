package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"
)

// SQLiteArticleDatasource providing an implementation of ArticleDatasourceService for SQLite
type SQLiteCategoryDatasource struct {
	SQLConn *sql.DB
}

func (rdb SQLiteCategoryDatasource) Create(c *Category) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO category (name, slug, last_modified, user_id) "+
		"VALUES (?, ?, ?, ?)",
		c.Name,
		c.Slug,
		time.Now(),
		c.Author.ID)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (rdb SQLiteCategoryDatasource) List(fc FilterCriteria) ([]Category, error) {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT c.id, c.name, c.slug, c.last_modified, ")
	stmt.WriteString("u.id, u.display_name, u.username, u.email, u.is_admin ")
	stmt.WriteString("FROM category as c ")
	stmt.WriteString("INNER JOIN user as u ")
	stmt.WriteString("ON c.user_id = u.id ")

	if fc == CategoriesWithArticles {
		stmt.WriteString("INNER JOIN article as a ")
		stmt.WriteString("ON c.id = a.category_id ")
	} else if fc == CategoriesWithoutArticles {
		stmt.WriteString("LEFT JOIN article as a ")
		stmt.WriteString("ON c.id = a.category_id ")
		stmt.WriteString("WHERE a.categorie_id IS NULL ")
	}

	stmt.WriteString("ORDER BY c.last_modified DESC ")

	rows, err := rdb.SQLConn.Query(stmt.String(), args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	cs := []Category{}

	for rows.Next() {
		var c Category
		var ru User

		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.LastModified, &ru.ID, &ru.DisplayName, &ru.Username, &ru.Email, &ru.IsAdmin); err != nil {
			return nil, err
		}

		c.Author = &ru

		cs = append(cs, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cs, nil
}

func (rdb SQLiteCategoryDatasource) Count(fc FilterCriteria) (int, error) {
	var total int

	if err := rdb.SQLConn.QueryRow("SELECT count(id) FROM category ").Scan(&total); err != nil {
		return -1, err
	}

	return total, nil
}

func (rdb SQLiteCategoryDatasource) Get(categoryID int) (*Category, error) {
	var stmt bytes.Buffer

	stmt.WriteString("SELECT c.id, c.name, c.slug, c.last_modified, ")
	stmt.WriteString("u.id, u.display_name, u.username, u.email, u.is_admin ")
	stmt.WriteString("FROM category as c ")
	stmt.WriteString("INNER JOIN user as u ")
	stmt.WriteString("ON u.id = c.user_id ")
	stmt.WriteString("WHERE c.id=? ")

	var c Category
	var ru User

	if err := rdb.SQLConn.QueryRow(stmt.String(), categoryID).Scan(&c.ID, &c.Name, &c.Slug, &c.LastModified, &ru.ID,
		&ru.DisplayName, &ru.Username, &ru.Email, &ru.IsAdmin); err != nil {
		return nil, err
	}

	c.Author = &ru

	return &c, nil
}

func (rdb SQLiteCategoryDatasource) GetBySlug(slug string) (*Category, error) {
	var stmt bytes.Buffer

	stmt.WriteString("SELECT c.id, c.name, c.slug, c.last_modified, ")
	stmt.WriteString("u.id, u.display_name, u.username, u.email, u.is_admin ")
	stmt.WriteString("FROM category as c ")
	stmt.WriteString("INNER JOIN user as u ")
	stmt.WriteString("ON u.id = c.user_id ")
	stmt.WriteString("WHERE c.slug=? ")

	fmt.Println("yes")
	var c Category
	var ru User

	if err := rdb.SQLConn.QueryRow(stmt.String(), slug).Scan(&c.ID, &c.Name, &c.Slug, &c.LastModified, &ru.ID,
		&ru.DisplayName, &ru.Username, &ru.Email, &ru.IsAdmin); err != nil {
		return nil, err
	}

	c.Author = &ru

	return &c, nil
}

func (rdb SQLiteCategoryDatasource) Update(c *Category) error {
	_, err := rdb.SQLConn.Exec("UPDATE category SET name=?, slug=?, last_modified=?, user_id=? WHERE id=?"+
		c.Name,
		c.Slug,
		time.Now(),
		c.Author.ID,
		c.ID)

	if err != nil {
		return err
	}

	return nil
}

func (rdb SQLiteCategoryDatasource) Delete(categoryID int) error {
	if _, err := rdb.SQLConn.Exec("DELETE FROM category WHERE id=?", categoryID); err != nil {
		return err
	}
	return nil
}
