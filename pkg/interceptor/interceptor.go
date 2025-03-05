package interceptor

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ask-23/go-wx/internal/models"
	"github.com/ask-23/go-wx/pkg/config"
	"github.com/ask-23/go-wx/pkg/database"
)

// Interceptor represents a service that listens for and processes weather data
type Interceptor struct {
	config     *config.CollectorConfig
	db         *database.Database
	server     *http.Server
	latestData *models.WeatherData
	mutex      sync.RWMutex
	running    bool
}

// NewInterceptor creates a new data interceptor
func NewInterceptor(cfg config.CollectorConfig, db *database.Database) (*Interceptor, error) {
	return &Interceptor{
		config:     &cfg,
		db:         db,
		latestData: &models.WeatherData{},
		mutex:      sync.RWMutex{},
		running:    false,
	}, nil
}

// Start begins listening for incoming weather data
func (i *Interceptor) Start() error {
	if i.running {
		return fmt.Errorf("interceptor is already running")
	}

	// Set up the HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", i.handleWeatherData)

	addr := fmt.Sprintf(":%d", i.config.Device.Port)
	i.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	i.running = true

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting interceptor on port %d", i.config.Device.Port)
		if err := i.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the interceptor
func (i *Interceptor) Stop() error {
	if !i.running {
		return nil
	}

	i.running = false

	// Shutdown the HTTP server
	return i.server.Close()
}

// GetLatestData returns the most recent weather data
func (i *Interceptor) GetLatestData() *models.WeatherData {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// Return a copy to prevent race conditions
	data := *i.latestData
	return &data
}

// handleWeatherData processes incoming weather data from the Ecowitt device
func (i *Interceptor) handleWeatherData(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request (most Ecowitt devices use POST)
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Debug: log all form values
	log.Printf("Received weather data: %v", r.Form)

	// Process the data based on device type
	switch strings.ToLower(i.config.Device.Type) {
	case "ecowitt":
		i.processEcowittData(r.Form)
	default:
		log.Printf("Unsupported device type: %s", i.config.Device.Type)
	}

	// Send a response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processEcowittData extracts weather data from Ecowitt format
func (i *Interceptor) processEcowittData(form map[string][]string) {
	// Create a new weather data point
	data := &models.WeatherData{
		Timestamp: time.Now(),
	}

	// Extract values from the form data
	extractFloatParam(form, "tempf", &data.Temperature)
	extractFloatParam(form, "humidity", &data.Humidity)
	extractFloatParam(form, "baromabsin", &data.Pressure)
	extractFloatParam(form, "windspeedmph", &data.WindSpeed)
	extractFloatParam(form, "winddir", &data.WindDirection)
	extractFloatParam(form, "rainratein", &data.Rain)
	extractFloatParam(form, "uv", &data.UVIndex)
	extractFloatParam(form, "cloudbase", &data.CloudBase)

	// Calculate derived values (dew point, wind chill, heat index)
	data.CalculateDerivedValues()

	// Update the latest data
	i.mutex.Lock()
	i.latestData = data
	i.mutex.Unlock()

	// Save to database
	if err := i.db.SaveWeatherData(data); err != nil {
		log.Printf("Error saving weather data: %v", err)
	}
}

// extractFloatParam safely extracts a float parameter from form data
func extractFloatParam(form map[string][]string, paramName string, target *float64) {
	if val, ok := form[paramName]; ok && len(val) > 0 {
		if parsedVal, err := strconv.ParseFloat(val[0], 64); err == nil {
			*target = parsedVal
		}
	}
}
