package models

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"git.hoogi.eu/snafu/go-blog/httperror"
	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/settings"
)

// File represents a file
type File struct {
	ID           int
	UniqueName   string    `json:"unique_name"`
	FullFilename string    `json:"full_name"`
	Link         string    `json:"link"`
	ContentType  string    `json:"content_type"`
	Inline       bool      `json:"inline"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	Data         []byte    `json:"-"`
	FileInfo     FileInfo
	Author       *User
}

// FileInfo contains Path, Name and Extension of a file.
// Use SplitFilename to split the information from a filename
type FileInfo struct {
	Path      string
	Name      string
	Extension string
}

// FileDatasourceService defines an interface for CRUD operations of files
type FileDatasourceService interface {
	Create(f *File) (int, error)
	Get(fileID int, u *User) (*File, error)
	GetByUniqueName(uniqueName string, u *User) (*File, error)
	List(u *User, p *Pagination) ([]File, error)
	Count(u *User) (int, error)
	Update(f *File) error
	Delete(fileID int) error
}

// validate validates if mandatory file fields are set
func (f *File) validate() error {
	if len(f.FullFilename) == 0 {
		return httperror.ValueRequired("filename")
	}

	if len(f.FullFilename) > 255 {
		return httperror.ValueTooLong("filename", 255)
	}

	return nil
}

func (f File) randomFilename() string {
	var buf bytes.Buffer

	sanFilename := sanitizeFilename(f.FileInfo.Name)

	if len(sanFilename) == 0 {
		sanFilename = "unnamed"
	}

	buf.WriteString(sanFilename)
	buf.WriteString(f.FileInfo.Extension)

	return buf.String()
}

func SplitFilename(filename string) FileInfo {
	base := filepath.Base(filename)
	base = strings.TrimLeft(base, ".")

	ext := filepath.Ext(base)

	idx := strings.LastIndex(base, ".")

	var name string
	if idx > 0 {
		name = base[:idx]
	} else {
		name = base
	}

	path := filepath.Dir(filename)

	return FileInfo{
		Name:      name,
		Extension: ext,
		Path:      path,
	}
}

// FileService containing the service to interact with files
type FileService struct {
	Datasource FileDatasourceService
	Config     settings.File
}

// GetByID returns the file based on the fileID; it the user is given and it is a non admin
// only file specific to this user is returned
func (fs *FileService) GetByID(fileID int, u *User) (*File, error) {
	return fs.Datasource.Get(fileID, u)
}

// GetByUniqueName returns the file based on the unique name; it the user is given and it is a non admin
// only file specific to this user is returned
func (fs *FileService) GetByUniqueName(uniqueName string, u *User) (*File, error) {
	return fs.Datasource.GetByUniqueName(uniqueName, u)
}

// List returns a list of files based on the filename; it the user is given and it is a non admin
// only files specific to this user are returned
func (fs *FileService) List(u *User, p *Pagination) ([]File, error) {
	return fs.Datasource.List(u, p)
}

// Count returns a number of files based on the filename; it the user is given and it is a non admin
// only files specific to this user are counted
func (fs *FileService) Count(u *User) (int, error) {
	return fs.Datasource.Count(u)
}

func (fs *FileService) ToggleInline(fileID int, u *User) error {
	f, err := fs.Datasource.Get(fileID, u)

	if err != nil {
		return err
	}

	f.FileInfo = SplitFilename(f.FullFilename)
	newName := f.randomFilename()
	f.Inline = !f.Inline

	err = os.Rename(filepath.Join(fs.Config.Location, f.UniqueName), filepath.Join(fs.Config.Location, newName))

	f.UniqueName = newName

	if err != nil {
		return err
	}

	return fs.Datasource.Update(f)
}

// Delete deletes a file based on fileID; users which are not the owner are not allowed to remove files; except admins
func (fs *FileService) Delete(fileID int, u *User) error {
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

	return os.Remove(filepath.Join(fs.Config.Location, file.UniqueName))
}

// Upload uploaded files will be saved at the configured file location, filename is saved in the database
func (fs *FileService) Upload(f *File) (int, error) {
	if err := f.validate(); err != nil {
		return -1, err
	}

	f.FileInfo = SplitFilename(f.FullFilename)

	if len(f.FileInfo.Extension) == 0 && !strings.HasPrefix(f.ContentType, "text/plain") {
		return -1, httperror.New(
			http.StatusUnprocessableEntity,
			"The file has no extension and does not contain plain text.",
			fmt.Errorf("the file %s has no extension and does not contain plain text, content type is: %s", f.FullFilename, f.ContentType))
	}

	if len(f.FileInfo.Extension) > 0 {
		if _, ok := fs.Config.AllowedFileExtensions[f.FileInfo.Extension]; !ok {
			return -1, httperror.New(
				http.StatusUnprocessableEntity,
				"The file type is not supported.",
				fmt.Errorf("error during upload, the file type %s is not supported", f.FileInfo.Extension))
		}
	}

	f.UniqueName = f.randomFilename()

	file, err := fs.GetByUniqueName(f.UniqueName, nil)

	if err != nil {
		if err != sql.ErrNoRows {
			return -1, err
		}
	}

	if file != nil {
		return -1, httperror.New(
			http.StatusUnprocessableEntity,
			"A file with this filename already exist. Please choose another filename.",
			errors.New("a file with this filename already exist"))
	}

	fi := filepath.Join(fs.Config.Location, f.UniqueName)

	if err = ioutil.WriteFile(fi, f.Data, 0640); err != nil {
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

	f.Data = nil

	return i, nil
}

var filenameSubs = map[rune]string{
	'/':  "",
	'\\': "",
	':':  "",
	'*':  "",
	'?':  "",
	'"':  "",
	'<':  "",
	'>':  "",
	'|':  "",
	' ':  "",
}

func isDot(r rune) bool {
	return '.' == r
}

// sanitizeFilename sanitizes a filename for safe use when serving file
func sanitizeFilename(s string) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.TrimFunc(s, unicode.IsSpace)

	s = strings.Map(func(r rune) rune {
		if _, ok := filenameSubs[r]; ok {
			return -1
		}
		return r
	}, s)

	s = strings.TrimLeftFunc(s, isDot)
	return s
}
