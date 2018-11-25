package controllers_test

import (
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/models"
)

func TestCategoryWorkflow(t *testing.T) {
	setup(t)

	defer teardown()

	c := &models.Category{
		Name: "My Category",
	}

	id, err := doAdminCategoryNewRequest(rAdminUser, c)

	if err != nil {
		t.Fatal(err)
	}

	rcvCategory, err := doAdminGetCategoryRequest(rAdminUser, id)

	if err != nil {
		t.Fatal(err)
	}

	if rcvCategory.Name != c.Name {
		t.Fatalf("the category name is wrong. expected: %s, actual: %s", c.Name, rcvCategory.Name)
	}

	c.ID = id
	c.Name = "Updated Category"

	err = doAdminCategoryEditRequest(rAdminUser, c)

	if err != nil {
		t.Fatal(err)
	}

	rcvCategory, err = doAdminGetCategoryRequest(rAdminUser, id)

	if err != nil {
		t.Fatal(err)
	}

	if rcvCategory.Name != c.Name {
		t.Fatalf("the category name is wrong. expected: %s, actual: %s", c.Name, rcvCategory.Name)
	}

	err = doAdminDeleteCategoryRequest(rAdminUser, id)

	if err != nil {
		t.Fatal(err)
	}

	rcvCategory, err = doAdminGetCategoryRequest(rAdminUser, id)

	if err == nil {
		t.Fatalf("removed category, but got a category %v", rcvCategory)
	}

}

func doAdminGetCategoryRequest(user reqUser, categoryID int) (*models.Category, error) {
	r := request{
		url:    "/admin/category/" + strconv.Itoa(categoryID),
		user:   user,
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "categoryID",
				value: strconv.Itoa(categoryID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminGetCategoryHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["category"].(*models.Category), nil
}

func doAdminListCategoriesRequest(user reqUser) ([]models.Category, error) {
	r := request{
		url:    "/admin/categories/",
		user:   user,
		method: "GET",
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminListCategoriesHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["categories"].([]models.Category), nil
}

func doAdminCategoryNewRequest(user reqUser, c *models.Category) (int, error) {
	values := url.Values{}
	addValue(values, "name", c.Name)
	r := request{
		url:    "/admin/category/new",
		user:   user,
		method: "POST",
		values: values,
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminCategoryNewPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return -1, tpl.Err
	}
	return tpl.Data["categoryID"].(int), nil
}

func doAdminCategoryEditRequest(user reqUser, c *models.Category) error {
	values := url.Values{}
	addValue(values, "name", c.Name)
	r := request{
		url:    "/admin/category/edit/" + strconv.Itoa(c.ID),
		user:   user,
		method: "POST",
		values: values,
		pathVar: []pathVar{
			pathVar{
				key:   "categoryID",
				value: strconv.Itoa(c.ID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminCategoryEditPostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}
	return nil
}

func doAdminDeleteCategoryRequest(user reqUser, categoryID int) error {
	r := request{
		url:    "/admin/category/" + strconv.Itoa(categoryID),
		user:   user,
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "categoryID",
				value: strconv.Itoa(categoryID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := controllers.AdminCategoryDeletePostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}
	return nil
}
