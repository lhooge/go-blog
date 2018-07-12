// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"plugin"
	"time"

	"git.hoogi.eu/cfg"
	"git.hoogi.eu/go-blog/components/database"
	"git.hoogi.eu/go-blog/components/logger"
	"git.hoogi.eu/go-blog/components/mail"
	m "git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/models"
	"git.hoogi.eu/go-blog/routers"
	"git.hoogi.eu/go-blog/settings"
	"git.hoogi.eu/session"
)

var (
	BuildVersion = "develop"
	GitHash      = ""
)

func main() {
	var exitCode int

	defer func() {
		os.Exit(exitCode)
	}()

	configFiles := []cfg.File{
		cfg.File{
			Name:     "go-blog.conf",
			Path:     ".",
			Required: true,
		},
		cfg.File{
			Name:     "go-blog.conf",
			Path:     "./custom",
			Required: false,
		},
	}

	cfg, err := settings.MergeConfigs(configFiles)

	if err != nil {
		exitCode = 1
		fmt.Println(err)
		return
	}

	if err = cfg.CheckConfig(); err != nil {
		exitCode = 1
		fmt.Println(err)
		return
	}

	if cfg.Environment == "prod" {
		logFile, err := os.OpenFile(cfg.Log.File, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

		if err != nil {
			fmt.Println(err)
			exitCode = 1
		}
		logger.InitLogger(logFile, cfg.Log.Level)
	} else {
		logger.InitLogger(os.Stderr, cfg.Log.Level)
	}

	csrf, err := cfg.GenerateCSRF()

	if err != nil {
		exitCode = 1
		logger.Log.Error(err)
		return
	}

	if csrf {
		logger.Log.Info("a random key for CSRF protection was generated")
	}

	logger.Log.Infof("Go-Blog version: %s, commit: %s", BuildVersion, GitHash)
	logger.Log.Infof("running in %s mode", cfg.Environment)

	var db *sql.DB
	if cfg.Database.Engine == settings.MySQL {
		mysqlConf := database.MySQLConfig{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Database: cfg.Database.Name,
		}

		db, err = mysqlConf.Open()
	} else {
		sqliteConf := database.SQLiteConfig{
			File: cfg.Database.File,
		}

		db, err = sqliteConf.Open()
	}

	if err != nil {
		logger.Log.Error(err)
		exitCode = 1
		return
	}

	defer func() {
		if db != nil {
			err = db.Close()
			logger.Log.Error(err)
		}
	}()

	ctx, err := context(db, cfg)

	if err != nil {
		logger.Log.Error(err)
		exitCode = 1
		return
	}

	r := routers.InitRoutes(ctx, cfg)

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Log.Infof("server will start at %s on port %d", cfg.Server.Address, cfg.Server.Port)

	if cfg.Server.UseTLS {
		err := s.ListenAndServeTLS(cfg.Server.Cert, cfg.Server.Key)
		if err != nil {
			exitCode = 1
			logger.Log.Error("failed to serve TLS server: ", err)
			return
		}
	} else {
		err := s.ListenAndServe()

		if err != nil {
			exitCode = 1
			logger.Log.Error("failed to start server ", err)
			return
		}
	}
}

func context(db *sql.DB, cfg *settings.Settings) (*m.AppContext, error) {
	ic := loadUserInterceptor(cfg.InterceptorPlugin)

	var uds models.UserDatasourceService
	var ads models.ArticleDatasourceService
	var sds models.SiteDatasourceService
	var fds models.FileDatasourceService
	var tds models.TokenDatasourceService

	if cfg.Database.Engine == settings.MySQL {
		uds = models.MySQLUserDatasource{
			SQLConn: db,
		}

		ads = models.MySQLArticleDatasource{
			SQLConn: db,
		}

		sds = models.MySQLSiteDatasource{
			SQLConn: db,
		}

		fds = models.MySQLFileDatasource{
			SQLConn: db,
		}

		tds = models.MySQLTokenDatasource{
			SQLConn: db,
		}
	} else {
		uds = models.SQLiteUserDatasource{
			SQLConn: db,
		}

		ads = models.SQLiteArticleDatasource{
			SQLConn: db,
		}

		sds = models.SQLiteSiteDatasource{
			SQLConn: db,
		}

		fds = models.SQLiteFileDatasource{
			SQLConn: db,
		}

		tds = models.SQLiteTokenDatasource{
			SQLConn: db,
		}
	}

	userService := models.UserService{
		Datasource:      uds,
		Config:          cfg.User,
		UserInterceptor: ic,
	}

	articleService := models.ArticleService{
		Datasource: ads,
	}

	siteService := models.SiteService{
		Datasource: sds,
	}

	fileService := models.FileService{
		Datasource: fds,
	}

	tokenService := models.TokenService{
		Datasource: tds,
	}

	smtpConfig := mail.SMTPConfig{
		Address:  cfg.Mail.Host,
		Port:     cfg.Mail.Port,
		User:     cfg.Mail.User,
		Password: []byte(cfg.Mail.Password),
	}

	mailService := mail.Service{
		SMTPConfig:    smtpConfig,
		From:          cfg.Mail.SenderAddress,
		SubjectPrefix: cfg.Mail.SubjectPrefix,
	}

	templates := m.Templates{
		Directory: "./templates",
		FuncMap:   m.FuncMap(siteService, cfg),
	}

	tpl, err := templates.Load()

	if err != nil {
		return nil, err
	}

	sessionService := session.SessionService{
		Path:            cfg.Session.CookiePath,
		Name:            cfg.Session.CookieName,
		Secure:          cfg.Session.CookieSecure,
		HTTPOnly:        true,
		SessionProvider: session.NewInMemoryProvider(),
		IdleSessionTTL:  cfg.Session.TTL.Nanoseconds() / 1e9,
	}

	ticker := time.NewTicker(cfg.Session.GarbageCollection)
	sessionService.InitGC(ticker, cfg.Session.TTL)

	return &m.AppContext{
		Templates:      tpl,
		UserService:    userService,
		ArticleService: articleService,
		SiteService:    siteService,
		FileService:    fileService,
		TokenService:   tokenService,
		MailService:    mailService,
		SessionService: &sessionService,
		ConfigService:  cfg,
	}, nil
}

func loadUserInterceptor(pluginFile string) models.UserInterceptor {
	if len(pluginFile) == 0 {
		return nil
	}

	p, err := plugin.Open(pluginFile)

	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	symbol, err := p.Lookup("GetUserInterceptor")
	if err != nil {
		logger.Log.Error(err)
		return nil
	}

	return symbol.(func() models.UserInterceptor)()
}
