// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	"git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

func TestCreateGetArticle(t *testing.T) {
	expectedArticle := getSampleArticle()

	artID, err := doCreateArticleRequest(expectedArticle)

	if err != nil {
		t.Fatal(err)
	}

	rcvArticle, err2 := ctx.ArticleService.GetArticleByID(artID, dummyUser(), models.All)
	if err2 != nil {
		t.Fatal(err)
	}

	if err = checkArticle(rcvArticle, expectedArticle); err != nil {
		t.Fatal(err)
	}

	rcvArticle, err = doGetArticleRequest(rcvArticle)

	if err != nil {
		t.Fatal(err)
	}

	if err = checkArticle(rcvArticle, expectedArticle); err != nil {
		t.Fatal(err)
	}

	expectedArticle = &models.Article{
		ID:       artID,
		Slug:     rcvArticle.Slug,
		Headline: "a new headline",
		Teaser:   "A sample teaser",
		Content:  "A new h1 header\n============\nthis is sample new content...",
	}

	if err := doEditArticleRequest(artID, expectedArticle); err != nil {
		t.Fatal(err)
	}

	rcvArticle, err = doGetArticleRequest(expectedArticle)

	if err != nil {
		t.Fatal(err)
	}

	if err = checkArticle(rcvArticle, expectedArticle); err != nil {
		t.Fatal(err)
	}
}

func checkArticle(article *models.Article, expectedArticle *models.Article) error {
	if article.Headline != expectedArticle.Headline {
		return fmt.Errorf("got an unexpected headline. expected: %s, actual: %s", expectedArticle.Headline, article.Headline)
	}
	if article.Content != expectedArticle.Content {
		return fmt.Errorf("got an unexpected content. expected: %s, actual: %s", expectedArticle.Content, article.Content)
	}
	if article.Published != expectedArticle.Published {
		return fmt.Errorf("the article published differs. expected: %t, actual: %t", expectedArticle.Published, article.Published)
	}
	if article.Author.ID != dummyUser().ID {
		return fmt.Errorf("the author id is wrong. expected: %d, actual: %d", expectedArticle.Author.ID, article.Author.ID)
	}
	return nil
}

func getSampleArticle() *models.Article {
	return &models.Article{
		Headline: "a sample headline",
		Teaser:   "A sample teaser",
		Content:  "An h1 header\n============\nthis is sample content...",
	}
}

func doEditArticleRequest(articleID int, article *models.Article) error {
	values := url.Values{}
	setValues(values, "headline", article.Headline)
	setValues(values, "teaser", article.Teaser)
	setValues(values, "content", article.Content)

	req, err := postRequest("/admin/article/edit", values)
	if err != nil {
		return err
	}

	setHeader(req, "articleID", strconv.Itoa(articleID))

	parent := req.Context()
	reqCtx := context.WithValue(parent, middleware.UserContextKey, dummyUser())

	rw := httptest.NewRecorder()

	tpl := controllers.AdminArticleEditPostHandler(ctx, rw, req.WithContext(reqCtx))

	if tpl.Err != nil {
		return tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return fmt.Errorf("The success message is empty")
	}
	return nil
}

func doCreateArticleRequest(article *models.Article) (int, error) {
	values := url.Values{}
	setValues(values, "headline", article.Headline)
	setValues(values, "teaser", article.Teaser)
	setValues(values, "content", article.Content)

	req, err := postRequest("/admin/article/new", values)
	if err != nil {
		return 0, err
	}

	reqCtx := context.WithValue(req.Context(), middleware.UserContextKey, dummyUser())
	rw := httptest.NewRecorder()
	tpl := controllers.AdminArticleNewPostHandler(ctx, rw, req.WithContext(reqCtx))

	if tpl.Err != nil {
		return 0, tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return -1, fmt.Errorf("there is no success message returned")
	}
	return tpl.Data["articleID"].(int), nil
}

func doGetArticleRequest(article *models.Article) (*models.Article, error) {
	split := strings.Split(article.Slug, "/")
	req, err := getRequest("/article", nil)
	setHeader(req, "year", split[0])
	setHeader(req, "month", split[1])
	setHeader(req, "slug", split[2])

	if err != nil {
		return nil, err
	}

	rw := httptest.NewRecorder()
	tpl := controllers.GetArticleHandler(ctx, rw, req)

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["article"].(*models.Article), nil
}

type inMemoryArticle struct {
	sync.RWMutex
	articles map[int]*models.Article
}

func (ima *inMemoryArticle) Create(a *models.Article) (int, error) {
	ima.Lock()
	defer ima.Unlock()
	artID := len(ima.articles) + 1
	ima.articles[artID] = a
	return artID, nil
}

func (ima *inMemoryArticle) List(user *models.User, pg *models.Pagination, pc models.PublishedCriteria) ([]models.Article, error) {
	ima.RLock()
	defer ima.RUnlock()

	arts := make([]models.Article, 0, len(ima.articles))

	for _, a := range ima.articles {
		arts = append(arts, *a)
	}
	return arts, nil
}

func (ima *inMemoryArticle) Count(user *models.User, publishedCriteria models.PublishedCriteria) (int, error) {
	return -1, nil
}

func (ima *inMemoryArticle) Publish(a *models.Article) error {
	return nil
}

func (ima *inMemoryArticle) Get(articleID int, user *models.User, pc models.PublishedCriteria) (*models.Article, error) {
	ima.RLock()
	defer ima.RUnlock()

	if k, ok := ima.articles[articleID]; ok {
		return k, nil
	}

	return nil, sql.ErrNoRows
}

func (ima *inMemoryArticle) GetBySlug(slug string, user *models.User, pc models.PublishedCriteria) (*models.Article, error) {
	ima.RLock()
	defer ima.RUnlock()

	for _, m := range ima.articles {
		if m.Slug == slug {
			return m, nil
		}
	}

	return nil, sql.ErrNoRows
}
func (ima *inMemoryArticle) Update(a *models.Article) error {
	ima.RLock()
	if k, ok := ima.articles[a.ID]; ok {
		ima.RUnlock()
		ima.Lock()
		a.Slug = k.Slug
		ima.articles[a.ID] = a
		ima.Unlock()
		return nil
	}
	return sql.ErrNoRows
}
func (ima *inMemoryArticle) Delete(articleID int) error {
	ima.Lock()
	defer ima.Unlock()
	delete(ima.articles, articleID)
	return nil
}
