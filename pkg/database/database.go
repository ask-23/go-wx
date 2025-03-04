package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ask-23/go-wx/internal/models"
	"github.com/ask-23/go-wx/pkg/config"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
)

// Database represents a database connection and operations
type Database struct {
	db     *sql.DB
	config *config.DatabaseConfig
}

// NewDatabase creates a new database connection
func NewDatabase(cfg config.DatabaseConfig) (*Database, error) {
	var db *sql.DB
	var err error
	var connStr string

	// Create connection string based on database type
	switch cfg.Type {
	case "mariadb":
		// Format: username:password@tcp(host:port)/dbname
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		db, err = sql.Open("mysql", connStr)
	case "postgres":
		// Format: postgres://username:password@host:port/dbname?sslmode=disable
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		db, err = sql.Open("postgres", connStr)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	database := &Database{
		db:     db,
		config: &cfg,
	}

	// Initialize database schema
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// SaveWeatherData saves weather data to the database
func (d *Database) SaveWeatherData(data *models.WeatherData) error {
	// SQL query to insert weather data
	var query string
	var args []interface{}

	switch d.config.Type {
	case "mariadb":
		query = `INSERT INTO weather_data (
			timestamp, temperature, humidity, pressure, wind_speed, 
			wind_direction, rain, uv_index, cloud_base
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		args = []interface{}{
			data.Timestamp, data.Temperature, data.Humidity, data.Pressure,
			data.WindSpeed, data.WindDirection, data.Rain, data.UVIndex, data.CloudBase,
		}
	case "postgres":
		query = `INSERT INTO weather_data (
			timestamp, temperature, humidity, pressure, wind_speed, 
			wind_direction, rain, uv_index, cloud_base
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		args = []interface{}{
			data.Timestamp, data.Temperature, data.Humidity, data.Pressure,
			data.WindSpeed, data.WindDirection, data.Rain, data.UVIndex, data.CloudBase,
		}
	}

	_, err := d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to save weather data: %w", err)
	}

	return nil
}

// GetLatestWeatherData retrieves the most recent weather data
func (d *Database) GetLatestWeatherData() (*models.WeatherData, error) {
	query := `SELECT 
		timestamp, temperature, humidity, pressure, wind_speed, 
		wind_direction, rain, uv_index, cloud_base
	FROM weather_data 
	ORDER BY timestamp DESC 
	LIMIT 1`

	var data models.WeatherData
	err := d.db.QueryRow(query).Scan(
		&data.Timestamp, &data.Temperature, &data.Humidity, &data.Pressure,
		&data.WindSpeed, &data.WindDirection, &data.Rain, &data.UVIndex, &data.CloudBase,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no weather data available")
		}
		return nil, fmt.Errorf("failed to get latest weather data: %w", err)
	}

	return &data, nil
}

// GetWeatherDataRange retrieves weather data for a specific time range
func (d *Database) GetWeatherDataRange(start, end time.Time) ([]*models.WeatherData, error) {
	// SQL query to get weather data for a time range
	var query string
	var args []interface{}

	switch d.config.Type {
	case "mariadb":
		query = `SELECT 
			timestamp, temperature, humidity, pressure, wind_speed, 
			wind_direction, rain, uv_index, cloud_base
		FROM weather_data 
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC`
		args = []interface{}{start, end}
	case "postgres":
		query = `SELECT 
			timestamp, temperature, humidity, pressure, wind_speed, 
			wind_direction, rain, uv_index, cloud_base
		FROM weather_data 
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp ASC`
		args = []interface{}{start, end}
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query weather data range: %w", err)
	}
	defer rows.Close()

	var results []*models.WeatherData
	for rows.Next() {
		var data models.WeatherData
		err := rows.Scan(
			&data.Timestamp, &data.Temperature, &data.Humidity, &data.Pressure,
			&data.WindSpeed, &data.WindDirection, &data.Rain, &data.UVIndex, &data.CloudBase,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan weather data row: %w", err)
		}
		results = append(results, &data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating weather data rows: %w", err)
	}

	return results, nil
}

// initSchema initializes the database schema if it doesn't exist
func (d *Database) initSchema() error {
	var createTableSQL string

	// Create tables based on database type
	switch d.config.Type {
	case "mariadb":
		createTableSQL = `
		CREATE TABLE IF NOT EXISTS weather_data (
			id INT AUTO_INCREMENT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			temperature FLOAT,
			humidity FLOAT,
			pressure FLOAT,
			wind_speed FLOAT,
			wind_direction FLOAT,
			rain FLOAT,
			uv_index FLOAT,
			cloud_base FLOAT,
			INDEX idx_timestamp (timestamp)
		)`
	case "postgres":
		createTableSQL = `
		CREATE TABLE IF NOT EXISTS weather_data (
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMP NOT NULL,
			temperature FLOAT,
			humidity FLOAT,
			pressure FLOAT,
			wind_speed FLOAT,
			wind_direction FLOAT,
			rain FLOAT,
			uv_index FLOAT,
			cloud_base FLOAT
		);
		CREATE INDEX IF NOT EXISTS idx_timestamp ON weather_data (timestamp);`
	}

	_, err := d.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}

	return nil
}
