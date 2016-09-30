package main

import (
	. "github.com/goldenspider/log4go"
)

func test() string {
	return "good luck"
}

func main() {
	//default use config.toml from current dir
	//console and file
	//StartLogServer("config.toml")

	StartLogServer()
	defer StopLogServer()

	LogInfof("This is good start. %s", test())
	LogWarn(test(), " Are you ready now? ", "OK")
}
