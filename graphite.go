package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/marpaia/graphite-golang"
)

func newGraphiteClient() *graphite.Graphite {
	// try to connect a graphiteClient server
	graphiteClient, err := graphite.NewGraphite(cfg.Graphite.Configuration.Host, cfg.Graphite.Configuration.Port)
	if err != nil {
		log.Fatal("An error has occurred while trying to create a Graphite connector:", err)
		os.Exit(1)
	}

	graphiteClient.Prefix = cfg.Graphite.Configuration.Prefix

	log.Printf("Loaded Graphite connection: %#v", graphiteClient)

	return graphiteClient
}

func sendGraphite(graphite *graphite.Graphite, input chan *Metric) {
	for {
		message := <-input

		log.Printf("Graphite Sending to '%s': %#v", message.GraphiteName(), message)

		if err := graphite.Connect(); err != nil {
			log.Println(err)
			continue
		}

		graphite.SimpleSend(message.GraphiteName(), message.GraphiteValue())

		if err := graphite.Disconnect(); err != nil {
			log.Println(err)
		}
	}
}

func (m *Metric) GraphiteValue() string {
	return fmt.Sprintf("%f", m.Value)
}

func (m *Metric) GraphiteName() string {
	return strings.Join([]string{m.Name, m.Type, "value"}, ".")
}
