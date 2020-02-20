// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Package cfg parses and validates the configuration
package settings

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"git.hoogi.eu/snafu/cfg"
	"git.hoogi.eu/snafu/go-blog/components/logger"
	"git.hoogi.eu/snafu/go-blog/utils"
)

type LoginMethod int

const (
	Username = iota
	EMail
)

func (lm *LoginMethod) Unmarshal(value string) error {
	if strings.ToLower(value) == "mail" {
		*lm = LoginMethod(EMail)
		return nil
	} else if strings.ToLower(value) == "username" {
		*lm = LoginMethod(Username)
		return nil
	}
	return fmt.Errorf("unexpected config value for login method %s", value)
}

type AllowedFileExts map[string]string

func (afe *AllowedFileExts) Unmarshal(value string) error {
	kv := make(map[string]string)

	splitted := strings.Split(value, ",")

	for _, v := range splitted {
		trimmed := strings.TrimSpace(v)

		if len(trimmed) > 0 {
			if trimmed[0] != '.' {
				trimmed = "." + trimmed
			}
		}

		kv[trimmed] = trimmed

		*afe = AllowedFileExts(kv)
	}

	return nil
}

type Settings struct {
	Environment  string `cfg:"environment" default:"prod"`
	BuildVersion string `cfg:"-"`
	BuildGitHash string `cfg:"-"`

	Application
	Blog
	User
	File
	Server
	Database
	Mail
	Session
	CSRF
	Log
}

type Server struct {
	Address string `cfg:"server_address" default:"127.0.0.1"`
	Port    int    `cfg:"server_port" default:"4730"`

	UseTLS bool   `cfg:"use_tls" default:"yes"`
	Cert   string `cfg:"ssl_certificate_file"`
	Key    string `cfg:"ssl_certificate_key_file"`
}

type Application struct {
	Title        string `cfg:"application_title"`
	Language     string `cfg:"application_language"`
	Description  string `cfg:"application_description"`
	Domain       string `cfg:"application_domain"`
	Favicon      string `cfg:"application_favicon" default:"assets/favicon.ico"`
	RobotsTxt    string `cfg:"application_robots_txt"`
	CustomCSS    string `cfg:"application_custom_css"`
	OverwriteCSS bool   `cfg:"application_overwrite_default_css" default:"false"`
}

type Database struct {
	File string `cfg:"sqlite_file" default:"data/goblog.sqlite"`
}

type File struct {
	Location              string          `cfg:"file_location" default:"/srv/goblog/files/"`
	MaxUploadSize         cfg.FileSize    `cfg:"file_max_upload_size" default:"10MB"`
	AllowedFileExtensions AllowedFileExts `cfg:"file_allowed_extensions"`
}

type Blog struct {
	ArticlesPerPage int `cfg:"blog_articles_per_page" default:"20"`
	RSSFeedItems    int `cfg:"blog_rss_feed_items" default:"10"`
}

type User struct {
	MinPasswordLength int         `cfg:"user_min_password_length" default:"12"`
	InterceptorPlugin string      `cfg:"user_interceptor_plugin"`
	LoginMethod       LoginMethod `cfg:"user_login_method" default:"username"`
}

type Mail struct {
	Host     string `cfg:"mail_smtp_host" default:"127.0.0.1"`
	Port     int    `cfg:"mail_smtp_port" default:"25"`
	User     string `cfg:"mail_smtp_user"`
	Password string `cfg:"mail_smtp_password"`

	SenderAddress string `cfg:"mail_sender_address"`
	SubjectPrefix string `cfg:"mail_subject_prefix"`
}

type Session struct {
	TTL               time.Duration `cfg:"session_time_to_live" default:"2h"`
	GarbageCollection time.Duration `cfg:"session_garbage_collection" default:"5m"`
	CookieName        string        `cfg:"session_cookie_name" default:"goblog"`
	CookieSecure      bool          `cfg:"session_cookie_secure" default:"true"`
	CookiePath        string        `cfg:"session_cookie_path" default:"/admin"`
}

type CSRF struct {
	CookieName   string `cfg:"csrf_cookie_name" default:"csrf"`
	CookieSecure bool   `cfg:"csrf_cookie_secure" default:"true"`
	CookiePath   string `cfg:"csrf_cookie_path" default:"/admin"`
	RandomKey    string `cfg:"csrf_random_key"`
}

type Log struct {
	Level      string `cfg:"log_level" default:"info"`
	File       string `cfg:"log_file" default:"/var/log/goblog/error.log"`
	Access     bool   `cfg:"log_access" default:"true"`
	AccessFile string `cfg:"log_access_file" default:"/var/log/goblog/access.log"`
}

const csrfTokenFilename = ".csrftoken"

func MergeConfigs(configs []cfg.File) (*Settings, error) {
	c := cfg.ConfigFiles{}

	for _, cp := range configs {
		c.AddConfig(cp.Path, cp.Name, cp.Required)
	}

	settings := new(Settings)

	def, err := c.MergeConfigsInto(settings)

	for k, d := range def {
		logger.Log.Warnf("config: no config value for key '%s' found in any config - assuming default value: '%v'", k, d.Value)
	}

	return settings, err
}

