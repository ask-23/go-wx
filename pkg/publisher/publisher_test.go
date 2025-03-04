package publisher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ask-23/go-wx/internal/models"
	"github.com/ask-23/go-wx/pkg/config"
)

// TestCustomPublisher tests the custom publisher implementation
func TestCustomPublisher(t *testing.T) {
	// Create a test server to receive the published data
	var receivedData map[string]interface{}
	var receivedHeaders http.Header

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Save the received headers for later verification
		receivedHeaders = r.Header

		// Parse the JSON body
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&receivedData)
		if err != nil {
			t.Errorf("Failed to decode request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer testServer.Close()

	// Create a custom publisher config
	cfg := config.CustomPublisherConfig{
		URL:    testServer.URL,
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-API-Key":    "test-api-key",
		},
		Format: "json",
	}

	// Create a custom publisher
	pub, err := NewCustomPublisher(cfg)
	if err != nil {
		t.Fatalf("Failed to create custom publisher: %v", err)
	}

	// Create a test weather data
	now := time.Now()
	data := &models.WeatherData{
		Timestamp:     now,
		Temperature:   21.5,
		Humidity:      45.0,
		Pressure:      1012.5,
		WindSpeed:     5.5,
		WindDirection: 180.0,
		Rain:          0.0,
		UVIndex:       5.0,
		CloudBase:     1500.0,
		DewPoint:      9.5,
		WindChill:     21.5,
		HeatIndex:     21.5,
	}

	// Publish the data
	err = pub.Publish(data)
	if err != nil {
		t.Fatalf("Failed to publish data: %v", err)
	}

	// Verify the published data
	if receivedData == nil {
		t.Fatalf("No data received by the test server")
	}

	// Check the required fields
	if receivedData["temperature"] != 21.5 {
		t.Errorf("Expected temperature 21.5, got %v", receivedData["temperature"])
	}

	if receivedData["humidity"] != 45.0 {
		t.Errorf("Expected humidity 45.0, got %v", receivedData["humidity"])
	}

	if receivedData["pressure"] != 1012.5 {
		t.Errorf("Expected pressure 1012.5, got %v", receivedData["pressure"])
	}

	// Check that timestamp was formatted correctly (should be a string in RFC3339 format)
	if timestamp, ok := receivedData["timestamp"].(string); !ok {
		t.Errorf("Expected timestamp to be a string, got %T", receivedData["timestamp"])
	} else {
		// Verify the timestamp format (should be RFC3339)
		_, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Errorf("Invalid timestamp format: %v", err)
		}
	}

	// Verify the request headers
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header 'application/json', got '%s'", receivedHeaders.Get("Content-Type"))
	}

	if receivedHeaders.Get("X-API-Key") != "test-api-key" {
		t.Errorf("Expected X-API-Key header 'test-api-key', got '%s'", receivedHeaders.Get("X-API-Key"))
	}
}

