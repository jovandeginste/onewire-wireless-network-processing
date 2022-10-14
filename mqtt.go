package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func newMQTTClient() mqtt.Client {
	mqtt.ERROR = log.New(os.Stdout, "", 0)

	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTT.Host).SetClientID("onewire_logger")

	opts.SetKeepAlive(300 * time.Second)

	opts.SetPingTimeout(1 * time.Second)
	opts.Username = cfg.MQTT.Username
	opts.Password = cfg.MQTT.Password

	client := mqtt.NewClient(opts)

	log.Printf("Loaded MQTT connection: %s@%s", cfg.MQTT.Username, cfg.MQTT.Host)

	return client
}

func sendMQTT(client mqtt.Client, input chan *Metric) {
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for {
		message := <-input

		log.Printf("MQTT Sending to '%s': %#v", message.MQTTTopic(), message)

		token := client.Publish(message.MQTTTopic(), 0, true, message.MQTTValue())
		token.Wait()

		if token.Error() != nil {
			log.Println(token.Error())
		}
	}
}

func (m *Metric) MQTTTopic() string {
	return path.Join(cfg.MQTT.TopicPrefix, m.Name, m.Type)
}

func (m *Metric) MQTTValue() string {
	u, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(u)
}
