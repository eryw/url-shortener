package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"url-shortener/internal/models"

	"github.com/gorilla/sessions"
)

type AuthHandler struct {
	db        *sql.DB
	store     *sessions.CookieStore
	templates *template.Template
}

func NewAuthHandler(db *sql.DB, store *sessions.CookieStore, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		db:        db,
		store:     store,
		templates: templates,
	}
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	exists, err := models.AdminExists(h.db)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Error": "",
	}

	if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	admin, err := models.GetAdmin(h.db, username)
	if err != nil {
		data := map[string]interface{}{
			"Error": "Invalid username or password",
		}
		h.templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	if !admin.CheckPassword(password) {
		data := map[string]interface{}{
			"Error": "Invalid username or password",
		}
		h.templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	session, _ := h.store.Get(r, "session")
	session.Values["authenticated"] = true
	session.Values["username"] = username
	session.Save(r, w)

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "session")
	session.Values["authenticated"] = false
	delete(session.Values, "username")
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
