// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Package cfg parses and validates the configuration
package settings

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"git.hoogi.eu/cfg"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/utils"
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

type DatabaseEngine int

const (
	MySQL = iota
	SQLite
)

func (de *DatabaseEngine) Unmarshal(value string) error {
	if strings.ToLower(value) == "mysql" {
		*de = DatabaseEngine(MySQL)
		return nil
	} else if strings.ToLower(value) == "sqlite" {
		*de = DatabaseEngine(SQLite)
		return nil
	}

	return fmt.Errorf("unexpected config value for database engine %s", value)
}

type Settings struct {
	Environment string `cfg:"environment" default:"prod"`

	AppVersion string
	BuildDate  string

	Title    string `cfg:"title"`
	Subtitle string `cfg:"subtitle"`

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

	Domain string `cfg:"domain"`
}

type Database struct {
	Engine   DatabaseEngine `cfg:"database_engine"`
	Host     string         `cfg:"mysql_host" default:"127.0.0.1"`
	Port     int            `cfg:"mysql_port" default:"3306"`
	User     string         `cfg:"mysql_user" default:"root"`
	Password string         `cfg:"mysql_password" default:""`
	Name     string         `cfg:"mysql_database" default:"go_blog"`
	File     string         `cfg:"sqlite_file"`
}
type File struct {
	Location      string       `cfg:"file_location"`
	MaxUploadSize cfg.FileSize `cfg:"file_max_upload_size"`
}

type Blog struct {
	ArticlesPerPage int `cfg:"blog_articles_per_page" default:"12"`
}

type User struct {
	MinPasswordLength int         `cfg:"user_min_password_length" default:"12"`
	InterceptorPlugin string      `cfg:"user_interceptor_plugin"`
	LoginMethod       LoginMethod `cfg:"user_login_method"`
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
	TTL               time.Duration `cfg:"session_time_to_live"`
	GarbageCollection time.Duration `cfg:"session_garbage_collection"`
	CookieName        string        `cfg:"session_cookie_name"`
	CookieSecure      bool          `cfg:"session_cookie_secure"`
	CookiePath        string        `cfg:"session_cookie_path"`
}

type CSRF struct {
	CookieName   string `cfg:"csrf_cookie_name"`
	CookieSecure bool   `cfg:"csrf_cookie_secure"`
	CookiePath   string `cfg:"csrf_cookie_path"`
	RandomKey    string `cfg:"csrf_random_key"`
}

type Log struct {
	Level      string `cfg:"log_level"`
	Path       string `cfg:"log_path"`
	AccessLogs bool   `cfg:"log_access_logs"`
}

const csrfTokenFilename = ".csrftoken"

func MergeConfigs(configs []cfg.File) (*Settings, error) {
	c := cfg.Config{}

	for _, cp := range configs {
		c.AddConfig(cp.Path, cp.Name)
	}

	settings := new(Settings)
	def, err := c.MergeConfigsInto(settings)

	for k, d := range def {
		logger.Log.Warnf("no config value for %s key found in any config - assuming default value %v", k, d.Value)
	}

	return settings, err
}

func LoadConfig(filename string) (*Settings, error) {
	settings := new(Settings)
	def, err := cfg.LoadConfigInto(filename, settings)

	for k, d := range def {
		logger.Log.Warnf("no config value for %s key found in any config - assuming default value %v", k, d.Value)
	}

	return settings, err
}

func (cfg *Settings) CheckConfig() error {
	//check log file is rw in production mode
	if cfg.Environment != "dev" {
		if _, err := os.Open(cfg.Log.Path); err != nil {
			return fmt.Errorf("could not open log file %s", cfg.Log.Path)
		}
	}

	//server settings
	if cfg.Server.UseTLS {
		if _, err := os.Open(cfg.Server.Cert); err != nil {
			return fmt.Errorf("could not open certificate %s", cfg.Server.Cert)
		}
		if _, err := os.Open(cfg.Server.Key); err != nil {
			return fmt.Errorf("could not open private key file %s", cfg.Server.Key)
		}
	}

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid port setting server_port is %d", cfg.Server.Port)
	}

	if _, err := os.Open(cfg.File.Location); err != nil {
		return fmt.Errorf("could not open file path %s", cfg.File.Location)
	}

	return nil
}

func (cfg *Settings) GenerateCSRF() (bool, error) {
	if len(cfg.CSRF.RandomKey) == 0 {

		var b []byte

		if _, err := os.Stat(csrfTokenFilename); os.IsNotExist(err) {
			//create a random csrf token
			r := utils.RandomSource{CharsToGen: utils.AlphaUpperLowerNumericSpecial}
			b = r.RandomSequence(32)

			err := ioutil.WriteFile(csrfTokenFilename, b, 0644)

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
