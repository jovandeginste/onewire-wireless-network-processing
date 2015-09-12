package main

import (
	"bufio"
	"github.com/marpaia/graphite-golang"
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
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

func read_config(filename string) (config Config, err error) {
	data, _ := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = yaml.Unmarshal([]byte(data), &config)
	return
}

func reset_tty(port_str string, baud_rate int) (err error) {
	binary, err := exec.LookPath("stty")
	if err != nil {
		return
	}
	args := []string{"-F", port_str, strconv.Itoa(baud_rate), "-hup", "raw", "-echo"}

	_, err = exec.Command(binary, args...).Output()

	if err != nil {
		log.Println("error occured")
		log.Printf("%s", err)
		return
	}
	return
}

func main() {
	config, err := read_config(os.Args[1])

	port_str := config.Receiver.Port_str
	baud_rate := config.Receiver.Baud_rate

	reset_tty(port_str, baud_rate)

	sif, err := serial.OpenPort(&serial.Config{Name: port_str, Baud: baud_rate})
	if err != nil {
		log.Fatal(err)
	}

	// try to connect a graphite server
	Graphite, err := graphite.NewGraphite(config.Collector.Configuration.Host, config.Collector.Configuration.Port)
	Graphite.Prefix = config.Collector.Configuration.Prefix

	log.Printf("Value: %#v\n", config.Name_mapping["00000d0000000001"])

	log.Printf("Loaded Graphite connection: %#v", Graphite)
	Graphite.SimpleSend("stats.graphite_loaded", "1")

	reader := bufio.NewReader(sif)
	reply, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	log.Println(reply)
}
