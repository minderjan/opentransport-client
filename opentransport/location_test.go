package opentransport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestLocationService_SearchWithType(t *testing.T) {
	srv, client, terminate := prepare()
	defer terminate()

	fixture, err := readFixture("location_search")
	if err != nil {
		t.Error(err)
	}

	srv.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, string(fixture))
	})

	// Search with Type
	typedLocations, typedErr := client.Location.SearchWithType(context.Background(), "Zürich", TypeAll)
	if err != nil {
		t.Errorf("Failed to search a location by name")
	}

	// Search
	locations, err := client.Location.Search(context.Background(), "Zürich")
	if err != nil {
		t.Errorf("Failed to search a location by name")
	}

	// Test values
	testValues := []struct {
		locations []Location
		err       error
	}{
		{typedLocations, typedErr},
		{locations, err},
	}

	// Check the returned struct against a static struct
	var staticResult LocationResult
	_ = json.Unmarshal(fixture, &staticResult)

	for _, tv := range testValues {
		// Check amount of locations returned by the api
		if got, want := len(tv.locations), 10; got != want {
			t.Errorf("Got %d locations from the api but want %d locations", got, want)
		}

		// Compare the expected and returned struct
		if got, want := tv.locations, staticResult.Stations; !reflect.DeepEqual(got, want) {
			t.Errorf("The proceeded response does not equals the static fixture")
		}
	}
}

func TestLocationService_SearchWithCoordinates(t *testing.T) {
	srv, client, terminate := prepare()
	defer terminate()

	fixture, err := readFixture("location_search_coordinates")
	if err != nil {
		t.Error(err)
	}

	srv.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, string(fixture))
	})

	// Search for a location by coordinates
	locations, err := client.Location.SearchWithCoordinates(context.Background(), 47.476001, 8.306130)
	if err != nil {
		t.Errorf("Failed to search a location by coordinates: %s", err)
	}

	if got, want := len(locations), 10; got != want {
		t.Errorf("No location objects were returned")
	}

	// check the returned struct against a static struct
	var staticResult LocationResult
	_ = json.Unmarshal(fixture, &staticResult)

	if got, want := locations, staticResult.Stations; !reflect.DeepEqual(got, want) {
		t.Errorf("The proceeded response does not equals the static fixture")
	}
}

func TestLocationService_SearchWithCoordinates_API_Error(t *testing.T) {
	srv, client, terminate := prepare()
	defer terminate()

	srv.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})

	// Search for a location by coordinates
	_, err := client.Location.SearchWithCoordinates(context.Background(), 47.476001, 8.306130)
	if err == nil {
		t.Errorf("The method should return an error")
	}
}

func TestLocationService_queryFailed(t *testing.T) {
	srv, client, terminate := prepare()
	defer terminate()

	srv.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	// Test valid input parameters
	testValues := []struct {
		in   string
		want string
	}{
		{"", "request path can not be empty"},
		{"locations", "failed to perform the http request"},
	}

	for _, v := range testValues {
		_, err := client.Location.query(context.Background(), v.in)
		if err == nil {
			t.Errorf("The location query should return an error when the url path is %s", v.in)
		} else {
			if !strings.Contains(err.Error(), v.want) {
				t.Errorf("The location query returned an unexpected error message")
			}
		}
	}
}

func TestLocation_Station(t *testing.T) {
	address := Location{
		Id:    "",
		Name:  "China Garden, Zermatt, Bahnhofstr. 18",
		Score: 0,
		Coordinate: Coordinate{
			Type: "WGS84",
			X:    0,
			Y:    0,
		},
		Distance: 0,
		Icon:     "",
	}

	if got := address.Station(); got != false {
		t.Errorf("The location was not recognized as address")
	}

	station := Location{
		Id:    "8503000",
		Name:  "Zürich HB",
		Score: 0,
		Coordinate: Coordinate{
			Type: "WGS84",
			X:    47.377847,
			Y:    8.540502,
		},
		Distance: 0,
		Icon:     "train",
	}

	if got := station.Station(); got != true {
		t.Errorf("The location was not recognized as station")
	}
}
