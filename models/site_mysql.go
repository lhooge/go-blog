package models

import (
	"bytes"
	"database/sql"
	"time"

	"git.hoogi.eu/go-blog/components/logger"
)

//MySQLSiteDatasource providing an implementation of SiteDatasourceService using MariaDB
type MySQLSiteDatasource struct {
	SQLConn *sql.DB
}

//List returns a array of sites
func (rdb MySQLSiteDatasource) List(pc PublishedCriteria, p *Pagination) ([]Site, error) {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT s.id, s.title, s.link, s.content, s.published, s.published_on, s.last_modified, s.order_no, u.id, u.display_name, u.email, u.username ")
	stmt.WriteString("FROM site s ")
	stmt.WriteString("INNER JOIN user u ON (s.user_id = u.id) ")
	stmt.WriteString("WHERE ")

	if pc == All {
		stmt.WriteString("(s.published=false OR s.published=true) ")
	} else if pc == NotPublished {
		stmt.WriteString("s.published = false ")
	} else {
		stmt.WriteString("s.published = true ")
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

	defer rows.Close()

	sites := []Site{}

	var s Site
	var u User

	for rows.Next() {
		if err = rows.Scan(&s.ID, &s.Title, &s.Link, &s.Content, &s.Published, &s.PublishedOn, &s.LastModified, &s.OrderNo, &u.ID, &u.DisplayName, &u.Email, &u.Username); err != nil {
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

//Get returns a site based on the site id
func (rdb MySQLSiteDatasource) Get(siteID int, pc PublishedCriteria) (*Site, error) {
	var stmt bytes.Buffer
	var args []interface{}

	stmt.WriteString("SELECT s.id, s.title, s.link, s.content, s.published, s.published_on, s.last_modified, s.order_no FROM site as s WHERE id=? ")
	args = append(args, siteID)

	if pc == NotPublished {
		stmt.WriteString("AND published = false ")
	} else if pc == OnlyPublished {
		stmt.WriteString("AND published = true ")
	}

	var s Site

	if err := rdb.SQLConn.QueryRow(stmt.String(), siteID).Scan(&s.ID, &s.Title, &s.Link, &s.Content, &s.Published, &s.PublishedOn, &s.LastModified, &s.OrderNo); err != nil {
		return nil, err
	}
	return &s, nil
}

//GetByLink returns a site based on the provided link
func (rdb MySQLSiteDatasource) GetByLink(link string, pc PublishedCriteria) (*Site, error) {
	var stmt bytes.Buffer
	var args []interface{}

	stmt.WriteString("SELECT s.id, s.title, s.link, s.content, s.published, s.published_on, s.order_no, s.last_modified FROM site as s WHERE link=? ")
	args = append(args, link)

	if pc == NotPublished {
		stmt.WriteString("AND published = false ")
	} else if pc == OnlyPublished {
		stmt.WriteString("AND published = true ")
	}

	var s Site

	if err := rdb.SQLConn.QueryRow(stmt.String(), link).Scan(&s.ID, &s.Title, &s.Link, &s.Content, &s.Published, &s.PublishedOn, &s.OrderNo, &s.LastModified); err != nil {
		return nil, err
	}
	return &s, nil
}

//Publish publishes or unpublishes a site
func (rdb MySQLSiteDatasource) Publish(s *Site) error {
	publishOn := NullTime{Valid: false}
	if !s.Published {
		publishOn = NullTime{Time: time.Now(), Valid: true}
	}

	if _, err := rdb.SQLConn.Exec("UPDATE site SET published=?, published_on=?, last_modified=? WHERE id=?", !s.Published, publishOn, time.Now(), s.ID); err != nil {
		return err
	}
	return nil
}

//Create creates a site
func (rdb MySQLSiteDatasource) Create(s *Site) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO site (title, link, content, published, published_on, last_modified, order_no, user_id) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		s.Title, s.Link, s.Content, s.Published, s.PublishedOn, time.Now(), s.OrderNo, s.Author.ID)

	if err != nil {
		return -1, err
	}

	i, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return int(i), nil
}

//Order moves a site up or down
func (rdb MySQLSiteDatasource) Order(id int, d Direction) error {
	tx, err := rdb.SQLConn.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			logger.Log.Error("error during ordering of sites ", err)
			tx.Rollback()
		}
	}()

	if err != nil {
		return err
	}

	if d == Up {
		if _, err = tx.Exec("UPDATE site AS s, (SELECT order_no-1 AS order_no FROM site WHERE id=?) AS n "+
			"SET s.order_no=n.order_no+1 "+
			"WHERE s.order_no=n.order_no ", id); err != nil {
			return err
		}

		if _, err = tx.Exec("UPDATE site SET order_no = order_no-1 WHERE id = ? AND order_no-1 > 0", id); err != nil {
			return err
		}
	} else if d == Down {
		var max int

		tx.QueryRow("SELECT MAX(order_no) AS max FROM site").Scan(&max)

		if _, err = tx.Exec("UPDATE site AS s, "+
			"(SELECT order_no+1 AS swap_el FROM site WHERE id=?) AS n "+
			"SET s.order_no=n.swap_el-1 "+
			"WHERE s.order_no=n.swap_el "+
			"AND n.swap_el <= ?", id, max); err != nil {
			return err
		}

		if _, err = tx.Exec("UPDATE site "+
			"SET order_no = order_no+1 "+
			"WHERE id = ? "+
			"AND order_no+1 <= ?", id, max); err != nil {
			return err
		}
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

//Update updates a site
func (rdb MySQLSiteDatasource) Update(s *Site) error {
	if _, err := rdb.SQLConn.Exec("UPDATE site SET title=?, link=?, content=?, last_modified=? WHERE id=?", s.Title, s.Link, s.Content, time.Now(), s.ID); err != nil {
		return err
	}

	return nil
}

//Count returns the amount of sites
func (rdb MySQLSiteDatasource) Count(pc PublishedCriteria) (int, error) {
	var stmt bytes.Buffer

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

//Max returns the maximum order number
func (rdb MySQLSiteDatasource) Max() (int, error) {
	var max sql.NullInt64

	if err := rdb.SQLConn.QueryRow("SELECT MAX(order_no) FROM site").Scan(&max); err != nil {
		return 0, err
	}

	if max.Valid == false {
		max.Int64 = 0
	}

	return int(max.Int64), nil
}

//Delete deletes a site and updates the order numbers
func (rdb MySQLSiteDatasource) Delete(s *Site) error {
	tx, err := rdb.SQLConn.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			logger.Log.Error("error during delete transaction", err)
			tx.Rollback()
			return
		}
	}()

	if _, err := tx.Exec("DELETE FROM site WHERE id=?", s.ID); err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE site SET order_no = order_no-1 WHERE order_no > ?", s.OrderNo); err != nil {
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}
