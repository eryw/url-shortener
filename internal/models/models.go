package models

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type URL struct {
	ID          int
	Code        string
	OriginalURL string
	Notes       string
	VisitCount  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Visit struct {
	ID        int
	URLID     int
	UserAgent string
	IPAddress string
	VisitedAt time.Time
}

func AdminExists(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM admin").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func CreateAdmin(db *sql.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO admin (username, password_hash) VALUES (?, ?)", username, string(hash))
	return err
}

func ResetAdmin(db *sql.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM admin")
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO admin (username, password_hash) VALUES (?, ?)", username, string(hash))
	return err
}

func GetAdmin(db *sql.DB, username string) (*Admin, error) {
	admin := &Admin{}
	err := db.QueryRow("SELECT id, username, password_hash, created_at FROM admin WHERE username = ?", username).
		Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.CreatedAt)
	if err != nil {
		return nil, err
	}
	return admin, nil
}

func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password))
	return err == nil
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func generateRandomCode(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func GenerateUniqueCode(db *sql.DB) (string, error) {
	length := 4
	maxAttempts := 3

	for {
		for attempt := 0; attempt < maxAttempts; attempt++ {
			code, err := generateRandomCode(length)
			if err != nil {
				return "", err
			}

			var exists int
			err = db.QueryRow("SELECT COUNT(*) FROM urls WHERE code = ?", code).Scan(&exists)
			if err != nil {
				return "", err
			}

			if exists == 0 {
				return code, nil
			}
		}
		length++
		if length > 20 {
			return "", errors.New("failed to generate unique code")
		}
	}
}

func CreateURL(db *sql.DB, code, originalURL, notes string) error {
	_, err := db.Exec("INSERT INTO urls (code, original_url, notes) VALUES (?, ?, ?)", code, originalURL, notes)
	return err
}

func GetURLByCode(db *sql.DB, code string) (*URL, error) {
	url := &URL{}
	err := db.QueryRow("SELECT id, code, original_url, notes, visit_count, created_at, updated_at FROM urls WHERE code = ?", code).
		Scan(&url.ID, &url.Code, &url.OriginalURL, &url.Notes, &url.VisitCount, &url.CreatedAt, &url.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func GetURLByID(db *sql.DB, id int) (*URL, error) {
	url := &URL{}
	err := db.QueryRow("SELECT id, code, original_url, notes, visit_count, created_at, updated_at FROM urls WHERE id = ?", id).
		Scan(&url.ID, &url.Code, &url.OriginalURL, &url.Notes, &url.VisitCount, &url.CreatedAt, &url.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func UpdateURLNotes(db *sql.DB, id int, notes string) error {
	_, err := db.Exec("UPDATE urls SET notes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", notes, id)
	return err
}

func DeleteURL(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM urls WHERE id = ?", id)
	return err
}

func SearchURLs(db *sql.DB, query, filter string, page, perPage int) ([]URL, int, error) {
	offset := (page - 1) * perPage

	var whereClause string
	var args []interface{}

	if query != "" {
		switch filter {
		case "original_url":
			whereClause = "WHERE original_url LIKE ?"
			args = append(args, "%"+query+"%")
		case "notes":
			whereClause = "WHERE notes LIKE ?"
			args = append(args, "%"+query+"%")
		default:
			whereClause = "WHERE original_url LIKE ? OR notes LIKE ?"
			args = append(args, "%"+query+"%", "%"+query+"%")
		}
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM urls " + whereClause
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := "SELECT id, code, original_url, notes, visit_count, created_at, updated_at FROM urls " + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, perPage, offset)

	rows, err := db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var url URL
		err := rows.Scan(&url.ID, &url.Code, &url.OriginalURL, &url.Notes, &url.VisitCount, &url.CreatedAt, &url.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		urls = append(urls, url)
	}

	return urls, total, nil
}

func IncrementVisitCount(db *sql.DB, code string) error {
	_, err := db.Exec("UPDATE urls SET visit_count = visit_count + 1 WHERE code = ?", code)
	return err
}

func RecordVisit(db *sql.DB, urlID int, userAgent, ipAddress string) (bool, error) {
	result, err := db.Exec("INSERT INTO visits (url_id, user_agent, ip_address) VALUES (?, ?, ?)", urlID, userAgent, ipAddress)
	if err != nil {
		// Check if it's a unique constraint violation (SQLite returns this as "UNIQUE constraint failed")
		if err.Error() == "UNIQUE constraint failed" ||
			(len(err.Error()) > 25 && err.Error()[:25] == "UNIQUE constraint failed") {
			return false, nil // Visit not recorded due to duplicate IP
		}
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func IsUserAuthenticated(r *http.Request, store sessions.Store) bool {
	session, _ := store.Get(r, "session")
	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}

func GetRealIP(r *http.Request) string {
	// Check for X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP from the list
		for i, char := range xff {
			if char == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr and extract IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If there's no port, return as-is
		return r.RemoteAddr
	}
	return ip
}
