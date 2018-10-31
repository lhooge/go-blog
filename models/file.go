package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.hoogi.eu/go-blog/components/httperror"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/utils"
)

//File represents a file
type File struct {
	ID           int
	Location     string
	Filename     string
	ContentType  string
	Size         int64
	LastModified time.Time
	Author       *User
}

//FileDatasourceService defines an interface for CRUD operations of files
type FileDatasourceService interface {
	Create(f *File) (int, error)
	Get(fileID int, u *User) (*File, error)
	GetByFilename(filename string, u *User) (*File, error)
	List(u *User, p *Pagination) ([]File, error)
	Count(u *User) (int, error)
	Delete(fileID int) error
}

// validate validates if mandatory file fields are set
// sanitizes the filename
func (f *File) validate() error {
	if len(f.Filename) > 255 {
		return httperror.ValueTooLong("filename", 255)
	}

	return nil
}

func (f *File) sanitizeFilename() string {
	idx := strings.LastIndex(f.Filename, ".")

	//ignore first dot
	if idx > 0 {
		extension := f.Filename[idx:len(f.Filename)]
		return utils.SanitizeFilename(f.Filename[:idx]) + extension
	}
	return utils.SanitizeFilename(f.Filename)
}

//FileService containing the service to interact with files
type FileService struct {
	Datasource FileDatasourceService
}

//GetFileByID returns the file based on the fileID; it the user is given and it is a non admin
//only file specific to this user is returned
func (fs FileService) GetFileByID(fileID int, u *User) (*File, error) {
	return fs.Datasource.Get(fileID, u)
}

//GetFileByName returns the file based on the filename; it the user is given and it is a non admin
//only file specific to this user is returned
func (fs FileService) GetFileByName(filename string, u *User) (*File, error) {
	return fs.Datasource.GetByFilename(filename, u)
}

//ListFiles returns a list of files based on the filename; it the user is given and it is a non admin
//only files specific to this user are returned
func (fs FileService) ListFiles(u *User, p *Pagination) ([]File, error) {
	return fs.Datasource.List(u, p)
}

//CountFiles returns a number of files based on the filename; it the user is given and it is a non admin
//only files specific to this user are counted
func (fs FileService) CountFiles(u *User) (int, error) {
	return fs.Datasource.Count(u)
}

//DeleteFile deletes a file based on fileID; users which are not the owner are not allowed to remove files; except admins
func (fs FileService) DeleteFile(fileID int, location string, u *User) error {
	file, err := fs.Datasource.Get(fileID, u)

	if err != nil {
		return err
	}

	if !u.IsAdmin {
		if file.Author.ID != u.ID {
			return httperror.PermissionDenied("delete", "file", fmt.Errorf("could not remove file %d user %d has no permission", fileID, u.ID))
		}
	}

	err = fs.Datasource.Delete(fileID)

	if err != nil {
		return err
	}

	return os.Remove(filepath.Join(location, file.Filename))
}

//UploadFile uploaded files will be saved at the configured file location, filename is saved in the database
func (fs FileService) UploadFile(f *File, data []byte) (int, error) {
	if err := f.validate(); err != nil {
		return -1, err
	}
	f.Filename = f.sanitizeFilename()

	fi := filepath.Join(f.Location, f.Filename)

	err := ioutil.WriteFile(fi, data, 0640)
	if err != nil {
		return -1, err
	}

	i, err := fs.Datasource.Create(f)

	if err != nil {
		err2 := os.Remove(fi)
		if err2 != nil {
			logger.Log.Error(err2)
		}
		return -1, err
	}

	data = nil

	return i, nil
}
