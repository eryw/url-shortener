package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Init(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS admin (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		notes TEXT,
		visit_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS visits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url_id INTEGER NOT NULL,
		user_agent TEXT,
		ip_address TEXT,
		visited_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
		UNIQUE(url_id, ip_address)
	);

	CREATE INDEX IF NOT EXISTS idx_urls_code ON urls(code);
	CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
	CREATE INDEX IF NOT EXISTS idx_visits_url_id ON visits(url_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	var columnExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM pragma_table_info('urls') 
		WHERE name='visit_count'
	`).Scan(&columnExists)

	if err != nil {
		return err
	}

	if !columnExists {
		_, err = db.Exec("ALTER TABLE urls ADD COLUMN visit_count INTEGER DEFAULT 0")
		if err != nil {
			return err
		}
	}

	// Check if visits table exists and create it if it doesn't
	var tableExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='visits'
	`).Scan(&tableExists)

	if err != nil {
		return err
	}

	if !tableExists {
		_, err = db.Exec(`
			CREATE TABLE visits (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				url_id INTEGER NOT NULL,
				user_agent TEXT,
				ip_address TEXT,
				visited_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
				UNIQUE(url_id, ip_address)
			)
		`)
		if err != nil {
			return err
		}

		_, err = db.Exec("CREATE INDEX idx_visits_url_id ON visits(url_id)")
		if err != nil {
			return err
		}
	} else {
		// Check if unique constraint exists and add it if it doesn't
		var constraintExists bool
		err = db.QueryRow(`
			SELECT COUNT(*) > 0 
			FROM pragma_index_list('visits') 
			WHERE name='sqlite_autoindex_visits_1'
		`).Scan(&constraintExists)

		if err != nil {
			return err
		}

		if !constraintExists {
			// For SQLite, we need to recreate the table to add unique constraint
			_, err = db.Exec(`
				CREATE TABLE visits_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					url_id INTEGER NOT NULL,
					user_agent TEXT,
					ip_address TEXT,
					visited_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
					UNIQUE(url_id, ip_address)
				)
			`)
			if err != nil {
				return err
			}

			_, err = db.Exec(`
				INSERT INTO visits_new (id, url_id, user_agent, ip_address, visited_at)
				SELECT id, url_id, user_agent, ip_address, visited_at FROM visits
			`)
			if err != nil {
				return err
			}

			_, err = db.Exec("DROP TABLE visits")
			if err != nil {
				return err
			}

			_, err = db.Exec("ALTER TABLE visits_new RENAME TO visits")
			if err != nil {
				return err
			}

			_, err = db.Exec("CREATE INDEX idx_visits_url_id ON visits(url_id)")
			if err != nil {
				return err
			}
		}
	}

	return nil
}
