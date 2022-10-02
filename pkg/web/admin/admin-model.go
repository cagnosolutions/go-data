package admin

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cagnosolutions/go-data/pkg/web/utils"
)

// adminModel is an admin controller
type adminModel struct {
	user *utils.SystemUser
	sess *utils.SessionStore
	tmpl *template.Template
}

// NewAdminController creates and returns a new *adminModel
func NewAdminController() *adminModel {
	return &adminModel{
		user: &utils.SystemUser{
			Username: "admin",
			Password: "admin",
			Role:     "admin",
		},
		sess: utils.NewSessionStore(
			&utils.SessionStoreConfig{
				SessionID: "webapp-sess-id",
				Domain:    "localhost",
				Timeout:   time.Duration(600) * time.Second,
			},
		),
		tmpl: AdminTemplates,
	}
}

// Routes are the routes for the admin controller
func (a *adminModel) Register(mux *http.ServeMux) {
	mux.Handle("/admin", a.handleAdmin())
	mux.Handle("/admin/", a.handleAdminHome())
	mux.Handle("/admin/login", a.handleAdminLogin())
	mux.Handle("/admin/logout", a.handleAdminLogout())
}

func (a *adminModel) RegisterLocalStaticHandler(mux *http.ServeMux) {
	// get current dir
	wd, err := os.Getwd()
	if err != nil {
		log.Panicf("woops: %s\n", err)
	}
	// get root path
	root := filepath.ToSlash(filepath.Clean(filepath.Join(wd, "resources/static/")))
	// return static
	mux.Handle("/static/", http.FileServer(http.Dir(root)))
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
		err := a.tmpl.ExecuteTemplate(w, "admin-home.gohtml", M{"LoggedIn": true})
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
