# RSS Server - Podcast RSS Feed Generator

> **Disclaimer**: This project is 100% vibe coded. It was built entirely by AI as an experiment in rapid prototyping. Use at your own risk, but it actually works!

A simple, file-based podcast RSS feed generator with a web interface. Upload MP3 episodes, customize your podcast metadata, and generate a standards-compliant RSS 2.0 + iTunes feed.

## Features

- **Web-based Dashboard**: Upload and manage episodes via browser
- **RSS 2.0 + iTunes**: Standards-compliant podcast feeds
- **File-centric Architecture**: No database - XML file is source of truth
- **Episode Management**: Upload, list, and delete episodes
- **Podcast Customization**: Configure title, author, artwork, category, and more
- **Audio Streaming**: Built-in HTTP audio file serving
- **HTMX Interface**: Fast, responsive UI without complex JavaScript

## Prerequisites

- **Go 1.21+** ([download](https://go.dev/dl/))
- MP3 audio files for podcast episodes

## Quick Start

### 1. Clone and Install

```bash
git clone https://github.com/example/rss-server.git
cd rss-server
go mod download
```

### 2. Start the Server

```bash
go run cmd/server/main.go
```

The server will start on http://localhost:8080

### 3. Open the Dashboard

Navigate to http://localhost:8080 in your browser to:
- Upload episodes
- Customize podcast metadata
- Manage your podcast feed

### 4. View Your RSS Feed

Access your RSS feed at http://localhost:8080/feed.xml

## Docker Deployment

### Quick Start with Docker

**Using Docker Compose (Recommended)**:

```bash
# Clone the repository
git clone https://github.com/example/rss-server.git
cd rss-server

# Start with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

**Using Docker directly**:

```bash
# Build the image
docker build -t rss-server:latest .

# Run the container
docker run -d \
  --name rss-server \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  rss-server:latest

# View logs
docker logs -f rss-server

# Stop and remove
docker stop rss-server && docker rm rss-server
```

### Docker Image Details

- **Base Image**: Alpine Linux (minimal, secure)
- **Size**: ~20MB (multi-stage build)
- **User**: Non-root user `podcast` for security
- **Health Check**: Automatic health monitoring
- **Volumes**: `/app/data` for persistent storage

### Customization

Mount a custom config file:

```bash
docker run -d \
  --name rss-server \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e PORT=8080 \
  rss-server:latest
```

### Docker Environment Variables

- `PORT`: HTTP server port (default: 8080)

## Usage

### Upload an Episode (Web UI)

1. Open http://localhost:8080
2. Fill in the episode upload form:
   - Select your MP3 audio file
   - Enter title and description
   - Optionally add episode/season numbers
3. Click "Upload"

### Upload an Episode (API)

```bash
curl -X POST http://localhost:8080/api/episodes \
  -F "audio=@episode.mp3" \
  -F "title=My First Episode" \
  -F "description=This is my first podcast episode!"
```

### Customize Podcast Settings

1. Click "Settings" in the dashboard
2. Configure:
   - Podcast title and author
   - Description and website link
   - Artwork (1400x1400 to 3000x3000 px, JPG/PNG)
   - iTunes category
   - Language code (e.g., "en-us")
3. Click "Save Settings"

### Delete an Episode

Click the "Delete" button next to any episode in the dashboard, or use the API:

```bash
curl -X DELETE http://localhost:8080/api/episodes/{episode-id}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Web dashboard |
| `/feed.xml` | GET | RSS feed (XML) |
| `/api/episodes` | GET | List all episodes (JSON) |
| `/api/episodes` | POST | Upload new episode |
| `/api/episodes/{id}` | DELETE | Delete episode |
| `/api/podcast/settings` | GET | Get podcast settings (HTML) |
| `/api/podcast/settings` | POST | Update podcast settings |
| `/audio/{filename}` | GET | Stream audio file |

## Configuration

The server can be configured via environment variables:

- `PORT`: HTTP server port (default: 8080)

Configuration files:
- `config.yaml`: Server configuration (port, limits, directories)
- `data/podcast.xml`: RSS feed source of truth

## Project Structure

```
rss-server/
├── cmd/server/           # Server entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   ├── models/           # Data structures
│   ├── rss/              # RSS feed generation
│   └── storage/          # File operations and persistence
├── web/
│   ├── templates/        # HTML templates
│   └── static/           # CSS and static assets
├── data/
│   ├── audio/            # Episode audio files
│   ├── artwork/          # Podcast artwork
│   └── podcast.xml       # RSS feed (source of truth)
└── config.yaml           # Server configuration
```

## Validation

Validate your RSS feed with:
- **W3C Feed Validator**: https://validator.w3.org/feed/
- **Cast Feed Validator**: https://podba.se/validate/

## Production Deployment

### Docker Production Deployment (Recommended)

**Using Docker Compose**:

```yaml
# docker-compose.yml
version: '3.8'

services:
  rss-server:
    image: rss-server:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - PORT=8080
```

Deploy:
```bash
docker-compose up -d
```

**With Nginx Reverse Proxy**:

```yaml
version: '3.8'

services:
  rss-server:
    image: rss-server:latest
    restart: unless-stopped
    expose:
      - "8080"
    volumes:
      - ./data:/app/data
    networks:
      - podcast-net

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    networks:
      - podcast-net
    depends_on:
      - rss-server

networks:
  podcast-net:
```

### Build for Production

```bash
go build -o rss-server ./cmd/server/
./rss-server
```

### Run with Systemd

Create `/etc/systemd/system/rss-server.service`:

```ini
[Unit]
Description=Podcast RSS Server
After=network.target

[Service]
Type=simple
User=podcast
WorkingDirectory=/opt/rss-server
ExecStart=/opt/rss-server/rss-server
Restart=on-failure
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable rss-server
sudo systemctl start rss-server
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 443 ssl;
    server_name podcast.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Submitting to Podcast Directories

### Apple Podcasts
1. Visit [Podcasts Connect](https://podcastsconnect.apple.com/)
2. Add your RSS feed URL
3. Validate and submit

### Spotify
1. Visit [Spotify for Podcasters](https://podcasters.spotify.com/)
2. Enter your RSS feed URL
3. Claim your podcast

### Google Podcasts
- Add feed to [Google Search Console](https://search.google.com/search-console)
- Google will automatically discover it

## Troubleshooting

### Upload Fails
- Check file is MP3 format
- Verify file size is under 500MB
- Ensure sufficient disk space

### Feed Not Updating
- Check `data/podcast.xml` was modified
- Verify file permissions
- Clear browser cache (Ctrl+F5)

### Audio Files Not Playing
- Verify files exist in `data/audio/`
- Check file permissions
- Ensure correct base URL in configuration

## Development

### Dependencies
- `github.com/eduncan911/podcast` - RSS feed generation library
- Go standard library (net/http, html/template, encoding/xml)

### Testing
```bash
go test ./...
```

## Architecture

**File-Centric Design**: The RSS feed XML file (`data/podcast.xml`) is the single source of truth. All operations read from and write to this file atomically.

**No Database**: Simplicity and reliability through filesystem-based storage.

**Standards Compliance**: Generates valid RSS 2.0 feeds with iTunes podcast extensions.

## License

MIT

## Support

- **Issues**: https://github.com/example/rss-server/issues
- **Documentation**: See `specs/001-podcast-rss-webapp/` for detailed specifications

## Acknowledgments

Built with:
- Go 1.21+
- [github.com/eduncan911/podcast](https://github.com/eduncan911/podcast)
- HTMX for dynamic UI

---

**RSS Server** - Simple, reliable podcast hosting.
