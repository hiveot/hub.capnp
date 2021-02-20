package logger

// import gateway src/gateway
// arguments: host, configFile, certFolder
func main() {
	plugin := NewLoggerPlugin()
	StartPlugin(plugin)

}
