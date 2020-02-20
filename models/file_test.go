package models_test

import (
	"testing"

	"git.hoogi.eu/snafu/go-blog/models"
)

func TestFilenameSplit(t *testing.T) {
	var testcases = []struct {
		in  string
		out models.FileInfo
	}{
		{".bashrc", models.FileInfo{Name: "bashrc", Path: "."}},
		{"test.zip", models.FileInfo{Name: "test", Extension: ".zip", Path: "."}},
		{"/home/me/.myfile.zip", models.FileInfo{Name: "myfile", Extension: ".zip", Path: "/home/me"}},
		{"~/me/test.zip", models.FileInfo{Name: "test", Extension: ".zip", Path: "~/me"}},
		{"", models.FileInfo{Name: "", Extension: "", Path: "."}},
	}

	for _, v := range testcases {
		actual := models.SplitFilename(v.in)

		if actual.Name != v.out.Name {
			t.Errorf("wrong name: '%s'; want '%s'", actual.Name, v.out.Name)
		}

		if actual.Extension != v.out.Extension {
			t.Errorf("wrong extension: '%s'; want '%s'", actual.Extension, v.out.Extension)
		}

		if actual.Path != v.out.Path {
			t.Errorf("wrong path: '%s'; want '%s'", actual.Path, v.out.Path)
		}
	}
}
