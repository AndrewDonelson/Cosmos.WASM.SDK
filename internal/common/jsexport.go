package common

import (
	"fmt"
	"syscall/js"
)

func jsSKDVersion(this js.Value, args []js.Value) interface{} {
	return SDKVersion("GWF-SDK")
}

func ExportJS() {
	// This is a channel that we will use to block the main thread
	c := make(chan bool)

	// Register the functions to be called from JS
	js.Global().Set("SDKVersion", js.FuncOf(jsSKDVersion))

	fmt.Println(SDKVersion("COSM-SDK"), "ready")

	// Block the main thread forever
	<-c

	// This is a hack to prevent the program from exiting
	fmt.Println(SDKVersion("COSM-SDK"), "exited")
}
