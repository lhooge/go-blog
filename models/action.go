// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

// Action This type is used for YES/NO actions
// Title is shown as the headline
// ActionURL defines where the form should be sent
// BackLinkURL defines where to go back
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
