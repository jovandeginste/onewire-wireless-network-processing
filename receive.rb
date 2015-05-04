#!/usr/bin/ruby

DEBUG = false
VERBOSE = true

require "serialport"
require "yaml"
require 'graphite-api'

configuration = YAML.load(File.read("config.yaml"))

config_receiver = configuration[:receiver] || {}
config_collector = configuration[:collector] || {}
config_collector_type = config_collector[:type] || "none"
config_collector_config = config_collector[:configuration] || {}
config_mapping = configuration[:name_mapping] || {}

prefix = config_collector_config[:prefix]
port_str = config_receiver[:port_str]
baud_rate = config_receiver[:baud_rate]
data_bits = config_receiver[:data_bits]
stop_bits = config_receiver[:stop_bits]
parity = config_receiver[:parity]

puts "Receiver configuration: #{config_receiver.inspect}" if DEBUG
command = ["stty -F", port_str, baud_rate, "-hup raw -echo"].join(" ")
system(command)
sp = SerialPort.new(port_str, baud_rate, data_bits, stop_bits, parity)

puts "Collector configuration: #{config_collector_config.inspect}" if DEBUG
client = GraphiteAPI.new(config_collector_config)

$stderr.puts "OneWire JeeLink initialized, now waiting for input..." if VERBOSE

#just read forever
while true do
	while (i = sp.gets.gsub(/\0/, "").chomp) do

		data = i.split(" ")
		result = {
			status: data.shift,
			timestamp: data.shift,
			id: data.shift(8).map{|i| "%02x" % i.to_i},
			data: data.map(&:to_i),
		}
		result[:name] = config_mapping[result[:id].join] || result[:id].join

		case result[:id].join(":")
		when /^00:00:/
			data = Array.new(result[:data])
			type = data.shift

			result[:type] = case type
					when 1
						"heartbeat"
					else
						"unknown"
					end

			result[:value] = data.inject([0, 0]){|r, i|
				e, index = r
				e = e + i * 2**index
				index += 8
				[e, index]
			}.first
		when /^28:/
			low, high = result[:data].first(2)

			high = high << 8
			t = high + low

			sign = t & 32768

			if sign == 0
				s = 1
			else
				s = -1
				t = (t ^ 65535) + 1
			end
			t = (s * t / 16.0).round(1)

			result[:value] = t
			result[:type] = "temperature"
		end

		$stderr.puts result.inspect if VERBOSE
		if result[:type] && result[:value]
			client.metrics(
				[result[:name], result[:type], "value"].join(".")  => result[:value],
			)
			$stdout.flush
		end
	end
end

sp.close
