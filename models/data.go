// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

// JSONData represents arbritary JSON data
type JSONData struct {
	Data interface{} `json:"data,-" xml:"data,-"`
}

// XMLData represents arbritary XML data
type XMLData struct {
	Data      interface{} `xml:"data,-"`
	HexEncode bool        `xml:"-"`
}
