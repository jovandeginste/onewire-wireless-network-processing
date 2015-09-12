package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/marpaia/graphite-golang"
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

type Metric struct {
	metric string
	value  string
}

var config Config

func read_config(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return err
	}

	return nil
}

func reset_tty(port_str string, baud_rate int) error {
	binary, err := exec.LookPath("stty")
	if err != nil {
		return err
	}
	args := []string{"-F", port_str, strconv.Itoa(baud_rate), "-hup", "raw", "-echo"}

	_, err = exec.Command(binary, args...).Output()

	if err != nil {
		log.Println("error occured")
		log.Printf("%s", err)
		return err
	}
	return nil
}

func read_from_tty(sif io.Reader, tty_input chan string) {
	var message string
	var err error
	reader := bufio.NewReader(sif)

	for {
		message, err = reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		tty_input <- strings.TrimSpace(message)
	}
}

func main() {
	err := read_config(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	port_str := config.Receiver.Port_str
	baud_rate := config.Receiver.Baud_rate

	err = reset_tty(port_str, baud_rate)
	if err != nil {
		log.Fatal(err)
	}

	sif, err := serial.OpenPort(&serial.Config{Name: port_str, Baud: baud_rate})
	if err != nil {
		log.Fatal(err)
	}

	// try to connect a graphite server
	graphite, err := graphite.NewGraphite(config.Collector.Configuration.Host, config.Collector.Configuration.Port)
	graphite.Prefix = config.Collector.Configuration.Prefix

	log.Printf("Loaded Graphite connection: %#v", graphite)

	tty_input := make(chan string, 10)
	graphite_output := make(chan Metric, 10)
	go read_from_tty(sif, tty_input)
	go send_to_graphite(graphite, graphite_output)
	parse_input(tty_input, graphite_output)
}

func send_to_graphite(graphite *graphite.Graphite, input chan Metric) {
	var message Metric
	for {
		message = <-input
		log.Printf("Sending: %v", message)
		graphite.SimpleSend(message.metric, message.value)
	}
}

func id_to_name(id string) string {
	return config.Name_mapping[id]
}

func parse_input(input chan string, output chan Metric) {
	var message string
	for {
		message = <-input
		log.Println(message)
		data := strings.Split(message, " ")
		status := data[0]
		timestamp := data[1]
		id, _ := integer_strings_to_hexstring(data[2:10])
		payload, _ := integer_strings_to_integers(data[10:])
		name := id_to_name(id)

		log.Println("Status:", status)
		log.Println("TS:", timestamp)
		log.Println("ID:", id)
		log.Println("Payload:", payload)

		var p_type string
		var p_value string
		if strings.HasPrefix(id, "0000") {
			p_type, p_value = payload_node(payload)
		} else if strings.HasPrefix(id, "28") {
			p_type, p_value = payload_ds18b20(payload)
		} else {
			p_type = "unknown"
			p_value = "-"
		}

		log.Println("Payload type:", p_type)
		log.Println("Payload value:", p_value)
		output <- Metric{fmt.Sprintf("%v.%v.value", name, p_type), p_value}
	}
}

func integer_strings_to_integers(integer_strings []string) ([]int, error) {
	var ints = []int{}

	for _, i := range integer_strings {
		j, err := strconv.Atoi(i)
		if err != nil {
			return ints, err
		}
		ints = append(ints, j)
	}
	return ints, nil
}

func integer_strings_to_hexstring(integer_strings []string) (string, error) {
	var buffer bytes.Buffer

	for _, i := range integer_strings {
		j, err := strconv.Atoi(i)
		if err != nil {
			return buffer.String(), err
		}
		buffer.WriteString(fmt.Sprintf("%02x", j))
	}
	return buffer.String(), nil
}

func payload_node(payload []int) (string, string) {
	payload_type_int := payload[0]

	var payload_type string
	if payload_type_int == 1 {
		payload_type = "heartbeat"
	} else {
		payload_type = "unknown"
	}

	payload_value := 0
	index := 1

	for _, i := range payload[1:] {
		payload_value = payload_value + (i * index)
		index = index * 256
	}
	return payload_type, fmt.Sprintf("%d", payload_value)
}

func payload_ds18b20(payload []int) (string, string) {
	low := payload[0]
	high := payload[1]

	high = high << 8
	t := high + low

	sign := t & 32768

	var s int
	if sign == 0 {
		s = 1
	} else {
		s = -1
		t = (t ^ 65535) + 1
	}
	temp := fmt.Sprintf("%f", RoundN(float64(s*t)/16.0, 1))

	return "temperature", temp
}

func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func RoundN(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
