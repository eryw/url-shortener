package handlers

import (
	"database/sql"
	"net/http"
	"url-shortener/internal/models"

	"github.com/gorilla/sessions"
)

type RedirectHandler struct {
	db    *sql.DB
	store sessions.Store
}

func NewRedirectHandler(db *sql.DB, store sessions.Store) *RedirectHandler {
	return &RedirectHandler{
		db:    db,
		store: store,
	}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]

	if code == "" {
		http.NotFound(w, r)
		return
	}

	url, err := models.GetURLByCode(h.db, code)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	isAuthenticated := models.IsUserAuthenticated(r, h.store)

	// Only count visit if user is not authenticated
	if !isAuthenticated {
		// Get user agent and IP address
		userAgent := r.Header.Get("User-Agent")
		ipAddress := models.GetRealIP(r)

		// Record detailed visit information
		visitRecorded, err := models.RecordVisit(h.db, url.ID, userAgent, ipAddress)
		if err != nil {
			// Log error but don't fail the redirect
			// In production, you might want to log this
		}

		// Only increment the visit count if this is a new unique visit
		if visitRecorded {
			models.IncrementVisitCount(h.db, code)
		}
	}

	http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}
