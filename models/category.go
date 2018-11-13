package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/utils"
)

type Category struct {
	ID           int
	Name         string
	Slug         string
	LastModified time.Time
	Author       *User
}

type FilterCriteria int

const (
	CategoriesWithArticles = iota
	CategoriesWithoutArticles
	AllCategories
)

// validate validates if mandatory article fields are set
func (c *Category) validate() error {
	c.Name = strings.TrimSpace(c.Name)

	if len(c.Name) == 0 {
		return httperror.ValueRequired("name")
	}

	if c.Author == nil {
		return httperror.InternalServerError(errors.New("category validation failed - the author is missing"))
	}
	return nil
}

type CategoryDatasourceService interface {
	Create(c *Category) (int, error)
	List(fc FilterCriteria) ([]Category, error)
	Count(fc FilterCriteria) (int, error)
	Get(categoryID int) (*Category, error)
	GetBySlug(slug string) (*Category, error)
	Update(c *Category) error
	Delete(categoryID int) error
}

//CategoryService containing the service to access categories
type CategoryService struct {
	Datasource CategoryDatasourceService
}

//SlugEscape escapes the slug for use in URLs
func (c Category) SlugEscape() string {
	return url.PathEscape(c.Slug)
}

func (cs CategoryService) GetCategoryBySlug(s string) (*Category, error) {
	c, err := cs.Datasource.GetBySlug(s)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("category", fmt.Errorf("the category with slug %s was not found", s))
		}
		return nil, err
	}

	return c, nil
}

func (cs CategoryService) GetCategoryByID(id int) (*Category, error) {
	c, err := cs.Datasource.Get(id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("category", fmt.Errorf("the category with id %d was not found", id))
		}
		return nil, err
	}

	return c, nil
}

func (cs CategoryService) CountCategories(fc FilterCriteria) (int, error) {
	return cs.Datasource.Count(fc)
}

func (cs CategoryService) ListCategories(fc FilterCriteria) ([]Category, error) {
	return cs.Datasource.List(fc)
}

// CreateCategory creates a category
func (cs CategoryService) CreateCategory(c *Category) (int, error) {
	for i := 0; i < 10; i++ {
		c.Slug = utils.CreateURLSafeSlug(c.Slug, i)
		_, err := cs.Datasource.GetBySlug(c.Slug)

		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return -1, err
		}
	}

	if err := c.validate(); err != nil {
		return 0, err
	}

	cid, err := cs.Datasource.Create(c)
	if err != nil {
		return 0, err
	}

	return cid, nil
}

//UpdateCategory updates a category
func (cs CategoryService) UpdateCategory(c *Category) error {
	if err := c.validate(); err != nil {
		return err
	}

	return cs.Datasource.Update(c)
}

//DeleteCategory removes a category
func (cs CategoryService) DeleteCategory(id int) error {
	c, err := cs.Datasource.Get(id)

	if err != nil {
		return err
	}

	return cs.Datasource.Delete(c.ID)
}
