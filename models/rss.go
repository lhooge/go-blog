package models

import (
	"encoding/xml"
	"time"
)

type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`

	Items []RSSItem `xml:"item"`
}

type RSSItem struct {
	GUID        string  `xml:"guid"`
	Author      string  `xml:"author"`
	Title       string  `xml:"title"`
	Link        string  `xml:"link"`
	Description string  `xml:"description"`
	PubDate     RSSTime `xml:"pubDate"`
}

type RSSTime time.Time

func (r RSSTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	t := time.Time(r)
	v := t.Format("Mon, 2 Jan 2007 15:04:05 GMT")
	return e.EncodeElement(v, start)
}
