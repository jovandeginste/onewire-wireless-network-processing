package main

import (
	"log"
	"os"
)

var cfg config

func main() {
	if err := read_configuration(os.Args[1]); err != nil {
		log.Fatal("An error has occurred while read configuration file:", err)
		os.Exit(1)
	}

	sif := newTTYReceiver()
	mqttClient := newMQTTClient()
	graphiteClient := newGraphiteClient()

	ttyInput := make(chan string, 10)
	graphiteOutput := make(chan *Metric, 10)
	mqttOutput := make(chan *Metric, 10)

	go read_from_tty(sif, ttyInput)
	go send_to_graphite(graphiteClient, graphiteOutput)
	go send_to_mqtt(mqttClient, mqttOutput)

	parse_input(ttyInput, graphiteOutput, mqttOutput)
}
