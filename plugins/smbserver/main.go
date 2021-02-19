package main

var config = msgbus.Config{}

// parse commandline, load configuration and start the service
func main() {
	args := ParseCommandline()
	err := LoadConfig(args, &config)
	StartSimpleMessageBus(&config)
	WaitForSignal()
}
