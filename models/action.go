// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

// Action this type is used for YES/NO actions see template/admin/action.html
// Title is shown in the headline
// ActionURL defines where the form should be sent
// BackLinkURL defines where to go back (if clicking on cancel)
// WarnMsg defines an optional warning which is shown above the description
// Description describes what action the user has to decide
type Action struct {
	ID          string
	Title       string
	ActionURL   string
	BackLinkURL string
	WarnMsg     string
	Description string
}
