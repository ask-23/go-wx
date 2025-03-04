package interceptor

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ask-23/go-wx/internal/models"
	"github.com/ask-23/go-wx/pkg/config"
)

// MockDatabase implements a mock of the database interface for testing
type MockDatabase struct {
	SavedData *models.WeatherData
	SaveCalls int
}

func (m *MockDatabase) SaveWeatherData(data *models.WeatherData) error {
	m.SavedData = data
	m.SaveCalls++
	return nil
}

func (m *MockDatabase) GetLatestWeatherData() (*models.WeatherData, error) {
	return m.SavedData, nil
}

func (m *MockDatabase) GetWeatherDataRange(start, end time.Time) ([]*models.WeatherData, error) {
	return []*models.WeatherData{m.SavedData}, nil
}

func (m *MockDatabase) Close() error {
	return nil
}

// MockPublisher implements a mock of the publisher interface for testing
type MockPublisher struct {
	PublishedData *models.WeatherData
	PublishCalls  int
}

func (m *MockPublisher) Publish(data *models.WeatherData) error {
	m.PublishedData = data
	m.PublishCalls++
	return nil
}

// TestInterceptorHandleEcowittData tests the Ecowitt data interceptor
func TestInterceptorHandleEcowittData(t *testing.T) {
	// Create mock database and publisher
	mockDB := &MockDatabase{}
	mockPub := &MockPublisher{}

	// Create a test config
	cfg := config.InterceptorConfig{
		Port:     8000,
		Address:  "0.0.0.0",
		Interval: 60,
	}

	// Create a new interceptor
	interceptor, err := NewInterceptor(cfg, mockDB, []Publisher{mockPub})
	if err != nil {
		t.Fatalf("Failed to create interceptor: %v", err)
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(interceptor.handleEcowittData))
	defer server.Close()

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

	// Send a POST request with the form data
	resp, err := http.Post(server.URL, "application/x-www-form-urlencoded", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Verify that the data was saved to the database
	if mockDB.SaveCalls != 1 {
		t.Errorf("Expected 1 call to SaveWeatherData, got %d", mockDB.SaveCalls)
	}

	// Verify the saved data values
	if mockDB.SavedData == nil {
		t.Fatalf("SavedData is nil, expected WeatherData object")
	}

	// Use approximate comparison for floating point values
	// Convert Fahrenheit to Celsius: (F - 32) * 5/9
	expectedTempC := (70.5 - 32) * 5 / 9
	if !approximatelyEqual(mockDB.SavedData.Temperature, expectedTempC, 0.1) {
		t.Errorf("Expected temperature %.2f°C, got %.2f°C", expectedTempC, mockDB.SavedData.Temperature)
	}

	if mockDB.SavedData.Humidity != 45.0 {
		t.Errorf("Expected humidity 45%%, got %.1f%%", mockDB.SavedData.Humidity)
	}

	// Convert inHg to hPa: inHg * 33.86389
	expectedPressure := 29.92 * 33.86389
	if !approximatelyEqual(mockDB.SavedData.Pressure, expectedPressure, 0.1) {
		t.Errorf("Expected pressure %.2f hPa, got %.2f hPa", expectedPressure, mockDB.SavedData.Pressure)
	}

	// Convert mph to m/s: mph * 0.44704
	expectedWindSpeed := 5.5 * 0.44704
	if !approximatelyEqual(mockDB.SavedData.WindSpeed, expectedWindSpeed, 0.1) {
		t.Errorf("Expected wind speed %.2f m/s, got %.2f m/s", expectedWindSpeed, mockDB.SavedData.WindSpeed)
	}

	if mockDB.SavedData.WindDirection != 180.0 {
		t.Errorf("Expected wind direction 180°, got %.1f°", mockDB.SavedData.WindDirection)
	}

	// Verify that the data was published
	if mockPub.PublishCalls != 1 {
		t.Errorf("Expected 1 call to Publish, got %d", mockPub.PublishCalls)
	}

	if mockPub.PublishedData == nil {
		t.Fatalf("PublishedData is nil, expected WeatherData object")
	}
}

// TestParseWeatherData tests the parsing of Ecowitt form data into WeatherData
func TestParseWeatherData(t *testing.T) {
	// Create a form data map for testing
	formData := map[string][]string{
		"dateutc":        {"2023-05-01 12:00:00"},
		"tempf":          {"70.5"},
		"humidity":       {"45"},
		"baromrelin":     {"29.92"},
		"winddir":        {"180"},
		"windspeedmph":   {"5.5"},
		"windgustmph":    {"8.0"},
		"dailyrainin":    {"0.0"},
		"solarradiation": {"850.5"},
		"uv":             {"5"},
	}

	// Parse the weather data
	data, err := parseWeatherData(formData)
	if err != nil {
		t.Fatalf("Failed to parse weather data: %v", err)
	}

	// Verify the parsed data
	// Convert Fahrenheit to Celsius: (F - 32) * 5/9
	expectedTempC := (70.5 - 32) * 5 / 9
	if !approximatelyEqual(data.Temperature, expectedTempC, 0.1) {
		t.Errorf("Expected temperature %.2f°C, got %.2f°C", expectedTempC, data.Temperature)
	}

	if data.Humidity != 45.0 {
		t.Errorf("Expected humidity 45%%, got %.1f%%", data.Humidity)
	}

	// Convert inHg to hPa: inHg * 33.86389
	expectedPressure := 29.92 * 33.86389
	if !approximatelyEqual(data.Pressure, expectedPressure, 0.1) {
		t.Errorf("Expected pressure %.2f hPa, got %.2f hPa", expectedPressure, data.Pressure)
	}

	// Test parsing with missing data
	formDataMissing := map[string][]string{
		"dateutc":  {"2023-05-01 12:00:00"},
		"tempf":    {"70.5"},
		"humidity": {"45"},
		// Missing other fields
	}

	// Parse the incomplete weather data
	dataMissing, err := parseWeatherData(formDataMissing)
	if err != nil {
		t.Fatalf("Failed to parse incomplete weather data: %v", err)
	}

	// Verify default values for missing fields
	if dataMissing.Pressure != 0.0 {
		t.Errorf("Expected default pressure 0.0 hPa, got %.2f hPa", dataMissing.Pressure)
	}

	if dataMissing.WindSpeed != 0.0 {
		t.Errorf("Expected default wind speed 0.0 m/s, got %.2f m/s", dataMissing.WindSpeed)
	}
}

// approximatelyEqual compares two float64 values within a given tolerance
func approximatelyEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
