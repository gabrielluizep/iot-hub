#include <WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>
#include <Adafruit_Sensor.h>
#include <DHT.h>
#include <DHT_U.h>
#include <Adafruit_BME280.h>
#include <WiFiClientSecure.h>

// WiFi credentials
const char* ssid = "Casa da Injecao";
const char* password = "eronildo";

// MQTT Broker settings
const char* mqtt_server = ""; // MQTT broker IP from .env
const int mqtt_port = 8883;  // Port for SSL/TLS
const char* mqtt_username = ""; // MQTT username from .env
const char* mqtt_password = ""; // MQTT password from .env
const char* mqtt_client_id = ""; // MQTT client ID
const char* topic = "sensor/readings";

// DHT Sensor configuration
#define LDRPIN 15  
#define DHTPIN 12  // GPIO pin where the DHT sensor is connected
#define DHTTYPE DHT11  // DHT 22 (AM2302)
DHT dht(DHTPIN, DHTTYPE);

// Replace with your root CA certificate
const char* rootCACertificate = ""; // Root CA certificate from `echo | openssl s_client -connect <mqtt_broker>:<port> -showcerts`


// Initialize WiFiClientSecure
WiFiClientSecure wifiClient;
PubSubClient client(wifiClient);

void setup() {
    Serial.begin(115200);

    // Initialize sensors
    dht.begin();   

    // Connect to Wi-Fi
    WiFi.begin(ssid, password);
    while (WiFi.status() != WL_CONNECTED) {
        delay(1000);
        Serial.println("Connecting to WiFi...");
    }
    Serial.println("Connected to WiFi");

    // Set root CA certificate for SSL/TLS
    wifiClient.setCACert(rootCACertificate);

    // Configure the MQTT client
    client.setServer(mqtt_server, mqtt_port);

    while (!client.connected()) {
        Serial.println("Connecting to MQTT broker...");
        if (client.connect(mqtt_client_id, mqtt_username, mqtt_password)) {
            Serial.println("Connected to MQTT broker");
        } else {
            Serial.print("Failed with state ");
            Serial.print(client.state());
            delay(2000);
        }
    }
}

void loop() {
    delay(2000);

    // Read sensor data
    float temperature = dht.readTemperature();
    float humidity = dht.readHumidity();
    float luminosity = analogRead(LDRPIN);

    if(isnan(temperature)) {
        Serial.println("Failed to read temperature from DHT sensor!");
        return;
    }
    Serial.printf("Temperature: %.2f°C\n", temperature);

    if(isnan(humidity)) {
        Serial.println("Failed to read humidity from DHT sensor!");
        return;
    }
    Serial.printf("Humidity: %.2f%%\n", humidity);

    if(isnan(luminosity)) {
        Serial.println("Failed to read luminosity from LDR sensor!");
        return;
    }
    Serial.printf("Luminosity: %.2f\n", luminosity);
    return;

    long timestamp = time(nullptr);

    // Prepare JSON payload
    StaticJsonDocument<256> doc;
    doc["id"] = 1;
    doc["timestamp"] = timestamp;
    doc["temperature"] = temperature;
    doc["humidity"] = humidity;
    doc["luminosity"] = luminosity;

    char jsonBuffer[256];
    serializeJson(doc, jsonBuffer);

    // Publish JSON message
    client.publish(topic, jsonBuffer);
    Serial.printf("Message sent (%s): %s\n", topic, jsonBuffer);
}
