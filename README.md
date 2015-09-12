# About

You can use this to send data from OneWire sensors to graphite (https://github.com/graphite-project)

# Configuration

An example configuration is included

Basically you need to configure the receiver, the collector and map the names. The receiver receives
data from the serial (USB), the collector is the part that actually stores the data, and the mapping
converts the ids to 'human' names.

1. receiver

	This is basically the configuration for the serial port. The given example is probably sufficient,
	you might want to doublecheck the USB port number. Most parameters are useless at the moment.

2. collector

	For the time being, this is highly oriented to graphite.

3. mapping

	Here you map any the unique hex id to a graphite (sub)path.

	The unit will send a heartbeat every cycle, which will have '0000XX0000000001' as id (XX is it's
	unique id programmed via the firmware).

	Every OneWire sensor will be transmitted over via the OneWire unique id; eg. for DS18B20 temperature
	sensors, the id starts with '28'.

	The data should be sent over as is to minimize power consumption on the sensors, and is processed
	by this daemon.

At some point, more collector options will be added.
