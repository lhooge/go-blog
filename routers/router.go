// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"net/http"
	"os"

	"git.hoogi.eu/snafu/go-blog/handler"
	m "git.hoogi.eu/snafu/go-blog/middleware"
	"git.hoogi.eu/snafu/go-blog/settings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

// InitRoutes initializes restricted and public routes
func InitRoutes(ctx *m.AppContext, cfg *settings.Settings) *mux.Router {
	router := mux.NewRouter()
	router = router.StrictSlash(false)
	sr := router.PathPrefix("/").Subrouter()

	rf :=
		csrf.Protect([]byte(cfg.CSRF.RandomKey),
			csrf.Secure(cfg.CSRF.CookieSecure),
			csrf.FieldName(cfg.CSRF.CookieName),
			csrf.Path(cfg.CSRF.CookiePath),
			csrf.CookieName(cfg.CSRF.CookieName))

	chain := alice.New()

	if cfg.Log.Access {
		if cfg.Environment == "dev" {
			chain = chain.Append(stdOutLoggingHandler)
		} else {
			chain = chain.Append(fileLoggingHandler(cfg.Log.AccessFile))
		}
	}

	publicRoutes(ctx, sr, chain)

	ar := router.PathPrefix("/admin").Subrouter()

	restrictedChain := chain.Append(rf).Append(ctx.AuthHandler)

	restrictedRoutes(ctx, ar, restrictedChain)

	router.NotFoundHandler = chain.Then(useTemplateHandler(ctx, m.NotFound))

	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.Application.Favicon)
	})

	if len(cfg.Application.RobotsTxt) > 0 {
		router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, cfg.Application.RobotsTxt)
		})
	}

	if len(cfg.Application.CustomCSS) > 0 {
		router.HandleFunc("/assets/css/custom.css", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, cfg.Application.CustomCSS)
		})
	}

	http.Handle("/", router)

	// File handler for static files
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	return router
}

func stdOutLoggingHandler(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, h)
}

