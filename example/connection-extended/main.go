// The command connection-extended demonstrates the usage of an extended connection search.
package main

import (
	"context"
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"os"
	"time"
)

func main() {

	from := "Zürich HB"
	to := "Bern"
	when := time.Now()

	// Advanced filter options
	reqOpts := opentransport.ConnOpts{
		Via:             []string{"Aarau"},
		Direct:          false,
		Transportations: []opentransport.Transportation{opentransport.Train},
		Accessibility:   opentransport.IndependentBoarding,
		Limit:           2,
	}

	// Crate a new opentransport client
	client := opentransport.NewClient()
	result, err := client.Connection.SearchWithOpts(context.Background(), from, to, when, &reqOpts)

	if err != nil {
		fmt.Printf("Failed to search connection from %s to %s via %s at %s: %s",  from, to, reqOpts.Via, when.Format("2006-01-02 15:04"), err)
		os.Exit(1)
	} else {
		for _, c := range result.Connections {

			// get first part of the connection
			s := c.Sections[0]

			fmt.Printf("%s %s at %s on platform %s \n",
				s.Journey.Category,
				s.Journey.Number,
				s.Departure.Departure.Time.Format("15:04"),
				s.Departure.Platform)
		}
	}

}
