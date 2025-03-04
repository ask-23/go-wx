package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ask-23/go-wx/internal/models"
)

// TestFormDataHandling tests handling of form data similar to what an Ecowitt device would send
func TestFormDataHandling(t *testing.T) {
	// Create a test handler for form data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get a few key fields
		tempF := r.FormValue("tempf")
		if tempF != "70.5" {
			t.Errorf("Expected tempf=70.5, got %s", tempF)
		}

		humidity := r.FormValue("humidity")
		if humidity != "45" {
			t.Errorf("Expected humidity=45, got %s", humidity)
		}

		// Return success
		w.WriteHeader(http.StatusOK)
	})

	// Create a form data payload simulating an Ecowitt device
	formData := url.Values{}
	formData.Set("PASSKEY", "12345")
	formData.Set("stationtype", "GW1000")
	formData.Set("dateutc", "2023-05-01 12:00:00")
	formData.Set("tempf", "70.5")
	formData.Set("humidity", "45")
	formData.Set("baromrelin", "29.92")
	formData.Set("baromabsin", "29.92")
	formData.Set("winddir", "180")
	formData.Set("windspeedmph", "5.5")
	formData.Set("windgustmph", "8.0")
	formData.Set("dailyrainin", "0.0")
	formData.Set("weeklyrainin", "0.5")
	formData.Set("monthlyrainin", "2.5")
	formData.Set("yearlyrainin", "10.0")
	formData.Set("solarradiation", "850.5")
	formData.Set("uv", "5")

	// Create a request
	req, err := http.NewRequest(
		"POST",
		"/data",
		bytes.NewBufferString(formData.Encode()),
	)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set appropriate headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, http.StatusOK)
	}
}

// TestWeatherDataProcessing tests the conversion and processing of weather data
func TestWeatherDataProcessing(t *testing.T) {
	// Create a sample weather data object
	data := &models.WeatherData{
		Timestamp:   time.Now(),
		Temperature: 21.5,    // 21.5°C (70.7°F)
		Humidity:    45.0,    // 45%
		Pressure:    1013.25, // 1013.25 hPa (29.92 inHg)
	}

	// Convert Celsius to Fahrenheit: °F = (°C × 9/5) + 32
	// 21.5°C should be about 70.7°F
	tempF := (data.Temperature * 9 / 5) + 32
	if tempF < 70.6 || tempF > 70.8 {
		t.Errorf("Temperature conversion incorrect, expected ~70.7°F, got %.1f°F", tempF)
	}

	// Convert hPa to inHg: inHg = hPa / 33.86389
	// 1013.25 hPa should be about 29.92 inHg
	inHg := data.Pressure / 33.86389
	if inHg < 29.91 || inHg > 29.93 {
		t.Errorf("Pressure conversion incorrect, expected ~29.92 inHg, got %.2f inHg", inHg)
	}
}
