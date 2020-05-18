package opentransport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestConnectionService_SearchWithOpts(t *testing.T) {
	srv, client, terminate := prepare()
	defer terminate()

	fixture, err := readFixture("connection_search")
	if err != nil {
		t.Errorf("Could not read fixture: %s", err)
		t.Skip("Further tests will be skipped du of missing fixture")
	}

	srv.HandleFunc("/connections", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, string(fixture))
	})

	d, _ := time.Parse("2006-01-02", "2020-04-25")
	connResult, err := client.Connection.SearchWithOpts(context.Background(), "Zürich, Sternen Oerlikon", "Paradeplatz 1, Zürich", d, &ConnOpts{})
	if err != nil {
		t.Errorf("Failed to search connection: %s", err)
	}

	// Check amount of connections returned by the api
	if got, want := len(connResult.Connections), 4; got != want {
		t.Errorf("Got %d connections from the api but want %d connections", got, want)
	}

	// check the returned struct against a static struct
	var staticResult ConnectionResult
	_ = json.Unmarshal(fixture, &staticResult)

	if got, want := connResult, &staticResult; !reflect.DeepEqual(got, want) {
		t.Errorf("The proceeded response does not equals the static fixture")
	}
}

func TestConnectionService_SearchWithZeroDate(t *testing.T) {
	_, client, terminate := prepare()
	defer terminate()

	_, err := client.Connection.Search(context.Background(), "Zürich", "Bern", time.Time{})
	if err == nil {
		t.Errorf("A zero datetime as search parameter is accepted but should not")
	}
}

func TestConnectionService_SearchVia(t *testing.T) {
	_, client, terminate := prepare()
	defer terminate()

	_, err := client.Connection.SearchVia(context.Background(), "Zürich", "Bern", time.Now(), []string{"", ""})
	if err == nil {
		t.Errorf("One or more Via stations are empty. This should not be possible.")
	}
}

func TestConnectionService_formatDate(t *testing.T) {
	// Create empty connection service. A client instance is not needed for this operation.
	connSvc := newConnectionService(nil)
	d, _ := time.Parse(time.RFC3339, "2020-04-23T14:30:00.000Z")
	connDate, connTime, err := connSvc.formatDate(d)
	if err != nil {
		t.Errorf("Failed to format the time.Time to requried date format format")
	}

	if got, want := string(connDate), "2020-04-23"; !strings.Contains(got, want) {
		t.Errorf("Got formatted connection date %s but does not fit the required format %s", got, want)
	}

	if got, want := string(connTime), "14:30"; !strings.Contains(got, want) {
		t.Errorf("Got formatted connection time %s but does not fit the required format %s", got, want)
	}
}

func TestConnectionService_buildUrlPath(t *testing.T) {
	_, client, terminate := prepare()
	defer terminate()

	connOpts := &ConnOpts{
		Via:           []string{"Zürich, Limmatplatz", "Zürich, Bahnhofstrasse"},
		Direct:        true,
		Transportations: []Transportation{Tram, Bus, Train},
		Accessibility: IndependentBoarding,
	}

	inputDate, _ := time.Parse(time.RFC3339, "2020-04-23T14:30:00.000Z")
	date, time, err := client.Connection.formatDate(inputDate)
	if err != nil {
		t.Errorf("Failed to convert input date %s to formatted date and time", inputDate)
	}

	want := "connections?from=Z%C3%BCrich%2C%20Sternen%20Oerlikon&" +
		"to=Paradeplatz%201%2C%20Z%C3%BCrich&date=2020-04-23&time=14:30&isArrivalTime=0&" +
		"direct=1&bike=0&sleeper=0&couchette=0&via[]=Z%C3%BCrich%2C%20Limmatplatz&" +
		"via[]=Z%C3%BCrich%2C%20Bahnhofstrasse&transportations[]=tram" +
		"&transportations[]=bus&transportations[]=train"

	got, err := client.Connection.buildUrlPath("Zürich, Sternen Oerlikon", "Paradeplatz 1, Zürich", date, time, connOpts)
	if !strings.Contains(got, want) {
		t.Errorf("The builded url %s path does not fit the wantet one %s", got, want)
	}
}

func TestConnectionService_parseResult(t *testing.T) {
	_, client, terminate := prepare()
	defer terminate()

	// Test valid input parameters
	testValues := []struct {
		in []byte
		out string
	}{
		{[]byte("{Invalid: Json}"), "failed to parse response"},
		{[]byte{}, "response buffer is empty"},
	}

	for _, v := range testValues {
		_, err := client.Connection.parseResponse(v.in)
		if err == nil {
			t.Errorf("The response parser should provide an error message but was empty")
		}
		if got, want := err.Error(), v.out; !strings.Contains(got, want) {
			t.Errorf("The response parser got error message '%s' but want '%s'", got, want)
		}
	}
}

func TestConnectionService_queryFailed(t *testing.T) {
	srv, client, terminate := prepare()
	defer terminate()

	srv.HandleFunc("/connections", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	// Test valid input parameters
	testValues := []struct {
		in   string
		want string
	}{
		{"", "request path can not be empty"},
		{"connections", "failed to perform the http"},
	}

	for _, v := range testValues {
		_, err := client.Connection.query(context.Background(), v.in)
		if err == nil {
			t.Errorf("The location query should return an error when the url path is %s", v.in)
		} else {
			if !strings.Contains(err.Error(), v.want) {
				t.Errorf("The connection query returned an unexpected error message")
			}
		}
	}
}