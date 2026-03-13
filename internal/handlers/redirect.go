package handlers

import (
	"database/sql"
	"net/http"
	"url-shortener/internal/models"
)

type RedirectHandler struct {
	db *sql.DB
}

func NewRedirectHandler(db *sql.DB) *RedirectHandler {
	return &RedirectHandler{db: db}
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

	models.IncrementVisitCount(h.db, code)

	http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}
