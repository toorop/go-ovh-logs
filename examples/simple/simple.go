package main

import (
	ovhlogs "github.com/toorop/go-ovh-logs"
)

func main() {
	logs := ovhlogs.New("STREAM_TOKEN", ovhlogs.GelfUDP, ovhlogs.CompressNone, false)

	entry := ovhlogs.Entry{
		Host:        "localhost",
		FullMessage: "helo world",
		Level:       6,
	}
	logs.Send(entry)
}
