package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.hoogi.eu/snafu/go-blog/httperror"
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
	CategoriesWithPublishedArticles = iota
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
	Get(categoryID int, fc FilterCriteria) (*Category, error)
	GetBySlug(slug string, fc FilterCriteria) (*Category, error)
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

func (cs CategoryService) GetBySlug(s string, fc FilterCriteria) (*Category, error) {
	c, err := cs.Datasource.GetBySlug(s, fc)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("category", fmt.Errorf("the category with slug %s was not found", s))
		}
		return nil, err
	}

	return c, nil
}

func (cs CategoryService) GetByID(id int, fc FilterCriteria) (*Category, error) {
	c, err := cs.Datasource.Get(id, fc)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperror.NotFound("category", fmt.Errorf("the category with id %d was not found", id))
		}
		return nil, err
	}

	return c, nil
}

func (cs CategoryService) Count(fc FilterCriteria) (int, error) {
	return cs.Datasource.Count(fc)
}

func (cs CategoryService) List(fc FilterCriteria) ([]Category, error) {
	return cs.Datasource.List(fc)
}

//Create creates a category
func (cs CategoryService) Create(c *Category) (int, error) {
	for i := 0; i < 10; i++ {
		c.Slug = CreateURLSafeSlug(c.Name, i)
		_, err := cs.Datasource.GetBySlug(c.Slug, AllCategories)

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

//Update updates a category
func (cs CategoryService) Update(c *Category) error {
	if err := c.validate(); err != nil {
		return err
	}

	return cs.Datasource.Update(c)
}

//Delete removes a category
func (cs CategoryService) Delete(id int) error {
	c, err := cs.Datasource.Get(id, AllCategories)

	if err != nil {
		return err
	}

	return cs.Datasource.Delete(c.ID)
}
