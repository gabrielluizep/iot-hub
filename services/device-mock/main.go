package main

import (
	"crypto/tls"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	fmt.Println("Initializing publisher")

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Set to false in production to verify the server's certificate
		ClientAuth:         tls.NoClientCert,
	}

	opts := mqtt.NewClientOptions().
		AddBroker("ssl://8f6cec80afc747f49167bf47720d82e8.s1.eu.hivemq.cloud:8883").
		SetClientID("publisher").
		SetUsername("mock").
		SetPassword("ne#`%I<^N3bX-:03,fb0Yx]fP7)n9=ELR#>oV6yrnZ~>eZ,fiI").
		SetTLSConfig(tlsConfig)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected to MQTT broker")

	topic := "test/topic"
	payload := "Hello, this is a test message!"

	token := c.Publish(topic, 0, false, payload)
	token.Wait()

	fmt.Printf("Published message to topic: %s => Message: %s\n", topic, payload)

	// Wait for a moment before disconnecting to ensure the message is sent
	time.Sleep(1 * time.Second)

	c.Disconnect(250)
	fmt.Println("Disconnected from MQTT broker")
}
