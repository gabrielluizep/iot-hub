package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

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
	topic := os.Getenv("MQTT_TOPIC")

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

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected to MQTT broker")

	payload := "Hello, this is a test message from the publisher!"

	token := c.Publish(topic, 0, false, payload)
	token.Wait()

	// Generate random values for temperature, humidity, and luminosity
	temperature := rand.Float64() * 100
	humidity := rand.Float64() * 100
	luminosity := rand.Float64() * 100
	timestamp := time.Now().Unix() // Use int64 for the Unix timestamp
	id := rand.Intn(1000)          // id should remain as an int

	// Create a map to hold the data with the correct types
	data := map[string]interface{}{
		"id":          id,          // Keep as int
		"timestamp":   timestamp,   // Keep as int64
		"temperature": temperature, // float64
		"humidity":    humidity,    // float64
		"luminosity":  luminosity,  // float64
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Publish the JSON payload
	token = c.Publish(topic, 0, false, string(jsonData))
	token.Wait()

	fmt.Printf("Published message to topic: %s => Message: %s\n", topic, string(jsonData))

	// Wait for a moment before disconnecting to ensure the message is sent
	time.Sleep(1 * time.Second)

	c.Disconnect(250)
	fmt.Println("Disconnected from MQTT broker")
}
