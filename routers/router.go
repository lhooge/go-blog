// Copyright 2018 Lars Hoogestraat
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routers

import (
	"net/http"
	"os"

	c "git.hoogi.eu/go-blog/controllers"
	"git.hoogi.eu/go-blog/controllers/json"
	m "git.hoogi.eu/go-blog/middleware"
	"git.hoogi.eu/go-blog/settings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

//InitRoutes initializes restricted and public routes
func InitRoutes(ctx *m.AppContext, cfg *settings.Settings) *mux.Router {
	router := mux.NewRouter()
	router = router.StrictSlash(true)
	sr := router.PathPrefix("/").Subrouter()

	csrf :=
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

	restrictedChain := chain.Append(csrf).Append(ctx.AuthHandler)

	restrictedRoutes(ctx, ar, restrictedChain)

	router.NotFoundHandler = chain.Then(useTemplateHandler(ctx, m.NotFound))
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
	//article routes
	router.Handle("/articles", chain.Then(useTemplateHandler(ctx, c.AdminListArticlesHandler))).Methods("GET")
	router.Handle("/articles/page/{page}", chain.Then(useTemplateHandler(ctx, c.AdminListArticlesHandler))).Methods("GET")
	router.Handle("/article/new", chain.Then(useTemplateHandler(ctx, c.AdminArticleNewHandler))).Methods("GET")
	router.Handle("/article/new", chain.Then(useTemplateHandler(ctx, c.AdminArticleNewPostHandler))).Methods("POST")
	router.Handle("/article/edit/{articleID}", chain.Then(useTemplateHandler(ctx, c.AdminArticleEditHandler))).Methods("GET")
	router.Handle("/article/edit/{articleID}", chain.Then(useTemplateHandler(ctx, c.AdminArticleEditPostHandler))).Methods("POST")
	router.Handle("/article/publish/{articleID}", chain.Then(useTemplateHandler(ctx, c.AdminArticlePublishHandler))).Methods("GET")
	router.Handle("/article/publish/{articleID}", chain.Then(useTemplateHandler(ctx, c.AdminArticlePublishPostHandler))).Methods("POST")
	router.Handle("/article/delete/{articleID}", chain.Then(useTemplateHandler(ctx, c.AdminArticleDeleteHandler))).Methods("GET")
	router.Handle("/article/delete/{articleID}", chain.Then(useTemplateHandler(ctx, c.AdminArticleDeletePostHandler))).Methods("POST")

	//user routes
	router.Handle("/user/profile", chain.Then(useTemplateHandler(ctx, c.AdminProfileHandler))).Methods("GET")
	router.Handle("/user/profile", chain.Then(useTemplateHandler(ctx, c.AdminProfilePostHandler))).Methods("POST")
	router.Handle("/users", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminUsersHandler))).Methods("GET")
	router.Handle("/users/page/{page}", chain.Then(useTemplateHandler(ctx, c.AdminUsersHandler))).Methods("GET")
	router.Handle("/user/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminUserNewHandler))).Methods("GET")
	router.Handle("/user/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminUserNewPostHandler))).Methods("POST")
	router.Handle("/user/edit/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminUserEditHandler))).Methods("GET")
	router.Handle("/user/edit/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminUserEditPostHandler))).Methods("POST")
	router.Handle("/user/delete/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminUserDeleteHandler))).Methods("GET")
	router.Handle("/user/delete/{userID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminUserDeletePostHandler))).Methods("POST")

	//site routes
	router.Handle("/sites", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSitesHandler))).Methods("GET")
	router.Handle("/site/page/{page}", chain.Then(useTemplateHandler(ctx, c.AdminSitesHandler))).Methods("GET")
	router.Handle("/site/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSiteNewHandler))).Methods("GET")
	router.Handle("/site/new", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSiteNewPostHandler))).Methods("POST")
	router.Handle("/site/publish/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSitePublishHandler))).Methods("GET")
	router.Handle("/site/publish/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSitePublishPostHandler))).Methods("POST")
	router.Handle("/site/edit/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSiteEditHandler))).Methods("GET")
	router.Handle("/site/edit/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSiteEditPostHandler))).Methods("POST")
	router.Handle("/site/delete/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSiteDeleteHandler))).Methods("GET")
	router.Handle("/site/delete/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSiteDeletePostHandler))).Methods("POST")
	router.Handle("/site/order/{siteID}", chain.Append(ctx.RequireAdmin).Then(useTemplateHandler(ctx, c.AdminSiteOrderHandler))).Methods("POST")

	//file routes
	router.Handle("/files", chain.Then(useTemplateHandler(ctx, c.AdminListFilesHandler))).Methods("GET")
	router.Handle("/files/page/{page}", chain.Then(useTemplateHandler(ctx, c.AdminListFilesHandler))).Methods("GET")
	router.Handle("/file/upload", chain.Then(useTemplateHandler(ctx, c.AdminUploadFileHandler))).Methods("GET")
	router.Handle("/file/upload", chain.Then(useTemplateHandler(ctx, c.AdminUploadFilePostHandler))).Methods("POST")
	router.Handle("/file/delete/{fileID}", chain.Then(useTemplateHandler(ctx, c.AdminUploadDeleteHandler))).Methods("GET")
	router.Handle("/file/delete/{fileID}", chain.Then(useTemplateHandler(ctx, c.AdminUploadDeletePostHandler))).Methods("POST")

	router.Handle("/logout", chain.Then(useTemplateHandler(ctx, c.LogoutHandler))).Methods("GET")

	router.Handle("/json/session/keep-alive", chain.Then(useJSONHandler(ctx, json.KeepAliveSessionHandler))).Methods("GET")
}

func publicRoutes(ctx *m.AppContext, router *mux.Router, chain alice.Chain) {
	router.Handle("/", chain.Then(useTemplateHandler(ctx, c.ListArticlesHandler))).Methods("GET")
	router.Handle("/index", chain.Then(useTemplateHandler(ctx, c.IndexArticlesHandler))).Methods("GET")
	router.Handle("/site/{site}", chain.Then(useTemplateHandler(ctx, c.SiteHandler))).Methods("GET")
	router.Handle("/articles/page/{page}", chain.Then(useTemplateHandler(ctx, c.ListArticlesHandler))).Methods("GET")
	router.Handle("/article/{year}/{month}/{slug}", chain.Then(useTemplateHandler(ctx, c.GetArticleHandler))).Methods("GET")

	router.Handle("/file/{filename}", chain.Then(c.FileGetHandler(ctx))).Methods("GET")

	router.Handle("/admin", chain.Then(useTemplateHandler(ctx, c.LoginHandler))).Methods("GET")
	router.Handle("/admin", chain.Then(useTemplateHandler(ctx, c.LoginPostHandler))).Methods("POST")

	router.Handle("/admin/forgot-password", chain.Then(useTemplateHandler(ctx, c.ForgotPasswordHandler))).Methods("GET")
	router.Handle("/admin/forgot-password", chain.Then(useTemplateHandler(ctx, c.ForgotPasswordPostHandler))).Methods("POST")

	router.Handle("/admin/reset-password/{hash}", chain.Then(useTemplateHandler(ctx, c.ResetPasswordHandler))).Methods("GET")
	router.Handle("/admin/reset-password/{hash}", chain.Then(useTemplateHandler(ctx, c.ResetPasswordPostHandler))).Methods("POST")

}

func useTemplateHandler(ctx *m.AppContext, handler m.Handler) m.TemplateHandler {
	return m.TemplateHandler{AppCtx: ctx, Handler: handler}
}

func useJSONHandler(ctx *m.AppContext, handler m.JHandler) m.JSONHandler {
	return m.JSONHandler{AppCtx: ctx, Handler: handler}
}
