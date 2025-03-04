package publisher

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ask-23/go-wx/pkg/config"
	"github.com/ask-23/go-wx/pkg/database"
)

// WundergroundPublisher publishes weather data to Weather Underground
type WundergroundPublisher struct {
	BasePublisher
	stationID string
	apiKey    string
}

// NewWundergroundPublisher creates a new Weather Underground publisher
func NewWundergroundPublisher(cfg config.PublisherConfig, db *database.Database) (Publisher, error) {
	// Validate required configuration
	if cfg.StationID == "" {
		return nil, fmt.Errorf("Weather Underground publisher requires station_id")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Weather Underground publisher requires api_key")
	}

	return &WundergroundPublisher{
		BasePublisher: BasePublisher{
			config:  cfg,
			db:      db,
			running: false,
		},
		stationID: cfg.StationID,
		apiKey:    cfg.APIKey,
	}, nil
}

// publish sends weather data to Weather Underground
func (w *WundergroundPublisher) publish() error {
	// Get the latest weather data from the database
	data, err := w.db.GetLatestWeatherData()
	if err != nil {
		return fmt.Errorf("failed to get latest weather data: %w", err)
	}

	// Check if data is too old (older than 10 minutes)
	if time.Since(data.Timestamp) > 10*time.Minute {
		return fmt.Errorf("weather data is too old for publishing (timestamp: %v)", data.Timestamp)
	}

	// Build the Weather Underground API URL
	baseURL := "https://weatherstation.wunderground.com/weatherstation/updateweatherstation.php"

	// Create the query parameters
	params := url.Values{}
	params.Set("ID", w.stationID)
	params.Set("PASSWORD", w.apiKey)
	params.Set("dateutc", data.Timestamp.UTC().Format("2006-01-02 15:04:05"))
	params.Set("action", "updateraw")

	// Add weather data
	params.Set("tempf", strconv.FormatFloat(data.Temperature, 'f', 1, 64))
	params.Set("humidity", strconv.FormatFloat(data.Humidity, 'f', 1, 64))
	params.Set("baromin", strconv.FormatFloat(data.Pressure*0.02953, 'f', 3, 64)) // Convert mbar to inHg
	params.Set("windspeedmph", strconv.FormatFloat(data.WindSpeed, 'f', 1, 64))
	params.Set("winddir", strconv.FormatFloat(data.WindDirection, 'f', 0, 64))
	params.Set("rainin", strconv.FormatFloat(data.Rain, 'f', 2, 64))
	params.Set("dewptf", strconv.FormatFloat(data.DewPoint, 'f', 1, 64))

	// Add UV index if available
	if data.UVIndex > 0 {
		params.Set("UV", strconv.FormatFloat(data.UVIndex, 'f', 1, 64))
	}

	// Make the HTTP request
	apiURL := baseURL + "?" + params.Encode()

	// Use a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request returned non-OK status: %s", resp.Status)
	}

	log.Printf("Successfully published to Weather Underground (station ID: %s)", w.stationID)
	return nil
}
