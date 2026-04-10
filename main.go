package main

import (
	"bufio"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/handlers"
	"url-shortener/internal/middleware"

	"github.com/gorilla/sessions"
)

//go:embed templates/*
var templatesFS embed.FS

func generateSecureSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func handleInit() error {
	envPath := ".env"

	if _, err := os.Stat(envPath); err == nil {
		fmt.Print(".env file already exists. Overwrite? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Init cancelled.")
			return nil
		}
	}

	secret, err := generateSecureSecret()
	if err != nil {
		return fmt.Errorf("failed to generate secure secret: %w", err)
	}

	envContent := fmt.Sprintf(`BASE_URL=http://localhost:8080
PORT=8080
DB_PATH=./data/urls.db
SESSION_SECRET=%s
RESET_ADMIN=false
`, secret)

	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	fmt.Printf("Created .env file with secure SESSION_SECRET.\n")
	return nil
}

func main() {
	initFlag := flag.Bool("init", false, "Generate .env file with default values and exit")
	flag.Parse()

	if *initFlag {
		if err := handleInit(); err != nil {
			log.Fatalf("Init failed: %v", err)
		}
		return
	}
	cfg := config.Load()

	db, err := database.Init(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}

	// Parse templates from embedded filesystem
	templatesSubFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(templatesSubFS, "*.html"))
	tmpl = template.Must(tmpl.ParseFS(templatesSubFS, "partials/*.html"))

	authMiddleware := middleware.NewAuthMiddleware(store, db)

	setupHandler := handlers.NewSetupHandler(db, store, tmpl, cfg.ResetAdmin)
	authHandler := handlers.NewAuthHandler(db, store, tmpl)
	dashboardHandler := handlers.NewDashboardHandler(db, store, tmpl)
	urlHandler := handlers.NewURLHandler(db, store, tmpl, cfg.BaseURL)
	redirectHandler := handlers.NewRedirectHandler(db, store)

	mux := http.NewServeMux()

	mux.HandleFunc("/setup", setupHandler.ShowSetup)
	mux.HandleFunc("/setup/create", setupHandler.CreateAdmin)
	mux.HandleFunc("/setup/reset", setupHandler.ShowReset)
	mux.HandleFunc("/setup/reset/confirm", setupHandler.ResetAdmin)

	mux.HandleFunc("/login", authMiddleware.RedirectIfAuth(authHandler.ShowLogin))
	mux.HandleFunc("/login/submit", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)

	mux.HandleFunc("/admin/dashboard", authMiddleware.RequireAuth(dashboardHandler.ShowDashboard))
	mux.HandleFunc("/admin/urls", authMiddleware.RequireAuth(dashboardHandler.ShowDashboard))
	mux.HandleFunc("/admin/urls/new", authMiddleware.RequireAuth(urlHandler.ShowAddForm))
	mux.HandleFunc("/admin/urls/create", authMiddleware.RequireAuth(urlHandler.CreateURL))
	mux.HandleFunc("/admin/urls/edit", authMiddleware.RequireAuth(urlHandler.ShowEditForm))
	mux.HandleFunc("/admin/urls/update", authMiddleware.RequireAuth(urlHandler.UpdateURL))
	mux.HandleFunc("/admin/urls/delete", authMiddleware.RequireAuth(dashboardHandler.DeleteURL))
	mux.HandleFunc("/admin/urls/info", authMiddleware.RequireAuth(urlHandler.GetURLInfo))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/admin/") ||
			strings.HasPrefix(r.URL.Path, "/login") ||
			strings.HasPrefix(r.URL.Path, "/logout") ||
			strings.HasPrefix(r.URL.Path, "/setup") {
			http.NotFound(w, r)
			return
		}

		redirectHandler.Redirect(w, r)
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Base URL: %s", cfg.BaseURL)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
