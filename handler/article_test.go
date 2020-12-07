// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handler_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"

	"strconv"
	"testing"

	"git.hoogi.eu/snafu/go-blog/handler"
	"git.hoogi.eu/snafu/go-blog/models"
)

func TestArticleWorkflow(t *testing.T) {
	setup(t)

	defer teardown()

	artID, err := doAdminCreateArticleRequest(rAdminUser, getSampleArticle())

	if err != nil {
		t.Fatal(err)
	}

	rcvArticle, err2 := doAdminGetArticleByIDRequest(rAdminUser, artID)

	if err2 != nil {
		t.Fatal(err)
	}

	if err = checkArticle(rcvArticle, getSampleArticle()); err != nil {
		t.Fatal(err)
	}

	updatedArticle := &models.Article{
		ID:       artID,
		Slug:     rcvArticle.Slug,
		Headline: "a new headline",
		Teaser:   "A sample teaser",
		Content:  "A new h1 header\n============\nthis is sample new content...",
	}

	if err := doAdminEditArticleRequest(rAdminUser, artID, updatedArticle); err != nil {
		t.Fatal(err)
	}

	rcvArticle, err = doAdminGetArticleByIDRequest(rAdminUser, artID)

	if err != nil {
		t.Fatal(err)
	}

	//Guest user should not see unpublished articles
	rcvArticle, err = doGetArticleByIDRequest(rGuest, artID)
	if err == nil {
		t.Fatal(err)
	}

	err = doAdminPublishArticleRequest(rAdminUser, artID)

	if err != nil {
		t.Fatal(err)
	}

	rcvArticle, err = doGetArticleByIDRequest(rAdminUser, artID)

	if err != nil {
		t.Fatal(err)
	}

	updatedArticle.Published = true

	if err = checkArticle(rcvArticle, updatedArticle); err != nil {
		t.Fatal(err)
	}

	err = doAdminRemoveArticleRequest(rAdminUser, artID)

	if err != nil {
		t.Fatal(err)
	}

	rcvArticle, err = doGetArticleByIDRequest(rAdminUser, artID)

	if err == nil {
		t.Fatalf("removed article, but got a category %v", rcvArticle)
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
	if article.Author.ID != dummyAdminUser().ID {
		return fmt.Errorf("the author id is wrong. expected: %d, actual: %d", expectedArticle.Author.ID, article.Author.ID)
	}
	return nil
}

func getSampleArticle() *models.Article {
	return &models.Article{
		Headline: "a sample headline",
		Teaser:   "A sample teaser",
		Content:  "An h1 header\n============\nthis is sample content...",
		Author:   dummyAdminUser(),
	}
}

func doGetArticleBySlugRequest(user reqUser, article *models.Article) (*models.Article, error) {
	split := strings.Split(article.Slug, "/")

	if len(split) != 3 {
		return nil, fmt.Errorf("invalid slug length %v", article.Slug)
	}

	r := request{
		url:    "/articles/" + split[0] + "/" + split[1] + "/" + split[2] + "/",
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "year",
				value: split[0],
			}, pathVar{
				key:   "month",
				value: split[1],
			}, pathVar{
				key:   "slug",
				value: split[2],
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := handler.GetArticleHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["article"].(*models.Article), nil
}

func doGetArticleByIDRequest(user reqUser, articleID int) (*models.Article, error) {
	r := request{
		url:    "/article/by-id/" + strconv.Itoa(articleID),
		method: "GET",
		user:   user,
		pathVar: []pathVar{
			pathVar{
				key:   "articleID",
				value: strconv.Itoa(articleID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := handler.GetArticleByIDHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["article"].(*models.Article), nil
}

func doAdminEditArticleRequest(user reqUser, articleID int, article *models.Article) error {
	values := url.Values{}
	addValue(values, "headline", article.Headline)
	addValue(values, "teaser", article.Teaser)
	addValue(values, "content", article.Content)

	r := request{
		url:    "/admin/article/edit/" + strconv.Itoa(articleID),
		user:   user,
		method: "POST",
		values: values,
		pathVar: []pathVar{
			pathVar{
				key:   "articleID",
				value: strconv.Itoa(articleID),
			},
		},
	}

	rw := httptest.NewRecorder()

	tpl := handler.AdminArticleEditPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return fmt.Errorf("The success message is empty")
	}
	return nil
}

func doAdminCreateArticleRequest(user reqUser, article *models.Article) (int, error) {
	values := url.Values{}
	addValue(values, "headline", article.Headline)
	addValue(values, "teaser", article.Teaser)
	addValue(values, "content", article.Content)

	r := request{
		url:    "/admin/article/new",
		user:   user,
		method: "POST",
		values: values,
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminArticleNewPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return 0, tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return -1, fmt.Errorf("there is no success message returned")
	}

	return tpl.Data["articleID"].(int), nil
}

func doAdminGetArticleByIDRequest(user reqUser, articleID int) (*models.Article, error) {
	r := request{
		url:    "/admin/article" + strconv.Itoa(articleID),
		method: "GET",
		user:   user,
		pathVar: []pathVar{
			pathVar{
				key:   "articleID",
				value: strconv.Itoa(articleID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminPreviewArticleByIDHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["article"].(*models.Article), nil
}

func doAdminListArticleRequest(user reqUser) ([]models.Article, error) {
	r := request{
		url:    "/admin/articles",
		user:   user,
		method: "GET",
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminListArticlesHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["articles"].([]models.Article), nil
}

func doAdminPublishArticleRequest(user reqUser, articleID int) error {
	r := request{
		url:    "/admin/article/publish/" + strconv.Itoa(articleID),
		user:   user,
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "articleID",
				value: strconv.Itoa(articleID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminArticlePublishPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}

func doAdminRemoveArticleRequest(user reqUser, articleID int) error {
	r := request{
		url:    "/admin/article/remove/" + strconv.Itoa(articleID),
		user:   user,
		method: "POST",
		pathVar: []pathVar{
			pathVar{
				key:   "articleID",
				value: strconv.Itoa(articleID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminArticleDeletePostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}
