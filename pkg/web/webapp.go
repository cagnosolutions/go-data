package web

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/cagnosolutions/go-data/pkg/web/middleware"
	"github.com/cagnosolutions/go-data/pkg/web/utils"
)

type WebAppConfig struct {

	// template configuration items
	*utils.TemplateCacheConfig
	// TemplateBasePath string
	// TemplateStubPath string
	// TemplateFuncMap  template.FuncMap

	// session configuration items
	*utils.SessionStoreConfig
	// SessionID      string
	// SessionTimeout int

	// default users
	DefaultSystemUsers []*utils.SystemUser

	// muxer configuration items
	HttpStaticPath string
	HttpErrorsPath string

	// server configuration items
	LoggingLevel utils.LogLevel
	ListenAddr   string
	AppName      string
}

type WebApp struct {

	// template cache
	templates *utils.TemplateCache

	// session store
	sessions *utils.SessionStore

	// default users
	users []*utils.SystemUser

	// muxer configuration items
	ServeMux       *http.ServeMux
	httpStaticPath string
	httpErrorsPath string

	// server configuration items
	logging    *utils.Logger
	listenAddr string
	appName    string
}

func NewWebApp(conf *WebAppConfig) *WebApp {
	if conf == nil {
		panic("webapp configuration: requires a valid configuration")
	}
	app := &WebApp{
		ServeMux: http.NewServeMux(),
	}
	if conf.TemplateCacheConfig != nil {
		app.templates = utils.NewTemplateCache(conf.TemplateCacheConfig)
	}
	if conf.SessionStoreConfig != nil {
		app.sessions = utils.NewSessionStore(conf.SessionStoreConfig)
	}
	if conf.DefaultSystemUsers != nil {
		if conf.SessionStoreConfig == nil {
			panic("webapp configuration: cannot configure system users without sessions enabled")
		}
		app.users = conf.DefaultSystemUsers
	}
	if conf.HttpStaticPath != "" {
		app.httpStaticPath = conf.HttpStaticPath
		staticHandler := middleware.HandleStatic("/static/", app.httpStaticPath)
		log.Printf("serving static: %s, %s\n", filepath.Base(app.httpStaticPath), app.httpStaticPath)
		app.ServeMux.Handle("/static/", staticHandler)
	}
	if conf.HttpErrorsPath != "" {
		app.httpErrorsPath = conf.HttpErrorsPath
	}
	if conf.LoggingLevel < utils.LevelOff {
		app.logging = utils.NewLogger(conf.LoggingLevel)
	}
	if conf.ListenAddr != "" {
		app.listenAddr = conf.ListenAddr
	}
	if conf.AppName != "" {
		app.appName = conf.AppName
	}
	return app
}

func (app *WebApp) ListenAndServe() error {
	utils.HandleSignalInterrupt("%q started, and running on %s...\n", app.appName, app.listenAddr)
	return http.ListenAndServe(app.listenAddr, app.ServeMux)
}

func (app *WebApp) RegisterModel(reg Registerer) {
	reg.Register(app.ServeMux)
}

type Registerer interface {
	Register(mux *http.ServeMux)
}
