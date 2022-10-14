package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/marpaia/graphite-golang"
)

func send_to_graphite(graphite *graphite.Graphite, input chan *Metric) {
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
