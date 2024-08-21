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
	LightOn     bool    `json:"lightOn"`
}

type LightOn struct {
	LightOn bool `json:"lightOn"`
}

// Connection string
var connStr string

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

func msgRcvd(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Received message on topic: %s => Message: %s\n", message.Topic(), message.Payload())

	// Parse JSON data
	var sensorData SensorData
	err := json.Unmarshal(message.Payload(), &sensorData)
	if err != nil {
		fmt.Println("Error parsing JSON data:", err)
		return
	}

	// Insert data into the psql database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT INTO sensor_data (id, timestamp, temperature, humidity, luminosity, light_on) VALUES ($1, $2, $3, $4, $5, false)`, sensorData.ID, sensorData.Timestamp, sensorData.Temperature, sensorData.Humidity, sensorData.Luminosity)
	if err != nil {
		fmt.Println("Error inserting data into database:", err)
		return
	}

	fmt.Println("Data successfully stored in database")
}

func main() {
	// Retrieve environment variables
	broker := os.Getenv("MQTT_BROKER")
	clientID := os.Getenv("MQTT_CLIENT_ID")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")
	topic := os.Getenv("MQTT_READINGS_TOPIC")
	connStr = os.Getenv("POSTGRES_CONN_STR")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// create table for sensor_data
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sensor_data (
		id INT, 
		timestamp BIGINT, 
		temperature DOUBLE PRECISION, 
		humidity DOUBLE PRECISION, 
		luminosity DOUBLE PRECISION,
		light_on BOOLEAN,
		PRIMARY KEY (id, timestamp))`)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

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

	if token := mqttContext.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected to MQTT broker")

	if token := mqttContext.Subscribe(topic, 0, msgRcvd); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	} else {
		fmt.Println("Subscribed to topic:", topic)
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))

	// add default headers
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	})

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
			err := rows.Scan(&data.ID, &data.Timestamp, &data.Temperature, &data.Humidity, &data.Luminosity, &data.LightOn)
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
			err := rows.Scan(&data.ID, &data.Timestamp, &data.Temperature, &data.Humidity, &data.Luminosity, &data.LightOn)
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
