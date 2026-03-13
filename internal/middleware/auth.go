package middleware

import (
	"database/sql"
	"net/http"
	"url-shortener/internal/models"

	"github.com/gorilla/sessions"
)

type AuthMiddleware struct {
	store *sessions.CookieStore
	db    *sql.DB
}

func NewAuthMiddleware(store *sessions.CookieStore, db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{
		store: store,
		db:    db,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exists, err := models.AdminExists(m.db)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Redirect(w, r, "/setup", http.StatusSeeOther)
			return
		}

		session, _ := m.store.Get(r, "session")
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

func (m *AuthMiddleware) RedirectIfAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := m.store.Get(r, "session")
		auth, ok := session.Values["authenticated"].(bool)
		if ok && auth {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
