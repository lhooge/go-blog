package models

import (
	"database/sql"
	"strings"
	"time"

	"git.hoogi.eu/snafu/go-blog/logger"
)

// SQLiteSiteDatasource providing an implementation of SiteDatasourceService for sqlite
type SQLiteSiteDatasource struct {
	SQLConn *sql.DB
}

// List returns a array of sites
func (rdb *SQLiteSiteDatasource) List(pc PublishedCriteria, p *Pagination) ([]Site, error) {
	var stmt strings.Builder
	var args []interface{}

	stmt.WriteString("SELECT s.id, s.title, s.link, s.section, s.content, s.published, s.published_on, s.last_modified, s.order_no, u.id, u.display_name, u.email, u.username ")
	stmt.WriteString("FROM site s ")
	stmt.WriteString("INNER JOIN user u ON (s.user_id = u.id) ")
	stmt.WriteString("WHERE ")

	if pc == All {
		stmt.WriteString("(s.published='0' OR s.published='1') ")
	} else if pc == NotPublished {
		stmt.WriteString("s.published = '0' ")
	} else {
		stmt.WriteString("s.published = '1' ")
	}

	stmt.WriteString("ORDER BY order_no ASC ")

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

	var sites []Site
	var s Site
	var u User

	for rows.Next() {
		if err = rows.Scan(&s.ID, &s.Title, &s.Link, &s.Section, &s.Content, &s.Published, &s.PublishedOn, &s.LastModified, &s.OrderNo, &u.ID, &u.DisplayName, &u.Email, &u.Username); err != nil {
			return nil, err
		}

		s.Author = &u

		sites = append(sites, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sites, nil
}

// Get returns a site based on the site id
func (rdb *SQLiteSiteDatasource) Get(siteID int, pc PublishedCriteria) (*Site, error) {
	var stmt strings.Builder
	var args []interface{}

	stmt.WriteString("SELECT s.id, s.title, s.link, s.section, s.content, s.published, s.published_on, s.last_modified, s.order_no, u.id, u.display_name, u.email, u.username FROM site as s ")
	stmt.WriteString("INNER JOIN user u ON (s.user_id = u.id) ")
	stmt.WriteString("WHERE s.id=? ")

	args = append(args, siteID)

	if pc == NotPublished {
		stmt.WriteString("AND s.published = '0' ")
	} else if pc == OnlyPublished {
		stmt.WriteString("AND s.published = '1' ")
	}

	var s Site
	var u User

	if err := rdb.SQLConn.QueryRow(stmt.String(), siteID).Scan(&s.ID, &s.Title, &s.Link, &s.Section, &s.Content, &s.Published, &s.PublishedOn, &s.LastModified, &s.OrderNo, &u.ID, &u.DisplayName, &u.Email, &u.Username); err != nil {
		return nil, err
	}

	s.Author = &u

	return &s, nil
}

// GetByLink returns a site based on the provided link
func (rdb *SQLiteSiteDatasource) GetByLink(link string, pc PublishedCriteria) (*Site, error) {
	var stmt strings.Builder
	var args []interface{}

	stmt.WriteString("SELECT s.id, s.title, s.link, s.section, s.content, s.published, s.published_on, s.order_no, s.last_modified, u.id, u.display_name, u.email, u.username FROM site as s ")
	stmt.WriteString("INNER JOIN user u ON (s.user_id = u.id) ")
	stmt.WriteString("WHERE s.link=? ")

	args = append(args, link)

	if pc == NotPublished {
		stmt.WriteString("AND s.published = '0' ")
	} else if pc == OnlyPublished {
		stmt.WriteString("AND s.published = '1' ")
	}

	var s Site
	var u User

	if err := rdb.SQLConn.QueryRow(stmt.String(), link).Scan(&s.ID, &s.Title, &s.Link, &s.Section, &s.Content, &s.Published,
		&s.PublishedOn, &s.OrderNo, &s.LastModified, &u.ID, &u.DisplayName, &u.Email, &u.Username); err != nil {
		return nil, err
	}

	s.Author = &u

	return &s, nil
}

// Publish publishes or unpublishes a site
func (rdb *SQLiteSiteDatasource) Publish(s *Site) error {
	publishOn := NullTime{Valid: false}

	if !s.Published {
		publishOn = NullTime{Time: time.Now(), Valid: true}
	}

	if _, err := rdb.SQLConn.Exec("UPDATE site SET published=?, published_on=?, last_modified=? WHERE id=?", !s.Published, publishOn, time.Now(), s.ID); err != nil {
		return err
	}
	return nil
}

// Create creates a site
func (rdb *SQLiteSiteDatasource) Create(s *Site) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO site (title, link, section, content, published, published_on, last_modified, order_no, user_id) "+
		"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)",
		s.Title, s.Link, s.Section, s.Content, s.Published, s.PublishedOn, time.Now(), s.OrderNo, s.Author.ID)

	if err != nil {
		return -1, err
	}

	i, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return int(i), nil
}

