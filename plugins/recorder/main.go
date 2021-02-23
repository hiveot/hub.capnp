package main

import (
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/plugins/recorder/internal"
)

func main() {
	internal.StartRecorder("")
	lib.WaitForSignal()
	internal.StopRecorder()
}
