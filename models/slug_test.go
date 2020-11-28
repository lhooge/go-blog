// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models_test

import (
	"testing"

	"git.hoogi.eu/snafu/go-blog/utils"
)

func TestCreateURLSafeSlug(t *testing.T) {
	testcases := []string{
		"This is a simple headline with umlauts\x00 ä ö ß ü and non printables \x1f\x00",
		"this-is-a-simple-headline-with-umlauts-ä-ö-ß-ü-and-non-printables",

		"A headline & a sample / ",
		"a-headline-a-sample",

		"A headline / a sample ",
		"a-headline-a-sample",

		"A headline / a sample ",
		"a-headline-a-sample",
	}

	for i := 0; i < len(testcases)-1; i = i + 2 {
		actual := utils.CreateURLSafeSlug(testcases[i], 0)

		if actual != testcases[i+1] {
			t.Errorf("Got: '%s'; want '%s'", actual, testcases[i+1])
		}
	}
}
