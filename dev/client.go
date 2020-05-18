// The dev-client is a helper application for developers to debug their code.
package main

import (
	"context"
	"fmt"
	"github.com/minderjan/opentransport-client/opentransport"
	"io"
	"os"
	"time"
)

func main() {

	// Eg. local version of transport.opendata.ch or a mocked server
	apiUrl := "http://localhost:3001/v1/"

	// Create client against custom remote api
	client, err := opentransport.NewClientWithUrl(nil, apiUrl)
	if err != nil {
		fmt.Printf("Could not create the opentransport client with url base url %s: %s",  apiUrl, err)
		os.Exit(1)
	}

	// Enable logs to stdout
	client.EnableLogs(nil)

	// logsWithFile(client)

	// Change User Agent
	client.UserAgent("OpenTransport Development")

	// Configure error behaviour
	httpRetryOptions(client)

	// use one of the following functions to test parts of the library
	// ----------------------------------------------------------------
	// locationWithName(client)
	// locationWithCoordinates(client)
	// locationWithType(client)
	//
	// connectionWithName(client)
	// connectionWithVia(client)
	// connectionWithOpts(client)
	//
	// stationboardWithName(client)
	// stationboardWithDate(client)
	// stationboardWithType(client)
	// stationboardWithOpts(client)

}

// Change retry configs. These will be used if the server resonpds with an > http 500 or a go error.
func httpRetryOptions(client *opentransport.Client) {

	// If an http error occur, the client try
	err := client.MaxRetry(2, 1)
	if err != nil {
		fmt.Printf("Failed to set max retry options: %s", err)
	}

	// Define Timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	// Try to get locations
	locations, err := client.Location.Search(ctx, "Zürich")
	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	for _, l := range locations {
		fmt.Printf("%s - %s\n", l.Name, l.Icon)
	}

}

// Logs will be written to a logfile
func logsWithFile(client *opentransport.Client) {
	// Create a logfile
	f, err := os.OpenFile("opentransport.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	// Check if the client could be created
	if err != nil {
		fmt.Printf("Could not access the logfile %s", err)
	}

	// Create new multi writer
	multi := io.MultiWriter(f, os.Stdout)

	// Default output os.Stdout and os.Stderr
	client.EnableLogs(multi)
}

// Search a location by its name
func locationWithName(client *opentransport.Client) {
	loc := "Zürich Bürkliplatz"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	locations, err := client.Location.Search(ctx, loc)
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

// Search a location by coordinates
func locationWithCoordinates(client *opentransport.Client) {
	// Coordinates of the location
	lat := 47.377847
	long := 8.540502

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

// Search a location by type. Currently not supported by the API.
func locationWithType(client *opentransport.Client) {
	loc := "Zürich Bürkliplatz"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	locations, err := client.Location.SearchWithType(ctx, loc, opentransport.TypePoi)
	if err != nil {
		fmt.Printf("Could not search location for %s: %s", loc, err)
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

// Search a connection with from, to and time
func connectionWithName(client *opentransport.Client) {
	from := "Zürich HB"
	to := "Bern"
	when := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	result, err := client.Connection.Search(ctx, from, to, when)
	if err != nil {
		fmt.Printf("Failed to search connection from %s to %s at %s: %s",  from, to, when.Format("2006-01-02 15:04"), err)
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

// Search a connection with from, to, time and via
func connectionWithVia(client *opentransport.Client) {
	from := "Zürich HB"
	to := "Bern"
	via := []string{"Aarau"}
	when := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	result, err := client.Connection.SearchVia(ctx, from, to, when, via)
	if err != nil {
		fmt.Printf("Failed to search connection from %s to %s at %s: %s",  from, to, when.Format("2006-01-02 15:04"), err)
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

// Search a connection with filter options
func connectionWithOpts(client *opentransport.Client) {
	from := "Zürich HB"
	to := "Bern"
	when := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// Advanced filter options
	reqOpts := opentransport.ConnOpts{
		Via:             []string{"Aarau"},
		Direct:          false,
		Transportations: []opentransport.Transportation{opentransport.Train},
		Accessibility:   opentransport.IndependentBoarding,
		Limit:           2,
	}
	result, err := client.Connection.SearchWithOpts(ctx, from, to, when, &reqOpts)
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

// Search a stationboard only with the name of a station
func stationboardWithName(client *opentransport.Client) {
	station := "Zürich HB"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	result, err := client.Stationboard.Search(ctx, station)
	if err != nil {
		fmt.Printf("Could not get Stationboard for %s: %s", station, err)
		os.Exit(1)
	}

	for _, j := range result.Journeys {
		fmt.Printf("Departure at %s (%s)%s to %s\n",
			j.Stop.Departure.Time.Format("15:04"),
			j.Category,
			j.Number,
			j.To)
	}
}

// Search stationboard with a date parameter
func stationboardWithDate(client *opentransport.Client) {
	station := "Zürich HB"
	when := time.Now().Add(20 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	result, err := client.Stationboard.SearchWithDate(ctx, station, when)
	if err != nil {
		fmt.Printf("Could not get Stationboard for %s: %s", station, err)
		os.Exit(1)
	}

	for _, j := range result.Journeys {
		fmt.Printf("Departure at %s (%s)%s to %s\n",
			j.Stop.Departure.Time.Format("15:04"),
			j.Category,
			j.Number,
			j.To)
	}
}

// Search stationboard with a type filter
func stationboardWithType(client *opentransport.Client) {
	station := "Zürich HB"
	when := time.Now().Add(20 * time.Minute)
	transports := []opentransport.Transportation{opentransport.Tram, opentransport.Bus}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	result, err := client.Stationboard.SearchWithType(ctx, station, when, transports)
	if err != nil {
		fmt.Printf("Could not get Stationboard for %s: %s", station, err)
		os.Exit(1)
	}

	for _, j := range result.Journeys {
		fmt.Printf("Departure at %s (%s)%s to %s\n",
			j.Stop.Departure.Time.Format("15:04"),
			j.Category,
			j.Number,
			j.To)
	}
}

// Search stationboard with a date parameter
func stationboardWithOpts(client *opentransport.Client) {
	station := "Zürich, Schmiede Wiedikon"
	when := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// Advanced filter options
	stbOpts := opentransport.StbOpts{
		Transportations: []opentransport.Transportation{opentransport.Bus},
		DateTime:        when,
		Arrival:         true,
		Limit:           2,
	}
	result, err := client.Stationboard.SearchWithOpts(ctx, station, stbOpts)
	if err != nil {
		fmt.Printf("Could not get Stationboard for %s: %s", station, err)
		os.Exit(1)
	}

	for _, j := range result.Journeys {
		fmt.Printf("Arrival at %s (%s)%s to %s\n",
			j.Stop.Departure.Time.Format("15:04"),
			j.Category,
			j.Number,
			j.To)
	}
}
