package main

import (
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/marpaia/graphite-golang"
	"github.com/tarm/serial"
)

var (
	cfg config
	f   mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("TOPIC: %s\n", msg.Topic())
		log.Printf("MSG: %s\n", msg.Payload())
	}
)

func main() {
	if err := read_configuration(os.Args[1]); err != nil {
		log.Fatal("An error has occurred while read configuration file:", err)
		os.Exit(1)
	}

	mqtt.ERROR = log.New(os.Stdout, "", 0)

	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTT.Host).SetClientID("onewire_logger")

	opts.SetKeepAlive(60 * time.Second)
	// Set the message callback handler
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)
	opts.Username = cfg.MQTT.Username
	opts.Password = cfg.MQTT.Password

	mqttClient := mqtt.NewClient(opts)

	portStr := cfg.Receiver.PortStr
	baudRate := cfg.Receiver.BaudRate

	if err := reset_tty(portStr, baudRate); err != nil {
		log.Fatal("An error has occurred while resetting tty:", err)
		os.Exit(1)
	}

	sif, err := serial.OpenPort(&serial.Config{Name: portStr, Baud: baudRate})
	if err != nil {
		log.Fatal("An error has occurred while trying to open the tty:", err)
		os.Exit(1)
	}

	// try to connect a graphite server
	graphite, err := graphite.NewGraphite(cfg.Graphite.Configuration.Host, cfg.Graphite.Configuration.Port)
	if err != nil {
		log.Fatal("An error has occurred while trying to create a Graphite connector:", err)
		os.Exit(1)
	}

	graphite.Prefix = cfg.Graphite.Configuration.Prefix

	log.Printf("Loaded Graphite connection: %#v", graphite)

	ttyInput := make(chan string, 10)
	graphiteOutput := make(chan *Metric, 10)
	mqttOutput := make(chan *Metric, 10)

	go read_from_tty(sif, ttyInput)
	go send_to_graphite(graphite, graphiteOutput)
	go send_to_mqtt(mqttClient, mqttOutput)

	parse_input(ttyInput, graphiteOutput, mqttOutput)
}
