// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/middleware"
	"git.hoogi.eu/snafu/go-blog/models"
)

// GetArticleHandler returns a specific article
// URL Parameters are in the form /year/month/slug e.g. 2016/3/my-
func GetArticleHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	year := getVar(r, "year")
	month := getVar(r, "month")
	headline := getVar(r, "slug")

	slug := year + "/" + month + "/" + headline

	a, err := ctx.ArticleService.GetBySlug(slug, nil, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name: tplArticle,
			Err:  err,
		}
	}

	c, err := ctx.CategoryService.List(models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	return &middleware.Template{
		Name: tplArticle,
		Data: map[string]interface{}{
			"article":    a,
			"categories": c,
		}}
}

// GetArticleByIDHandler returns a specific article by the ID
func GetArticleByIDHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	id, err := parseInt(getVar(r, "articleID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  httperror.ParameterMissing("articleID", err),
		}
	}

	a, err := ctx.ArticleService.GetByID(id, nil, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name: tplArticle,
			Err:  err,
		}
	}

	c, err := ctx.CategoryService.List(models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	return &middleware.Template{
		Name: tplArticle,
		Data: map[string]interface{}{
			"article":    a,
			"categories": c,
		}}
}

// ListArticlesHandler returns all published articles
func ListArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	page := getPageParam(r)

	t, err := ctx.ArticleService.Count(nil, nil, models.OnlyPublished)

	p := &models.Pagination{
		Total:       t,
		Limit:       ctx.ConfigService.ArticlesPerPage,
		CurrentPage: page,
		RelURL:      "articles/page",
	}

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	a, err := ctx.ArticleService.List(nil, nil, p, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	c, err := ctx.CategoryService.List(models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	return &middleware.Template{
		Name:   tplArticles,
		Active: "articles",
		Data: map[string]interface{}{
			"articles":   a,
			"categories": c,
			"pagination": p,
		},
	}
}

// ListArticlesCategoryHandler returns all published articles in a category
func ListArticlesCategoryHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	page := getPageParam(r)

	category := getVar(r, "categorySlug")

	c, err := ctx.CategoryService.GetBySlug(category, models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	t, err := ctx.ArticleService.Count(nil, c, models.OnlyPublished)

	p := &models.Pagination{
		Total:       t,
		Limit:       ctx.ConfigService.ArticlesPerPage,
		CurrentPage: page,
		RelURL:      "articles/page",
	}

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	a, err := ctx.ArticleService.List(nil, c, p, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	cs, err := ctx.CategoryService.List(models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	return &middleware.Template{
		Name:   tplArticles,
		Active: "articles",
		Data: map[string]interface{}{
			"articles":   a,
			"categories": cs,
			"catActive":  c.SlugEscape(),
			"pagination": p,
		},
	}
}

// IndexArticlesHandler returns articles for the index page
func IndexArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	a, err := ctx.ArticleService.Index(nil, nil, nil, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name:   tplIndexArticles,
			Active: "index",
			Err:    err,
		}
	}

	c, err := ctx.CategoryService.List(models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplIndexArticles,
			Active: "index",
			Err:    err,
		}
	}

	return &middleware.Template{
		Name:   tplIndexArticles,
		Active: "index",
		Data: map[string]interface{}{
			"articles":   a,
			"categories": c,
		},
	}
}

// IndexArticlesCategoryHandler returns articles for the index page grouped by categories
func IndexArticlesCategoryHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	cs, err := ctx.CategoryService.List(models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplIndexArticles,
			Active: "index",
			Err:    err,
		}
	}

	category := getVar(r, "categorySlug")

	c, err := ctx.CategoryService.GetBySlug(category, models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplIndexArticles,
			Active: "index",
			Err:    err,
			Data: map[string]interface{}{
				"categories": cs,
			},
		}
	}

	a, err := ctx.ArticleService.Index(nil, c, nil, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name:   tplIndexArticles,
			Active: "index",
			Err:    err,
			Data: map[string]interface{}{
				"categories": cs,
			},
		}
	}

	return &middleware.Template{
		Name:   tplIndexArticles,
		Active: "index",
		Data: map[string]interface{}{
			"articles":   a,
			"categories": cs,
		},
	}
}

// RSSFeed returns XML list of published articles for the RSS feed
func RSSFeed(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) (*models.XMLData, error) {
	p := &models.Pagination{
		Limit: ctx.ConfigService.RSSFeedItems,
	}

	rss, err := ctx.ArticleService.RSSFeed(p, models.OnlyPublished)

	if err != nil {
		return nil, err
	}

	return &models.XMLData{
		Data:      rss,
		HexEncode: true,
	}, nil
}

// AdminListArticlesHandler returns all articles, also not yet published articles
func AdminListArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	t, err := ctx.ArticleService.Count(u, nil, models.All)

	if err != nil {
		return &middleware.Template{
			Active: "articles",
			Name:   tplAdminArticles,
			Err:    err,
		}
	}

	p := &models.Pagination{
		Total:       t,
		Limit:       20,
		CurrentPage: getPageParam(r),
		RelURL:      "admin/articles/page",
	}

	a, err := ctx.ArticleService.List(u, nil, p, models.All)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticles,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminArticles,
		Active: "articles",
		Data: map[string]interface{}{
			"articles":   a,
			"pagination": p,
		}}
}

// AdminPreviewArticleByIDHandler returns a specific article, renders it on the front page used for the preview
func AdminPreviewArticleByIDHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	id, err := parseInt(getVar(r, "articleID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  httperror.ParameterMissing("articleID", err),
		}
	}

	a, err := ctx.ArticleService.GetByID(id, u, models.All)

	if err != nil {
		return &middleware.Template{
			Name: tplArticle,
			Err:  err,
		}
	}

	if a.Published == false {
		a.PublishedOn = models.NullTime{Valid: true, Time: a.LastModified}
	}

	c, err := ctx.CategoryService.List(models.CategoriesWithPublishedArticles)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	return &middleware.Template{
		Name: tplArticle,
		Data: map[string]interface{}{
			"article":    a,
			"categories": c,
		}}
}

