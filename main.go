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

	"git.hoogi.eu/snafu/cfg"
	"git.hoogi.eu/snafu/go-blog/database"
	"git.hoogi.eu/snafu/go-blog/logger"
	"git.hoogi.eu/snafu/go-blog/mail"
	m "git.hoogi.eu/snafu/go-blog/middleware"
	"git.hoogi.eu/snafu/go-blog/models"
	"git.hoogi.eu/snafu/go-blog/routers"
	"git.hoogi.eu/snafu/go-blog/settings"
	"git.hoogi.eu/snafu/session"
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
		{
			Name:     "go-blog.conf",
			Path:     ".",
			Required: true,
		},
		{
			Name:     "go-blog.conf",
			Path:     "./custom",
			Required: false,
		},
	}

	config, err := settings.MergeConfigs(configFiles)

	if err != nil {
		exitCode = 1
		fmt.Println(err)
		return
	}

	config.BuildVersion = BuildVersion
	config.BuildGitHash = GitHash

	if err = config.CheckConfig(); err != nil {
		exitCode = 1
		fmt.Println(err)
		return
	}

	if config.Environment == "prod" {
		logFile, err := os.OpenFile(config.Log.File, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

		if err != nil {
			fmt.Println(err)
			exitCode = 1
		}

		logger.InitLogger(logFile, config.Log.Level)
	} else {
		logger.InitLogger(os.Stdout, config.Log.Level)
	}

	csrf, err := config.GenerateCSRF()

	if err != nil {
		exitCode = 1
		logger.Log.Error(err)
		return
	}

	if csrf {
		logger.Log.Info("a random key for CSRF protection was generated")
	}

	logger.Log.Infof("Go-Blog version: %s, commit: %s", BuildVersion, GitHash)
	logger.Log.Infof("running in %s mode", config.Environment)

	dbConf := database.SQLiteConfig{
		File: config.Database.File,
	}

	db, err := dbConf.Open()

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

	ctx, err := context(db, config)

	if err != nil {
		logger.Log.Error(err)
		exitCode = 1
		return
	}

	r := routers.InitRoutes(ctx, config)

	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port),
		Handler:        r,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Log.Infof("server will start at %s on port %d", config.Server.Address, config.Server.Port)

	if config.Server.UseTLS {
		err = s.ListenAndServeTLS(config.Server.Cert, config.Server.Key)

	} else {
		err = s.ListenAndServe()
	}

	if err != nil {
		exitCode = 1
		logger.Log.Errorf("failed to start server: %v", err)
		return
	}
}

func context(db *sql.DB, cfg *settings.Settings) (*m.AppContext, error) {
	ic := loadUserInterceptor(cfg.User.InterceptorPlugin)

	userService := models.UserService{
		Datasource: models.SQLiteUserDatasource{
			SQLConn: db,
		},
		Config:          cfg.User,
		UserInterceptor: ic,
	}

	userInviteService := models.UserInviteService{
		Datasource: models.SQLiteUserInviteDatasource{
			SQLConn: db,
		},
		UserService: userService,
	}

	articleService := models.ArticleService{
		AppConfig: cfg.Application,
		Datasource: models.SQLiteArticleDatasource{
			SQLConn: db,
		},
	}

	siteService := models.SiteService{
		Datasource: models.SQLiteSiteDatasource{
			SQLConn: db,
		},
	}

	fileService := models.FileService{
		Config: cfg.File,
		Datasource: models.SQLiteFileDatasource{
			SQLConn: db,
		},
	}

	categoryService := models.CategoryService{
		Datasource: models.SQLiteCategoryDatasource{
			SQLConn: db,
		},
	}

	tokenService := models.TokenService{
		Datasource: models.SQLiteTokenDatasource{
			SQLConn: db,
		},
	}

	smtpConfig := mail.SMTPConfig{
		Address:  cfg.Mail.Host,
		Port:     cfg.Mail.Port,
		User:     cfg.Mail.User,
		Password: []byte(cfg.Mail.Password),
	}

	sender := mail.NewMailService(cfg.Mail.SubjectPrefix, cfg.Mail.SenderAddress, smtpConfig)

	mailer := models.Mailer{
		Sender:    &sender,
		AppConfig: &cfg.Application,
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
		Templates:         tpl,
		UserService:       userService,
		UserInviteService: userInviteService,
		ArticleService:    articleService,
		CategoryService:   categoryService,
		SiteService:       siteService,
		FileService:       fileService,
		TokenService:      tokenService,
		Mailer:            mailer,
		SessionService:    &sessionService,
		ConfigService:     cfg,
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

	ui, err := symbol.(func() (models.UserInterceptor, error))()

	if err != nil {
		logger.Log.Error("unexpected type from module symbol", err)
		return nil

	}

	return ui
}
