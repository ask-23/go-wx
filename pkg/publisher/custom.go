package publisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ask-23/go-wx/internal/models"
	"github.com/ask-23/go-wx/pkg/config"
	"github.com/ask-23/go-wx/pkg/database"
)

// CustomPublisher publishes weather data to a custom endpoint
type CustomPublisher struct {
	BasePublisher
	url     string
	method  string
	headers map[string]string
}

// NewCustomPublisher creates a new custom publisher
func NewCustomPublisher(cfg config.PublisherConfig, db *database.Database) (Publisher, error) {
	// Validate required configuration
	if cfg.URL == "" {
		return nil, fmt.Errorf("custom publisher requires URL")
	}

	// Default to POST method if not specified
	method := cfg.Method
	if method == "" {
		method = "POST"
	}

	return &CustomPublisher{
		BasePublisher: BasePublisher{
			config:  cfg,
			db:      db,
			running: false,
		},
		url:     cfg.URL,
		method:  method,
		headers: cfg.Headers,
	}, nil
}

// CustomWeatherData is a struct for formatting weather data for the custom API
type CustomWeatherData struct {
	Timestamp     string  `json:"timestamp"`
	Temperature   float64 `json:"temperature"`
	Humidity      float64 `json:"humidity"`
	Pressure      float64 `json:"pressure"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection float64 `json:"wind_direction"`
	Rain          float64 `json:"rain"`
	UVIndex       float64 `json:"uv_index"`
	DewPoint      float64 `json:"dew_point"`
	WindChill     float64 `json:"wind_chill"`
	HeatIndex     float64 `json:"heat_index"`
}

// publish sends weather data to a custom endpoint
func (c *CustomPublisher) publish() error {
	// Get the latest weather data from the database
	data, err := c.db.GetLatestWeatherData()
	if err != nil {
		return fmt.Errorf("failed to get latest weather data: %w", err)
	}

	// Check if data is too old (older than 10 minutes)
	if time.Since(data.Timestamp) > 10*time.Minute {
		return fmt.Errorf("weather data is too old for publishing (timestamp: %v)", data.Timestamp)
	}

	// Format the data for the custom API
	formattedData := formatWeatherData(data)

	// Convert to JSON
	jsonData, err := json.Marshal(formattedData)
	if err != nil {
		return fmt.Errorf("failed to marshal weather data: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest(c.method, c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Use a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("HTTP request returned non-success status: %s, body: %s", resp.Status, body)
	}

	log.Printf("Successfully published to custom endpoint: %s", c.url)
	return nil
}

// formatWeatherData converts the internal weather data model to a format suitable for the custom API
func formatWeatherData(data *models.WeatherData) CustomWeatherData {
	return CustomWeatherData{
		Timestamp:     data.Timestamp.UTC().Format(time.RFC3339),
		Temperature:   data.Temperature,
		Humidity:      data.Humidity,
		Pressure:      data.Pressure,
		WindSpeed:     data.WindSpeed,
		WindDirection: data.WindDirection,
		Rain:          data.Rain,
		UVIndex:       data.UVIndex,
		DewPoint:      data.DewPoint,
		WindChill:     data.WindChill,
		HeatIndex:     data.HeatIndex,
	}
}
