// The command logging demonstrates the usage of the logging functionality of the library.
package main

import (
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"io"
	"os"
)

func main() {

	// Crate a new opentransport client
	client := opentransport.NewClient()

	// ------------- Log to Console --------------------------

	// Enable logs (will be printed to stdout)
	client.EnableLogs(nil)

	// ------------- Log to File --------------------------

	f, err := os.OpenFile("opentransport.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	// Check if the file could be opened
	if err != nil {
		fmt.Printf("Could not access the logfile %s", err)
	}

	// Create new multi writer (logs will be written to multiple locations)
	multi := io.MultiWriter(f, os.Stdout)

	// Enable advanced logging
	client.EnableLogs(multi)

}
