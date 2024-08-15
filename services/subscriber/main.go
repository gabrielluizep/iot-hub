package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

// Data structure to represent the incoming data
type SensorData struct {
	ID          int64   `json:"id"`
	Timestamp   int64   `json:"timestamp"` // Using int64 for Unix timestamp
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Luminosity  float64 `json:"luminosity"`
}

// Supabase client
var supabaseClient *supabase.Client

func msgRcvd(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Received message on topic: %s => Message: %s\n", message.Topic(), message.Payload())

	// Parse JSON data
	var sensorData SensorData
	err := json.Unmarshal(message.Payload(), &sensorData)
	if err != nil {
		fmt.Println("Error parsing JSON data:", err)
		return
	}

	// Insert data into the Supabase table
	_, _, err = supabaseClient.From("sensor_data").Insert(sensorData, true, "", "", "").Execute()
	if err != nil {
		fmt.Println("Error inserting data into Supabase:", err)
		return
	}

	fmt.Println("Data successfully stored in Supabase")
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
	topic := os.Getenv("MQTT_TOPIC")
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	// Initialize Supabase client
	supabaseClient, err = supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		fmt.Println("cannot initalize client", err)
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
