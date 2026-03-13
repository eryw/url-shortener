package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"url-shortener/internal/models"

	"github.com/gorilla/sessions"
)

type SetupHandler struct {
	db            *sql.DB
	store         *sessions.CookieStore
	templates     *template.Template
	resetAdminEnv bool
}

func NewSetupHandler(db *sql.DB, store *sessions.CookieStore, templates *template.Template, resetAdminEnv bool) *SetupHandler {
	return &SetupHandler{
		db:            db,
		store:         store,
		templates:     templates,
		resetAdminEnv: resetAdminEnv,
	}
}

func (h *SetupHandler) ShowSetup(w http.ResponseWriter, r *http.Request) {
	exists, err := models.AdminExists(h.db)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Error": "",
	}

	if err := h.templates.ExecuteTemplate(w, "setup.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *SetupHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	exists, err := models.AdminExists(h.db)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		data := map[string]interface{}{
			"Error": "Username and password are required",
		}
		h.templates.ExecuteTemplate(w, "setup.html", data)
		return
	}

	if err := models.CreateAdmin(h.db, username, password); err != nil {
		data := map[string]interface{}{
			"Error": "Failed to create admin account",
		}
		h.templates.ExecuteTemplate(w, "setup.html", data)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *SetupHandler) ShowReset(w http.ResponseWriter, r *http.Request) {
	if !h.resetAdminEnv {
		http.Error(w, "Reset admin is disabled. Set RESET_ADMIN=true in .env", http.StatusForbidden)
		return
	}

	data := map[string]interface{}{
		"Error": "",
	}

	if err := h.templates.ExecuteTemplate(w, "setup.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *SetupHandler) ResetAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.resetAdminEnv {
		http.Error(w, "Reset admin is disabled. Set RESET_ADMIN=true in .env", http.StatusForbidden)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		data := map[string]interface{}{
			"Error": "Username and password are required",
		}
		h.templates.ExecuteTemplate(w, "setup.html", data)
		return
	}

	if err := models.ResetAdmin(h.db, username, password); err != nil {
		data := map[string]interface{}{
			"Error": "Failed to reset admin account",
		}
		h.templates.ExecuteTemplate(w, "setup.html", data)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
