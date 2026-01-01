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

### config.yaml

The server is configured using `config.yaml` in the repository root. 

**Important**: The `base_url` field is **required** for podcast feeds to work correctly. Without a valid base URL, audio files and artwork will not be accessible to podcast clients.

```yaml
# Base URL for the server (REQUIRED)
# This must be the public URL where your podcast is hosted
# Examples: http://podcast.example.com, https://mypodcast.com
base_url: "http://localhost:8080"

server:
  port: 8080
  host: "0.0.0.0"

upload:
  max_file_size_mb: 500
  allowed_extensions:
    - ".mp3"

paths:
  data_dir: "./data"
  audio_dir: "./data/audio"
  artwork_dir: "./data/artwork"
  rss_file: "./data/podcast.xml"

podcast:
  default_title: "My Podcast"
  default_author: "Podcast Creator"
  default_description: "A podcast about interesting topics"
  default_language: "en-us"
  default_explicit: "no"
  default_category: "Technology"
```

### Configuration Fields

#### base_url (REQUIRED)
The public URL where your podcast server is hosted. This is used to generate absolute URLs in the RSS feed for audio files and artwork.

- **Format**: Must include protocol (`http://` or `https://`) and hostname
- **Examples**: 
  - Development: `http://localhost:8080`
  - Production: `https://podcast.example.com`
  - With port: `http://example.com:3000`
- **Validation**: Server will not start without a valid base_url

#### server
- `port`: HTTP server port (default: 8080)
- `host`: Listen address (default: "0.0.0.0" for all interfaces)

#### upload
- `max_file_size_mb`: Maximum allowed audio file size in MB (default: 500)
- `allowed_extensions`: List of allowed file extensions (default: [".mp3"])

#### paths
- `data_dir`: Base directory for data files
- `audio_dir`: Directory for episode audio files
- `artwork_dir`: Directory for podcast and episode artwork
- `rss_file`: Path to the RSS feed XML file

#### podcast
Default metadata used when creating a new podcast:
- `default_title`: Podcast title
- `default_author`: Author/host name
- `default_description`: Podcast description
- `default_language`: Language code (e.g., "en-us")
- `default_explicit`: Explicit content flag ("yes", "no", or "clean")
- `default_category`: iTunes category

### Environment Variables

The server can be configured via environment variables:

- `PORT`: HTTP server port (default: 8080) - **Note**: This overrides the `server.port` value in config.yaml

### Configuration Files
- `config.yaml`: Server configuration (port, limits, directories, **base URL**)
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

### Pre-Deployment Checklist

Before deploying to production, verify the following:

- [ ] **Configure base_url**: Set `base_url` in `config.yaml` to your public domain
  ```yaml
  base_url: "https://podcast.yourdomain.com"
  ```
- [ ] **Test base_url**: Verify server starts without errors:
  ```bash
  go run cmd/server/main.go
  ```
- [ ] **Validate RSS feed**: Check that feed contains absolute URLs:
  ```bash
  curl https://podcast.yourdomain.com/feed.xml | grep enclosure
  ```
- [ ] **Test audio playback**: Verify audio files are accessible:
  ```bash
  curl -I https://podcast.yourdomain.com/audio/test-episode.mp3
  ```
- [ ] **SSL/TLS Certificate**: If using HTTPS, ensure valid SSL certificate is installed
- [ ] **Firewall Rules**: Open port 8080 (or configured port) in firewall
- [ ] **Reverse Proxy**: Configure Nginx/Apache if using reverse proxy
- [ ] **File Permissions**: Ensure `data/` directory is writable by server process
- [ ] **Backup Strategy**: Set up automated backups of `data/podcast.xml` and `data/audio/`
- [ ] **Monitoring**: Configure health checks and uptime monitoring
- [ ] **RSS Validation**: Test feed with [Cast Feed Validator](https://podba.se/validate/)

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

### Server Won't Start

**Error: "Configuration error: base_url is required in configuration"**
- Add the `base_url` field to your `config.yaml` file
- Example: `base_url: "http://localhost:8080"`
- The base_url must include the protocol (http:// or https://)

**Error: "Configuration validation failed: base_url must use http or https scheme"**
- Ensure your base_url starts with `http://` or `https://`
- Invalid examples: `ftp://example.com`, `example.com`, `www.example.com`
- Valid examples: `http://example.com`, `https://podcast.example.com:8080`

**Error: "Configuration validation failed: base_url must include a host"**
- The base_url must include a hostname or IP address
- Invalid: `http://` or `https://`
- Valid: `http://localhost:8080`, `https://example.com`

### Upload Fails
- Check file is MP3 format
- Verify file size is under 500MB (or configured `max_file_size_mb`)
- Ensure sufficient disk space
- Check file permissions on `data/audio/` directory

### Feed Not Updating
- Check `data/podcast.xml` was modified
- Verify file permissions on `data/` directory
- Clear browser cache (Ctrl+F5)
- Validate RSS feed with [W3C Feed Validator](https://validator.w3.org/feed/)

### Audio Files Not Playing in Podcast Clients
- **Most Common Issue**: Incorrect or missing `base_url` in `config.yaml`
- Verify files exist in `data/audio/` directory
- Check file permissions (should be readable)
- Ensure `base_url` matches your server's public URL
- Test by accessing audio file directly: `{base_url}/audio/filename.mp3`
- Validate RSS feed URLs are absolute (not relative): 
  ```bash
  curl http://localhost:8080/feed.xml | grep enclosure
  ```
  Should show: `http://your-domain.com/audio/file.mp3` (absolute)
  NOT: `/audio/file.mp3` (relative)

### RSS Feed Validation Errors

**"Relative URLs in feed"**
- Check that `base_url` is configured in `config.yaml`
- Restart the server after changing `base_url`
- Verify the feed URL by visiting `/feed.xml` in your browser

**"Invalid URL encoding"**
- The server automatically encodes special characters in URLs
- If you have files with spaces or special characters, they will be encoded as `%20`, etc.

### Dashboard Shows Wrong Feed URL
- Verify `base_url` in `config.yaml` matches your deployment URL
- Restart the server after changing configuration
- For Docker deployments, ensure `base_url` reflects the external URL (not `localhost`)

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
