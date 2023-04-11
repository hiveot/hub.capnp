//go:build js

package main

import (
	"github.com/hiveot/hub/cmd/wasm/wsjs"
	"syscall/js"
)

var wait = make(chan bool)

func Gostop(this js.Value, args []js.Value) any {
	go func() {
		wait <- true
		println("Stopping Gomain")
	}()
	return true
}

// hapi main
func main() {
	// logrus.SetReportCaller(true)
	println("Entering Go main")
	hapi := wsjs.NewHubAPI()
	// Register the Go Gateway API for use by JS
	js.Global().Set("connect", js.FuncOf(hapi.Connect))
	js.Global().Set("login", js.FuncOf(hapi.Login))
	js.Global().Set("pubEvent", js.FuncOf(hapi.PubEvent))
	js.Global().Set("subActions", js.FuncOf(hapi.SubActions))
	js.Global().Set("gostop", js.FuncOf(Gostop))
	//time.Sleep(time.Millisecond * 100)

	// Prevent the program from exit
	<-wait
	println("Gomain has stopped")
}
