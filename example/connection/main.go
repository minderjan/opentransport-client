// The command connection demonstrates the usage of a basic connection search.
package main

import (
	"context"
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"os"
	"time"
)

func main() {

	// Prepare connection details
	from := "Zürich HB"
	to := "Bern"
	when := time.Now()

	// Crate a new opentransport client
	client := opentransport.NewClient()

	// Search for connections between {from} and {to} at {when}
	result, err := client.Connection.Search(context.Background(), from, to, when)
	if err != nil {
		fmt.Printf("Failed to search connection from %s to %s at %s: %s",  from, to, when.Format("2006-01-02 15:04"), err)
		os.Exit(1)
	} else {

		// The result contains a list of connections
		for _, c := range result.Connections {

			// First part of the connection
			s := c.Sections[0]

			fmt.Printf("%s %s at %s on platform %s \n",
				s.Journey.Category,
				s.Journey.Number,
				s.Departure.Departure.Time.Format("15:04"),
				s.Departure.Platform)
		}
	}

}