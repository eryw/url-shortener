package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"url-shortener/internal/models"

	"github.com/gorilla/sessions"
)

type URLHandler struct {
	db        *sql.DB
	store     *sessions.CookieStore
	templates *template.Template
	baseURL   string
}

func NewURLHandler(db *sql.DB, store *sessions.CookieStore, templates *template.Template, baseURL string) *URLHandler {
	return &URLHandler{
		db:        db,
		store:     store,
		templates: templates,
		baseURL:   baseURL,
	}
}

func (h *URLHandler) ShowAddForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Error": "",
	}

	if err := h.templates.ExecuteTemplate(w, "modal_add.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *URLHandler) CreateURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	originalURL := r.FormValue("original_url")
	notes := r.FormValue("notes")

	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		data := map[string]interface{}{
			"Error": "Original URL is required",
		}
		h.templates.ExecuteTemplate(w, "modal_add.html", data)
		return
	}

	if _, err := url.ParseRequestURI(originalURL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data := map[string]interface{}{
			"Error": "Invalid URL format",
		}
		h.templates.ExecuteTemplate(w, "modal_add.html", data)
		return
	}

	code, err := models.GenerateUniqueCode(h.db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data := map[string]interface{}{
			"Error": "Failed to generate unique code",
		}
		h.templates.ExecuteTemplate(w, "modal_add.html", data)
		return
	}

	if err := models.CreateURL(h.db, code, originalURL, notes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data := map[string]interface{}{
			"Error": "Failed to create URL",
		}
		h.templates.ExecuteTemplate(w, "modal_add.html", data)
		return
	}

	w.Header().Set("HX-Redirect", "/admin/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (h *URLHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	urlObj, err := models.GetURLByID(h.db, id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"URL":   urlObj,
		"Error": "",
	}

	if err := h.templates.ExecuteTemplate(w, "modal_edit.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *URLHandler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	notes := r.FormValue("notes")

	if err := models.UpdateURLNotes(h.db, id, notes); err != nil {
		urlObj, _ := models.GetURLByID(h.db, id)
		w.WriteHeader(http.StatusInternalServerError)
		data := map[string]interface{}{
			"URL":   urlObj,
			"Error": "Failed to update URL",
		}
		h.templates.ExecuteTemplate(w, "modal_edit.html", data)
		return
	}

	w.Header().Set("HX-Redirect", "/admin/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (h *URLHandler) GetURLInfo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	urlObj, err := models.GetURLByID(h.db, id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           urlObj.ID,
		"code":         urlObj.Code,
		"original_url": urlObj.OriginalURL,
		"short_url":    h.baseURL + "/" + urlObj.Code,
		"notes":        urlObj.Notes,
	})
}
