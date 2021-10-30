// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/settings"
	"git.hoogi.eu/snafu/go-blog/slug"
)

// Article represents an article
type Article struct {
	ID           int
	Headline     string
	PublishedOn  NullTime
	Published    bool
	Teaser       string
	Content      string
	Slug         string
	LastModified time.Time
	Author       *User

	//duplicate category struct to support left joins with nulls
	//TODO: find a better solution
	CID   sql.NullInt64
	CName sql.NullString
}

// ArticleDatasourceService defines an interface for CRUD operations of articles
type ArticleDatasourceService interface {
	Create(a *Article) (int, error)
	List(u *User, c *Category, p *Pagination, pc PublishedCriteria) ([]Article, error)
	Count(u *User, c *Category, pc PublishedCriteria) (int, error)
	Get(articleID int, u *User, pc PublishedCriteria) (*Article, error)
	GetBySlug(slug string, u *User, pc PublishedCriteria) (*Article, error)
	Publish(a *Article) error
	Update(a *Article) error
	Delete(articleID int) error
}

const (
	maxHeadlineSize = 150
)

// SlugEscape escapes the slug for use in URLs
func (a Article) SlugEscape() string {
	spl := strings.Split(a.Slug, "/")
	return fmt.Sprintf("%s/%s/%s", spl[0], spl[1], url.PathEscape(spl[2]))
}

func (a *Article) buildSlug(now time.Time, suffix int) string {
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(now.Year()))
	sb.WriteString("/")
	sb.WriteString(strconv.Itoa(int(now.Month())))
	sb.WriteString("/")
	sb.WriteString(slug.CreateURLSafeSlug(a.Headline, suffix))
	return sb.String()
}

func (a *Article) slug(as ArticleService, now time.Time) error {
	for i := 0; i < 10; i++ {
		a.Slug = a.buildSlug(now, i)

		if _, err := as.Datasource.GetBySlug(a.Slug, nil, All); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			return err
		}
	}
	return nil
}

// validate validates if mandatory article fields are set
func (a *Article) validate() error {
	a.Headline = strings.TrimSpace(a.Headline)
	a.Content = strings.TrimSpace(a.Content)

	if len(a.Headline) == 0 {
		return httperror.ValueRequired("headline")
	}

	if len([]rune(a.Headline)) > maxHeadlineSize {
		return httperror.ValueTooLong("headline", maxHeadlineSize)
	}

	if len(a.Teaser) == 0 {
		return httperror.ValueRequired("teaser")
	}

	if a.Author == nil {
		return httperror.InternalServerError(errors.New("article validation failed - the author is missing"))
	}
	return nil
}

// ArticleService containing the service to access articles
type ArticleService struct {
	Datasource ArticleDatasourceService
	AppConfig  settings.Application
}

// Create creates an article
func (as ArticleService) Create(a *Article) (int, error) {
	now := time.Now()

	a.PublishedOn = NullTime{Time: now, Valid: true}

	if err := a.validate(); err != nil {
		return 0, err
	}

	if err := a.slug(as, now); err != nil {
		return -1, err
	}

	return as.Datasource.Create(a)
}

// Update updates an article
func (as ArticleService) Update(a *Article, u *User, updateSlug bool) error {
	if err := a.validate(); err != nil {
		return err
	}

	oldArt, err := as.Datasource.Get(a.ID, a.Author, All)

	if err != nil {
		return err
	}

	if !updateSlug {
		a.Slug = oldArt.Slug
	} else {
		now := time.Now()

		if err := a.slug(as, now); err != nil {
			return err
		}
	}

	if !u.IsAdmin {
		if oldArt.Author.ID != u.ID {
			return httperror.PermissionDenied("update", "article", fmt.Errorf("could not update article %d user %d has no permission", a.ID, u.ID))
		}
	}

	return as.Datasource.Update(a)
}

// Publish publishes or 'unpublishes' an article
func (as ArticleService) Publish(id int, u *User) error {
	a, err := as.Datasource.Get(id, nil, All)

	if err != nil {
		return err
	}

	if !u.IsAdmin {
		if a.Author.ID != u.ID {
			return httperror.PermissionDenied("publish", "article", fmt.Errorf("could not publish article %d user %d has no permission", a.ID, u.ID))
		}
	}

	return as.Datasource.Publish(a)
}

