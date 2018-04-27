package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
)

const (
	tplAdminFiles      = "admin/files"
	tplAdminFileUpload = "admin/file_upload"
)

//FileGetHandler serves the file based on the url filename
func FileGetHandler(ctx *middleware.AppContext) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		filename := getVar(r, "filename")

		file, err := ctx.FileService.GetFileByName(filename, nil)

		if err != nil {
			http.Error(w, "the file was not found", http.StatusNotFound)
			return
		}

		fileLoc := filepath.Join(ctx.ConfigService.Location, file.Filename)

		w.Header().Set("Content-Type", file.ContentType)
		w.Header().Set("Content-Disposition", "attachment")

		rf, err := os.Open(fileLoc)

		if err != nil {
			if os.IsNotExist(err) {
				logger.Log.Errorf("the file %s was not found - %v", fileLoc, err)
				http.Error(w, "404 page not found", http.StatusNotFound)
			}
			if os.IsPermission(err) {
				logger.Log.Errorf("not permitted to read file %s - %v", fileLoc, err)
				http.Error(w, "404 page not found", http.StatusForbidden)
			}
			logger.Log.Errorf("an internal error while reading file %s - %v", fileLoc, err)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		}

		defer rf.Close()

		http.ServeContent(w, r, fileLoc, file.LastModified, rf)
	}
	return http.HandlerFunc(fn)
}

//AdminListFilesHandler returns the template which lists alle uploaded files belonging to a user, admins will see all files
func AdminListFilesHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	page := getPageParam(r)

	total, err := ctx.FileService.CountFiles(user)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFiles,
			Active: "files",
			Err:    err,
		}
	}

	pagination := &models.Pagination{
		Total:       total,
		Limit:       20,
		CurrentPage: page,
		RelURL:      "admin/files/page",
	}

	files, err := ctx.FileService.ListFiles(user, pagination)

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
			"files":      files,
			"pagination": pagination,
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
	r.Body = http.MaxBytesReader(w, r.Body, int64(ctx.ConfigService.MaxUploadSize))

	err := r.ParseMultipartForm(20 * 1024)
	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    err,
		}
	}

	user, _ := middleware.User(r)

	cf, header, err := r.FormFile("file")

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    err,
		}
	}

	defer cf.Close()

	data, err := ioutil.ReadAll(cf)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFileUpload,
			Active: "files",
			Err:    err,
		}
	}

	nf := r.FormValue("newfilename")

	ct := http.DetectContentType(data)

	file := &models.File{
		ContentType: ct,
		Location:    ctx.ConfigService.File.Location,
		Author:      user,
		Size:        int64(len(data)),
	}

	if len(nf) > 0 {
		idx := strings.LastIndex(nf, ".")

		//check if an extension were provided in the new file name; if not try to use extension from form data
		if idx > 0 {
			ue := filepath.Ext(header.Filename)

			file.Filename = fmt.Sprintf("%s%s", nf, ue)
		} else {
			file.Filename = nf
		}
	} else {
		file.Filename = header.Filename
	}

	_, err = ctx.FileService.UploadFile(file, data)

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
	user, _ := middleware.User(r)

	reqVar := getVar(r, "fileID")

	fileID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFiles,
			Err:    err,
			Active: "files",
		}
	}

	file, err := ctx.FileService.GetFileByID(fileID, user)

	if err != nil {
		return &middleware.Template{
			Name:   tplAdminFiles,
			Err:    err,
			Active: "files",
		}
	}

	deleteInfo := models.Action{
		ID:          "deleteFile",
		ActionURL:   fmt.Sprintf("/admin/file/delete/%d", file.ID),
		Description: fmt.Sprintf("%s %s?", "Do you want to delete the file ", file.Filename),
		Title:       "Confirm removal of file",
	}

	return &middleware.Template{
		Name:   tplAdminAction,
		Active: "articles",
		Data: map[string]interface{}{
			"action": deleteInfo,
		},
	}
}

//AdminUploadDeletePostHandler removes a file
func AdminUploadDeletePostHandler(ctx *middleware.AppContext, w http.ResponseWriter, r *http.Request) *middleware.Template {
	user, _ := middleware.User(r)

	reqVar := getVar(r, "fileID")

	fileID, err := parseInt(reqVar)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "/admin/files",
			Err:          err,
			Active:       "files",
		}
	}

	err = ctx.FileService.DeleteFile(fileID, ctx.ConfigService.File.Location, user)

	if err != nil {
		return &middleware.Template{
			RedirectPath: "/admin/files",
			Err:          err,
			Active:       "files",
		}
	}

	return &middleware.Template{
		Active:       "files",
		RedirectPath: "admin/files",
		SuccessMsg:   "File successfully deleted",
	}
}
