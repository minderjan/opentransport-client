package opentransport

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Package global test functions

// Reads a json file from ./testdata directory.
// You only have to provide the filename without the file ending ".json"
//
// Returns the content of the file as string
func readFixture(name string) ([]byte, error) {
	path := filepath.Join("testdata", fmt.Sprintf("%s.json", name))
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Prepares a new test server and an instance of the opentransport client
// Reutrns the server, the client and a close function of the server
func prepare() (server *http.ServeMux, client *Client, terminate func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fixture, _ := readFixture("root")
		_, _ = fmt.Fprintln(w, fixture)
	})

	ts := httptest.NewServer(mux)

	c, _ := NewClientWithUrl(nil, ts.URL)
	c.EnableLogs(nil)
	_ = c.MaxRetry(2, 1)

	return mux, c, ts.Close
}

func TestNewClient(t *testing.T) {
	c := NewClient()

	// Test api url
	if got, want := c.cfg.apiUrl.String(), DefaultApiURL; got != want {
		t.Errorf("New opentransport client has a wrong api url. Got %s instead of wanted %s.", got, want)
	}

	// Test user agent
	if got, want := c.cfg.userAgent, DefaultUserAgent; got != want {
		t.Errorf("New opentransport client has a wrong user agent. Got %s instead of wanted %s.", got, want)
	}

	// Test if the http client will be recreated, each time a new client will be created
	c2 := NewClient()
	if c.httpClient == c2.httpClient {
		t.Errorf("New opentransport clients have all the same httpclient, but they should have different ones.")
	}
}

func TestNewClientWithUrl(t *testing.T) {
	goodURL := "https://localhost:3001/v1"

	c, err := NewClientWithUrl(nil, goodURL)

	// Test if the client can be created without a library error
	if err != nil {
		t.Errorf("Failed to create new opentransport client with custom url: %s", err)
		t.Skip("Followed tests can not be executed, when the client can not be created")
	}

	// Test if the returned client has the correct api url configured
	if got, want := c.cfg.apiUrl.String(), fmt.Sprintf("%s/", goodURL); got != want {
		t.Errorf("New opentransport client does not have the correct api url configured. Got %s but wanted %s", got, want)
	}

	// Test if a client can be created with different bad URLs
	var badURLs = []string{
		"https:///v1",
		"//localhost:3001/v1",
		"",
		"+.https:/localhost/v1",
	}

	for _, badURL := range badURLs {
		_, err = NewClientWithUrl(nil, badURL)

		// Test if the client can be created without a library error
		if err == nil {
			t.Errorf("A client with a bad api url could be created: bad URL (%s)", badURL)
		}
	}
}

func TestClient_SetUserAgent(t *testing.T) {
	ua := "Testing"
	c := NewClient()

	// Test an empty user agent change
	c.UserAgent("")

	if got, want := c.cfg.userAgent, DefaultUserAgent; got != want {
		t.Errorf("Failed to set a custom user agent to the client")
	}

	// Test a correct user agent change
	c.UserAgent(ua)

	if got, want := c.cfg.userAgent, ua; got != want {
		t.Errorf("Failed to set a custom user agent to the client")
	}
}

func TestClient_DefaultLoggerOutput(t *testing.T) {

	// Initiate new Client and assign created logger
	c := NewClient()
	c.EnableLogs(nil)

	if c.debug == nil {
		t.Errorf("No default io.Writer configured for the debug logger")
		t.Skip("Skip default logging tests, because of missing output target")
	}

	if c.error == nil {
		t.Errorf("No default io.Writer configured for the error logger")
		t.Skip("Skip default logging tests, because of missing output target")
	}

	// Check if the logger has a default output
	debugOut := c.debug.Writer()
	errorOut := c.error.Writer()

	// Check if the debug output is a type of os.File
	if _, ok := debugOut.(*os.File); !ok {
		t.Errorf("Debug logs does not write to a file output")
	}

	// Check if the error output is a type of os.File
	if _, ok := errorOut.(*os.File); !ok {
		t.Errorf("Error logs does not write to a file output")
	}

}

func TestClient_CustomLoggerOutput(t *testing.T) {

	// Define buffer where the logs should be written in
	var out bytes.Buffer

	// Initiate new Client and assign created logger
	c := NewClient()
	c.EnableLogs(&out)

	// Enable Logging (triggers also a log message)
	c.EnableLogs(&out)

	// Check if the logger did not wrote messages to the buffer
	c.debug.Printf("Test Debug Message")
	if out.Len() == 0 {
		t.Errorf("logs are enabled but defined debug log output will not be used for logging")
	}

	out.Reset() // clear buffer output

	// Check if the logger did not wrote messages to the buffer
	c.error.Printf("Test Error Message")
	if out.Len() == 0 {
		t.Errorf("logs are enabled but defined error log output will not be used for logging")
	}

	out.Reset() // clear buffer output
}

