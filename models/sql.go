package models

import (
	"database/sql/driver"
	"time"
)

// AdminCriteria specifies which type of users should be considered
type AdminCriteria int

const (
	// OnlyAdmins consider only admins
	OnlyAdmins = iota
	// NoAdmins consider no admins
	NoAdmins
	// AllUser consider all users
	AllUser
)

// PublishedCriteria specifies which entries should be shown
type PublishedCriteria int

const (
	// OnlyPublished consider only published
	OnlyPublished = iota
	// NotPublished consider only not published
	NotPublished
	// All consider both published and not published
	All
)

// NullTime represents a time which may not valid if time is null
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
