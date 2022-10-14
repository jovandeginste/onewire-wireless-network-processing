package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type Metric struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value float64 `json:"value"`
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
			pValue float64
		)

		if strings.HasPrefix(id, "0000") {
			pType, pValue = payload_node(payload)
		} else if strings.HasPrefix(id, "28") {
			pType, pValue = payload_ds18b20(payload)
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

func payload_node(payload []int) (string, float64) {
	payload_type_int := payload[0]
	payload_type := "unknown"

	if payload_type_int == 1 {
		payload_type = "heartbeat"
	}

	payload_value := 0

	for i := len(payload) - 1; i >= 1; i-- {
		payload_value = payload_value<<8 + payload[i]
	}

	return payload_type, float64(payload_value)
}

func payload_ds18b20(payload []int) (string, float64) {
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