func fileLoggingHandler(accessLogPath string) (flh func(http.Handler) http.Handler) {
	al, _ := os.OpenFile(accessLogPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	flh = func(h http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(al, h)
	}
	return
}

func restrictedRoutes(ctx *m.AppContext, router *mux.Router, chain alice.Chain) {
	// article
	router.Handle("/articles", chain.Then(useTemplateHandler(ctx, handler.AdminListArticlesHandler))).Methods("GET")
	router.Handle("/articles/page/{page}", chain.Then(useTemplateHandler(ctx, handler.AdminListArticlesHandler))).Methods("GET")
	router.Handle("/article/new", chain.Then(useTemplateHandler(ctx, handler.AdminArticleNewHandler))).Methods("GET")
	router.Handle("/article/new", chain.Then(useTemplateHandler(ctx, handler.AdminArticleNewPostHandler))).Methods("POST")
	router.Handle("/article/edit/{articleID}", chain.Then(useTemplateHandler(ctx, handler.AdminArticleEditHandler))).Methods("GET")
	router.Handle("/article/edit/{articleID}", chain.Then(useTemplateHandler(ctx, handler.AdminArticleEditPostHandler))).Methods("POST")
	router.Handle("/article/publish/{articleID}", chain.Then(useTemplateHandler(ctx, handler.AdminArticlePublishHandler))).Methods("GET")
	router.Handle("/article/publish/{articleID}", chain.Then(useTemplateHandler(ctx, handler.AdminArticlePublishPostHandler))).Methods("POST")
	router.Handle("/article/delete/{articleID}", chain.Then(useTemplateHandler(ctx, handler.AdminArticleDeleteHandler))).Methods("GET")
	router.Handle("/article/delete/{articleID}", chain.Then(useTemplateHandler(ctx, handler.AdminArticleDeletePostHandler))).Methods("POST")
	router.Handle("/article/{articleID}", chain.Then(useTemplateHandler(ctx, handler.AdminPreviewArticleByIDHandler))).Methods("GET")

	// user
	router.Handle("/user/profile", chain.Then(useTemplateHandler(ctx, handler.AdminProfileHandler))).Methods("GET")
	router.Handle("/user/profile", chain.Then(useTemplateHandler(ctx, handler.AdminProfilePostHandler))).Methods("POST")
	router.Handle("/users", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUsersHandler))).Methods("GET")
	router.Handle("/users/page/{page}", chain.Then(useTemplateHandler(ctx, handler.AdminUsersHandler))).Methods("GET")
	router.Handle("/user/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserNewHandler))).Methods("GET")
	router.Handle("/user/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserNewPostHandler))).Methods("POST")
	router.Handle("/user/edit/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserEditHandler))).Methods("GET")
	router.Handle("/user/edit/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserEditPostHandler))).Methods("POST")
	router.Handle("/user/delete/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserDeleteHandler))).Methods("GET")
	router.Handle("/user/delete/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserDeletePostHandler))).Methods("POST")

	// user invites
	router.Handle("/user-invite/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserInviteNewHandler))).Methods("GET")
	router.Handle("/user-invite/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserInviteNewPostHandler))).Methods("POST")
	router.Handle("/user-invite/resend/{inviteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserInviteResendPostHandler))).Methods("POST")
	router.Handle("/user-invite/delete/{inviteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserInviteDeleteHandler))).Methods("GET")
	router.Handle("/user-invite/delete/{inviteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminUserInviteDeletePostHandler))).Methods("POST")

	// site
	router.Handle("/sites", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSitesHandler))).Methods("GET")
	router.Handle("/site/page/{page}", chain.Then(useTemplateHandler(ctx, handler.AdminSitesHandler))).Methods("GET")
	router.Handle("/site/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSiteNewHandler))).Methods("GET")
	router.Handle("/site/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSiteNewPostHandler))).Methods("POST")
	router.Handle("/site/publish/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSitePublishHandler))).Methods("GET")
	router.Handle("/site/publish/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSitePublishPostHandler))).Methods("POST")
	router.Handle("/site/edit/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSiteEditHandler))).Methods("GET")
	router.Handle("/site/edit/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSiteEditPostHandler))).Methods("POST")
	router.Handle("/site/delete/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSiteDeleteHandler))).Methods("GET")
	router.Handle("/site/delete/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSiteDeletePostHandler))).Methods("POST")
	router.Handle("/site/order/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, handler.AdminSiteOrderHandler))).Methods("POST")
	router.Handle("/site/{siteID:[0-9]+}}", chain.Then(useTemplateHandler(ctx, handler.AdminGetSiteHandler))).Methods("GET")

	// article
	router.Handle("/categories", chain.Then(useTemplateHandler(ctx, handler.AdminListCategoriesHandler))).Methods("GET")
	router.Handle("/category/{categoryID:[0-9]+}}", chain.Then(useTemplateHandler(ctx, handler.AdminGetCategoryHandler))).Methods("POST")
	router.Handle("/category/new", chain.Then(useTemplateHandler(ctx, handler.AdminCategoryNewHandler))).Methods("GET")
	router.Handle("/category/new", chain.Then(useTemplateHandler(ctx, handler.AdminCategoryNewPostHandler))).Methods("POST")
	router.Handle("/category/edit/{categoryID}", chain.Then(useTemplateHandler(ctx, handler.AdminCategoryEditHandler))).Methods("GET")
	router.Handle("/category/edit/{categoryID}", chain.Then(useTemplateHandler(ctx, handler.AdminCategoryEditPostHandler))).Methods("POST")
	router.Handle("/category/delete/{categoryID}", chain.Then(useTemplateHandler(ctx, handler.AdminCategoryDeleteHandler))).Methods("GET")
	router.Handle("/category/delete/{categoryID}", chain.Then(useTemplateHandler(ctx, handler.AdminCategoryDeletePostHandler))).Methods("POST")

	// file
	router.Handle("/files", chain.Then(useTemplateHandler(ctx, handler.AdminListFilesHandler))).Methods("GET")
	router.Handle("/files/page/{page}", chain.Then(useTemplateHandler(ctx, handler.AdminListFilesHandler))).Methods("GET")
	router.Handle("/file/upload", chain.Then(useTemplateHandler(ctx, handler.AdminUploadFileHandler))).Methods("GET")
	router.Handle("/file/upload", chain.Then(useTemplateHandler(ctx, handler.AdminUploadFilePostHandler))).Methods("POST")
	router.Handle("/file/toggleInline/{fileID}", chain.Then(useTemplateHandler(ctx, handler.AdminToggleInlineFilePostHandler))).Methods("POST")
	router.Handle("/file/delete/{fileID}", chain.Then(useTemplateHandler(ctx, handler.AdminUploadDeleteHandler))).Methods("GET")
	router.Handle("/file/delete/{fileID}", chain.Then(useTemplateHandler(ctx, handler.AdminUploadDeletePostHandler))).Methods("POST")

	router.Handle("/logout", chain.Then(useTemplateHandler(ctx, handler.LogoutHandler))).Methods("GET")

	router.Handle("/json/session/keep-alive", chain.Then(useJSONHandler(ctx, handler.KeepAliveSessionHandler))).Methods("POST")
	router.Handle("/json/file/upload", chain.Then(useJSONHandler(ctx, handler.AdminUploadJSONFilePostHandler))).Methods("POST")
}

func publicRoutes(ctx *m.AppContext, router *mux.Router, chain alice.Chain) {
	fh := handler.FileHandler{
		Context: ctx,
	}

	router.Handle("/", chain.Then(useTemplateHandler(ctx, handler.ListArticlesHandler))).Methods("GET")
	router.Handle("/articles/category/{categorySlug}", chain.Then(useTemplateHandler(ctx, handler.ListArticlesCategoryHandler))).Methods("GET")
	router.Handle("/articles/category/{categorySlug}/{page}", chain.Then(useTemplateHandler(ctx, handler.ListArticlesCategoryHandler))).Methods("GET")
	router.Handle("/index", chain.Then(useTemplateHandler(ctx, handler.IndexArticlesHandler))).Methods("GET")
	router.Handle("/index/category/{categorySlug}", chain.Then(useTemplateHandler(ctx, handler.IndexArticlesCategoryHandler))).Methods("GET")

	router.Handle("/articles/page/{page}", chain.Then(useTemplateHandler(ctx, handler.ListArticlesHandler))).Methods("GET")
	router.Handle("/article/{year}/{month}/{slug}", chain.Then(useTemplateHandler(ctx, handler.GetArticleHandler))).Methods("GET")
	router.Handle("/article/by-id/{articleID}", chain.Then(useTemplateHandler(ctx, handler.GetArticleByIDHandler))).Methods("GET")

	router.Handle("/rss.xml", chain.Then(useXMLHandler(ctx, handler.RSSFeed))).Methods("GET")

	router.Handle("/site/{site}", chain.Then(useTemplateHandler(ctx, handler.GetSiteHandler))).Methods("GET")

	router.Handle("/file/{uniquename}", chain.ThenFunc(fh.FileGetHandler)).Methods("GET")

	router.Handle("/admin", chain.Then(useTemplateHandler(ctx, handler.LoginHandler))).Methods("GET")
	router.Handle("/admin", chain.Then(useTemplateHandler(ctx, handler.LoginPostHandler))).Methods("POST")

	router.Handle("/admin/forgot-password", chain.Then(useTemplateHandler(ctx, handler.ForgotPasswordHandler))).Methods("GET")
	router.Handle("/admin/forgot-password", chain.Then(useTemplateHandler(ctx, handler.ForgotPasswordPostHandler))).Methods("POST")

	router.Handle("/admin/reset-password/{hash}", chain.Then(useTemplateHandler(ctx, handler.ResetPasswordHandler))).Methods("GET")
	router.Handle("/admin/reset-password/{hash}", chain.Then(useTemplateHandler(ctx, handler.ResetPasswordPostHandler))).Methods("POST")

	router.Handle("/admin/activate-account/{hash}", chain.Then(useTemplateHandler(ctx, handler.ActivateAccountHandler))).Methods("GET")
	router.Handle("/admin/activate-account/{hash}", chain.Then(useTemplateHandler(ctx, handler.ActivateAccountPostHandler))).Methods("POST")
}

func useTemplateHandler(ctx *m.AppContext, handler m.Handler) m.TemplateHandler {
	return m.TemplateHandler{AppCtx: ctx, Handler: handler}
}

func useJSONHandler(ctx *m.AppContext, handler m.JHandler) m.JSONHandler {
	return m.JSONHandler{AppCtx: ctx, Handler: handler}
}

func useXMLHandler(ctx *m.AppContext, handler m.XHandler) m.XMLHandler {
	return m.XMLHandler{AppCtx: ctx, Handler: handler}
}
