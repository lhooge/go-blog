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

const (
	tplArticle       = "front/article"
	tplArticles      = "front/articles"
	tplIndexArticles = "front/index"

	tplAdminArticles    = "admin/articles"
	tplAdminArticleNew  = "admin/article_add"
	tplAdminArticleEdit = "admin/article_edit"
)

//GetArticleHandler returns a specific article
//Parameters in the url form 2016/03/my-headline are used for obtaining the article
func GetArticleHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	year := getVar(r, "year")
	month := getVar(r, "month")
	slug := getVar(r, "slug")

	article, err := ctx.ArticleService.GetArticleBySlug(nil, utils.AppendString(year, "/", month, "/", slug), models.OnlyPublished)

	if err != nil {
		return &middleware.Template{
			Name: tplArticle,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name: tplArticle,
		Data: map[string]interface{}{
			"article": article,
		}}
}

//ListArticlesHandler returns the template which contains all published articles
func ListArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	page := getPageParam(r)

	total, err := ctx.ArticleService.CountArticles(nil, models.OnlyPublished)

	pagination := &models.Pagination{
		Total:       total,
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

	article, err := ctx.ArticleService.ListArticles(nil, pagination, models.OnlyPublished)

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
			"articles":   article,
			"pagination": pagination,
		}}
}

//IndexArticlesHandler returns the template information for the index page
func IndexArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	article, err := ctx.ArticleService.IndexArticles(nil, nil, models.OnlyPublished)

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
			"indexed_articles": article,
		}}
}

//AdminListArticlesHandler returns all articles, also not yet published articles will be shown
func AdminListArticlesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	page := getPageParam(r)

	total, err := ctx.ArticleService.CountArticles(user, models.All)

	if err != nil {
		return &middleware.Template{
			Active: "articles",
			Name:   tplAdminArticles,
			Err:    err,
		}
	}

	pagination := &models.Pagination{
		Total:       total,
		Limit:       20,
		CurrentPage: page,
		RelURL:      "admin/articles/page",
	}

	articles, err := ctx.ArticleService.ListArticles(user, pagination, models.All)

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
			"articles":   articles,
			"pagination": pagination,
		}}
}

// AdminArticleNewHandler returns the template which shows the form to create a new article
func AdminArticleNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Active: "articles",
		Name:   tplAdminArticleNew,
	}
}

// AdminArticleNewPostHandler handles the creation of a new article
func AdminArticleNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	article := &models.Article{
		Headline: r.FormValue("headline"),
		Teaser:   r.FormValue("teaser"),
		Content:  r.FormValue("content"),
		Author:   user,
	}

	if r.FormValue("action") == "preview" {
		return previewArticle(article)
	}

	articleID, err := ctx.ArticleService.CreateArticle(article)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticleNew,
			Active: "articles",
			Err:    err,
			Data: map[string]interface{}{
				"article": article,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
		Active:       "articles",
		SuccessMsg:   "Article successfully saved",
		Data: map[string]interface{}{
			"articleID": articleID,
		}}
}

//AdminArticleEditHandler shows the form for changing an article
func AdminArticleEditHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	articleID, err := parseInt(getVar(r, "articleID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  httperror.ParameterMissing("articleID", err),
		}
	}

	article, err := ctx.ArticleService.GetArticleByID(user, articleID, models.All)

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
			"article": article,
		}}
}

//AdminArticleEditPostHandler handles the update of an article
func AdminArticleEditPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	articleID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplArticles,
			Active: "articles",
			Err:    err,
		}
	}

	article := &models.Article{
		ID:       articleID,
		Headline: r.FormValue("headline"),
		Teaser:   r.FormValue("teaser"),
		Content:  r.FormValue("content"),
		Author:   user,
	}

	if r.FormValue("action") == "preview" {
		return previewArticle(article)
	}

	if err = ctx.ArticleService.UpdateArticle(article, user); err != nil {
		return &middleware.Template{
			Name:   tplAdminArticleEdit,
			Err:    err,
			Active: "articles",
			Data: map[string]interface{}{
				"article": article,
			}}
	}

	return &middleware.Template{
		RedirectPath: "admin/articles",
		Active:       "articles",
		SuccessMsg:   "Article successfully updated",
	}
}

//AdminArticlePublishHandler returns the action template which asks the user if the article should be published / unpublished
func AdminArticlePublishHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	articleID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticles,
			Err:    err,
			Active: "articles",
		}
	}

	a, err := ctx.ArticleService.GetArticleByID(user, articleID, models.All)

	var publishInfo models.Action

	if a.Published {
		publishInfo = models.Action{
			ID:          "unpublishSite",
			ActionURL:   fmt.Sprintf("/admin/article/publish/%d", a.ID),
			BackLinkURL: "/admin/articles",
			Description: fmt.Sprintf("%s %s?", "Do you want to unpublish the article ", a.Headline),
			Title:       "Confirm unpublishing of article",
		}
	} else {
		publishInfo = models.Action{
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
			"action": publishInfo,
		},
	}
}

// AdminArticlePublishPostHandler publishes or "depublishes" an article
func AdminArticlePublishPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	articleID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	if err := ctx.ArticleService.PublishArticle(articleID, user); err != nil {
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
	user, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	articleID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticles,
			Err:    err,
			Active: "articles",
		}
	}

	article, err := ctx.ArticleService.GetArticleByID(user, articleID, models.All)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminArticles,
			Err:    err,
			Active: "articles",
		}
	}

	deleteInfo := models.Action{
		ID:          "deleteArticle",
		ActionURL:   fmt.Sprintf("/admin/article/delete/%d", article.ID),
		BackLinkURL: "/admin/articles",
		Description: fmt.Sprintf("%s %s?", "Do you want to delete the article", article.Headline),
		Title:       "Confirm removal of article",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "articles",
		Data: map[string]interface{}{
			"action": deleteInfo,
		},
	}
}

//AdminArticleDeletePostHandler handles the removing of an article
func AdminArticleDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	reqVar := getVar(r, "articleID")

	articleID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/articles",
			Err:          err,
		}
	}

	err = ctx.ArticleService.DeleteArticle(articleID, user)
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

func previewArticle(article *models.Article) *middleware.Template {
	return &middleware.Template{
		Name: tplArticle,
		Data: map[string]interface{}{
			"article": article,
		},
	}
}
