package main

import (
	. "github.com/goldenspider/log4go"
)

func main() {
	//default use config.toml from current dir
	//console and file
	StartLogServer("config.toml")
	defer StopLogServer()

	LogInfof("This is good start. %s", "Yes")
	LogWarn("Are you ready now? ", "OK")
}
