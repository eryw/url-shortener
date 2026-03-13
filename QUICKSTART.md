# Quick Start Guide

## Run Locally (Development)

1. **Start the server**:
```bash
./url-shortener
```

2. **Open your browser**:
   - Go to: `http://localhost:8080`

3. **First-time setup**:
   - You'll be redirected to `/setup`
   - Create your admin username and password
   - Click "Create Account"

4. **Login**:
   - Enter your credentials
   - Click "Sign in"

5. **Start creating short URLs**:
   - Click "Add New URL" button
   - Enter the original URL (e.g., `https://example.com/very/long/url`)
   - Optionally add notes
   - Click "Create"
   - Your short URL will be generated automatically!

## Test the Short URL

After creating a short URL with code `abc123`:
```bash
curl -I http://localhost:8080/abc123
```

You should see a `301 Moved Permanently` redirect to your original URL.

## Features to Try

### Search & Filter
- Use the search box to find URLs
- Filter by: "Anything", "Original URL", or "Notes"
- Results update instantly with HTMX

### Edit Notes
- Click "Edit" on any URL
- Modify the notes field
- Short code and original URL are read-only

### Delete URLs
- Click "Delete" on any URL
- Confirm the deletion
- Row disappears instantly

### Pagination
- Create 50+ URLs to see pagination
- Navigate between pages

## Production Deployment

See `README.md` for detailed production deployment instructions with:
- systemd service setup
- Caddy reverse proxy configuration
- Security best practices

## Reset Admin Password

If you forget your password:

1. Edit `.env`:
```bash
RESET_ADMIN=true
```

2. Restart the server:
```bash
./url-shortener
```

3. Visit: `http://localhost:8080/setup/reset`

4. Create new credentials

5. Set `RESET_ADMIN=false` in `.env` and restart

## Environment Variables

Edit `.env` to configure:
- `BASE_URL`: Your short URL domain
- `PORT`: Server port (default: 8080)
- `DB_PATH`: SQLite database location
- `SESSION_SECRET`: Cookie encryption key
- `RESET_ADMIN`: Enable password reset

## Troubleshooting

**Port already in use**:
```bash
# Change PORT in .env
PORT=8081
```

**Database locked**:
```bash
# Stop all instances and restart
pkill url-shortener
./url-shortener
```

**Can't access from other machines**:
- The server binds to all interfaces (0.0.0.0)
- Check firewall settings
- Update BASE_URL in .env to your actual domain/IP
