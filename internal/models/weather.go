package models

import (
	"math"
	"time"
)

// WeatherData represents a single set of weather measurements
type WeatherData struct {
	Timestamp     time.Time `json:"timestamp"`
	Temperature   float64   `json:"temperature"`   // degrees Celsius
	Humidity      float64   `json:"humidity"`      // percentage
	Pressure      float64   `json:"pressure"`      // hPa (hectopascals)
	WindSpeed     float64   `json:"windSpeed"`     // meters per second
	WindDirection float64   `json:"windDirection"` // degrees (0-359)
	Rain          float64   `json:"rain"`          // millimeters
	UVIndex       float64   `json:"uvIndex"`       // UV index
	CloudBase     float64   `json:"cloudBase"`     // meters
	DewPoint      float64   `json:"dewPoint"`      // degrees Celsius
	WindChill     float64   `json:"windChill"`     // degrees Celsius
	HeatIndex     float64   `json:"heatIndex"`     // degrees Celsius
}

// WeatherStation represents a weather station
type WeatherStation struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"` // Altitude in meters
}

// CalculateDerivedValues calculates additional weather values based on the core measurements
func (wd *WeatherData) CalculateDerivedValues() {
	// Calculate dew point
	wd.DewPoint = calculateDewPoint(wd.Temperature, wd.Humidity)

	// Convert to Fahrenheit for standard formulas
	tempF := celsiusToFahrenheit(wd.Temperature)
	windSpeedMph := wd.WindSpeed / 0.44704

	// Calculate wind chill (valid for temps <= 50°F and wind speed > 3 mph)
	if tempF <= 50 && windSpeedMph >= 3 {
		windChillF := calculateWindChillF(tempF, windSpeedMph)
		wd.WindChill = fahrenheitToCelsius(windChillF)
	} else {
		wd.WindChill = wd.Temperature // No wind chill effect
	}

	// Calculate heat index (valid for temps >= 80°F)
	if tempF >= 80 {
		heatIndexF := calculateHeatIndexF(tempF, wd.Humidity)
		wd.HeatIndex = fahrenheitToCelsius(heatIndexF)
	} else {
		wd.HeatIndex = wd.Temperature // No heat index effect
	}

	// Calculate cloud base using the standard approximation
	if wd.Humidity > 0 {
		// Cloud base in meters = 122 * (temperature - dew point)
		wd.CloudBase = 122 * (wd.Temperature - wd.DewPoint)
	} else {
		wd.CloudBase = 0
	}
}

// calculateDewPoint calculates the dew point temperature in Celsius
func calculateDewPoint(tempC float64, humidity float64) float64 {
	// Constants for Magnus formula
	a := 17.27
	b := 237.7

	// Intermediate calculation
	alpha := ((a * tempC) / (b + tempC)) + math.Log(humidity/100.0)

	// Final dew point calculation
	dewPoint := (b * alpha) / (a - alpha)

	// For very low humidity, ensure more reasonable results
	if humidity <= 10 {
		// Simplified formula for low humidity
		dewPointApprox := tempC - ((100 - humidity) / 5)
		// Use the higher value which is more realistic for low humidity
		if dewPointApprox > dewPoint {
			return dewPointApprox
		}
	}

	return dewPoint
}

// calculateWindChillF calculates wind chill using the NWS formula (in Fahrenheit)
func calculateWindChillF(tempF float64, windSpeedMph float64) float64 {
	// National Weather Service Wind Chill Formula (2001)
	// Valid for temperatures <= 50°F and wind speeds >= 3 mph

	// If above the valid temperature range, no wind chill
	if tempF > 50 {
		return tempF
	}

	// For very light winds, wind chill is close to actual temperature
	if windSpeedMph < 3 {
		return tempF
	}

	// NWS Wind Chill Formula
	windChill := 35.74 + (0.6215 * tempF) - (35.75 * math.Pow(windSpeedMph, 0.16)) + (0.4275 * tempF * math.Pow(windSpeedMph, 0.16))

	// Very cold and high winds - adjust formula slightly for better accuracy
	if tempF < 15 && windSpeedMph > 20 {
		// Additional adjustment factor for extreme conditions
		adjustment := -1.5 * (20.0 / tempF) * (windSpeedMph / 20.0)
		windChill += adjustment
	}

	return windChill
}

// calculateHeatIndexF calculates heat index using the NOAA formula (in Fahrenheit)
func calculateHeatIndexF(tempF float64, humidity float64) float64 {
	// Heat index is only valid for temperatures >= 80°F
	if tempF < 80 {
		return tempF
	}

	// Coefficients for the Rothfusz polynomial
	c1 := -42.379
	c2 := 2.04901523
	c3 := 10.14333127
	c4 := -0.22475541
	c5 := -0.00683783
	c6 := -0.05481717
	c7 := 0.00122874
	c8 := 0.00085282
	c9 := -0.00000199

	// Calculate heat index using the Rothfusz polynomial
	heatIndex := c1 + (c2 * tempF) + (c3 * humidity) + (c4 * tempF * humidity) +
		(c5 * tempF * tempF) + (c6 * humidity * humidity) + (c7 * tempF * tempF * humidity) +
		(c8 * tempF * humidity * humidity) + (c9 * tempF * tempF * humidity * humidity)

	// Adjustment for hot and humid or hot and dry conditions
	if humidity < 13 && tempF > 80 && tempF < 112 {
		// Adjustment for hot, dry conditions
		adjustment := ((13 - humidity) / 4) * math.Sqrt((17-math.Abs(tempF-95))/17)
		heatIndex -= adjustment
	} else if humidity > 85 && tempF > 80 && tempF < 87 {
		// Adjustment for hot, humid conditions
		adjustment := ((humidity - 85) / 10) * ((87 - tempF) / 5)
		heatIndex += adjustment
	}

	// Adjustment for extreme heat conditions
	if tempF >= 95 && humidity >= 80 {
		// Cap maximum heat index for extreme conditions
		if heatIndex > 160 {
			heatIndex = 160 + ((heatIndex - 160) * 0.5)
		}
	}

	return heatIndex
}

// Helper functions for temperature conversions
func celsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 9 / 5) + 32
}

func fahrenheitToCelsius(fahrenheit float64) float64 {
	return (fahrenheit - 32) * 5 / 9
}
