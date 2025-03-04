package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ask-23/go-wx/internal/models"
)

// MockWeatherData returns a sample weather data record for testing
func MockWeatherData() *models.WeatherData {
	return &models.WeatherData{
		Timestamp:     time.Now(),
		Temperature:   72.5,
		Humidity:      45.0,
		Pressure:      1012.5,
		WindSpeed:     8.0,
		WindDirection: 180.0,
		Rain:          0.0,
		UVIndex:       5.0,
		CloudBase:     4500.0,
		DewPoint:      50.0,
		WindChill:     72.5,
		HeatIndex:     72.5,
	}
}

// TestJSONResponse tests a simple JSON response
func TestJSONResponse(t *testing.T) {
	// Create a test handler that returns mock weather data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := MockWeatherData()
		json.NewEncoder(w).Encode(data)
	})

	// Create a request
	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, http.StatusOK)
	}

	// Check the content type
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Handler returned wrong content type: got %v, expected %v", contentType, expectedContentType)
	}

	// Parse the response body
	var responseData models.WeatherData
	if err := json.Unmarshal(rr.Body.Bytes(), &responseData); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify the response data
	if responseData.Temperature != 72.5 {
		t.Errorf("Expected temperature 72.5, got %.1f", responseData.Temperature)
	}

	if responseData.Humidity != 45.0 {
		t.Errorf("Expected humidity 45.0, got %.1f", responseData.Humidity)
	}
}
