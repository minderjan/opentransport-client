// The command stationboard-extended demonstrates the usage of a advanced stationboard queries.
package main

import (
	"context"
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"os"
	"time"
)

func main() {

	station := "Zürich, Schmiede Wiedikon"
	when := time.Now()

	// Advanced filter options
	stbOpts := opentransport.StbOpts{
		Transportations: []opentransport.Transportation{opentransport.Bus},
		DateTime:        when,
		Arrival:         true,
		Limit:           2,
	}

	// Crate a new opentransport client
	client := opentransport.NewClient()

	// Enable logs to stdout
	client.EnableLogs(nil)

	// Define a timeout of 40 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*40)
	defer cancel()

	// Search a stationbard based on the advanced filters
	result, err := client.Stationboard.SearchWithOpts(ctx, station, stbOpts)
	if err != nil {
		fmt.Printf("Could not get Stationboard for %s: %s", station, err)
		os.Exit(1)
	}

	// The search returns a list of journeys with the found connections in it
	for _, j := range result.Journeys {
		fmt.Printf("Arrival at %s (%s)%s to %s\n",
			j.Stop.Departure.Time.Format("15:04"),
			j.Category,
			j.Number,
			j.To)
	}

}
