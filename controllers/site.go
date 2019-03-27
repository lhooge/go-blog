// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

//GetSiteHandler returns the site template - only published sites are considered
func GetSiteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	site, err := ctx.SiteService.GetByLink(getVar(r, "site"), models.OnlyPublished)

	if err != nil {
		if err == sql.ErrNoRows {
			return &middleware.Template{
				Name: tplSite,
				Err:  httperror.NotFound("site", err),
			}
		}
		return &middleware.Template{
			Name: tplSite,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name: tplSite,
		Data: map[string]interface{}{
			"site": site,
		},
	}
}

//AdminGetSiteHandler returns the template containing the sites
func AdminGetSiteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "siteID")

	id, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name: tplSite,
			Err:  err,
		}
	}

	site, err := ctx.SiteService.GetByID(id, models.All)

	if err != nil {
		if err == sql.ErrNoRows {
			return &middleware.Template{
				Name: tplSite,
				Err:  httperror.NotFound("site", err),
			}
		}
		return &middleware.Template{
			Name: tplSite,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name: tplSite,
		Data: map[string]interface{}{
			"site": site,
		},
	}
}

//AdminSitesHandler returns the template containing the sites overview in the administration
func AdminSitesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	page := getPageParam(r)

	total, err := ctx.SiteService.Count(models.All)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Active: "sites",
			Err:    err,
		}
	}

	pagination := &models.Pagination{
		Total:       total,
		Limit:       20,
		CurrentPage: page,
		RelURL:      "admin/sites/page",
	}

	sites, err := ctx.SiteService.List(models.All, pagination)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
			Data: map[string]interface{}{
				"sites":      sites,
				"pagination": pagination,
			},
		}
	}

	return &middleware.Template{
		Name:   tplAdminSites,
		Active: "sites",
		Data: map[string]interface{}{
			"sites":      sites,
			"pagination": pagination,
		},
	}
}

//AdminSiteNewHandler returns the template for adding a new site
func AdminSiteNewHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name:   tplAdminSiteNew,
		Active: "sites",
	}
}

//AdminSiteNewPostHandler receives the form values and creating the site; on success the user is redirected with a success message
//to the site overview
func AdminSiteNewPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	s := &models.Site{
		Title:     r.FormValue("title"),
		Link:      r.FormValue("link"),
		Content:   r.FormValue("content"),
		Section:   r.FormValue("section"),
		Published: false,
		Author:    user,
	}

	if r.FormValue("action") == "preview" {
		return previewSite(s)
	}

	siteID, err := ctx.SiteService.Create(s)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSiteNew,
			Err:    err,
			Active: "sites",
			Data: map[string]interface{}{
				"site": s,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/sites",
		Active:       "sites",
		SuccessMsg:   "Successfully added site " + s.Title,
		Data: map[string]interface{}{
			"siteID": siteID,
		},
	}
}

//AdminSiteEditHandler returns the template for editing an existing site
func AdminSiteEditHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	siteID, err := parseInt(getVar(r, "siteID"))

	if err != nil {
		return &middleware.Template{
			Name: tplAdminSites,
			Err:  err,
		}
	}

	s, err := ctx.SiteService.GetByID(siteID, models.All)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminSites,
			Err:  err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminSiteEdit,
		Active: "sites",
		Data: map[string]interface{}{
			"site": s,
		},
	}
}

//AdminSiteEditPostHandler receives the form values and updates the site; on success the user is redirected with a success message
//to the site overview
func AdminSiteEditPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	siteID, err := parseInt(getVar(r, "siteID"))

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Active: "sites",
			Err:    err,
		}
	}

	s := &models.Site{
		ID:      siteID,
		Title:   r.FormValue("title"),
		Link:    r.FormValue("link"),
		Content: r.FormValue("content"),
		Section: r.FormValue("section"),
		Author:  u,
	}

	if r.FormValue("action") == "preview" {
		return previewSite(s)
	}

	if err := ctx.SiteService.Update(s); err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
			Data: map[string]interface{}{
				"site": s,
			},
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/sites",
		Active:       "sites",
		SuccessMsg:   fmt.Sprintf("%s %s.", "Successfully edited site", s.Title),
	}
}

//AdminSiteOrderHandler moves the site with site id down or up
func AdminSiteOrderHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	siteID, err := parseInt(getVar(r, "siteID"))

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Active: "sites",
			Err:    err,
		}
	}

	order := r.FormValue("direction")

	var d models.Direction
	if order == "up" {
		d = models.Up
	} else if order == "down" {
		d = models.Down
	} else {
		return &middleware.Template{
			Name:   tplAdminSites,
			Active: "sites",
			Err:    errors.New("invalid"),
		}
	}

	if err := ctx.SiteService.Order(siteID, d); err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/sites",
		Active:       "sites",
		SuccessMsg:   "Site successfully reordered.",
	}
}

//AdminSitePublishHandler returns the action template which asks the user if the site should be published / unpublished
func AdminSitePublishHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "siteID")

	siteID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	s, err := ctx.SiteService.GetByID(siteID, models.All)

	if err != nil {
		return &middleware.Template{
			Name: tplAdminSites,
			Err:  err,
		}
	}

	var publishInfo models.Action

	if s.Published {
		publishInfo = models.Action{
			ID:          "unpublishSite",
			ActionURL:   fmt.Sprintf("/admin/site/publish/%d", s.ID),
			BackLinkURL: "/admin/sites",
			Description: fmt.Sprintf("%s %s?", "Do you want to unpublish the site", s.Title),
			Title:       "Confirm unpublishing of site",
		}
	} else {
		publishInfo = models.Action{
			ID:          "publishSite",
			ActionURL:   fmt.Sprintf("/admin/site/publish/%d", s.ID),
			BackLinkURL: "/admin/sites",
			Description: fmt.Sprintf("%s %s?", "Do you want to publish the site", s.Title),
			Title:       "Confirm publishing of site",
		}
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "sites",
		Data: map[string]interface{}{
			"action": publishInfo,
		},
	}
}

//AdminSitePublishPostHandler handles the un-/publishing of a site
func AdminSitePublishPostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "siteID")

	siteID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	if err := ctx.SiteService.Publish(siteID); err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/sites",
		Active:       "sites",
		SuccessMsg:   "Site successfully published.",
	}
}

//AdminSiteDeleteHandler returns the action template which asks the user if the site should be removed
func AdminSiteDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "siteID")

	siteID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	site, err := ctx.SiteService.GetByID(siteID, models.All)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	deleteInfo := models.Action{
		ID:          "deleteSite",
		ActionURL:   fmt.Sprintf("/admin/site/delete/%d", site.ID),
		BackLinkURL: "/admin/sites",
		Description: fmt.Sprintf("%s %s?", "Do you want to delete the site ", site.Title),
		Title:       "Confirm removal of site",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "sites",
		Data: map[string]interface{}{
			"action": deleteInfo,
		},
	}
}

//AdminSiteDeletePostHandler handles the removing of a site
func AdminSiteDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	reqVar := getVar(r, "siteID")

	siteID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	err = ctx.SiteService.Delete(siteID)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminSites,
			Err:    err,
			Active: "sites",
		}
	}

	return &middleware.Template{
		RedirectPath: "admin/sites",
		Active:       "sites",
		SuccessMsg:   "Site successfully deleted.",
	}
}

func previewSite(s *models.Site) *middleware.Template {
	return &middleware.Template{
		Name: tplSite,
		Data: map[string]interface{}{
			"site": s,
		},
	}
}
