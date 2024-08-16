package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

// Connection string
var connStr string

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

	_, err = db.Exec(`INSERT INTO sensor_data (id, timestamp, temperature, humidity, luminosity) VALUES ($1, $2, $3, $4, $5)`, sensorData.ID, sensorData.Timestamp, sensorData.Temperature, sensorData.Humidity, sensorData.Luminosity)
	if err != nil {
		fmt.Println("Error inserting data into database:", err)
		return
	}

	fmt.Println("Data successfully stored in database")
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Retrieve environment variables
	broker := os.Getenv("MQTT_BROKER")
	clientID := os.Getenv("MQTT_CLIENT_ID")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")
	topic := os.Getenv("MQTT_READINGS_TOPIC")

	// Connection string
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
						PRIMARY KEY (id, timestamp))`)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	fmt.Println("Initializing subscriber")

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

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected to MQTT broker")

	if token := c.Subscribe(topic, 0, msgRcvd); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	} else {
		fmt.Println("Subscribed to topic:", topic)
	}

	// Keep the subscriber running
	for {
		time.Sleep(1 * time.Second)
	}
}
