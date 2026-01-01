# Quick Start Guide: Uploading Your First Episode

## Step 1: Start the Server

```bash
cd /home/devin/Projects/rss-server
go run cmd/server/main.go
```

You should see:
```
RSS Server starting...
Loaded podcast feed successfully
Server listening on http://localhost:8080
RSS Feed: http://localhost:8080/feed.xml
```

## Step 2: Access the Web Dashboard

Open your web browser and navigate to:
**http://localhost:8080**

You'll see the Podcast RSS Server dashboard with:
- 🎙️ Header with navigation
- Upload form section
- Episodes list (empty initially)

## Step 3: Upload an Episode (Web UI Method)

### Using the Web Form:

1. **Audio File**: Click "Choose File" and select an MP3 file (max 500MB)
2. **Episode Title**: Enter a title (e.g., "My First Episode")
3. **Episode Description**: Add a description
4. **Optional Fields**:
   - Publication Date: Leave empty to use current time
   - Episode Number: e.g., 1
   - Season Number: e.g., 1
   - Explicit Content: Select "No", "Yes", or "Clean"
   - Episode Type: Select "Full", "Trailer", or "Bonus"

5. Click **"Upload Episode"** button

The page will show "Uploading..." indicator, then the new episode will appear in the list below!

## Step 4: Upload an Episode (API Method)

### Using curl:

```bash
curl -X POST http://localhost:8080/api/episodes \
  -F "audio=@/path/to/your/episode.mp3" \
  -F "title=My First Episode" \
  -F "description=This is my first podcast episode"
```

### With optional fields:

```bash
curl -X POST http://localhost:8080/api/episodes \
  -F "audio=@/path/to/episode.mp3" \
  -F "title=Episode 1: Getting Started" \
  -F "description=In this episode, we discuss getting started with podcasting" \
  -F "episodeNumber=1" \
  -F "seasonNumber=1" \
  -F "explicit=no" \
  -F "episodeType=full"
```

## Step 5: View Your RSS Feed

After uploading, visit:
**http://localhost:8080/feed.xml**

You'll see a valid RSS 2.0 feed with your episode!

## Step 6: Listen to Your Episode

The audio file is available at:
**http://localhost:8080/audio/[filename]**

Or click the "Play" button next to the episode in the dashboard.

## Step 7: Validate Your Feed

Copy your feed URL and paste it into:
- **W3C Feed Validator**: https://validator.w3.org/feed/
- **Cast Feed Validator**: https://castfeedvalidator.com/

Your feed should pass validation! ✅

## Example: Quick Test with a Sample File

If you don't have an MP3 file handy, you can create a tiny test MP3:

```bash
# Create a test MP3 file (requires ffmpeg)
ffmpeg -f lavfi -i "sine=frequency=440:duration=5" -ac 2 /tmp/test-episode.mp3

# Upload it
curl -X POST http://localhost:8080/api/episodes \
  -F "audio=@/tmp/test-episode.mp3" \
  -F "title=Test Episode" \
  -F "description=This is a test episode with a 5-second tone"
```

## Troubleshooting

### "Only MP3 files are supported"
- Make sure your file ends with `.mp3` extension
- The server only accepts MP3 format

### "File too large (max 500 MB)"
- Your audio file exceeds the 500MB limit
- Try compressing your MP3 or split into multiple episodes

### "Title and description required"
- Both fields are mandatory
- Make sure you fill them in before uploading

### Server not starting / Port in use
- Another process is using port 8080
- Use a different port: `PORT=8081 go run cmd/server/main.go`
- Then access at http://localhost:8081

## What Happens Behind the Scenes

When you upload an episode:

1. ✅ Server validates the file (MP3, size < 500MB)
2. ✅ Generates a unique filename with timestamp
3. ✅ Saves audio file to `data/audio/`
4. ✅ Generates episode ID from title + date
5. ✅ Updates `data/podcast.xml` atomically
6. ✅ Episode appears in RSS feed immediately
7. ✅ Episode shows in dashboard with HTMX

If upload fails, the server automatically cleans up partial files!

## Next Steps

- **Customize Podcast**: Settings page coming in Phase 4 (podcast title, author, artwork)
- **Delete Episodes**: Delete button coming in Phase 5
- **RSS Feed**: Your feed is ready to submit to Apple Podcasts, Spotify, Google Podcasts!

---

**Need help?** The server logs show detailed info about each operation.
