package main

import (
	"log"
	"os"
)

var cfg config

func main() {
	if err := readConfiguration(os.Args[1]); err != nil {
		log.Fatal("An error has occurred while read configuration file:", err)
		os.Exit(1)
	}

	sif := newTTYReceiver()
	mqttClient := newMQTTClient()
	graphiteClient := newGraphiteClient()

	ttyInput := make(chan string, 10)
	graphiteOutput := make(chan *Metric, 10)
	mqttOutput := make(chan *Metric, 10)

	go readFromTTY(sif, ttyInput)
	go sendGraphite(graphiteClient, graphiteOutput)
	go sendMQTT(mqttClient, mqttOutput)

	parseInput(ttyInput, graphiteOutput, mqttOutput)
}
