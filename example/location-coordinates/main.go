// The command location-coordinates demonstrates the usage of location search by coordinates.
package main

import (
	"context"
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"os"
)

func main() {

	// Coordinates of the location
	lat := 47.377847
	long := 8.540502

	// Crate a new opentransport client
	client := opentransport.NewClient()

	// Search for locations with coordinates (lat / long)
	locations, err := client.Location.SearchWithCoordinates(context.Background(), lat, long)
	if err != nil {
		fmt.Printf("Could not search location for coordinates: %f / %f: %s", lat, long, err)
		os.Exit(1)
	}

	// The search returns a list of matching locations
	for _, l := range locations {
		if l.Station() {
			fmt.Printf("Station: %s\n", l.Name)
		} else {
			fmt.Printf("Address: %s\n", l.Name)
		}
	}

}