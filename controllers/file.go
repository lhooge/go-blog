package controllers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

type FileHandler struct {
	Context *middleware.AppContext
}

//FileGetHandler serves the file based on the url filename
func (fh FileHandler) FileGetHandler(w http.ResponseWriter, r *http.Request) {
	rv := getVar(r, "filename")

	f, err := fh.Context.FileService.GetByName(rv, nil)

	if err != nil {
		http.Error(w, "the file was not found", http.StatusNotFound)
		return
	}

	loc := filepath.Join(fh.Context.ConfigService.Location, f.FullFilename)

	w.Header().Set("Content-Type", f.ContentType)
	w.Header().Set("Content-Disposition", "inline")

	rf, err := os.Open(loc)

	if err != nil {
		if os.IsNotExist(err) {
			logger.Log.Errorf("the file %s was not found - %v", loc, err)
			http.Error(w, "404 page not found", http.StatusNotFound)
		}
		if os.IsPermission(err) {
			logger.Log.Errorf("not permitted to read file %s - %v", loc, err)
			http.Error(w, "404 page not found", http.StatusForbidden)
		}
		logger.Log.Errorf("an internal error while reading file %s - %v", loc, err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}

	defer rf.Close()

	http.ServeContent(w, r, loc, f.LastModified, rf)
}

//AdminListFilesHandler returns the template which lists alle uploaded files belonging to a user, admins will see all files
func AdminListFilesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	page := getPageParam(r)

	t, err := ctx.FileService.Count(u)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFiles,
			Active: "files",
			Err:    err,
		}
	}

	p := &models.Pagination{
		Total:       t,
		Limit:       20,
		CurrentPage: page,
		RelURL:      "admin/files/page",
	}

	fs, err := ctx.FileService.List(u, p)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFiles,
			Active: "files",
			Err:    err,
		}
	}

	return &middleware.Template{
		Name:   tplAdminFiles,
		Active: "files",
		Data: map[string]interface{}{
			"files":      fs,
			"pagination": p,
		}}
}

//AdminUploadFileHandler returns the form for uploading a file
func AdminUploadFileHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	return &middleware.Template{
		Name:   tplAdminFileUpload,
		Active: "files",
	}
}

//AdminUploadFilePostHandler handles the upload
func AdminUploadFilePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	if r.ContentLength > int64(ctx.ConfigService.MaxUploadSize) {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    httperror.New(http.StatusUnprocessableEntity, "Filesize too large", errors.New("filesize too large")),
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(ctx.ConfigService.MaxUploadSize))

	err := r.ParseMultipartForm(1024)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    err,
		}
	}

	u, _ := middleware.User(r)

	ff, h, err := r.FormFile("file")

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    err,
		}
	}

	defer ff.Close()

	data, err := ioutil.ReadAll(ff)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    err,
		}
	}

	ct := http.DetectContentType(data)

	file := &models.File{
		ContentType: ct,
		Location:    ctx.ConfigService.File.Location,
		Author:      u,
		Size:        int64(len(data)),
	}

	var filename string
	var ext string

	ext = filepath.Ext(h.Filename[1:])

	if ext != "" {
		filename = h.Filename[0 : len(h.Filename)-len(ext)]
	} else {
		filename = h.Filename
	}

	file.Filename = filename
	file.Extension = ext

	_, err = ctx.FileService.Upload(file, data)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    err,
		}
	}

	return &middleware.Template{
		RedirectPath: "/admin/files",
		SuccessMsg:   "Successfully uploaded file",
		Active:       "files",
	}
}

//AdminUploadDeleteHandler returns the action template which asks the user if the file should be removed
func AdminUploadDeleteHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	rv := getVar(r, "fileID")

	id, err := parseInt(rv)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFiles,
			Err:    err,
			Active: "files",
		}
	}

	f, err := ctx.FileService.GetByID(id, u)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFiles,
			Err:    err,
			Active: "files",
		}
	}

	action := models.Action{
		ID:          "deleteFile",
		ActionURL:   fmt.Sprintf("/admin/file/delete/%d", f.ID),
		Description: fmt.Sprintf("%s %s?", "Do you want to delete the file ", f.Filename),
		Title:       "Confirm removal of file",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "articles",
		Data: map[string]interface{}{
			"action": action,
		},
	}
}

//AdminUploadDeletePostHandler removes a file
func AdminUploadDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	u, _ := middleware.User(r)

	rv := getVar(r, "fileID")

	id, err := parseInt(rv)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "/admin/files",
			Err:          err,
			Active:       "files",
		}
	}

	err = ctx.FileService.Delete(id, ctx.ConfigService.File.Location, u)

	warnMsg := ""
	if err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
			warnMsg = "File removed from database, but was not found in file system anymore"
		} else {
			return &middleware.Template{
				RedirectPath: "/admin/files",
				Err:          err,
				Active:       "files",
			}
		}
	}

	return &middleware.Template{
		Active:       "files",
		RedirectPath: "admin/files",
		SuccessMsg:   "File successfully deleted",
		WarnMsg:      warnMsg,
	}
}