// TestWundergroundPublisher tests the Weather Underground publisher implementation
func TestWundergroundPublisher(t *testing.T) {
	// Track the request URL and parameters
	var requestURL string
	var requestQuery string

	// Create a test server to receive the published data
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Save the request URL and query parameters
		requestURL = r.URL.Path
		requestQuery = r.URL.RawQuery

		// Send a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer testServer.Close()

	// Extract just the host:port from the test server URL (remove http://)
	serverURL := strings.TrimPrefix(testServer.URL, "http://")

	// Create a Weather Underground publisher config
	cfg := config.WundergroundConfig{
		StationID: "KTEST123",
		Password:  "testpassword",
		BaseURL:   fmt.Sprintf("http://%s", serverURL), // Use the test server
	}

	// Create a Weather Underground publisher
	pub, err := NewWundergroundPublisher(cfg)
	if err != nil {
		t.Fatalf("Failed to create Weather Underground publisher: %v", err)
	}

	// Create a test weather data (using Imperial units to match WU expectations)
	data := &models.WeatherData{
		Timestamp:     time.Now(),
		Temperature:   21.5, // Celsius
		Humidity:      45.0,
		Pressure:      1012.5, // hPa
		WindSpeed:     5.5,    // m/s
		WindDirection: 180.0,
		Rain:          2.5, // mm
		UVIndex:       5.0,
	}

	// Publish the data
	err = pub.Publish(data)
	if err != nil {
		t.Fatalf("Failed to publish data: %v", err)
	}

	// Verify the request URL path (should be /weatherstation/updateweatherstation.php)
	expectedPath := "/weatherstation/updateweatherstation.php"
	if requestURL != expectedPath {
		t.Errorf("Expected request path '%s', got '%s'", expectedPath, requestURL)
	}

	// Verify the query parameters
	if !strings.Contains(requestQuery, "ID=KTEST123") {
		t.Errorf("Query string missing station ID parameter, got: %s", requestQuery)
	}

	if !strings.Contains(requestQuery, "PASSWORD=testpassword") {
		t.Errorf("Query string missing password parameter, got: %s", requestQuery)
	}

	// Check for temperature in F (21.5°C = 70.7°F)
	if !strings.Contains(requestQuery, "tempf=70.7") {
		t.Errorf("Query string missing or incorrect tempf parameter, got: %s", requestQuery)
	}

	// Check for pressure in inches (1012.5 hPa = 29.89 inHg)
	expectedInHg := 1012.5 / 33.86389
	if !strings.Contains(requestQuery, fmt.Sprintf("baromin=%.2f", expectedInHg)) {
		t.Errorf("Query string missing or incorrect baromin parameter, got: %s", requestQuery)
	}

	// Check for wind speed in mph (5.5 m/s = 12.3 mph)
	expectedMph := 5.5 / 0.44704
	if !strings.Contains(requestQuery, fmt.Sprintf("windspeedmph=%.1f", expectedMph)) {
		t.Errorf("Query string missing or incorrect windspeedmph parameter, got: %s", requestQuery)
	}

	// Check for rainfall in inches (2.5 mm = 0.098 in)
	expectedInches := 2.5 / 25.4
	if !strings.Contains(requestQuery, fmt.Sprintf("rainin=%.3f", expectedInches)) {
		t.Errorf("Query string missing or incorrect rainin parameter, got: %s", requestQuery)
	}
}

// TestMultiPublisher tests the multi-publisher implementation
func TestMultiPublisher(t *testing.T) {
	// Create mock publishers to track calls
	mockPub1 := &mockPublisher{t: t, name: "Mock1"}
	mockPub2 := &mockPublisher{t: t, name: "Mock2"}

	// Create a multi-publisher with the mocks
	multiPub := NewMultiPublisher([]Publisher{mockPub1, mockPub2})

	// Create a test weather data
	data := &models.WeatherData{
		Timestamp:     time.Now(),
		Temperature:   21.5,
		Humidity:      45.0,
		Pressure:      1012.5,
		WindSpeed:     5.5,
		WindDirection: 180.0,
		Rain:          0.0,
		UVIndex:       5.0,
	}

	// Publish the data using the multi-publisher
	err := multiPub.Publish(data)
	if err != nil {
		t.Fatalf("Failed to publish data: %v", err)
	}

	// Verify that both mock publishers were called
	if mockPub1.callCount != 1 {
		t.Errorf("Expected Mock1 to be called once, got %d calls", mockPub1.callCount)
	}

	if mockPub2.callCount != 1 {
		t.Errorf("Expected Mock2 to be called once, got %d calls", mockPub2.callCount)
	}

	// Test error handling when one publisher fails
	mockPub1.shouldError = true

	// Publish again, with Mock1 set to return an error
	err = multiPub.Publish(data)

	// The overall operation should still succeed (errors are logged, not returned)
	if err != nil {
		t.Fatalf("Expected success despite one publisher failing, got error: %v", err)
	}

	// Verify call counts
	if mockPub1.callCount != 2 {
		t.Errorf("Expected Mock1 to be called twice, got %d calls", mockPub1.callCount)
	}

	if mockPub2.callCount != 2 {
		t.Errorf("Expected Mock2 to be called twice, got %d calls", mockPub2.callCount)
	}
}

// mockPublisher implements the Publisher interface for testing
type mockPublisher struct {
	t           *testing.T
	name        string
	callCount   int
	shouldError bool
}

func (m *mockPublisher) Publish(data *models.WeatherData) error {
	m.callCount++

	if m.shouldError {
		return fmt.Errorf("%s publisher error", m.name)
	}

	// Verify the data
	if data == nil {
		m.t.Errorf("%s received nil data", m.name)
	}

	return nil
}
