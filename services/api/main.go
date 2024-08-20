package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-contrib/cors"
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

type LightOn struct {
	LightOn bool `json:"lightOn"`
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

	// Retrieve environment variables
	broker := os.Getenv("MQTT_BROKER")
	clientID := os.Getenv("MQTT_CLIENT_ID")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	fmt.Println("Initializing publisher")

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Set to false in production to verify the server's certificate
		ClientAuth:         tls.NoClientCert,
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetTLSConfig(tlsConfig)

	mqttContext := mqtt.NewClient(opts)
	if token := mqttContext.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected to MQTT broker")

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5173", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	router.GET("/sensors", func(c *gin.Context) {
		// Corrected SQL query with PostgreSQL syntax
		rows, err := db.Query("SELECT DISTINCT id FROM sensor_data ORDER BY id")
		if err != nil {
			log.Fatalf("Error querying data: %v", err)
		}
		defer rows.Close()

		var ids []int
		for rows.Next() {
			var data int
			err := rows.Scan(&data)
			if err != nil {
				log.Fatalf("Error scanning row: %v", err)
			}

			ids = append(ids, data)
		}

		c.JSON(200, ids)
	})

	router.GET("/sensors/:id", func(c *gin.Context) {
		id := c.Param("id")

		// Corrected SQL query with PostgreSQL syntax
		rows, err := db.Query("SELECT * FROM sensor_data WHERE id = $1 ORDER BY timestamp DESC LIMIT 1", id)
		if err != nil {
			log.Fatalf("Error querying data: %v", err)
		}
		defer rows.Close()

		var sensorData SensorData
		for rows.Next() {
			var data SensorData
			err := rows.Scan(&data.ID, &data.Timestamp, &data.Temperature, &data.Humidity, &data.Luminosity)
			if err != nil {
				log.Fatalf("Error scanning row: %v", err)
			}

			sensorData = data
		}

		c.JSON(200, sensorData)
	})

	router.POST("/sensors/:id", func(c *gin.Context) {
		id := c.Param("id")

		// parse the incoming data
		var body LightOn
		err := c.BindJSON(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Create a map to hold the data with the correct types
		data := map[string]interface{}{
			"lightOn": body.LightOn,
		}

		// Convert the data to JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Fatalf("Error marshaling JSON: %v", err)
		}

		topic := "light/" + id

		token := mqttContext.Publish(topic, 0, false, string(jsonData))
		token.Wait()

		fmt.Printf("Message sent (%s): %s\n", topic, string(jsonData))
	})

	router.GET("/sensors/:id/readings", func(c *gin.Context) {
		id := c.Param("id")

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
		rows, err := db.Query("SELECT * FROM sensor_data WHERE timestamp >= $1 AND timestamp <= $2 AND id = $3", startUnix, endUnix, id)
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

	port := os.Getenv("PORT")
	router.Run(":" + port)
}
