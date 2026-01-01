package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/rss-server/internal/handlers"
	"github.com/example/rss-server/internal/storage"
)

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(lw, r)

		// Log the request
		duration := time.Since(start)
		log.Printf("%s %s %d %v %s", r.Method, r.URL.Path, lw.statusCode, duration, r.RemoteAddr)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func main() {
	// Configuration
	audioDir := "./data/audio"
	artworkDir := "./data/artwork"
	rssFile := "./data/podcast.xml"
	templatesDir := "./web/templates"
	maxUploadMB := int64(500)

	// Load RSS store
	store, err := storage.LoadRSSStore(rssFile)
	if err != nil {
		log.Fatalf("Failed to load RSS store: %v", err)
	}

	log.Println("RSS Server starting...")
	log.Println("Loaded podcast feed successfully")

	// Load templates
	tmpl, err := template.ParseGlob(templatesDir + "/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	tmpl, err = tmpl.ParseGlob(templatesDir + "/components/*.html")
	if err != nil {
		log.Fatalf("Failed to parse component templates: %v", err)
	}

	// Create handlers
	episodesHandler := handlers.NewEpisodesHandler(store, audioDir, artworkDir, maxUploadMB, tmpl)
	feedHandler := handlers.NewFeedHandler(store)
	staticHandler := handlers.NewStaticHandler(audioDir)
	webHandler, err := handlers.NewWebHandler(store, templatesDir)
	if err != nil {
		log.Fatalf("Failed to create web handler: %v", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Serve static files (CSS, images)
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve artwork files
	artworkFS := http.FileServer(http.Dir("./data/artwork"))
	mux.Handle("/static/artwork/", http.StripPrefix("/static/artwork/", artworkFS))

	// Web UI routes
	mux.HandleFunc("/", webHandler.HandleDashboard)

	// API routes
	mux.HandleFunc("/api/episodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			episodesHandler.HandleUpload(w, r)
		} else if r.Method == http.MethodGet {
			episodesHandler.HandleList(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// DELETE /api/episodes/{episodeId}
	mux.HandleFunc("/api/episodes/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			episodesHandler.HandleDelete(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/podcast/settings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			episodesHandler.HandleGetSettings(w, r)
		} else if r.Method == http.MethodPost {
			episodesHandler.HandleUpdateSettings(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// RSS feed route
	mux.HandleFunc("/feed.xml", feedHandler.HandleFeed)

	// Audio file serving route
	mux.HandleFunc("/audio/", staticHandler.HandleAudio)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Printf("RSS Feed: http://localhost%s/feed.xml", addr)

	// Wrap mux with logging middleware
	loggedMux := loggingMiddleware(mux)

	if err := http.ListenAndServe(addr, loggedMux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
