package models

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/utils"
)

//SiteDatasourceService defines an interface for CRUD operations on sites
type SiteDatasourceService interface {
	Create(s *Site) (int, error)
	List(pc PublishedCriteria, p *Pagination) ([]Site, error)
	Get(siteID int, pc PublishedCriteria) (*Site, error)
	GetByLink(link string, pc PublishedCriteria) (*Site, error)
	Publish(s *Site) error
	Update(s *Site) error
	Delete(s *Site) error
	Order(siteID int, dir Direction) error
	Max() (int, error)
	Count(pc PublishedCriteria) (int, error)
}

//Direction type to distinct if a site should be moved up or down
type Direction int

const (
	//Up for moving the site one up
	Up = iota
	//Down for moving the site one down
	Down
)

//Site represents a site
type Site struct {
	ID           int
	Title        string
	Link         string
	Content      string
	Published    bool
	PublishedOn  NullTime
	LastModified time.Time
	OrderNo      int
	Author       *User
}

// LinkEscape escapes a link for safe use in URLs
func (s Site) LinkEscape() string {
	if s.isExternal() {
		return s.Link
	}
	return utils.AppendString("/site/", url.PathEscape(s.Link))
}

func (s Site) safeLink() string {
	if s.isExternal() {
		return s.Link
	}
	return utils.CreateURLSafeSlug(s.Link, -1)
}

func (s Site) isExternal() bool {
	if len(s.Link) > 6 {
		if s.Link[:7] == "http://" {
			return true
		}
		if len(s.Link) > 7 {
			if s.Link[:8] == "https://" {
				return true
			}
		}
	}
	return false
}

// validate validates if mandatory site fields are set
func (s *Site) validate(ds SiteDatasourceService, changeLink bool) error {
	if len(s.Link) == 0 {
		return httperror.ValueRequired("link")
	}

	if len(s.Title) == 0 {
		return httperror.ValueRequired("title")
	}

	if s.isExternal() {
		//TODO() no further checks.
		return nil
	}

	l, err := ds.GetByLink(s.safeLink(), All)

	if changeLink {
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		}

		if l != nil {
			return httperror.New(http.StatusUnprocessableEntity, fmt.Sprintf("The link %s already exists.", s.Link), fmt.Errorf("the link %s already exits", s.Link))
		}
	}

	return nil
}

//SiteService containing the service to access site
type SiteService struct {
	Datasource SiteDatasourceService
}

// ListSites returns all sites
func (ss SiteService) ListSites(pc PublishedCriteria, p *Pagination) ([]Site, error) {
	return ss.Datasource.List(pc, p)
}

//PublishSite switches the publish state of the site
func (ss SiteService) PublishSite(siteID int) error {
	s, err := ss.Datasource.Get(siteID, All)

	if err != nil {
		return err
	}

	err = ss.Datasource.Publish(s)

	if err != nil {
		return err
	}

	return nil
}

// CreateSite creates a site
func (ss SiteService) CreateSite(s *Site) (int, error) {
	if err := s.validate(ss.Datasource, true); err != nil {
		return -1, err
	}

	s.Link = s.safeLink()

	m, err := ss.Datasource.Max()

	if err != nil {
		return -1, err
	}

	s.OrderNo = m + 1

	return ss.Datasource.Create(s)
}

//OrderSite reorder the site
func (ss SiteService) OrderSite(siteID int, dir Direction) error {
	return ss.Datasource.Order(siteID, dir)
}

//UpdateSite updates a site
func (ss SiteService) UpdateSite(s *Site) error {
	oldSite, err := ss.GetSiteByID(s.ID, All)

	if err != nil {
		return err
	}

	changeLink := false
	if oldSite.Link != s.Link {
		changeLink = true
		s.Link = s.safeLink()
	}

	if err := s.validate(ss.Datasource, changeLink); err != nil {
		return err
	}

	return ss.Datasource.Update(s)
}

//DeleteSite deletes a site
func (ss SiteService) DeleteSite(siteID int) error {
	s, err := ss.GetSiteByID(siteID, All)

	if err != nil {
		return err
	}

	return ss.Datasource.Delete(s)
}

// GetSiteByLink Get a site by the link.
func (ss SiteService) GetSiteByLink(link string, pc PublishedCriteria) (*Site, error) {
	return ss.Datasource.GetByLink(link, pc)

}

// GetSiteByID Get a site by the id.
func (ss SiteService) GetSiteByID(siteID int, pc PublishedCriteria) (*Site, error) {
	return ss.Datasource.Get(siteID, pc)

}

// CountSites returns the number of sites
func (ss SiteService) CountSites(pc PublishedCriteria) (int, error) {
	return ss.Datasource.Count(pc)
}
