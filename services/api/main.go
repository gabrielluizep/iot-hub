package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Data structure to represent the incoming data
type SensorData struct {
	ID          int64   `json:"id"`
	Timestamp   int64   `json:"timestamp"` // Using int64 for Unix timestamp
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Luminosity  float64 `json:"luminosity"`
}

// parseDate attempts to parse a date string using multiple formats.
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",          // Format for date only
		"2006-01-02T15:04:05", // Format for date and time
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Connection string
	connStr := os.Getenv("POSTGRES_CONN_STR")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// create table for sensor_data
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS sensor_data (id SERIAL PRIMARY KEY, timestamp BIGINT, temperature DOUBLE PRECISION, humidity DOUBLE PRECISION, luminosity DOUBLE PRECISION)")
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	router.GET("/sensor-data", func(c *gin.Context) {
		startStr := c.DefaultQuery("start", "2000-01-01")
		endStr := c.DefaultQuery("end", "2100-01-01")

		// Parse the start and end parameters using the parseDate function
		startTime, err := parseDate(startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date"})
			return
		}
		endTime, err := parseDate(endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date"})
			return
		}

		// Convert to Unix timestamps
		startUnix := startTime.Unix()
		endUnix := endTime.Unix()

		// Corrected SQL query with PostgreSQL syntax
		rows, err := db.Query("SELECT * FROM sensor_data WHERE timestamp >= $1 AND timestamp <= $2", startUnix, endUnix)
		if err != nil {
			log.Fatalf("Error querying data: %v", err)
		}
		defer rows.Close()

		var sensorData []SensorData
		for rows.Next() {
			var data SensorData
			err := rows.Scan(&data.ID, &data.Timestamp, &data.Temperature, &data.Humidity, &data.Luminosity)
			if err != nil {
				log.Fatalf("Error scanning row: %v", err)
			}

			sensorData = append(sensorData, data)
		}

		c.JSON(200, sensorData)
	})

	router.Run("localhost:8080")
}
