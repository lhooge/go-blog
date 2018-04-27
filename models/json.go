// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

// JSONData Represents arbritary JSON data
type JSONData struct {
	Data interface{} `json:"data,-"`
}
