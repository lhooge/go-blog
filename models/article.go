// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/go-blog/utils"
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
	Category     *Category
	Author       *User
}

//ArticleDatasourceService defines an interface for CRUD operations of articles
type ArticleDatasourceService interface {
	Create(a *Article) (int, error)
	List(u *User, p *Pagination, pc PublishedCriteria) ([]Article, error)
	Count(u *User, pc PublishedCriteria) (int, error)
	Get(articleID int, u *User, pc PublishedCriteria) (*Article, error)
	GetBySlug(slug string, u *User, pc PublishedCriteria) (*Article, error)
	Publish(a *Article) error
	Update(a *Article) error
	Delete(articleID int) error
}

const (
	maxHeadlineSize = 150
)

//SlugEscape escapes the slug for use in URLs
func (a Article) SlugEscape() string {
	spl := strings.Split(a.Slug, "/")
	return fmt.Sprintf("%s/%s/%s", spl[0], spl[1], url.PathEscape(spl[2]))
}

func (a Article) buildSafeSlug(now time.Time, suffix int) string {
	return utils.AppendString(strconv.Itoa(now.Year()), "/", strconv.Itoa(int(now.Month())), "/", utils.CreateURLSafeSlug(a.Headline, suffix))
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

//ArticleService containing the service to access articles
type ArticleService struct {
	Datasource ArticleDatasourceService
	AppConfig  settings.Application
}

// CreateArticle creates an article
func (as ArticleService) CreateArticle(a *Article) (int, error) {
	curTime := time.Now()

	for i := 0; i < 10; i++ {
		a.Slug = a.buildSafeSlug(curTime, i)
		_, err := as.Datasource.GetBySlug(a.Slug, nil, All)

		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return -1, err
		}
	}

	a.PublishedOn = NullTime{Time: curTime}
	a.Headline = strings.TrimSpace(a.Headline)

	if err := a.validate(); err != nil {
		return 0, err
	}

	artID, err := as.Datasource.Create(a)
	if err != nil {
		return 0, err
	}

	return artID, nil
}

//UpdateArticle updates an article
func (as ArticleService) UpdateArticle(a *Article, u *User) error {
	if err := a.validate(); err != nil {
		return err
	}

	art, err := as.Datasource.Get(a.ID, a.Author, All)
	if err != nil {
		return err
	}

	if !u.IsAdmin {
		if art.Author.ID != u.ID {
			return httperror.PermissionDenied("update", "article", fmt.Errorf("could not update article %d user %d has no permission", a.ID, u.ID))
		}
	}

	return as.Datasource.Update(a)
}

//PublishArticle publishes or 'unpublishes' an article
func (as ArticleService) PublishArticle(id int, u *User) error {
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

//DeleteArticle deletes an article
func (as ArticleService) DeleteArticle(id int, u *User) error {
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

// GetArticleBySlug gets a article by the slug.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) GetArticleBySlug(s string, u *User, pc PublishedCriteria) (*Article, error) {
	a, err := as.Datasource.GetBySlug(s, u, pc)

	if err != nil {
		if err == sql.ErrNoRows {
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

// GetArticleByID get a article by the id.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) GetArticleByID(id int, u *User, pc PublishedCriteria) (*Article, error) {
	a, err := as.Datasource.Get(id, u, pc)

	if err != nil {
		if err == sql.ErrNoRows {
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

// CountArticles returns the number of articles.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) CountArticles(u *User, pc PublishedCriteria) (int, error) {
	return as.Datasource.Count(u, pc)
}

// ListArticles returns all article by the slug.
// The publishedCriteria defines whether the published and/or unpublished articles should be considered
func (as ArticleService) ListArticles(u *User, p *Pagination, pc PublishedCriteria) ([]Article, error) {
	return as.Datasource.List(u, p, pc)
}

// RSSFeed receives a specified number of articles in RSS
func (as ArticleService) RSSFeed(p *Pagination, pc PublishedCriteria) (RSS, error) {
	c := RSSChannel{
		Title:       as.AppConfig.Title,
		Link:        as.AppConfig.Domain,
		Description: as.AppConfig.Description,
		Language:    as.AppConfig.Language,
	}

	articles, err := as.Datasource.List(nil, p, pc)

	if err != nil {
		return RSS{}, err
	}

	items := []RSSItem{}
	for _, a := range articles {
		fmt.Println(sanitize(a.Teaser))
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

func (as ArticleService) IndexArticles(u *User, p *Pagination, pc PublishedCriteria) ([]IndexArticle, error) {
	arts, err := as.Datasource.List(u, p, pc)

	if err != nil {
		return nil, err
	}

	var ias []IndexArticle
	var articles []Article

	idx := 0
	for i := 0; i < len(arts); i++ {
		if arts[i].PublishedOn.Valid {
			curYear := arts[i].PublishedOn.Time.Year()

			if i == len(arts)-1 {
				articles = append(articles, arts[i])

				ia := IndexArticle{
					Year:     arts[idx].PublishedOn.Time.Year(),
					Articles: articles,
				}
				ias = append(ias, ia)
				articles = nil
			} else if curYear == arts[idx].PublishedOn.Time.Year() {
				articles = append(articles, arts[i])
			} else {
				ia := IndexArticle{
					Year:     arts[idx].PublishedOn.Time.Year(),
					Articles: articles,
				}

				idx = i
				ias = append(ias, ia)

				articles = nil
			}
		}
	}

	return ias, nil
}
