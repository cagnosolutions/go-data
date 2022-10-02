package main

import (
	"log"

	"github.com/cagnosolutions/go-data/pkg/web"
	"github.com/cagnosolutions/go-data/pkg/web/admin"
)

const (
	serveAddr   = ":8080"
	serveStatic = "web/admin/resources/static/"
)

var appConf = &web.WebAppConfig{
	TemplateCacheConfig: nil,
	SessionStoreConfig:  nil,
	DefaultSystemUsers:  nil,
	HttpStaticPath:      "",
	HttpErrorsPath:      "",
	LoggingLevel:        0,
	ListenAddr:          serveAddr,
	AppName:             "",
}

func main() {
	app := web.NewWebApp(appConf)
	adminController := admin.NewAdminController()
	adminController.RegisterLocalStaticHandler(app.ServeMux)
	app.RegisterModel(adminController)
	log.Panic(app.ListenAndServe())
}

// func index() http.Handler {
// 	fn := func(w http.ResponseWriter, r *http.Request) {
// 		// implement handler logic here...
// 		web.Handle
// 	}
// 	return http.HandlerFunc(fn)
// }
//
// func index() http.Handler {
// 	fn := func(w http.ResponseWriter, r *http.Request) {
//
// 	}
// 	return http.HandlerFunc(fn)
// }
