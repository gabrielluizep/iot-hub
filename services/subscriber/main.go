package main

import (
	"crypto/tls"
	"fmt"
	"log"
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

	fmt.Printf("Published message to topic: %s => Message: %s\n", topic, payload)

	// Wait for a moment before disconnecting to ensure the message is sent
	time.Sleep(1 * time.Second)

	c.Disconnect(250)
	fmt.Println("Disconnected from MQTT broker")
}
