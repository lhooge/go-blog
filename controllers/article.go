// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers

import (
	"fmt"
	"net/http"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/utils"
)

//GetArticleHandler returns a specific article
//Parameters in the url form 2016/03/my-headline are used for obtaining the article
func GetArticleHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	year := getVar(r, "year")
	month := getVar(r, "month")
	slug := getVar(r, "slug")

	a, err := ctx.ArticleService.GetArticleBySlug(utils.AppendString(year, "/", month, "/", slug), nil, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name: tplArticle,
			Err:  err,
		}
	}

	c, err := ctx.CategoryService.ListCategories(models.CategoriesWithArticles)

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

//GetArticleHandler returns a specific article
//Parameters in the url form 2016/03/my-headline are used for obtaining the article
func GetArticleByIDHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	id, err := parseInt(getVar(r, "articleID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  httperror.ParameterMissing("articleID", err),
		}
	}

	a, err := ctx.ArticleService.GetArticleByID(id, nil, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name: tplArticle,
			Err:  err,
		}
	}

	c, err := ctx.CategoryService.ListCategories(models.CategoriesWithArticles)

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

//ListArticlesHandler returns the template which contains all published articles
func ListArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	page := getPageParam(r)

	t, err := ctx.ArticleService.CountArticles(nil, models.OnlyPublished)

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

	a, err := ctx.ArticleService.ListArticles(nil, p, models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	c, err := ctx.CategoryService.ListCategories(models.CategoriesWithArticles)

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

//IndexArticlesHandler returns the template information for the index page
func IndexArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	a, err := ctx.ArticleService.IndexArticles(nil, nil, models.OnlyPublished)

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
			"articles": a,
		},
	}
}

//GetArticleHandler returns a specific article
//Parameters in the url form 2016/03/my-headline are used for obtaining the article
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

//AdminListArticlesHandler returns all articles, also not yet published articles will be shown
func AdminListArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	t, err := ctx.ArticleService.CountArticles(u, models.All)

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

	a, err := ctx.ArticleService.ListArticles(u, p, models.All)

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

// AdminArticleNewHandler returns the template which shows the form to create a new article
func AdminArticleNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	c, err := ctx.CategoryService.ListCategories(models.AllCategories)

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

	cid, err := parseInt(r.FormValue("category"))

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticleNew,
			Active: "articles",
			Err:    err,
		}
	}

	c := &models.Category{
		ID: cid,
	}

	a := &models.Article{
		Headline: r.FormValue("headline"),
		Teaser:   r.FormValue("teaser"),
		Content:  r.FormValue("content"),
		Author:   u,
		Category: c,
	}

	if r.FormValue("action") == "preview" {
		return previewArticle(a)
	}

	id, err := ctx.ArticleService.CreateArticle(a)

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

//AdminArticleEditHandler shows the form for changing an article
func AdminArticleEditHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	id, err := parseInt(getVar(r, "articleID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  httperror.ParameterMissing("articleID", err),
		}
	}

	a, err := ctx.ArticleService.GetArticleByID(id, u, models.All)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminArticleEdit,
		Active: "articles",
		Data: map[string]interface{}{
			"article": a,
		},
	}
}

//AdminArticleEditPostHandler handles the update of an article
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

	if r.FormValue("action") == "preview" {
		return previewArticle(a)
	}

	if err = ctx.ArticleService.UpdateArticle(a, u); err != nil {
		return &middleware.Template{
			Name:   tplAdminArticleEdit,
			Err:    err,
			Active: "articles",
			Data: map[string]interface{}{
				"article": a,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
		Active:       "articles",
		SuccessMsg:   "Article successfully updated",
	}
}

//AdminArticlePublishHandler returns the action template which asks the user if the article should be published / unpublished
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

	a, err := ctx.ArticleService.GetArticleByID(id, u, models.All)

	var action models.Action

	if a.Published {
		action = models.Action{
			ID:          "unpublishSite",
			ActionURL:   fmt.Sprintf("/admin/article/publish/%d", a.ID),
			BackLinkURL: "/admin/articles",
			Description: fmt.Sprintf("%s %s?", "Do you want to unpublish the article ", a.Headline),
			Title:       "Confirm unpublishing of article",
		}
	} else {
		action = models.Action{
			ID:          "publishSite",
			ActionURL:   fmt.Sprintf("/admin/article/publish/%d", a.ID),
			BackLinkURL: "/admin/articles",
			Description: fmt.Sprintf("%s %s?", "Do you want to publish the article ", a.Headline),
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

// AdminArticlePublishPostHandler publishes or "depublishes" an article
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

	if err := ctx.ArticleService.PublishArticle(id, u); err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
		Active:       "articles",
		SuccessMsg:   "Article successfully published",
	}
}

//AdminArticleDeleteHandler returns the action template which asks the user if the article should be removed
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

	a, err := ctx.ArticleService.GetArticleByID(id, u, models.All)

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
		Description: fmt.Sprintf("%s %s?", "Do you want to delete the article", a.Headline),
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

//AdminArticleDeletePostHandler handles the removing of an article
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

	err = ctx.ArticleService.DeleteArticle(id, u)
	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	return &middleware.Template{
		Active:       "articles",
		RedirectPath: "admin/articles",
		SuccessMsg:   "Article successfully deleted",
	}
}

func previewArticle(a *models.Article) *middleware.Template {
	return &middleware.Template{
		Name: tplArticle,
		Data: map[string]interface{}{
			"article": a,
		},
	}
}
