package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ask-23/go-wx/internal/models"
)

// mockWeatherData returns a sample weather data record for testing
func mockWeatherData() *models.WeatherData {
	return &models.WeatherData{
		Timestamp:     time.Now(),
		Temperature:   22.5,   // in Celsius
		Humidity:      45.0,   // percentage
		Pressure:      1012.5, // hPa
		WindSpeed:     4.5,    // m/s
		WindDirection: 180.0,  // degrees
		Rain:          0.0,    // mm
		UVIndex:       5.0,
		CloudBase:     1500.0, // meters
		DewPoint:      10.3,   // Celsius
		WindChill:     22.0,   // Celsius
		HeatIndex:     23.0,   // Celsius
	}
}

// handleCurrentData returns mock current weather data
func handleCurrentData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := mockWeatherData()
	json.NewEncoder(w).Encode(data)
}

// handleHistoryData returns mock historical weather data
func handleHistoryData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Create sample historical data
	now := time.Now()
	history := []*models.WeatherData{
		{
			Timestamp:     now.Add(-2 * time.Hour),
			Temperature:   21.2,
			Humidity:      50.0,
			Pressure:      1011.2,
			WindSpeed:     3.5,
			WindDirection: 170.0,
			Rain:          0.0,
			UVIndex:       6.0,
			CloudBase:     1450.0,
			DewPoint:      10.0,
			WindChill:     21.0,
			HeatIndex:     22.0,
		},
		{
			Timestamp:     now.Add(-1 * time.Hour),
			Temperature:   22.0,
			Humidity:      47.0,
			Pressure:      1012.0,
			WindSpeed:     4.0,
			WindDirection: 175.0,
			Rain:          0.0,
			UVIndex:       5.5,
			CloudBase:     1480.0,
			DewPoint:      10.2,
			WindChill:     21.5,
			HeatIndex:     22.5,
		},
		mockWeatherData(),
	}

	json.NewEncoder(w).Encode(history)
}

// handleHomePage serves a simple HTML page
func handleHomePage(w http.ResponseWriter, r *http.Request) {
	htmlContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>go-wx Demo</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            line-height: 1.6;
        }
        h1 {
            color: #2c3e50;
            text-align: center;
        }
        .weather-data {
            background-color: #f7f7f7;
            border-radius: 10px;
            padding: 20px;
            margin-top: 20px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 10px;
            border-bottom: 1px solid #ddd;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
        .button {
            background-color: #3498db;
            color: white;
            padding: 10px 15px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            margin-top: 20px;
            display: inline-block;
        }
        #refresh-btn {
            display: block;
            margin: 20px auto;
        }
    </style>
</head>
<body>
    <h1>go-wx Weather Station Demo</h1>
    
    <div class="weather-data">
        <h2>Current Weather</h2>
        <div id="current-weather">Loading...</div>
        
        <button id="refresh-btn" class="button">Refresh Data</button>
        
        <h2>Historical Data</h2>
        <div id="historical-data">Loading...</div>
    </div>

    <script>
        // Function to fetch and display current weather data
        async function fetchCurrentWeather() {
            try {
                const response = await fetch('/api/current');
                const data = await response.json();
                
                // Format and display the data
                const timestamp = new Date(data.timestamp).toLocaleString();
                
                let html = '<table>';
                html += '<tr><th>Metric</th><th>Value</th></tr>';
                html += '<tr><td>Time</td><td>' + timestamp + '</td></tr>';
                html += '<tr><td>Temperature</td><td>' + data.temperature.toFixed(1) + '°C</td></tr>';
                html += '<tr><td>Humidity</td><td>' + data.humidity.toFixed(1) + '%</td></tr>';
                html += '<tr><td>Pressure</td><td>' + data.pressure.toFixed(1) + ' hPa</td></tr>';
                html += '<tr><td>Wind Speed</td><td>' + data.windSpeed.toFixed(1) + ' m/s</td></tr>';
                html += '<tr><td>Wind Direction</td><td>' + data.windDirection.toFixed(0) + '°</td></tr>';
                html += '<tr><td>Rain</td><td>' + data.rain.toFixed(1) + ' mm</td></tr>';
                html += '<tr><td>UV Index</td><td>' + data.uvIndex.toFixed(1) + '</td></tr>';
                html += '<tr><td>Cloud Base</td><td>' + data.cloudBase.toFixed(0) + ' m</td></tr>';
                html += '<tr><td>Dew Point</td><td>' + data.dewPoint.toFixed(1) + '°C</td></tr>';
                html += '<tr><td>Wind Chill</td><td>' + data.windChill.toFixed(1) + '°C</td></tr>';
                html += '<tr><td>Heat Index</td><td>' + data.heatIndex.toFixed(1) + '°C</td></tr>';
                html += '</table>';
                
                document.getElementById('current-weather').innerHTML = html;
            } catch (error) {
                console.error('Error fetching current weather:', error);
                document.getElementById('current-weather').innerHTML = 'Error loading data';
            }
        }
        
        // Function to fetch and display historical weather data
        async function fetchHistoricalWeather() {
            try {
                const response = await fetch('/api/history');
                const data = await response.json();
                
                // Create a table of historical data
                let html = '<table>';
                html += '<tr><th>Time</th><th>Temp (°C)</th><th>Humidity (%)</th><th>Pressure (hPa)</th><th>Wind (m/s)</th></tr>';
                
                // Add rows for each data point
                data.forEach(function(item) {
                    const timestamp = new Date(item.timestamp).toLocaleString();
                    html += '<tr>' +
                        '<td>' + timestamp + '</td>' +
                        '<td>' + item.temperature.toFixed(1) + '</td>' +
                        '<td>' + item.humidity.toFixed(1) + '</td>' +
                        '<td>' + item.pressure.toFixed(1) + '</td>' +
                        '<td>' + item.windSpeed.toFixed(1) + '</td>' +
                    '</tr>';
                });
                
                html += '</table>';
                document.getElementById('historical-data').innerHTML = html;
            } catch (error) {
                console.error('Error fetching historical weather:', error);
                document.getElementById('historical-data').innerHTML = 'Error loading data';
            }
        }
        
        // Initial data load
        fetchCurrentWeather();
        fetchHistoricalWeather();
        
        // Set up refresh button
        document.getElementById('refresh-btn').addEventListener('click', function() {
            fetchCurrentWeather();
            fetchHistoricalWeather();
        });
        
        // Auto-refresh every 30 seconds
        setInterval(function() {
            fetchCurrentWeather();
            fetchHistoricalWeather();
        }, 30000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent))
}

func main() {
	// Register handlers
	http.HandleFunc("/", handleHomePage)
	http.HandleFunc("/api/current", handleCurrentData)
	http.HandleFunc("/api/history", handleHistoryData)

	// Start the server
	port := 8080
	fmt.Printf("Starting go-wx demo server on http://localhost:%d\n", port)
	fmt.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
