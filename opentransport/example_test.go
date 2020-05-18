package opentransport

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ExampleNewClient() {

	// Create new default client
	client := NewClient()

	// Default output os.Stdout and os.Stderr
	client.EnableLogs(nil)

	// Use a service method
	stations, err := client.Location.Search(context.Background(), "Zürich")

	// Check if the query failed and get an error
	if err != nil {
		fmt.Printf("Failed to search location because of %s", err)
	}

	// Print the returned array of stations
	fmt.Printf("All returend stations %+v", stations)
}

func ExampleNewClientWithUrl() {

	// Create new default client
	client, err := NewClientWithUrl(&http.Client{}, "http://localhost:8080/v1/")

	// Check if the client could be created
	if err != nil {
		fmt.Printf("Failed to create a new client because of %s", err)
	}

	// Create a logfile
	f, err := os.OpenFile("opentransport.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)

	// Check if the client could be created
	if err != nil {
		fmt.Printf("Could not access the logfile %s", err)
	}

	// Create new multi writer
	multi := io.MultiWriter(f, os.Stdout)

	// Default output os.Stdout and os.Stderr
	client.EnableLogs(multi)

	// Use a service method
	stations, err := client.Location.Search(context.Background(), "Zürich")

	// Print the returned array of stations
	fmt.Printf("All returend stations %+v", stations)
}

func ExampleLocationService_SearchWithCoordinates() {
	// Create new OpenTransport Client
	client := NewClient()

	// Query a location by lat / lon coordinates
	locations, err := client.Location.SearchWithCoordinates(context.Background(), 47.368281, 8.537556)
	if err != nil {
		fmt.Printf("failed to search for location: %s", err)
	}

	// Print locations to stdout
	for i, s := range locations {
		fmt.Printf("Location[%d]: %s\n",i, s.Name)
	}
}

func ExampleLocationService_SearchWithType() {
	// Create new OpenTransport Client
	client := NewClient()

	// Query a location by lat / lon coordinates
	locations, err := client.Location.SearchWithType(context.Background(), "Zürich Bürkliplatz", TypePoi)
	if err != nil {
		fmt.Printf("failed to search for location: %s", err)
	}

	// Print locations to stdout
	for i, s := range locations {
		fmt.Printf("Location[%d]: %s\n",i, s.Name)
	}
}