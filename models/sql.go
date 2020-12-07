package models

import (
	"database/sql/driver"
	"time"
)

//AdminCriteria specifies which type of users should be considered
type AdminCriteria int

const (
	//OnlyAdmins conider only published
	OnlyAdmins = iota
	//NoAdmins conider no admins
	NoAdmins
	//AllUser conider all users
	AllUser
)

// PublishedCriteria specifies which entries should be shown
type PublishedCriteria int

const (
	// OnlyPublished conider only published
	OnlyPublished = iota
	// NotPublished conider only not published
	NotPublished
	// All conider both published and not published
	All
)

//NullTime represents a time which may not valid if time is null
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