func TestClient_convParamList(t *testing.T) {
	// Test valid input parameters
	testValues := []struct {
		name string
		in   []string
		want string
	}{
		{"via", []string{"Zürich, Hardbrücke", "Bern", "Olten"}, "&via[]=Z%C3%BCrich%2C%20Hardbr%C3%BCcke&via[]=Bern&via[]=Olten"},
		{"transportations", []string{"tram"}, "&transportations[]=tram"},
		{"via", []string{}, ""},
	}

	for _, val := range testValues {
		out, err := convListParam(val.in, val.name)
		if err != nil {
			t.Errorf("Failed to convert array to a flat query parameter string. %s. Got %s but want %s. ", err, out, val.want)
		}

		if got, want := out, val.want; !strings.Contains(got, want) {
			t.Errorf("Parameter convertion failed. Got %s but want %s.", got, val.want)
		}
	}

	// Test invalid input parameters
	testValues = []struct {
		name string
		in   []string
		want string
	}{
		{"transportations", []string{""}, "transportations filter at index 0 cannot be empty"},
	}

	for _, val := range testValues {
		out, err := convListParam(val.in, val.name)

		if err == nil {
			t.Errorf("An array with invalid input parameter should not be able to be converted to a valid query string. Got %s but want %s", out, val.want)
		}

		if got, want := err.Error(), val.want; !strings.Contains(got, want) {
			t.Errorf("The parameter convertion should return a error string like '%s' but got '%s'", want, got)
		}
	}
}

func TestClient_NewRequest(t *testing.T) {
	req, err := NewClient().NewRequest(nil, "http://transport.opendata.ch/v1/")
	if err != nil {
		t.Errorf("Failed to create new request because of %s", err)
	}

	if req.Context() != context.Background() {
		t.Errorf("A default context should be assigned, if the user does not provide one.")
	}
}

func TestClient_Do(t *testing.T) {
	req, err :=  http.NewRequest("POST", DefaultApiURL, nil)
	if err != nil {
		t.Errorf("Failed to create new raw request: %s", err)
	}

	_, err = NewClient().Do(req)
	if got, want := err.Error(), "invalid http request"; !strings.Contains(got, want) {
		t.Errorf("Only GET Methods are allowed.")
	}
}

func TestClient_validRequest(t *testing.T) {
	req1, _ :=  http.NewRequest("POST", DefaultApiURL, nil)
	req2, _ :=  http.NewRequest("DELETE", DefaultApiURL, nil)
	req3, _ :=  http.NewRequest("PUT", DefaultApiURL, nil)
	req4, _ :=  http.NewRequest("GET", DefaultApiURL, bytes.NewBuffer([]byte("request with a body")))
	req5, _ :=  http.NewRequest("GET", "http://", nil)
	req6, _ :=  http.NewRequest("GET", "transport.opendata.ch", nil)

	testValues := []struct {
		req *http.Request
		want string
	}{
		{req1, "invalid http method"},
		{req2, "invalid http method"},
		{req3, "invalid http method"},
		{req4, "should not contain a body"},
		{req5, "valid host"},
		{req6, "protocol scheme should"},
	}

	for _, tv := range testValues {
		_, err := validRequest(tv.req)
		if err ==  nil {
			t.Errorf("The request should be invalid but was not.")
			t.Skipf("Abort following tests because of the dependency to this.")
		}
		if got, want := err.Error(), tv.want; !strings.Contains(got, want) {
			t.Errorf("The request was invalid, but did trigger a wrong error message. Got %s but want %s", got, want)
		}
	}
}

func TestConnectionService_convSlice(t *testing.T) {
	transportations := []Transportation{Tram, Bus, Ship}

	want := []string{"tram", "bus", "ship"}
	got := convSlice(transportations)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Convertion from Transportation Slice failed. Got %s but want %s", got, want)
	}
}

func TestValidClientConfig(t *testing.T) {
	validUrl, _ := url.Parse("https://transport.opendata.ch/v1/")

	cfg := &clientConfig{
		apiUrl:        validUrl,
		userAgent:     "",
		maxRetry:      2,
		maxRetryPause: 1,
	}

	cfg, err := validClientConfig(cfg)
	if err != nil {
		t.Errorf("An empty useragent should not trigger an error. The default value should be used instead.")
	}

	if got, want := cfg.userAgent, DefaultUserAgent; !strings.Contains(got, want) {
		t.Errorf("A wrong default value was choosen for an empty user agent. Got %s but want %s", got, want)
	}
}

func TestValidClientConfigErrors(t *testing.T) {
	validUrl, _ := url.Parse("https://transport.opendata.ch/v1/")
	missingHost, _ := url.Parse(".com")
	missingScheme, _ := url.Parse("//transport.openendata.ch")

	// Test valid input parameters
	testValues := []struct {
		in   *clientConfig
		want string
	}{
		{&clientConfig{
			apiUrl:        missingHost,
			userAgent:     "OpenTransport Go/test",
			maxRetry:      2,
			maxRetryPause: 1,
		}, "provide a valid api url host"},
		{&clientConfig{
			apiUrl:        nil,
			userAgent:     "OpenTransport Go/test",
			maxRetry:      2,
			maxRetryPause: 1,
		}, "please provide a api url"},
		{&clientConfig{
			apiUrl:        missingScheme,
			userAgent:     "OpenTransport Go/test",
			maxRetry:      2,
			maxRetryPause: 1,
		}, "provide a valid api url scheme"},
		{&clientConfig{
			apiUrl:        validUrl,
			userAgent:     "OpenTransport Go/test",
			maxRetry:      30,
			maxRetryPause: 1,
		}, "provide a max retry between 0 and 10"},
		{&clientConfig{
			apiUrl:        validUrl,
			userAgent:     "OpenTransport Go/test",
			maxRetry:      3,
			maxRetryPause: 0,
		}, "more than 1 seconds"},
		{nil, "client config cannot be nil"},
	}

	for _, v := range testValues {
		_, err := validClientConfig(v.in)
		if err == nil {
			t.Errorf("The client config should be invalid: %+v", v.in)
		} else {
			if got, want := err.Error(), v.want; !strings.Contains(got, want) {
				t.Errorf("The client config validation should return '%s' but returned '%s'", want, got)
			}
		}
	}
}
