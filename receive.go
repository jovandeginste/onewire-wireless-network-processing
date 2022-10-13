package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/marpaia/graphite-golang"
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("TOPIC: %s\n", msg.Topic())
	log.Printf("MSG: %s\n", msg.Payload())
}

type config struct {
	Receiver struct {
		PortStr  string `yaml:"port_str"`
		BaudRate int    `yaml:"baud_rate"`
		DataBits int    `yaml:"data_bits"`
		StopBits int    `yaml:"stop_bits"`
		Parity   int    `yaml:"parity"`
	} `yaml:"receiver"`
	Collector struct {
		Type          string `yaml:"type"`
		Configuration struct {
			Host   string `yaml:"host"`
			Port   int    `yaml:"port"`
			Prefix string `yaml:"prefix"`
		}
	}
	MQTT struct {
		Host        string `yaml:"host"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		TopicPrefix string `yaml:"topic_prefix"`
	}
	NameMapping map[string]string `yaml:"name_mapping"`
}

type Metric struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

var cfg config

func read_configuration(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &cfg)
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
		return err
	}

	return nil
}

func read_from_tty(sif io.Reader, tty_input chan string) error {
	var message string
	var err error
	reader := bufio.NewReader(sif)

	for {
		message, err = reader.ReadString('\n')
		if err != nil {
			return err
		}

		tty_input <- strings.TrimSpace(message)
	}
}

func main() {
	if err := read_configuration(os.Args[1]); err != nil {
		log.Fatal("An error has occurred while read configuration file:", err)
		os.Exit(1)
	}

	mqtt.DEBUG = log.New(os.Stdout, "", 0)
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
	graphite, err := graphite.NewGraphite(cfg.Collector.Configuration.Host, cfg.Collector.Configuration.Port)
	if err != nil {
		log.Fatal("An error has occurred while trying to create a Graphite connector:", err)
		os.Exit(1)
	}

	graphite.Prefix = cfg.Collector.Configuration.Prefix

	log.Printf("Loaded Graphite connection: %#v", graphite)

	ttyInput := make(chan string, 10)
	graphiteOutput := make(chan *Metric, 10)
	mqttOutput := make(chan *Metric, 10)

	go read_from_tty(sif, ttyInput)
	go send_to_graphite(graphite, graphiteOutput)
	go send_to_mqtt(mqttClient, mqttOutput)

	parse_input(ttyInput, graphiteOutput, mqttOutput)
}

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
	return cfg.MQTT.TopicPrefix + "/" + m.Name
}

func (m *Metric) MQTTValue() string {
	u, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(u)
}

func send_to_graphite(graphite *graphite.Graphite, input chan *Metric) {
	for {
		message := <-input

		log.Printf("Graphite Sending to '%s': %#v", message.GraphiteName(), message)

		if err := graphite.Connect(); err != nil {
			log.Println(err)
			continue
		}

		graphite.SimpleSend(message.GraphiteName(), message.Value)

		if err := graphite.Disconnect(); err != nil {
			log.Println(err)
		}
	}
}

func id_to_name(id string) string {
	return cfg.NameMapping[id]
}

func parse_input(input chan string, outputs ...chan *Metric) {
	for {
		message := <-input

		log.Println(message)

		data := strings.Split(message, " ")
		/****
		Following information is calculated by tty but not used
		status := data[0]
		timestamp := data[1]
		****/
		id, _ := integer_strings_to_hexstring(data[2:10])
		payload, _ := integer_strings_to_integers(data[10:])
		name := id_to_name(id)

		var (
			pType  string
			pValue string
		)

		if strings.HasPrefix(id, "0000") {
			pType, pValue = payload_node(payload)
		} else if strings.HasPrefix(id, "28") {
			pType, pValue = payload_ds18b20(payload)
		} else {
			pType = "unknown"
			pValue = "-"
		}

		m := Metric{Name: name, Type: pType, Value: pValue}

		for _, o := range outputs {
			o <- &m
		}
	}
}

func (m *Metric) GraphiteName() string {
	return strings.Join([]string{m.Name, m.Type, "value"}, ".")
}

func integer_strings_to_integers(integer_strings []string) ([]int, error) {
	ints := []int{}

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
	payload_type := "unknown"

	if payload_type_int == 1 {
		payload_type = "heartbeat"
	}

	payload_value := 0

	for i := len(payload) - 1; i >= 1; i-- {
		payload_value = payload_value<<8 + payload[i]
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

func RoundN(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor((f*shift)+.5) / shift
}
