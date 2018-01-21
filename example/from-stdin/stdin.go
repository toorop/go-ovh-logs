// This example will send the first line read from stdin to OVH logs
// ex: echo "hello OVH logs" | fromSTDIN

package main

import (
	"bufio"
	"fmt"
	"os"

	ovhlogs "github.com/toorop/go-ovh-logs"
)

func main() {
	// get token from ENV
	token := os.Getenv("OVHLOGS_TOKEN")
	if token == "" {
		fmt.Println("OVHLOGS_TOKEN env var not found")
		os.Exit(1)
	}
	// init logger
	ovhlogger := ovhlogs.New(token, ovhlogs.GelfUDP, ovhlogs.CompressNone, false)

	// read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	// Send log
	ovhlogger.Print(scanner.Text())
}
