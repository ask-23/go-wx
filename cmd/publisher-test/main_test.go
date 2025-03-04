package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ask-23/go-wx/internal/models"
)

// TestCustomPublisherHttp tests publishing weather data to an HTTP endpoint
func TestCustomPublisherHttp(t *testing.T) {
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
		CloudBase:     1500.0,
		DewPoint:      9.5,
		WindChill:     21.5,
		HeatIndex:     21.5,
	}

	// Create the HTTP client
	client := &http.Client{}

	// Marshal the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data to JSON: %v", err)
	}

	// Create a request
	req, err := http.NewRequest(
		"POST",
		testServer.URL,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// Verify received data
	if receivedData == nil {
		t.Fatalf("No data received by the test server")
	}

	// Verify some key fields
	if temperature, ok := receivedData["temperature"].(float64); !ok || temperature != 21.5 {
		t.Errorf("Expected temperature 21.5, got %v", receivedData["temperature"])
	}

	if humidity, ok := receivedData["humidity"].(float64); !ok || humidity != 45.0 {
		t.Errorf("Expected humidity 45.0, got %v", receivedData["humidity"])
	}

	// Check headers
	if contentType := receivedHeaders.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type header 'application/json', got '%s'", contentType)
	}

	if apiKey := receivedHeaders.Get("X-API-Key"); apiKey != "test-api-key" {
		t.Errorf("Expected X-API-Key header 'test-api-key', got '%s'", apiKey)
	}
}
