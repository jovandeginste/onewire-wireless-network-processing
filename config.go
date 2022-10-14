package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	Receiver struct {
		PortStr  string `yaml:"port_str"`
		BaudRate int    `yaml:"baud_rate"`
		DataBits int    `yaml:"data_bits"`
		StopBits int    `yaml:"stop_bits"`
		Parity   int    `yaml:"parity"`
	} `yaml:"receiver"`
	Graphite struct {
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

func id_to_name(id string) string {
	return cfg.NameMapping[id]
}
