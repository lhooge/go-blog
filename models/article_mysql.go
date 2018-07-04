package models

import (
	"bytes"
	"database/sql"
	"time"
)

//MySQLArticleDatasource providing an implementation of ArticleDatasourceService using MariaDB
type MySQLArticleDatasource struct {
	SQLConn *sql.DB
}

//Create creates a article
func (rdb MySQLArticleDatasource) Create(a *Article) (int, error) {
	res, err := rdb.SQLConn.Exec("INSERT INTO article (headline, teaser, content, slug, published_on, published, last_modified, user_id) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		a.Headline,
		a.Teaser,
		a.Content,
		a.Slug,
		nil,
		false,
		time.Now(),
		a.Author.ID)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

//List returns a slice of articles; if the user is not nil the number of articles for this explcit user is returned
//the PublishedCritera specifies which articles should be considered
func (rdb MySQLArticleDatasource) List(u *User, p *Pagination, pc PublishedCriteria) ([]Article, error) {
	rows, err := selectArticlesStmt(rdb.SQLConn, u, p, pc)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	articles := []Article{}

	for rows.Next() {
		var a Article
		var ru User

		if err := rows.Scan(&a.ID, &a.Headline, &a.Teaser, &a.Content, &a.Published, &a.PublishedOn, &a.Slug, &a.LastModified, &ru.ID, &ru.DisplayName,
			&ru.Email, &ru.Username, &ru.IsAdmin); err != nil {
			return nil, err
		}

		a.Author = &ru

		articles = append(articles, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return articles, nil
}

//Count returns the number of article found; if the user is not nil the number of articles for this explcit user is returned
//the PublishedCritera specifies which articles should be considered
func (rdb MySQLArticleDatasource) Count(u *User, pc PublishedCriteria) (int, error) {
	var total int

	var args []interface{}

	var stmt bytes.Buffer
	stmt.WriteString("SELECT count(id) FROM article WHERE ")

	if u != nil {
		if !u.IsAdmin {
			stmt.WriteString("user_id=? AND ")
			args = append(args, u.ID)
		}
	}

	if pc == NotPublished {
		stmt.WriteString("published = false ")
	} else if pc == All {
		stmt.WriteString("(published=false OR published=true) ")
	} else {
		stmt.WriteString("published = true ")
	}

	if err := rdb.SQLConn.QueryRow(stmt.String(), args...).Scan(&total); err != nil {
		return -1, err
	}

	return total, nil
}

//Get returns a article by its id; if the user is not nil the article for this explcit user is returned
//the PublishedCritera specifies which articles should be considered
func (rdb MySQLArticleDatasource) Get(articleID int, u *User, pc PublishedCriteria) (*Article, error) {
	var a Article
	var ru User

	if err := selectArticleStmt(rdb.SQLConn, articleID, "", u, pc).Scan(&a.ID, &a.Headline, &a.PublishedOn, &a.Published, &a.Slug, &a.Teaser, &a.Content,
		&a.LastModified, &ru.ID, &ru.DisplayName, &ru.Email, &ru.Username, &ru.IsAdmin); err != nil {
		return nil, err
	}

	a.Author = &ru

	return &a, nil
}

//GetBySlug returns a article by its slug; if the user is not nil the article for this explcit user is returned
//the PublishedCritera specifies which articles should be considered
func (rdb MySQLArticleDatasource) GetBySlug(slug string, u *User, pc PublishedCriteria) (*Article, error) {
	var a Article
	var ru User

	if err := selectArticleStmt(rdb.SQLConn, -1, slug, u, pc).Scan(&a.ID, &a.Headline, &a.PublishedOn, &a.Published, &a.Slug, &a.Teaser, &a.Content,
		&a.LastModified, &ru.ID, &ru.DisplayName, &ru.Email, &ru.Username, &ru.IsAdmin); err != nil {
		return nil, err
	}

	a.Author = &ru

	return &a, nil
}

// Update updates an aricle
func (rdb MySQLArticleDatasource) Update(a *Article) error {
	if _, err := rdb.SQLConn.Exec("UPDATE article SET headline=?, teaser=?, content=?, last_modified=? WHERE id=? ", a.Headline, a.Teaser, a.Content,
		time.Now(), a.ID); err != nil {
		return err
	}

	return nil
}

// Publish checks if the article is published or not - switches the appropriate status
func (rdb MySQLArticleDatasource) Publish(a *Article) error {
	publishOn := NullTime{Valid: false}

	if !a.Published {
		publishOn = NullTime{Time: time.Now(), Valid: true}
	}

	if _, err := rdb.SQLConn.Exec("UPDATE article SET published=?, last_modified=?, published_on=? WHERE id=? ", !a.Published, time.Now(),
		publishOn, a.ID); err != nil {
		return err
	}

	return nil
}

// Delete deletes the article specified by the articleID
func (rdb MySQLArticleDatasource) Delete(articleID int) error {
	if _, err := rdb.SQLConn.Exec("DELETE FROM article WHERE id=?  ", articleID); err != nil {
		return err
	}
	return nil
}

func selectArticleStmt(db *sql.DB, articleID int, slug string, u *User, pc PublishedCriteria) *sql.Row {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT a.id, a.headline, a.published_on, a.published, a.slug, a.teaser, a.content, a.last_modified, ")
	stmt.WriteString("u.id, u.display_name, u.email, u.username, u.is_admin ")
	stmt.WriteString("FROM article a ")
	stmt.WriteString("INNER JOIN user u ON (a.user_id = u.id) ")
	stmt.WriteString("WHERE ")

	if pc == NotPublished {
		stmt.WriteString("a.published=false ")
	} else if pc == All {
		stmt.WriteString("(a.published=false OR a.published=true) ")
	} else {
		stmt.WriteString("a.published=true ")
	}

	if len(slug) > 0 {
		stmt.WriteString("AND a.slug = ? ")
		args = append(args, slug)
	} else {
		stmt.WriteString("AND a.id=? ")
		args = append(args, articleID)
	}
	if u != nil {
		if !u.IsAdmin {
			stmt.WriteString("AND a.user_id=? ")
			args = append(args, u.ID)
		}
	}
	stmt.WriteString("LIMIT 1")
	return db.QueryRow(stmt.String(), args...)
}

func selectArticlesStmt(db *sql.DB, u *User, p *Pagination, pc PublishedCriteria) (*sql.Rows, error) {
	var stmt bytes.Buffer

	var args []interface{}

	stmt.WriteString("SELECT a.id, a.headline, a.teaser, a.content, a.published, a.published_on, a.slug, a.last_modified, ")
	stmt.WriteString("u.id, u.display_name, u.email, u.username, u.is_admin ")
	stmt.WriteString("FROM article a ")
	stmt.WriteString("INNER JOIN user u ON (a.user_id = u.id) ")
	stmt.WriteString("WHERE ")

	if u != nil {
		if !u.IsAdmin {
			stmt.WriteString("a.user_id=? AND ")
			args = append(args, u.ID)
		}
	}

	if pc == NotPublished {
		stmt.WriteString("a.published=false ")
	} else if pc == All {
		stmt.WriteString("(a.published=false OR a.published=true) ")
	} else {
		stmt.WriteString("a.published=true ")
	}

	stmt.WriteString("ORDER BY a.published_on DESC, a.published ASC, last_modified DESC ")

	if p != nil {
		stmt.WriteString("LIMIT ? OFFSET ? ")
		args = append(args, p.Limit, p.Offset())
	}

	return db.Query(stmt.String(), args...)
}
