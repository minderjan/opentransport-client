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

func setupStationBoardTests(t *testing.T) (*Client, []byte, func()) {
	srv, client, terminate := prepare()

	fixture, err := readFixture("stationboard_search")
	if err != nil {
		t.Error(err)
	}

	srv.HandleFunc("/stationboard", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, string(fixture))
	})

	return client, fixture, terminate
}

func TestStationboardService_SearchWithOpts(t *testing.T) {
	client, fixture, terminate := setupStationBoardTests(t)
	defer terminate()

	date, _ := time.Parse("2006-01-02 15:04", "2020-05-02 20:00")
	stbOpts := StbOpts{
		Transportations: nil,
		DateTime:        date,
		Arrival:         false,
		Limit:           3,
	}
	stbResult, err := client.Stationboard.SearchWithOpts(context.Background(), "8591382", stbOpts)
	if err != nil {
		t.Errorf("Failed to search stationboard: %s", err)
		t.Skip("Skip tests because the result is not available")
	}

	// Check amount of connections returned by the api. The amount must not fit the limit because its not a hard limit.
	if got, want := len(stbResult.Journeys), 4; got != want {
		t.Errorf("Got %d connections from the api but want %d connections", got, want)
	}

	// check the returned struct against a static struct
	var staticResult StationboardResult
	_ = json.Unmarshal(fixture, &staticResult)

	if got, want := stbResult, &staticResult; !reflect.DeepEqual(got, want) {
		t.Errorf("The proceeded response does not equals the static fixture")
	}
}

func TestStationboardService_SearchWithZeroDate(t *testing.T) {
	client, _, terminate := setupStationBoardTests(t)
	defer terminate()

	_, err := client.Stationboard.SearchWithDate(context.Background(),"Zürich HB", time.Time{})
	if err == nil {
		t.Errorf("An empty date should result in an error but did not.")
	} else {
		if got, want := err.Error(), "provided date is zero"; !strings.Contains(got, want) {
			t.Errorf("The error messgae '%s' did not contain error '%s'", got, want)
		}
	}
}

func TestStationboardService_SearchWithEmptyTransportations(t *testing.T) {
	client, _, terminate := setupStationBoardTests(t)
	defer terminate()

	_, err := client.Stationboard.SearchWithType(context.Background(),"Zürich HB", time.Time{}, []Transportation{})
	if err == nil {
		t.Errorf("An empty transport filter should result in an error but did not.")
	} else {
		if got, want := err.Error(), "transportation filter is empty"; !strings.Contains(got, want) {
			t.Errorf("The error messgae '%s' did not contain error '%s'", got, want)
		}
	}

}

func TestStationboardService_SearchWithTransportations(t *testing.T) {
	client, fixture, terminate := setupStationBoardTests(t)
	defer terminate()

	date, _ := time.Parse("2006-01-02 15:04", "2020-05-02 20:00")
	stbResult, err := client.Stationboard.SearchWithType(context.Background(), "8591382", date, []Transportation{Tram, Train})
	if err != nil {
		t.Errorf("Failed to search stationboard: %s", err)
		t.Skip("Skip tests because the result is not available")
	}

	// Check amount of connections returned by the api. The amount must not fit the limit because its not a hard limit.
	if got, want := len(stbResult.Journeys), 4; got != want {
		t.Errorf("Got %d connections from the api but want %d connections", got, want)
	}

	// check the returned struct against a static struct
	var staticResult StationboardResult
	_ = json.Unmarshal(fixture, &staticResult)

	if got, want := stbResult, &staticResult; !reflect.DeepEqual(got, want) {
		t.Errorf("The proceeded response does not equals the static fixture")
	}
}

func TestStationboardService_buildUrlPath(t *testing.T) {

	stbSvc := newStationboardService(nil)

	inputDate, _ := time.Parse("2006-01-02 15:04", "2020-05-02 02:00")
	stbOpts := StbOpts{
		Transportations: nil,
		DateTime:        inputDate,
		Arrival:         false,
		Limit:           3,
	}

	want := "stationboard?station=Z%C3%BCrich%2C%20Sternen%20Oerlikon&limit=3&type=departure&datetime=2020-05-02%2002:00"

	got, err := stbSvc.buildUrlPath("Zürich, Sternen Oerlikon", stbOpts)
	if err != nil {
		t.Errorf("Build url path for stationboard failed: %s", err)
	}

	if !strings.Contains(got, want) {
		t.Errorf("The builded url path '%s' does not match the wanted one '%s'", got, want)
	}
}

func TestStationboardService_SearchWithEmptyLocation(t *testing.T) {
	client, _, terminate := setupStationBoardTests(t)
	defer terminate()

	_, err := client.Stationboard.Search(context.Background(),"")
	if err == nil {
		t.Errorf("An empty location name or id should result in an error but did not.")
	} else {
		if got, want := err.Error(), "no location name or id to search for"; !strings.Contains(got, want) {
			t.Errorf("The error messgae '%s' did not contain error '%s'", got, want)
		}
	}
}

func TestStationboardService_queryFailed(t *testing.T) {
	srv, client, terminate := prepare()
	defer terminate()

	srv.HandleFunc("/stationboard", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	// Test valid input parameters
	testValues := []struct {
		in   string
		want string
	}{
		{"", "request path can not be empty"},
		{"stationboard", "failed to perform the http request"},
	}

	for _, v := range testValues {
		_, err := client.Stationboard.query(context.Background(), v.in)
		if err == nil {
			t.Errorf("The stationboard query should return an error when the url path is %s", v.in)
		} else {
			if !strings.Contains(err.Error(), v.want) {
				t.Errorf("The stationboard query returned an unexpected error message %s", err)
			}
		}
	}
}

func TestStationboardService_parseResponseError(t *testing.T) {
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
		_, err := client.Stationboard.parseResponse(v.in)
		if err == nil {
			t.Errorf("The response parser should provide an error message but was empty")
		}
		if got, want := err.Error(), v.out; !strings.Contains(got, want) {
			t.Errorf("The response parser got error message '%s' but want '%s'", got, want)
		}
	}
}