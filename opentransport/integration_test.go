// +build integration

package opentransport

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)


func TestLocationIntegration(t *testing.T) {
	client := NewClient()

	locations, err := client.Location.SearchWithType(context.Background(), "Zürich HB", TypeAll)
	if err != nil {
		t.Errorf("Failed to search a location by name")
	}

	var zrh = Location{
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

	// Check if the first location match the expected location
	if got, want := locations[0], zrh; !reflect.DeepEqual(got, want) {
		t.Errorf("The proceeded response does not equals the static fixture")
	}
}

func TestConnectionIntegration(t *testing.T) {
	client := NewClient()

	from := "Zürich HB"
	to := "Bern"
	when := time.Now()

	connRes, err := client.Connection.Search(context.Background(), from, to, when)
	if err != nil {
		t.Errorf("Failed to search a connection from %s to %s at %s", from, to, when.Format("2006-01-02 15:04"))
		t.Skipf("Further tests will be skipped because of an error during the query: %s", err)
	}

	if len(connRes.Connections) == 0 {
		t.Errorf("No connections found but expected at minimum one")
	}

	// Check if the first location match the expected location
	if got, want := connRes.Connections[0].From.Station.Name, from; !strings.Contains(got, want){
		t.Errorf("The connection returend a wrong from location: got %s but want %s", got, want)
	}

	// Check if the target location match the expected location
	if got, want := connRes.Connections[0].To.Station.Name, to; !strings.Contains(got, want){
		t.Errorf("The connection returend a wrong from location: got %s but want %s", got, want)
	}

	// Check the amount of sections between these locations
	if got, wantMin := len(connRes.Connections[0].Sections), 1; got < wantMin {
		t.Errorf("The connection returend returned less than expected sections: got %d but want minimum %d", got, wantMin)
	}
}

func TestStationboardIntegration(t *testing.T) {
	client := NewClient()

	station := "Zürich HB"
	when := time.Now()

	stbOpts := StbOpts{
		Transportations: []Transportation{Train},
		DateTime:        when,
		Arrival:         true,
		Limit:           2,
	}

	stbRes, err := client.Stationboard.SearchWithOpts(context.Background(), station, stbOpts)
	if err != nil {
		t.Errorf("Could not get Stationboard for %s: %s", station, err)
		t.Skipf("Skip stationboard integration because of a query error: %s", err)
	}

	// Check if the first location match the expected location
	if got, want := stbRes.Journeys[0].Stop.Station.Name, station; !strings.Contains(got, want){
		t.Errorf("The connection returend a wrong from location: got %s but want %s", got, want)
	}

	// Check if the connection has journeys
	if got := len(stbRes.Journeys[0].Journey.Name); got == 0{
		t.Error("The connection returend an empty Journey Name")
	}
}
