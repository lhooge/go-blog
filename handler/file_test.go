package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"git.hoogi.eu/snafu/go-blog/handler"
	"git.hoogi.eu/snafu/go-blog/models"
)

func TestFileWorkflow(t *testing.T) {
	err := doAdminUploadFileRequest(rAdminUser, "testdata/color.png")

	if err != nil {
		t.Fatal(err)
	}

	files, err := doAdminListFilesRequest(rAdminUser)

	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("one file is uploaded; but list files returns %d file(s)", len(files))
	}

	rr, err := doAdminGetFileRequest(rGuest, files[0].UniqueName)

	if err != nil {
		t.Error(err)
	}

	if rr.Result().ContentLength != 1610 {
		t.Errorf("expected 1610 bytes, but got %d", rr.Result().ContentLength)
	}

	if rr.Result().Header.Get("Content-Type") != "image/png" {
		t.Errorf("expected image/png content type, but got %s", rr.Result().Header.Get("Content-Type"))
	}

	err = doAdminFileDeleteRequest(rAdminUser, files[0].ID)

	if err != nil {
		t.Error(err)
	}

	_, err = doAdminGetFileRequest(rGuest, files[0].UniqueName)

	if err == nil {
		t.Errorf("file should be removed, but file is there %s", files[0].UniqueName)
	}
}

func doAdminListFilesRequest(user reqUser) ([]models.File, error) {
	r := request{
		url:    "/admin/files",
		user:   user,
		method: "GET",
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminListFilesHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return nil, tpl.Err
	}

	return tpl.Data["files"].([]models.File), nil
}

func doAdminGetFileRequest(user reqUser, uniquename string) (*httptest.ResponseRecorder, error) {
	r := request{
		url:    "/file/" + uniquename,
		user:   user,
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "uniquename",
				value: uniquename,
			},
		},
	}

	rw := httptest.NewRecorder()

	fh := handler.FileHandler{
		Context: ctx,
	}

	fh.FileGetHandler(rw, r.buildRequest())

	if rw.Result().StatusCode != http.StatusOK {
		return rw, fmt.Errorf("got an invalid status code during file request /file/%s , code: %d, message %s", uniquename, rw.Result().StatusCode, rw.Result().Status)
	}

	return rw, nil
}

func doAdminUploadFileRequest(user reqUser, file string) error {
	mp := []multipartRequest{
		multipartRequest{
			key:  "file",
			file: file,
		},
	}

	r := request{
		url:          "/admin/file/upload",
		user:         user,
		method:       "POST",
		multipartReq: mp,
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminUploadFilePostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}

func doAdminFileDeleteRequest(user reqUser, fileID int) error {
	r := request{
		url:    "/admin/file/delete/" + strconv.Itoa(fileID),
		user:   user,
		method: "GET",
		pathVar: []pathVar{
			pathVar{
				key:   "fileID",
				value: strconv.Itoa(fileID),
			},
		},
	}

	rw := httptest.NewRecorder()
	tpl := handler.AdminUploadDeletePostHandler(ctx, rw, r.buildRequest())

	if tpl.Err != nil {
		return tpl.Err
	}

	return nil
}
