# Quickstart Guide: Podcast RSS Webapp

**Project**: RSS Server - Podcast RSS Feed Generator  
**Version**: 1.0.0  
**Last Updated**: 2025-12-31

## Overview

This guide will help you get the podcast RSS webapp running locally and publish your first episode in under 5 minutes.

## Prerequisites

- **Go 1.21+** installed ([download](https://go.dev/dl/))
- Basic command line familiarity
- An MP3 audio file to upload (for testing)

## Quick Start (5 Minutes)

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/example/rss-server.git
cd rss-server

# Install dependencies
go mod download

# Create data directories
mkdir -p data/audio data/artwork
```

### 2. Configure the Server

Edit `config.yaml` (or use defaults):

```yaml
server:
  port: 8080
  host: "localhost"
  baseURL: "http://localhost:8080"

storage:
  audioDir: "./data/audio"
  artworkDir: "./data/artwork"
  rssFile: "./data/podcast.xml"

limits:
  maxUploadSize: 524288000  # 500MB in bytes
  maxEpisodes: 1000
```

### 3. Start the Server

```bash
# Run the server
go run cmd/server/main.go

# You should see:
# RSS Server starting on http://localhost:8080
# RSS feed available at http://localhost:8080/feed.xml
```

### 4. Access the Dashboard

Open your browser to [http://localhost:8080](http://localhost:8080)

You'll see the podcast management dashboard with:
- Episode upload form
- Episode list (empty initially)
- Podcast settings link

### 5. Upload Your First Episode

**Via Web UI**:

1. Click "Upload Episode" button
2. Fill in the form:
   - **Audio File**: Select your MP3 file
   - **Title**: "My First Episode"
   - **Description**: "This is my first podcast episode!"
   - (Optional) Episode number, season, etc.
3. Click "Upload"
4. Episode appears in the list immediately

**Via API** (curl):

```bash
curl -X POST http://localhost:8080/api/episodes \
  -F "audio=@/path/to/episode.mp3" \
  -F "title=My First Episode" \
  -F "description=This is my first podcast episode!"
```

### 6. View Your RSS Feed

Open [http://localhost:8080/feed.xml](http://localhost:8080/feed.xml) in your browser.

You'll see valid RSS 2.0 XML with your episode:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
  <channel>
    <title>My Podcast</title>
    <link>http://localhost:8080</link>
    <description>A podcast about interesting topics</description>
    <item>
      <title>My First Episode</title>
      <description>This is my first podcast episode!</description>
      <enclosure url="http://localhost:8080/audio/episode.mp3" 
                 length="12345678" 
                 type="audio/mpeg" />
    </item>
  </channel>
</rss>
```

### 7. Validate Your Feed

Copy your feed URL and validate it:

**Option 1: W3C Feed Validator**
- Visit https://validator.w3.org/feed/
- Paste `http://localhost:8080/feed.xml`
- Click "Check"
- Should show "Valid RSS feed"

**Option 2: Cast Feed Validator** (for iTunes)
- Visit https://podba.se/validate/
- Paste your feed URL
- Click "Validate"
- Check for iTunes compatibility

### 8. Customize Podcast Metadata

Click "Podcast Settings" in the dashboard to configure:

- **Title**: Your podcast name
- **Author**: Your name
- **Description**: About your podcast
- **Artwork**: Upload square image (1400x1400 to 3000x3000 px)
- **Category**: Select iTunes category
- **Language**: Set language code (e.g., "en-us")

Changes appear immediately in the RSS feed.

## Common Tasks

### Add More Episodes

**Via Web UI**:
- Click "Upload Episode"
- Fill form and submit
- New episode appears at top of list

**Via API**:
```bash
curl -X POST http://localhost:8080/api/episodes \
  -F "audio=@episode2.mp3" \
  -F "title=Episode 2" \
  -F "description=My second episode" \
  -F "episodeNumber=2" \
  -F "seasonNumber=1"
```

### Delete an Episode

**Via Web UI**:
- Click "Delete" button next to episode
- Confirm deletion
- Episode removed from feed and filesystem

**Via API**:
```bash
curl -X DELETE http://localhost:8080/api/episodes/ep-20251231-episode-2
```

### List All Episodes

**Via API**:
```bash
curl http://localhost:8080/api/episodes
```

Returns JSON array of episodes.

### Download Audio File

Direct link to audio file:
```
http://localhost:8080/audio/{filename}
```

Example:
```
http://localhost:8080/audio/ep-20251231-my-first-episode.mp3
```

## Production Deployment

### 1. Set Base URL

Update `config.yaml` with your public domain:

```yaml
server:
  baseURL: "https://podcast.example.com"
```

This ensures RSS feed URLs are publicly accessible.

### 2. Use Reverse Proxy

Run behind nginx or Caddy for HTTPS:

**Nginx example**:
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

### 3. Run as Systemd Service

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

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable rss-server
sudo systemctl start rss-server
```

### 4. Submit to Podcast Directories

Once your feed is publicly accessible:

**Apple Podcasts**:
1. Go to [Podcasts Connect](https://podcastsconnect.apple.com/)
2. Click "Add a Show"
3. Enter your RSS feed URL
4. Validate and submit

**Spotify**:
1. Go to [Spotify for Podcasters](https://podcasters.spotify.com/)
2. Click "Get Started"
3. Enter RSS feed URL
4. Claim your podcast

**Google Podcasts**:
1. Ensure feed is in [Google Search Console](https://search.google.com/search-console)
2. Google will automatically discover the feed
3. Or manually submit via Podcast Manager

## Troubleshooting

### RSS Feed Validation Errors

**Problem**: Feed doesn't validate

**Solutions**:
- Check all required fields are filled (title, author, description)
- Ensure artwork is square and correct dimensions
- Validate iTunes category spelling
- Check for special characters in XML (should auto-escape)

### Audio Files Not Playing

**Problem**: Enclosure URL returns 404

**Solutions**:
- Verify audio file exists in `data/audio/` directory
- Check file permissions (readable by server)
- Ensure `baseURL` in config matches your domain
- Check Content-Type header is `audio/mpeg`

### Upload Fails

**Problem**: Episode upload returns error

**Solutions**:
- Check file size (must be < 500MB by default)
- Verify file is MP3 format
- Check disk space in `data/audio/` directory
- Review server logs for specific error

### Feed Not Updating

**Problem**: Changes don't appear in RSS feed

**Solutions**:
- Check `data/podcast.xml` file was modified
- Verify file permissions (writable by server)
- Clear browser cache (Ctrl+F5)
- Check for file lock errors in logs

## API Reference

See [contracts/openapi.yaml](./contracts/openapi.yaml) for complete API documentation.

**Quick Reference**:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/feed.xml` | GET | Serve RSS feed |
| `/api/episodes` | POST | Upload episode |
| `/api/episodes` | GET | List episodes |
| `/api/episodes/{id}` | DELETE | Delete episode |
| `/api/podcast/settings` | GET | Get podcast settings |
| `/api/podcast/settings` | POST | Update settings |
| `/audio/{filename}` | GET | Stream audio file |
| `/` | GET | Web dashboard |

## Configuration Reference

### Server Settings

```yaml
server:
  port: 8080              # HTTP port
  host: "0.0.0.0"         # Listen address (0.0.0.0 for all interfaces)
  baseURL: "http://localhost:8080"  # Public base URL
  readTimeout: 30s        # Request read timeout
  writeTimeout: 300s      # Response write timeout (long for uploads)
```

### Storage Settings

```yaml
storage:
  audioDir: "./data/audio"           # Audio file storage
  artworkDir: "./data/artwork"       # Artwork storage
  rssFile: "./data/podcast.xml"      # RSS feed source of truth
```

### Limits

```yaml
limits:
  maxUploadSize: 524288000  # 500MB in bytes
  maxEpisodes: 1000         # Maximum episodes in feed
```

## File Structure

After setup, your directory will look like:

```
rss-server/
├── cmd/
│   └── server/
│       └── main.go           # Server entry point
├── internal/
│   ├── handlers/             # HTTP handlers
│   ├── rss/                  # RSS generation
│   ├── storage/              # File operations
│   └── models/               # Data structures
├── web/
│   ├── templates/            # HTMX templates
│   └── static/               # CSS, images
├── data/
│   ├── audio/                # Uploaded MP3 files
│   ├── artwork/              # Podcast artwork
│   └── podcast.xml           # RSS feed (source of truth)
├── config.yaml               # Configuration
├── go.mod                    # Go dependencies
└── README.md                 # Project documentation
```

## Next Steps

1. **Customize branding**: Upload artwork and configure podcast metadata
2. **Add more episodes**: Build your episode library
3. **Test on podcast apps**: Validate playback in Apple Podcasts, Spotify, etc.
4. **Deploy to production**: Follow production deployment steps
5. **Submit to directories**: Get discovered by listeners

## Support

- **Issues**: https://github.com/example/rss-server/issues
- **Documentation**: https://github.com/example/rss-server/wiki
- **RSS Spec**: https://www.rssboard.org/rss-specification
- **iTunes Podcast Spec**: https://help.apple.com/itc/podcasts_connect/

## Constitution Alignment

This implementation follows the RSS Server Constitution v1.0.0:

- ✅ **RSS Standard Compliance**: Valid RSS 2.0 + iTunes feeds
- ✅ **File-Centric Architecture**: XML file is source of truth
- ✅ **HTTP-First Design**: All operations via HTTP endpoints
- ✅ **Testing Discipline**: Integration tests with RSS validators
- ✅ **Simplicity & Maintainability**: No database, minimal dependencies