func LoadConfig(filename string) (*Settings, error) {
	settings := new(Settings)
	def, err := cfg.LoadConfigInto(filename, settings)

	for k, d := range def {
		logger.Log.Warnf("config: no config value for %s key found in any config - assuming default value %v", k, d.Value)
	}

	return settings, err
}

func (cfg *Settings) CheckConfig() error {
	//check log file is rw in production mode
	if cfg.Environment != "dev" {
		if _, err := os.OpenFile(cfg.Log.File, os.O_RDONLY|os.O_CREATE, 0640); err != nil {
			return fmt.Errorf("config 'log_file': could not open log file %s error %v", cfg.Log.File, err)
		}
		if _, err := os.OpenFile(cfg.Log.AccessFile, os.O_RDONLY|os.O_CREATE, 0640); err != nil {
			return fmt.Errorf("config 'log_access_file': could not open access log file %s error %v", cfg.Log.AccessFile, err)
		}
	}

	if len(cfg.Application.Domain) == 0 {
		return errors.New("config 'application_domain': please specify a domain name")
	}

	_, err := url.ParseRequestURI(cfg.Application.Domain)
	if err != nil {
		return fmt.Errorf("config 'application_domain': invalid url setting for key 'application_domain' value '%s'", cfg.Application.Domain)
	}

	if len(cfg.Application.CustomCSS) > 0 {
		f, err := os.Open(cfg.Application.CustomCSS)

		if err != nil {
			return fmt.Errorf("config 'application_custom_css': could not open specified custom css file %s error %v", cfg.Application.CustomCSS, err)
		}

		fi, err := f.Stat()

		if err != nil {
			return fmt.Errorf("config 'application_custom_css': could not open specified custom css file %s error %v", cfg.Application.CustomCSS, err)
		}

		if fi.IsDir() {
			return fmt.Errorf("config 'application_custom_css': the custom css file '%s' is a directory", cfg.Application.CustomCSS)
		}
	}

	if len(cfg.Application.RobotsTxt) > 0 {
		f, err := os.Open(cfg.Application.RobotsTxt)

		if err != nil {
			return fmt.Errorf("config 'application_robots_txt': could not open specified robots txt file %s error %v", cfg.Application.RobotsTxt, err)
		}

		fi, err := f.Stat()

		if err != nil {
			return fmt.Errorf("config 'application_robots_txt': could not open specified robots txt file %s error %v", cfg.Application.RobotsTxt, err)
		}

		if fi.IsDir() {
			return fmt.Errorf("config 'application_robots_txt': the robots txt file '%s' is a directory", cfg.Application.RobotsTxt)
		}
	}

	if len(cfg.Application.Favicon) > 0 {
		f, err := os.Open(cfg.Application.Favicon)

		if err != nil {
			return fmt.Errorf("config 'application_favicon': could not open specified favicon file %s error %v", cfg.Application.Favicon, err)
		}

		fi, err := f.Stat()

		if err != nil {
			return fmt.Errorf("config 'application_favicon': could not open specified favicon file %s error %v", cfg.Application.Favicon, err)
		}

		if fi.IsDir() {
			return fmt.Errorf("config 'application_favicon': the favicon file '%s' is a directory", cfg.Application.Favicon)
		}
	} else {
		cfg.Application.Favicon = "assets/favicon.ico"
	}

	//server settings
	if cfg.Server.UseTLS {
		if _, err := os.Open(cfg.Server.Cert); err != nil {
			return fmt.Errorf("config: could not open certificate %s error %v", cfg.Server.Cert, err)
		}
		if _, err := os.Open(cfg.Server.Key); err != nil {
			return fmt.Errorf("config: could not open private key file %s error %v", cfg.Server.Key, err)
		}
	}

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("config: invalid port setting for key 'server_port' value %d", cfg.Server.Port)
	}

	if _, err := os.Open(cfg.File.Location); err != nil {
		return fmt.Errorf("config: could not open file path %s error %v", cfg.File.Location, err)
	}

	return nil
}

func (cfg *Settings) GenerateCSRF() (bool, error) {
	if len(cfg.CSRF.RandomKey) == 0 {

		var b []byte

		if _, err := os.Stat(csrfTokenFilename); os.IsNotExist(err) {
			//create a random csrf token
			r := utils.RandomSource{
				CharsToGen: utils.AlphaUpperLowerNumericSpecial,
			}

			b = r.RandomSequence(32)

			err := ioutil.WriteFile(csrfTokenFilename, b, 0640)

			if err != nil {
				return false, err
			}

			cfg.CSRF.RandomKey = string(b)

			return true, nil
		} else {
			//read existing csrf token
			b, err = ioutil.ReadFile(csrfTokenFilename)

			if err != nil {
				return false, err
			}

			cfg.CSRF.RandomKey = string(b)

			return false, nil
		}

	}
	return false, nil
}
