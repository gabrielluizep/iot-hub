package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
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

func main() {
	// Connection string
	connStr := "postgresql://postgres:u00ths5w83oiqyUD@tangibly-colossal-wildfowl.data-1.use1.tembo.io:5432/postgres"

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
		// // start and end query parameters for pagination
		// start := c.Query("start")
		// end := c.Query("end")

		// // parse to int
		// startTime, err := time.Parse("2006-01-02", start)
		// if err != nil {
		// 	log.Fatalf("Error parsing start date: %v", err)
		// }
		// startUnix := startTime.Unix()

		// endTime, err := time.Parse("2006-01-02", end)
		// if err != nil {
		// 	log.Fatalf("Error parsing end date: %v", err)
		// }
		// endUnix := endTime.Unix()

		// query data from sensor_data
		// rows, err := db.Query("SELECT * FROM sensor_data WHERE timestamp >= $1 AND timestamp <= $2", startUnix, endUnix)
		rows, err := db.Query("SELECT * FROM sensor_data")
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

			fmt.Printf("Sensor Data: %+v\n", data)

			sensorData = append(sensorData, data)
		}

		c.JSON(200, sensorData)
	})

	router.Run("localhost:8080")
}
