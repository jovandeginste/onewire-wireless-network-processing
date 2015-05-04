You can use this to send data to graphite (https://github.com/graphite-project)

Steps needed:

1. complete the config.yaml
   
   An example is included

   Basically you need to configure the receiver, the collector and map the names. The receiver receives
   data from the serial (USB), the collector is the part that actually stores the data, and the mapping
   converts the ids to 'human' names.

   1. receiver

      This is basically the configuration for the serial port. The given example is probably sufficient,
      you might want to doublecheck the USB port number.

   2. collector

      For the time being, this is highly oriented to graphite. This is the configuration as is passed to
      the initializer of the graphite API gem (https://github.com/kontera-technologies/graphite-api).
      It's fairly straight forward.

   3. mapping

      Here you map any the unique hex id to a graphite (sub)path.

      The unit will send a heartbeat every cycle, which will have '0000XX0000000001' as id (XX is it's
      unique id programmed via the firmware).

      Every OneWire sensor will be transmitted over via the OneWire unique id; eg. for DS18B20 temperature
      sensors, the id starts with '28'.

      The data should be sent over as is to minimize power consumption on the sensors, and is processed
      by the ruby script here.

2. install the ruby dependencies - either or not using rvm
   
   ```bash
   bundle install
   ```

3. try the script
   
   ```bash
   ruby receive.rb
   ```

At some point, more collector options will be added.
