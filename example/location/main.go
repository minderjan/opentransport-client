// The command location demonstrates the usage of a basic location search.
package main

import (
	"context"
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"os"
)

func main() {

	// Location name to search for
	loc := "Zürich Bürkliplatz"

	// Crate new opentransport client
	client := opentransport.NewClient()
	locations, err := client.Location.Search(context.Background(), loc)
	if err != nil {
		fmt.Printf("Could not search for %s: %s", loc, err)
		os.Exit(1)
	}

	for _, l := range locations {
		if l.Station() {
			fmt.Printf("Station: %s\n", l.Name)
		} else {
			fmt.Printf("Address: %s\n", l.Name)
		}
	}

}
