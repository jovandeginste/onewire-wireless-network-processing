package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tarm/serial"
)

type Metric struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

func newTTYReceiver() *serial.Port {
	portStr := cfg.Receiver.PortStr
	baudRate := cfg.Receiver.BaudRate

	if err := resetTTY(portStr, baudRate); err != nil {
		log.Fatal("An error has occurred while resetting tty:", err)
		os.Exit(1)
	}

	sif, err := serial.OpenPort(&serial.Config{Name: portStr, Baud: baudRate})
	if err != nil {
		log.Fatal("An error has occurred while trying to open the tty:", err)
		os.Exit(1)
	}

	return sif
}

func resetTTY(portStr string, baudRate int) error {
	binary, err := exec.LookPath("stty")
	if err != nil {
		return err
	}

	args := []string{"-F", portStr, strconv.Itoa(baudRate), "-hup", "raw", "-echo"}

	if _, err = exec.Command(binary, args...).Output(); err != nil {
		return err
	}

	return nil
}

func readFromTTY(sif io.Reader, ttyInput chan string) error {
	var message string
	var err error
	reader := bufio.NewReader(sif)

	for {
		message, err = reader.ReadString('\n')
		if err != nil {
			return err
		}

		ttyInput <- strings.TrimSpace(message)
	}
}

func parseInput(input chan string, outputs ...chan *Metric) {
	for {
		message := <-input

		log.Println(message)

		data := strings.Split(message, " ")
		/****
		Following information is calculated by tty but not used
		status := data[0]
		timestamp := data[1]
		****/
		id, _ := stringsToIntegerHexes(data[2:10])
		payload, _ := stringsToIntegers(data[10:])
		name := idToName(id)

		var (
			pType  string
			pValue float64
		)

		if strings.HasPrefix(id, "0000") {
			pType, pValue = payloadNode(payload)
		} else if strings.HasPrefix(id, "28") {
			pType, pValue = payloadDS18B20(payload)
		} else {
			pType = "unknown"
			pValue = 0
		}

		m := Metric{Name: name, Type: pType, Value: pValue}

		for _, o := range outputs {
			o <- &m
		}
	}
}

func stringsToIntegers(integerStrings []string) ([]int, error) {
	ints := []int{}

	for _, i := range integerStrings {
		j, err := strconv.Atoi(i)
		if err != nil {
			return ints, err
		}

		ints = append(ints, j)
	}

	return ints, nil
}

func stringsToIntegerHexes(integerStrings []string) (string, error) {
	var buffer bytes.Buffer

	for _, i := range integerStrings {
		j, err := strconv.Atoi(i)
		if err != nil {
			return buffer.String(), err
		}

		buffer.WriteString(fmt.Sprintf("%02x", j))
	}

	return buffer.String(), nil
}

func payloadNode(payload []int) (string, float64) {
	payloadTypeInt := payload[0]
	payloadType := "unknown"

	if payloadTypeInt == 1 {
		payloadType = "heartbeat"
	}

	payloadValue := 0

	for i := len(payload) - 1; i >= 1; i-- {
		payloadValue = payloadValue<<8 + payload[i]
	}

	return payloadType, float64(payloadValue)
}

func payloadDS18B20(payload []int) (string, float64) {
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

	temp := float64(s*t) / 16.0

	return "temperature", temp
}