// Order moves a site up or down
func (rdb *SQLiteSiteDatasource) Order(id int, d Direction) error {
	tx, err := rdb.SQLConn.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			logger.Log.Error("error during ordering of sites ", err)

			err := tx.Rollback()

			if err != nil {
				logger.Log.Error("error during transaction rollback ", err)
			}
		}
	}()

	if d == Up {
		if _, err = tx.Exec("UPDATE site "+
			"SET order_no=(SELECT order_no AS order_no FROM site WHERE id=?) "+
			"WHERE order_no=(SELECT order_no-1 AS order_no FROM site WHERE id=?) ", id, id); err != nil {
			return err
		}

		if _, err = tx.Exec("UPDATE site SET order_no = order_no - 1 WHERE id = ? AND order_no-1 > 0", id); err != nil {
			return err
		}
	} else if d == Down {
		var max int

		if err := tx.QueryRow("SELECT MAX(order_no) AS max FROM site").Scan(&max); err != nil {
			return err
		}

		if _, err = tx.Exec("UPDATE site "+
			"SET order_no=(SELECT order_no AS swap_el FROM site WHERE id=?) "+
			"WHERE order_no=(SELECT order_no+1 AS swap_el FROM site WHERE id=? AND swap_el <= ?) ", id, id, max); err != nil {
			return err
		}

		if _, err = tx.Exec("UPDATE site "+
			"SET order_no = order_no+1 "+
			"WHERE id = ? "+
			"AND order_no + 1 <= ?", id, max); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Update updates a site
func (rdb *SQLiteSiteDatasource) Update(s *Site) error {
	if _, err := rdb.SQLConn.Exec("UPDATE site SET title=?, link=?, section=?, content=?, last_modified=? WHERE id=?",
		s.Title, s.Link, s.Section, s.Content, time.Now(), s.ID); err != nil {
		return err
	}

	return nil
}

// Count returns the amount of sites
func (rdb *SQLiteSiteDatasource) Count(pc PublishedCriteria) (int, error) {
	var stmt strings.Builder

	stmt.WriteString("SELECT count(id) FROM site ")

	if pc == OnlyPublished {
		stmt.WriteString("WHERE published = true ")
	} else if pc == NotPublished {
		stmt.WriteString("WHERE published = false ")
	}

	var total int

	if err := rdb.SQLConn.QueryRow(stmt.String()).Scan(&total); err != nil {
		return -1, err
	}

	return total, nil
}

// Max returns the maximum order number
func (rdb *SQLiteSiteDatasource) Max() (int, error) {
	var max sql.NullInt64

	if err := rdb.SQLConn.QueryRow("SELECT MAX(order_no) FROM site").Scan(&max); err != nil {
		return 0, err
	}

	if max.Valid == false {
		max.Int64 = 0
	}

	return int(max.Int64), nil
}

// Delete deletes a site and updates the order numbers
func (rdb *SQLiteSiteDatasource) Delete(s *Site) error {
	tx, err := rdb.SQLConn.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			logger.Log.Errorf("error site removal not successful %v", err)
			if err := tx.Rollback(); err != nil {
				logger.Log.Errorf("could not rollback transaction during site removal %v", err)
				return
			}
			return
		}
	}()

	if _, err := tx.Exec("DELETE FROM site WHERE id=?", s.ID); err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE site SET order_no = order_no-1 WHERE order_no > ?", s.OrderNo); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
