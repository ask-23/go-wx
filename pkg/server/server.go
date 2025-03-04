package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ask-23/go-wx/internal/models"
	"github.com/ask-23/go-wx/pkg/config"
	"github.com/ask-23/go-wx/pkg/database"
)

// Server represents a web server for weather data
type Server struct {
	config  *config.ServerConfig
	db      *database.Database
	server  *http.Server
	station *config.StationConfig
}

// NewServer creates a new web server
func NewServer(cfg config.ServerConfig, db *database.Database) (*Server, error) {
	return &Server{
		config:  &cfg,
		db:      db,
		station: &cfg.Station,
	}, nil
}

// Start begins serving the web interface
func (s *Server) Start() error {
	// Create a new HTTP server
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/api/current", s.handleCurrentData)
	mux.HandleFunc("/api/history", s.handleHistoryData)

	// Serve static files
	staticDir := "/static/"
	mux.Handle(staticDir, http.StripPrefix(staticDir, http.FileServer(http.Dir("web/static"))))

	// Configure the server
	addr := fmt.Sprintf("%s:%d", s.config.Address, s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start the server
	log.Printf("Starting web server on %s", addr)
	var err error
	if s.config.SSL.Enabled {
		err = s.server.ListenAndServeTLS(s.config.SSL.CertFile, s.config.SSL.KeyFile)
	} else {
		err = s.server.ListenAndServe()
	}

	// If we get here, there was an error
	if err != http.ErrServerClosed {
		return fmt.Errorf("web server error: %w", err)
	}

	return nil
}

// Stop stops the web server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// handleHome serves the main dashboard page
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	// Ensure we're only handling the root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get the latest weather data
	data, err := s.db.GetLatestWeatherData()
	if err != nil {
		http.Error(w, "Error retrieving weather data", http.StatusInternalServerError)
		log.Printf("Error retrieving weather data: %v", err)
		return
	}

	// Get historical data for the graphs
	end := time.Now()
	start := end.Add(-24 * time.Hour)
	history, err := s.db.GetWeatherDataRange(start, end)
	if err != nil {
		http.Error(w, "Error retrieving historical data", http.StatusInternalServerError)
		log.Printf("Error retrieving historical data: %v", err)
		return
	}

	// Prepare template data
	templateData := struct {
		Current *models.WeatherData
		History []*models.WeatherData
		Station *config.StationConfig
	}{
		Current: data,
		History: history,
		Station: s.station,
	}

	// Parse and execute the template
	tmpl, err := template.ParseFiles(filepath.Join("web/templates", "dashboard.html"))
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Printf("Error loading template: %v", err)
		return
	}

	if err := tmpl.Execute(w, templateData); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
		return
	}
}

// handleCurrentData returns the current weather data as JSON
func (s *Server) handleCurrentData(w http.ResponseWriter, r *http.Request) {
	// Get the latest weather data
	data, err := s.db.GetLatestWeatherData()
	if err != nil {
		http.Error(w, "Error retrieving weather data", http.StatusInternalServerError)
		log.Printf("Error retrieving weather data: %v", err)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v", err)
		return
	}
}

// handleHistoryData returns historical weather data as JSON
func (s *Server) handleHistoryData(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	r.ParseForm()

	// Get the time range from query parameters or use defaults
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour) // Default to last 24 hours

	if startStr := r.Form.Get("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}

	if endStr := r.Form.Get("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	// Get historical data for the specified range
	history, err := s.db.GetWeatherDataRange(startTime, endTime)
	if err != nil {
		http.Error(w, "Error retrieving historical data", http.StatusInternalServerError)
		log.Printf("Error retrieving historical data: %v", err)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write JSON response
	if err := json.NewEncoder(w).Encode(history); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v", err)
		return
	}
}
