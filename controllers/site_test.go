package controllers_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/models"
)

func TestSiteWorkflow(t *testing.T) {
	setup(t)

	defer teardown()

	site := &models.Site{
		Title:   "test title",
		Link:    "a link",
		Content: "content",
		Section: "navigation",
	}

	id, err := doAdminSiteCreateRequest(rAdminUser, site)

	site.ID = id

	if err != nil {
		t.Error(err)
	}

	site, err = doAdminGetSiteRequest(rAdminUser, site.ID)

	if err != nil {
		t.Error(err)
	}

	_, err = doGetSiteRequest(rGuest, site.Link)

	if err == nil {
		t.Error("received an unpublished site as guest")
	}

	sites, err := doAdminListSitesRequest(rAdminUser)

	if err != nil {
		t.Error(err)
	}

	if len(sites) != 1 {
		t.Errorf("expected 1 site to be returned, bot got %d", len(sites))
	}

	err = doAdminSitePublishRequest(rAdminUser, site.ID)

	if err != nil {
		t.Error(err)
	}

	_, err = doGetSiteRequest(rGuest, site.Link)

	if err != nil {
		t.Error(err)
	}

	err = doAdminSiteDeleteRequest(rAdminUser, site.ID)

	if err != nil {
		t.Error(err)
	}

	_, err = doGetSiteRequest(rGuest, site.Link)

	if err == nil {
		t.Error("received an removed site")
	}
}

func doGetSiteRequest(user reqUser, link string) (*models.Site, error) {
	r := request{
		url:    "/site/" + link,
		user:   user,
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "site",
				value: link,
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.GetSiteHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["site"].(*models.Site), nil
}

func doAdminGetSiteRequest(user reqUser, siteID int) (*models.Site, error) {
	r := request{
		url:    "/admin/site/" + strconv.Itoa(siteID),
		user:   user,
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "siteID",
				value: strconv.Itoa(siteID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminGetSiteHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["site"].(*models.Site), nil
}

func doAdminListSitesRequest(user reqUser) ([]models.Site, error) {
	r := request{
		url:    "/admin/sites",
		user:   user,
		method: "GET",
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSitesHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["sites"].([]models.Site), nil
}

func doAdminSiteCreateRequest(user reqUser, s *models.Site) (int, error) {
	values := url.Values{}
	addValue(values, "title", s.Title)
	addValue(values, "link", s.Link)
	addValue(values, "content", s.Content)
	addValue(values, "section", s.Section)

	r := request{
		url:    "/admin/site/new",
		user:   user,
		method: "POST",
		values: values,
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSiteNewPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return 0, tpl.Err
	}

	if len(tpl.SuccessMsg) == 0 {
		return -1, fmt.Errorf("there is no success message returned")
	}

	return tpl.Data["siteID"].(int), nil

}

func doAdminSitePublishRequest(user reqUser, siteID int) error {

	r := request{
		url:    "/admin/site/publish/" + strconv.Itoa(siteID),
		user:   user,
		method: "POST",
		pathVar: []pathVar{
			pathVar{
				key:   "siteID",
				value: strconv.Itoa(siteID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSitePublishPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}

func doAdminSiteEditRequest(user reqUser, s *models.Site) error {
	values := url.Values{}
	addValue(values, "title", s.Title)
	addValue(values, "link", s.Link)
	addValue(values, "content", s.Content)
	addValue(values, "section", s.Section)

	r := request{
		url:    "/admin/site/edit" + strconv.Itoa(s.ID),
		user:   user,
		method: "POST",
		values: values,
		pathVar: []pathVar{
			pathVar{
				key:   "siteID",
				value: strconv.Itoa(s.ID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSiteEditPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}

func doAdminSiteDeleteRequest(user reqUser, siteID int) error {
	r := request{
		url:    "/admin/site/delete/" + strconv.Itoa(siteID),
		user:   user,
		method: "POST",
		pathVar: []pathVar{
			pathVar{
				key:   "siteID",
				value: strconv.Itoa(siteID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSiteDeletePostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}

func doAdminSiteOrderRequest(user reqUser, siteID int) error {
	r := request{
		url:    "/admin/site/publish/" + strconv.Itoa(siteID),
		user:   user,
		method: "POST",
		pathVar: []pathVar{
			pathVar{
				key:   "siteID",
				value: strconv.Itoa(siteID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminSiteOrderHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}