// Delete deletes an article
func (as ArticleService) Delete(id int, u *User) error {
	a, err := as.Datasource.Get(id, nil, All)

	if err != nil {
		return err
	}

	if !u.IsAdmin {
		if a.Author.ID != u.ID {
			return httperror.PermissionDenied("delete", "article", fmt.Errorf("could not delete article %d user %d has no permission", a.ID, u.ID))
		}
	}

	return as.Datasource.Delete(a.ID)
}

// GetBySlug gets a article by the slug.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) GetBySlug(s string, u *User, pc PublishedCriteria) (*Article, error) {
	a, err := as.Datasource.GetBySlug(s, u, pc)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NotFound("article", err)
		}
		return nil, err
	}

	if u != nil {
		if !u.IsAdmin {
			if a.Author.ID != u.ID {
				return nil, httperror.PermissionDenied("view", "article", fmt.Errorf("could not get article %s user %d has no permission", a.Slug, u.ID))
			}
		}
	}

	return a, nil
}

// GetByID get a article by the id.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) GetByID(id int, u *User, pc PublishedCriteria) (*Article, error) {
	a, err := as.Datasource.Get(id, u, pc)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperror.NotFound("article", fmt.Errorf("the article with id %d was not found", id))
		}
		return nil, err
	}

	if u != nil {
		if !u.IsAdmin {
			if a.Author.ID != u.ID {
				return nil, httperror.PermissionDenied("get", "article", fmt.Errorf("could not get article %d user %d has no permission", a.ID, u.ID))
			}
		}
	}

	return a, nil
}

// Count returns the number of articles.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) Count(u *User, c *Category, pc PublishedCriteria) (int, error) {
	return as.Datasource.Count(u, c, pc)
}

// List returns all article by the slug.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) List(u *User, c *Category, p *Pagination, pc PublishedCriteria) ([]Article, error) {
	return as.Datasource.List(u, c, p, pc)
}

// RSSFeed receives a specified number of articles in RSS
func (as ArticleService) RSSFeed(p *Pagination, pc PublishedCriteria) (RSS, error) {
	c := RSSChannel{
		Title:       as.AppConfig.Title,
		Link:        as.AppConfig.Domain,
		Description: as.AppConfig.Description,
		Language:    as.AppConfig.Language,
	}

	//TODO: categories in rss feeds
	articles, err := as.Datasource.List(nil, nil, p, pc)

	if err != nil {
		return RSS{}, err
	}

	var items []RSSItem

	for _, a := range articles {
		link := fmt.Sprint(as.AppConfig.Domain, "/article/by-id/", a.ID)
		item := RSSItem{
			GUID:        link,
			Link:        link,
			Title:       EscapeHTML(a.Headline),
			Author:      fmt.Sprintf("%s (%s)", EscapeHTML(a.Author.Email), EscapeHTML(a.Author.DisplayName)),
			Description: NewlineToBr(EscapeHTML(a.Teaser)),
			PubDate:     RSSTime(a.PublishedOn.Time),
		}

		items = append(items, item)
	}

	c.Items = items

	return RSS{
		Version: "2.0",
		Channel: c,
	}, nil
}

type IndexArticle struct {
	Year     int
	Articles []Article
}

func (as ArticleService) Index(u *User, c *Category, p *Pagination, pc PublishedCriteria) ([]IndexArticle, error) {
	articles, err := as.Datasource.List(u, c, p, pc)

	if err != nil {
		return nil, err
	}

	var keys []int
	amap := make(map[int][]Article)

	for _, v := range articles {
		if v.PublishedOn.Valid {
			year := v.PublishedOn.Time.Year()

			amap[year] = append(amap[year], v)
			keys = append(keys, year)
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	var ia []IndexArticle

	for _, year := range keys {
		v, _ := amap[year]

		a := IndexArticle{
			Year:     year,
			Articles: v,
		}
		ia = append(ia, a)
	}

	return ia, nil
}
