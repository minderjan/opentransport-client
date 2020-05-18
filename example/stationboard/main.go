// The command stationboard demonstrates the usage of a basic stationboard query.
package main

import (
	"context"
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"os"
)

func main() {

	station := "Zürich HB"

	// Crate a new opentransport client
	client := opentransport.NewClient()

	// Search for a stationboard of the given station
	result, err := client.Stationboard.Search(context.Background(), station)
	if err != nil {
		fmt.Printf("Could not get Stationboard for %s: %s", station, err)
		os.Exit(1)
	}

	// The stationboard contains a journey object. This holds all connection data.
	for _, j := range result.Journeys {
		fmt.Printf("Departure at %s (%s)%s to %s\n",
			j.Stop.Departure.Time.Format("15:04"),
			j.Category,
			j.Number,
			j.To)
	}

}
