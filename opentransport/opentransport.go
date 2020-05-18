package opentransport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// The default URL which points to the production API
const DefaultApiURL = "https://transport.opendata.ch/v1/"

// The default user agent includes the name of the library and its version number
const DefaultUserAgent = "Golang OpenTransport Client/v1.0"

// The default amount of retries during http queries.
const DefaultMaxRetry = 3

// The default pause in seconds between multiple retry requests
const DefaultRetryPause = 5

// The client config holds all values configurable by a user. The type itself will be used internally.
type clientConfig struct {
	// The url of the remote api. Default is DefaultApiURL.
	apiUrl *url.URL

	// The useragent which will be used for http requests.
	userAgent string

	// The amount of retries if the http request to the api fails. Default is 3
	maxRetry int

	// The amount of seconds to sleep between retries
	maxRetryPause int
}

// Transportation can be a Train, Bus, Tram, Ship or Cableway
type Transportation string

const (
	Train    Transportation = "train"
	Bus      Transportation = "bus"
	Tram     Transportation = "tram"
	Ship     Transportation = "ship"
	Cableway Transportation = "cableway"
)

// The client is the entry point of the library. It can be used to
// access various services of the OpenData Transport API.
type Client struct {
	// The instance of a http client, which will be used for all HTTP Requests to the API
	httpClient *http.Client

	// Configuration, which can be changed during initialization
	cfg   *clientConfig
	debug *log.Logger
	error *log.Logger

	// Services which can be used to query different parts of the API
	Location     *LocationService
	Connection   *ConnectionService
	Stationboard *StationboardService
}

// A custom date type to parse iso 8601 date strings
type isoDate struct {
	time.Time
}

// Creates a new opentrasport client, configured with default values
// returns a opentransport client
func NewClient() *Client {
	apiURL, _ := url.Parse(DefaultApiURL)

	cfg := clientConfig{
		apiUrl:    apiURL,
		userAgent: DefaultUserAgent,
		maxRetry:  DefaultMaxRetry,
		maxRetryPause: DefaultRetryPause,
	}

	c, _ := newClientWithConfig(&http.Client{}, &cfg)
	return c
}

// Creates a new opentransport client with a custom apiUrl
// returns a opentransport client object
func NewClientWithUrl(httpClient *http.Client, customURL string) (*Client, error) {
	if len(customURL) == 0 {
		return nil, fmt.Errorf("custom URL does not have to be empty")
	}

	pApiURL, err := url.Parse(customURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse api URL: %w", err)
	}

	if len(pApiURL.Scheme) == 0 {
		return nil, fmt.Errorf("could not parse api URL: empty scheme")
	}

	if len(pApiURL.Host) == 0 {
		return nil, fmt.Errorf("could not parse api URL: empty host")
	}

	cfg := clientConfig{
		apiUrl:    pApiURL,
		userAgent: DefaultUserAgent,
		maxRetry:  DefaultMaxRetry,
		maxRetryPause: DefaultRetryPause,
	}
	return newClientWithConfig(httpClient, &cfg)
}

// Creates a new opentransport client based on a clientConfig type
// returns a configured opentransport client object
func newClientWithConfig(httpClient *http.Client, cfg *clientConfig) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	cfg, err := validClientConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("opentransport: invalid client config: %w", err)
	}

	// Check if the apiUrl has a slash as suffix, otherwise add one
	if !strings.HasSuffix(cfg.apiUrl.Path, "/") {
		cfg.apiUrl.Path += "/"
	}

	// Init default disabled loggers (Logs are written to devNull(0)
	debugLogger := log.New(ioutil.Discard, "Debug:\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger := log.New(ioutil.Discard, "Error:\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Create basic client
	client := &Client{
		cfg:        cfg,
		httpClient: httpClient,
		debug:      debugLogger,
		error:      errorLogger,
	}

	// Init all services
	client.Location = newLocationService(client)
	client.Connection = newConnectionService(client)
	client.Stationboard = newStationboardService(client)

	return client, nil
}

// Validates a client config object and returns an error or a valid config.
func validClientConfig(cfg *clientConfig) (*clientConfig, error) {
	if cfg == nil {
		return nil, errors.New("client config cannot be nil")
	}

	if len(cfg.userAgent) == 0 {
		cfg.userAgent = DefaultUserAgent
	}

	if cfg.apiUrl == nil {
		return nil, errors.New("please provide a api url")
	}

	if len(cfg.apiUrl.Host) == 0 {
		return nil, errors.New("please provide a valid api url host")
	}

	if len(cfg.apiUrl.Scheme) == 0 {
		return nil, errors.New("please provide a valid api url scheme")
	}

	if _, err := validRetry(cfg.maxRetry, cfg.maxRetryPause); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validates the retry configuration
func validRetry(attempts int, pause int) (bool, error) {
	if attempts < 0 || attempts > 10 {
		return false, errors.New("please provide a max retry between 0 and 10")
	}

	if pause < 1 {
		return false, errors.New("please provide a max retry pause more than 1 seconds")
	}

	return true, nil
}

// Create new API Request based on a context and url path.
//
// Returns a pointer to a http.Request. If an error occur, it will be returned.
func (c *Client) NewRequest(ctx context.Context, path string) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Prepare full URL
	reqUrl := fmt.Sprintf("%s%s", c.cfg.apiUrl, path)
	c.debug.Printf("Request url: %s", reqUrl)

	// Prepare Request
	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	return req, nil
}

