package main

import (
	"encoding/json"
	"log"
	"path"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func send_to_mqtt(client mqtt.Client, input chan *Metric) {
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for {
		message := <-input

		log.Printf("MQTT Sending to '%s': %#v", message.MQTTTopic(), message)

		token := client.Publish(message.MQTTTopic(), 0, false, message.MQTTValue())
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
