# URL Shortener

A fast, lightweight URL shortener built with Go, SQLite, and HTMX.

## Features

- **Simple Setup**: First-run wizard to create admin account
- **Admin Dashboard**: Manage shortened URLs with search and pagination
- **Fast Redirects**: 301 permanent redirects with embedded SQLite
- **HTMX Frontend**: SPA-like experience without heavy JavaScript
- **Secure**: bcrypt password hashing, session-based authentication
- **Admin Reset**: Environment variable-based password reset

## Tech Stack

- **Backend**: Go with standard library + minimal dependencies
- **Database**: SQLite (modernc.org/sqlite - pure Go, no CGO)
- **Frontend**: HTMX + Tailwind CSS
- **Auth**: gorilla/sessions + bcrypt
- **Reverse Proxy**: Caddy

## Installation

### Prerequisites

- Go 1.21 or higher
- Caddy web server (optional, for production)

### Setup

1. **Clone and build**:
```bash
cd /home/projects/web/Ojrek/url-shortener
go mod download
go build -o url-shortener
```

2. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your settings
nano .env
```

3. **Run the application**:
```bash
./url-shortener
```

4. **First-run setup**:
   - Visit `http://localhost:8080/setup`
   - Create your admin account
   - Login at `http://localhost:8080/login`

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `BASE_URL` | Base URL for short links | - | Yes |
| `PORT` | Server port | `8080` | No |
| `DB_PATH` | SQLite database path | `./data/urls.db` | No |
| `SESSION_SECRET` | Cookie encryption key (32+ chars) | - | Yes |
| `RESET_ADMIN` | Enable admin reset endpoint | `false` | No |

### Generate Session Secret

```bash
openssl rand -base64 32
```

## Production Deployment

### Using systemd

1. **Build the binary**:
```bash
go build -o url-shortener
```

2. **Copy files to production directory**:
```bash
sudo mkdir -p /opt/url-shortener
sudo cp url-shortener /opt/url-shortener/
sudo cp .env /opt/url-shortener/
sudo chown -R www-data:www-data /opt/url-shortener
```

3. **Install systemd service**:
```bash
sudo cp url-shortener.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable url-shortener
sudo systemctl start url-shortener
```

4. **Check status**:
```bash
sudo systemctl status url-shortener
```

### Using Caddy

1. **Install Caddy** (if not already installed):
```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

2. **Configure Caddy**:
```bash
sudo cp Caddyfile /etc/caddy/Caddyfile
# Edit the domain name
sudo nano /etc/caddy/Caddyfile
```

3. **Reload Caddy**:
```bash
sudo systemctl reload caddy
```

## Usage

### Admin Dashboard

- **Login**: `/login`
- **Dashboard**: `/admin/dashboard`
- **Logout**: `/logout`

### Managing URLs

1. **Add New URL**:
   - Click "Add New URL" button
   - Enter the original URL
   - Optionally add notes
   - Short code is auto-generated (4+ lowercase alphanumeric characters)

2. **Edit URL**:
   - Click "Edit" on any URL row
   - Only notes can be edited
   - Short code and original URL are immutable

3. **Delete URL**:
   - Click "Delete" on any URL row
   - Confirm the deletion

4. **Search URLs**:
   - Use the search box and click "Search" button to filter URLs
   - Filter by: Anything, Original URL, or Notes
   - Results are paginated (50 per page) and search the entire database

### Short URL Format

Short URLs follow this pattern:
```
https://go.yourdomain.com/{code}
```

Where `{code}` is a unique lowercase alphanumeric string (4+ characters).

### Redirects

When a user visits a short URL, they are redirected with HTTP 301 (permanent redirect) to the original URL.

## Admin Password Reset

If you forget your admin password:

1. **Set environment variable**:
```bash
# In .env file
RESET_ADMIN=true
```

2. **Restart the service**:
```bash
sudo systemctl restart url-shortener
```

3. **Visit reset page**:
```
http://localhost:8080/setup/reset
```

4. **Create new credentials**

5. **Disable reset mode**:
```bash
# In .env file
RESET_ADMIN=false
```

6. **Restart again**:
```bash
sudo systemctl restart url-shortener
```

## Database

The application uses SQLite with the following schema:

### Tables

**admin**:
- `id`: Primary key
- `username`: Admin username (unique)
- `password_hash`: bcrypt hashed password
- `created_at`: Timestamp

**urls**:
- `id`: Primary key
- `code`: Short code (unique, indexed)
- `original_url`: Target URL (indexed)
- `notes`: Optional notes
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

### Backup

To backup your database:
```bash
cp ./data/urls.db ./data/urls.db.backup
```

## Development

### Run in development mode:
```bash
go run main.go
```

### Install dependencies:
```bash
go mod download
```

### Build:
```bash
go build -o url-shortener
```

## License

Apache License 2.0