// AdminArticleNewHandler returns the form to create a new article
func AdminArticleNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	c, err := ctx.CategoryService.List(models.AllCategories)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminCategories,
			Err:  err,
		}
	}

	return &middleware.Template{
		Active: "articles",
		Name:   tplAdminArticleNew,
		Data: map[string]interface{}{
			"categories": c,
		},
	}
}

// AdminArticleNewPostHandler handles the creation of a new article
func AdminArticleNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	a := &models.Article{
		Headline: r.FormValue("headline"),
		Teaser:   r.FormValue("teaser"),
		Content:  r.FormValue("content"),
		Author:   u,
	}

	cid, err := parseInt(r.FormValue("categoryID"))

	if err != nil {
		cid = 0
	}

	a.CID = sql.NullInt64{Int64: int64(cid), Valid: true}

	if r.FormValue("action") == "preview" {
		return previewArticle(a)
	}

	id, err := ctx.ArticleService.Create(a)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticleNew,
			Active: "articles",
			Err:    err,
			Data: map[string]interface{}{
				"article": a,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
		Active:       "articles",
		SuccessMsg:   "Article successfully saved",
		Data: map[string]interface{}{
			"articleID": id,
		},
	}
}

// AdminArticleEditHandler shows the form for changing an article
func AdminArticleEditHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	id, err := parseInt(getVar(r, "articleID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  httperror.ParameterMissing("articleID", err),
		}
	}

	a, err := ctx.ArticleService.GetByID(id, u, models.All)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  err,
		}
	}

	c, err := ctx.CategoryService.List(models.AllCategories)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminCategories,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminArticleEdit,
		Active: "articles",
		Data: map[string]interface{}{
			"article":    a,
			"categories": c,
		},
	}
}

// AdminArticleEditPostHandler handles the update of an article
func AdminArticleEditPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	a := &models.Article{
		ID:       id,
		Headline: r.FormValue("headline"),
		Teaser:   r.FormValue("teaser"),
		Content:  r.FormValue("content"),
		Author:   u,
	}

	updateSlug := convertCheckbox(r, "updateSlug")

	cid, err := parseInt(r.FormValue("categoryID"))

	if err != nil {
		cid = 0
	}

	a.CID = sql.NullInt64{Int64: int64(cid), Valid: true}

	if r.FormValue("action") == "preview" {
		return previewArticle(a)
	}

	if err = ctx.ArticleService.Update(a, u, updateSlug); err != nil {
		return &middleware.Template{
			Name:   tplAdminArticleEdit,
			Err:    err,
			Active: "articles",
			Data: map[string]interface{}{
				"article":    a,
				"updateSlug": updateSlug,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
		Active:       "articles",
		SuccessMsg:   "Article successfully updated",
	}
}

// AdminArticlePublishHandler returns the action template which asks the user if the article should be published / unpublished
func AdminArticlePublishHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticles,
			Err:    err,
			Active: "articles",
		}
	}

	a, err := ctx.ArticleService.GetByID(id, u, models.All)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticles,
			Err:    err,
			Active: "articles",
		}
	}

	var action models.Action

	if a.Published {
		action = models.Action{
			ID:          "unpublishArticle",
			ActionURL:   fmt.Sprintf("/admin/article/publish/%d", a.ID),
			BackLinkURL: "/admin/articles",
			Description: fmt.Sprintf("%s %s?", "Do you want to unpublish the article", a.Headline),
			Title:       "Confirm unpublishing of article",
		}
	} else {
		action = models.Action{
			ID:          "publishArticle",
			ActionURL:   fmt.Sprintf("/admin/article/publish/%d", a.ID),
			BackLinkURL: "/admin/articles",
			Description: fmt.Sprintf("%s %s?", "Do you want to publish the article", a.Headline),
			Title:       "Confirm publishing of article",
		}
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "articles",
		Data: map[string]interface{}{
			"action": action,
		},
	}
}

// AdminArticlePublishPostHandler publish or unpublish an article
func AdminArticlePublishPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	if err := ctx.ArticleService.Publish(id, u); err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
		Active:       "articles",
		SuccessMsg:   "Article successfully published.",
	}
}

// AdminArticleDeleteHandler returns the action template which asks the user if the article should be removed
func AdminArticleDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticles,
			Err:    err,
			Active: "articles",
		}
	}

	a, err := ctx.ArticleService.GetByID(id, u, models.All)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticles,
			Err:    err,
			Active: "articles",
		}
	}

	action := models.Action{
		ID:          "deleteArticle",
		ActionURL:   fmt.Sprintf("/admin/article/delete/%d", a.ID),
		BackLinkURL: "/admin/articles",
		Description: fmt.Sprintf("%s %s?", "Do you want to delete the article ", a.Headline),
		Title:       "Confirm removal of article",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "articles",
		Data: map[string]interface{}{
			"action": action,
		},
	}
}

// AdminArticleDeletePostHandler handles the removing of an article
func AdminArticleDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	err = ctx.ArticleService.Delete(id, u)
	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	return &middleware.Template{
		Active:       "articles",
		RedirectPath: "admin/articles",
		SuccessMsg:   "Article successfully deleted.",
	}
}

func previewArticle(a *models.Article) *middleware.Template {
	a.PublishedOn = models.NullTime{Time: time.Now(), Valid: true}

	return &middleware.Template{
		Name: tplArticle,
		Data: map[string]interface{}{
			"article": a,
		},
	}
}
