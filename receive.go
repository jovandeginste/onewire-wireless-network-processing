package main

import (
	"github.com/marpaia/graphite-golang"
	"io/ioutil"
	"os"
	//"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"log"
)

type Config struct {
	Receiver struct {
		Port_str  string
		Baud_rate int
		Data_bits int
		Stop_bits int
		Parity    int
	}
	Collector struct {
		Type          string
		Configuration struct {
			Host   string
			Port   int
			Prefix string
		}
	}
	Name_mapping map[string]string
}

func main() {
	filename := os.Args[1]
	var config Config

	data, _ := ioutil.ReadFile(filename)
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		panic(err)
	}

	// try to connect a graphite server
	Graphite, err := graphite.NewGraphite(config.Collector.Configuration.Host, config.Collector.Configuration.Port)

	// if you couldn't connect to graphite, use a nop
	if err != nil {
		Graphite = graphite.NewGraphiteNop(config.Collector.Configuration.Host, config.Collector.Configuration.Port)
	}

	log.Printf("Value: %#v\n", config.Name_mapping["00000d0000000001"])

	log.Printf("Loaded Graphite connection: %#v", Graphite)
	Graphite.SimpleSend("stats.graphite_loaded", "1")
}