// The function passed as parameter will be retried until the max attempts is reached or no error returned.
func (c *Client) retry(attempts int, sleep time.Duration, f func() ([]byte, error)) ([]byte, error) {
	r, err := f()
	if err != nil {
		if attempts--; attempts > 0 {
			c.error.Printf("Retry attempt %d of %d: %s", c.cfg.maxRetry - attempts, c.cfg.maxRetry, err)
			time.Sleep(sleep)
			return c.retry(attempts, sleep, f)
		}
	}
	return r, err
}

// Do the actual http request. Retries the http request, if an http 500 or a http error occur.
// The max retries and the pause between can be configured with MaxRetry Method.
//
// Returns a byte array of the body and an error if the request failed. When the server
// respond with a status which does not match HTTP 200 OK, an error will be returned
func (c *Client) Do(req *http.Request) ([]byte, error) {
	if ok, err := validRequest(req); !ok {
		return nil, fmt.Errorf("opentransport: invalid http request: %w", err)
	}

	pause :=  time.Duration(c.cfg.maxRetryPause) * time.Second
	var r, err = c.retry(c.cfg.maxRetry, pause, func() ([]byte, error) {
		r, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to proceed http request: %w", err)
		}

		if r != nil {
			defer r.Body.Close()
			c.debug.Printf("Server responded with status %s", r.Status)

			switch s := r.StatusCode; {
			case s >= http.StatusInternalServerError:
				return nil, fmt.Errorf("remote server responded with an error: %s", r.Status)
			case s == http.StatusOK:
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to parse response (%d): %v ", err, r.StatusCode)
				}
				return body, nil
			default:
				return nil, nil
			}
		}
		return nil, errors.New("server response of the http request is empty")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to perform the http request after %d retries", c.cfg.maxRetry)
	}

	return r, nil
}

// Enable debug and error logs to a specified output.
// When the output is nil, the debug logs will be written to os.Stdout
// and the error logs to os.Stderr
func (c *Client) EnableLogs(out io.Writer) {
	if out != nil {
		c.debug.SetOutput(out)
		c.error.SetOutput(out)
	} else {
		c.debug.SetOutput(os.Stdout)
		c.error.SetOutput(os.Stderr)
	}

	c.debug.Printf("Client is configured with target API: %s and useragent: '%s'", c.cfg.apiUrl, c.cfg.userAgent)
}

// Sets a custom user agent. If the provided user agent is empty, the default one will be used.
func (c *Client) UserAgent(userAgent string) {
	if len(userAgent) > 0 {
		c.cfg.userAgent = userAgent
	} else {
		c.cfg.userAgent = DefaultUserAgent
	}
}

// Sets the max attempts to retry and the pause between a http request.
//
// Returns an error if the provided value is invalid
func (c *Client) MaxRetry(attempts int, pause int) error {
	// Check if the user agent is empty
	if _, err := validRetry(attempts, pause); err != nil {
		return fmt.Errorf("failed to configure retry options: %w", err)
	}
	c.cfg.maxRetry = attempts
	c.cfg.maxRetryPause = pause
	return nil
}

// Parse date fields from format 2006-01-02T15:04:05Z0700 to time.Time. When
// the field is nil, an empty time.Time will be unmarshal. Returns an error if a
// invalid date format will be provided.
func (d *isoDate) UnmarshalJSON(raw []byte) error {
	i := string(raw)

	// Check if the string contains an empty string or `null`
	if strings.Contains(i, "null") || len(i) == 0 {
		return nil
	}

	i = strings.Trim(i, `"`) // Remove doubled quotes

	// Parse the date string to iso 8601
	t, err := time.Parse("2006-01-02T15:04:05Z0700", i)
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

// Generates a URL encoded string which can be appended to an url path.
// Example: &name[]=value1&name[]=value2
//
// Returns a url path string based on the input string array
func convListParam(values []string, name string) (string, error) {
	paramList := ""
	for i, v := range values {
		if len(v) == 0 {
			return "", fmt.Errorf("%s filter at index %d cannot be empty", name, i)
		}
		paramList = fmt.Sprintf("%s&%s[]=%s", paramList, name, url.PathEscape(v))
	}
	return paramList, nil
}

// Converts a typed transportation slice to a string slice
//
// Returns a plain string slice
func convSlice(list []Transportation) []string {
	conv := make([]string, len(list))
	for i, s := range list {
		conv[i] = string(s)
	}
	return conv
}

// Check if a string value is a valid id. The value can be converted to an integer and has more than 5 digits.
//
// Returns true if the value is a valid id and false if not
func isId(v string) bool {
	if len(v) > 5 {
		_, err := strconv.Atoi(v)
		if err == nil {
			return true
		}
	}

	return false
}

// Validates a http.Request against minimum requirements
//
// Returns true or false if the request is valid
func validRequest(req *http.Request) (bool, error) {
	if req.Method != "GET" {
		return false, fmt.Errorf("the request has an invalid http method %s. (only GET is allowed)", req.Method)
	}

	if req.Body != nil {
		return false, fmt.Errorf("the request should not contain a body")
	}

	if len(req.URL.Scheme) == 0 {
		return false, fmt.Errorf("a valid protocol scheme should be defined")
	}

	if len(req.Host) == 0 {
		return false, fmt.Errorf("a valid host should be defined")
	}

	return true, nil
}

// Converts a boolean value to a numeric value
//
// Returns 1 or 0
func boolToInt(value bool) uint8 {
	var out uint8
	if value {
		out = 1
	}
	return out
}
