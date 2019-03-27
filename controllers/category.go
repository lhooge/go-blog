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
)

func AdminListCategoriesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	c, err := ctx.CategoryService.List(models.AllCategories)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminCategories,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminCategories,
		Active: "categories",
		Data: map[string]interface{}{
			"categories": c,
		}}
}

func AdminGetCategoryHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "categoryID")
	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminCategories,
			Err:  err,
		}
	}

	c, err := ctx.CategoryService.GetByID(id, models.AllCategories)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminCategories,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminCategories,
		Active: "categories",
		Data: map[string]interface{}{
			"category": c,
		}}
}

// AdminCategoryNewHandler returns the template which shows the form to create a new article
func AdminCategoryNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Active: "categories",
		Name:   tplAdminCategoryNew,
	}
}

// AdminCategoryNewPostHandler handles the creation of a new article
func AdminCategoryNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	c := &models.Category{
		Name:   r.FormValue("name"),
		Author: u,
	}

	id, err := ctx.CategoryService.Create(c)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminCategoryNew,
			Active: "categories",
			Err:    err,
			Data: map[string]interface{}{
				"category": c,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/categories",
		Active:       "categories",
		SuccessMsg:   "Category successfully saved.",
		Data: map[string]interface{}{
			"categoryID": id,
		},
	}
}

//AdminCategoryEditHandler shows the form for changing an article
func AdminCategoryEditHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	id, err := parseInt(getVar(r, "categoryID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminArticleEdit,
			Err:  httperror.ParameterMissing("categoryID", err),
		}
	}

	c, err := ctx.CategoryService.GetByID(id, models.AllCategories)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminCategoryEdit,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminCategoryEdit,
		Active: "categories",
		Data: map[string]interface{}{
			"category": c,
		},
	}
}

//AdminArticleEditPostHandler handles the update of an article
func AdminCategoryEditPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	reqVar := getVar(r, "categoryID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminCategories,
			Active: "categories",
			Err:    err,
		}
	}

	c := &models.Category{
		ID:     id,
		Name:   r.FormValue("name"),
		Author: u,
	}

	if err = ctx.CategoryService.Update(c); err != nil {
		return &middleware.Template{
			Name:   tplAdminCategoryEdit,
			Err:    err,
			Active: "categories",
			Data: map[string]interface{}{
				"category": c,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/categories",
		Active:       "categories",
		SuccessMsg:   "Category successfully updated.",
	}
}

//AdminArticleDeleteHandler returns the action template which asks the user if the article should be removed
func AdminCategoryDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "categoryID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminCategories,
			Err:    err,
			Active: "categories",
		}
	}

	c, err := ctx.CategoryService.GetByID(id, models.AllCategories)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminCategories,
			Err:    err,
			Active: "categories",
		}
	}

	action := models.Action{
		ID:          "deleteCategory",
		ActionURL:   fmt.Sprintf("/admin/category/delete/%d", c.ID),
		BackLinkURL: "/admin/categories",
		Description: fmt.Sprintf("Do you want to delete the category %s?", c.Name),
		Title:       "Confirm removal of article",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "categories",
		Data: map[string]interface{}{
			"action": action,
		},
	}
}

//AdminArticleDeletePostHandler handles the removing of an article
func AdminCategoryDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "categoryID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/categories",
			Err:          err,
		}
	}

	err = ctx.CategoryService.Delete(id)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "admin/categories",
			Err:          err,
		}
	}

	return &middleware.Template{
		Active:       "categories",
		RedirectPath: "admin/categories",
		SuccessMsg:   "Category successfully deleted",
	}
}
