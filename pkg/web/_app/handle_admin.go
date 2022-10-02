package _app

import (
	"html/template"
	"net/http"
	"time"

	"github.com/cagnosolutions/go-data/pkg/web"
)

// adminModel is an admin controller
type adminModel struct {
	user   *web.SystemUser
	sess   *web.SessionStore
	tmpl   *template.Template
	routes *RouteMap
}

// NewAdminController creates and returns a new *adminModel
func NewAdminController() *adminModel {
	return &adminModel{
		user: &web.SystemUser{
			Username: "admin",
			Password: "admin",
			Role:     "admin",
		},
		sess: web.NewSessionStore(
			&web.SessionStoreConfig{
				SessionID: "webapp-sess-id",
				Domain:    "localhost",
				Timeout:   time.Duration(600) * time.Second,
			},
		),
		tmpl:   LoadTemplates(),
		routes: NewRouteMap(),
	}
}

// Routes are the routes for the admin controller
func (a *adminModel) Routes() *RouteMap {
	a.routes.Handle("/admin", a.handleAdmin())
	a.routes.Handle("/admin/", a.handleAdminHome())
	a.routes.Handle("/admin/login", a.handleAdminLogin())
	a.routes.Handle("/admin/logout", a.handleAdminLogout())
	return a.routes
}

// M is a data model map
type M map[string]interface{}

// handleAdmin GET -> /admin
func (a *adminModel) handleAdmin() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// reject post request
		if r.Method == http.MethodPost {
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			return
		}
		// check for session
		session, loggedIn := a.sess.Get(r)
		if !loggedIn {
			// if not logged in, redirect to login
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		// otherwise, update session and redirect to admin home
		a.sess.Save(w, r, session)
		http.Redirect(w, r, "/admin/", http.StatusFound)
		return
	}
	return http.HandlerFunc(fn)
}

// handleAdminHome GET -> /admin/
func (a *adminModel) handleAdminHome() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// reject post request
		if r.Method == http.MethodPost {
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			return
		}
		// check for session
		session, loggedIn := a.sess.Get(r)
		if !loggedIn {
			// if not logged in, redirect to login
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		// otherwise, update session and render admin home template
		a.sess.Save(w, r, session)
		err := a.tmpl.ExecuteTemplate(w, "admin-home.html", M{"LoggedIn": true})
		if err != nil {
			code := http.StatusExpectationFailed
			http.Error(w, http.StatusText(code), code)
		}
		return
	}
	return http.HandlerFunc(fn)
}

// handleAdminLogin -> GET/POST /admin/login
func (a *adminModel) handleAdminLogin() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// handle get request
		if r.Method == http.MethodGet {
			// attempt to render login template
			err := a.tmpl.ExecuteTemplate(w, "admin-login.gohtml", M{"LoggedIn": false})
			if err != nil {
				code := http.StatusExpectationFailed
				http.Error(w, http.StatusText(code), code)
			}
			return
		}
		// handle post request
		if r.Method == http.MethodPost {
			// check for form values
			un := r.FormValue("username")
			pw := r.FormValue("password")
			// attempt to authenticate
			if un != a.user.Username && pw != a.user.Password {
				// bad username or password
				code := http.StatusUnauthorized
				http.Error(w, http.StatusText(code), code)
				return
			}
			// authentication successful, start new session
			session := a.sess.New()
			session.Set("role", a.user.Role)
			a.sess.Save(w, r, session)
			// redirect to home
			http.Redirect(w, r, "/admin/", http.StatusFound)
			return
		}
	}
	return http.HandlerFunc(fn)
}

// handleAdmin GET -> /admin/logout
func (a *adminModel) handleAdminLogout() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// reject post request
		if r.Method == http.MethodPost {
			code := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(code), code)
			return
		}
		// check for session
		_, loggedIn := a.sess.Get(r)
		if !loggedIn {
			// if not logged in, redirect to login
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		// otherwise, update session and redirect to admin home
		a.sess.Save(w, r, nil)
		http.Redirect(w, r, "/admin/", http.StatusFound)
		return
	}
	return http.HandlerFunc(fn)
}
