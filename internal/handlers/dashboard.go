package handlers

import (
	"database/sql"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"url-shortener/internal/models"

	"github.com/gorilla/sessions"
)

type DashboardHandler struct {
	db        *sql.DB
	store     *sessions.CookieStore
	templates *template.Template
}

func NewDashboardHandler(db *sql.DB, store *sessions.CookieStore, templates *template.Template) *DashboardHandler {
	return &DashboardHandler{
		db:        db,
		store:     store,
		templates: templates,
	}
}

func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	filter := r.URL.Query().Get("filter")
	pageStr := r.URL.Query().Get("page")

	if filter == "" {
		filter = "any"
	}

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 50
	urls, total, err := models.SearchURLs(h.db, query, filter, page, perPage)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	isHTMX := r.Header.Get("HX-Request") == "true"

	data := map[string]interface{}{
		"URLs":       urls,
		"Query":      query,
		"Filter":     filter,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
	}

	if isHTMX {
		if err := h.templates.ExecuteTemplate(w, "url_table.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *DashboardHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteURL(h.db, id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete URL", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
