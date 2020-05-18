package opentransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// A connection represents a possible journey between two locations.
type Connection struct {
	From        Stop           `json:"from"`        // Specifies the departure location of the connection.
	To          Stop           `json:"to"`          // Specifies the arrival location of the connection.
	Duration    string         `json:"duration"`    // Specifies up to five via locations.
	Transfers   int            `json:"transfers"`   // Count of different vehicles.
	Service     ServiceDetails `json:"service"`     // Service information about how regular the connection operates.
	Products    []string       `json:"products"`    // List of transport products (e.g. IR, S9).
	Capacity1st int            `json:"capacity1st"` // The maximum estimated occupation load of 1st class coaches (e.g. 1).
	Capacity2nd int            `json:"capacity2nd"` // The maximum estimated occupation load of 2nd class coaches (e.g. 2).
	Sections    []Section      `json:"sections"`    // A list of sections.
}

// Operation information for a connection. Currently not available: https://github.com/OpendataCH/Transport/issues/159
type ServiceDetails struct {
	Regular   string `json:"regular"`   // Information about how regular a connection operates (e.g. daily).
	Irregular string `json:"irregular"` // Additional information about irregular operation dates (e.g. not 23., 24. Jun 2020).
}

// A checkpoint represents an arrival or a departure point (in time and space) of a connection.
type Stop struct {
	Station   Location  `json:"station"`   // A location object showing this line's stop at the requested station.
	Arrival   isoDate   `json:"arrival"`   // The arrival time to the checkpoint. If the value is null, 0001-01-01 00:00:00 +0000 UTC will be returned.
	Departure isoDate   `json:"departure"` // The departure time from the checkpoint. If the value is null, 0001-01-01 00:00:00 +0000 UTC will be returned.
	Delay     int       `json:"delay"`     // The delay at this checkpoint, can be null if no prognosis is available.
	Platform  string    `json:"platform"`  // The arrival/departure platform
	Prognosis Prognosis `json:"prognosis"` // status of a connection checkpoint in realtime
}

// A prognosis contains "realtime" information on the status of a connection checkpoint.
type Prognosis struct {
	Platform    string  `json:"platform"`    // The estimated arrival/departure platform (e.g. 8). Can be empty if no platform is available for this connection.
	Arrival     isoDate `json:"arrival"`     // The arrival time prognosis to the checkpoint, date format ISO 8601 (e.g. 2019-03-31T08:58:00+02:00).
	Departure   isoDate `json:"departure"`   // The departure time prognosis to the checkpoint,  date format ISO 8601 (e.g. 2019-03-31T08:58:00+02:00).
	Capacity1st int     `json:"capacity1st"` // The estimated occupation load of 1st class coaches (e.g. 1).
	Capacity2nd int     `json:"capacity2nd"` // The estimated occupation load of 2nd class coaches (e.g. 2).
}

// A connection consists of one or multiple sections.
type Section struct {
	Journey   Journey `json:"journey"`   // A journey, the transportation used by this section. Can be empty.
	Walk      Walk    `json:"walk"`      // Information about walking distance, if available. (eg. null)
	Departure Stop    `json:"departure"` // The departure checkpoint of the connection
	Arrival   Stop    `json:"arrival"`   // The arrival checkpoint of the connection
}

// The actual transportation of a section, e.g. a bus or a train between two stations.
type Journey struct {
	Name         string `json:"name"`         // The name of the connection (e.g. ICN 518).
	Category     string `json:"category"`     // The type of connection this is (e.g. ICN).
	Subcategory  string `json:"subcategory"`  // The sub type of connection this is (e.g. ICN).
	CategoryCode int    `json:"categoryCode"` // Currently not available: https://github.com/OpendataCH/Transport/issues/160
	Number       string `json:"number"`       // The number of the connection's line (e.g. 518).
	Operator     string `json:"operator"`     // The operator of the connection's line (e.g. ZVV).
	To           string `json:"to"`           // The final destination of this line (e.g. Zürich HB)
	PassList     []Stop `json:"passList"`     // Checkpoints the train passed on the journey.
	Capacity1st  int    `json:"capacity1st"`  // currently not available: https://github.com/OpendataCH/Transport/issues/163
	Capacity2nd  int    `json:"capacity2nd"`  // currently not available: https://github.com/OpendataCH/Transport/issues/163
}

// Information about walking distance, if available
type Walk struct {
	Duration int `json:"duration"` // Distance in meter to walk, from the latest station (eg. 130)
}

// Represents an answer from the API
type ConnectionResult struct {
	Connections []Connection `json:"connections"`
	From        Location     `json:"from"` // Specifies the departure location of the search.
	To          Location     `json:"to"`   // Specifies the arrival location of the search.
	Stations    struct {
		From []Location `json:"from"` // Specifies the departure station of the connection.
		To   []Location `json:"to"`   // Specifies the arrival station of the connection.
	}
}

// Provides access to query connections
type ConnectionService struct {
	client *Client
}

// Possible request option to search for a connection between two locations
type ConnOpts struct {
	IsArrival       bool             // defaults to false
	Transportations []Transportation // defaults to all
	Via             []string         // The via locations, which the connection should pass during transfer.
	Bike            bool             // currently not available: https://github.com/OpendataCH/Transport/issues/191
	Couchette       bool             // defaults to false, if set to true only night trains containing couchettes are allowed, implies Direct=true
	Sleeper         bool             // defaults to false, if set to true only night trains containing beds are allowed, implies Direct=true
	Direct          bool             // defaults to false, if set to true only direct connections are allowed
	Accessibility   Accessibility    // default is empty. You can set IndependentBoarding, AssistedBoarding or AdvancedNotice
	Limit           int              // 1 - 16. Specifies the number of connections to return. If several connections depart at the same time they are counted as 1. Default limit is 0 which means, no limit is set.
}

type Accessibility string

const (
	IndependentBoarding Accessibility = "independent_boarding"
	AssistedBoarding    Accessibility = "assisted_boarding"
	AdvancedNotice      Accessibility = "advanced_notice"
)

// Date string when a connection starts
type connDate string

// Time string when a connection starts
type connTime string

// Create a new ConnectionService
// returns a pointer to a ConnectionService
func newConnectionService(client *Client) *ConnectionService {
	cs := &ConnectionService{client: client}
	return cs
}

// Search for the next connections from a location to another.
// A non zero time.Time parameter defines a specific time of the departing location.
//
// Returns a ConnectionResult type which contains all data according to this query.
func (s *ConnectionService) Search(ctx context.Context, from string, to string, date time.Time) (*ConnectionResult, error) {
	return s.SearchWithOpts(ctx, from, to, date, &ConnOpts{})
}

// Search for the next connections from a location to another via one or more stations between.
// You have to provide at minimum one via station.
// A non zero time.Time parameter defines a specific time of the departing location.
//
// Returns a ConnectionResult type which contains all data according to this query.
func (s *ConnectionService) SearchVia(ctx context.Context, from string, to string, date time.Time, via []string, ) (*ConnectionResult, error) {
	connOpts := &ConnOpts{
		Via: via,
	}
	return s.SearchWithOpts(ctx, from, to, date, connOpts)
}

// Search for the next connections from a location to another.
// You can provide api parameters within an ConnOpts type.
// A non zero time.Time parameter defines a specific time of the departing location.
//
// Returns a ConnectionResult type which contains all data according to this query.
func (s *ConnectionService) SearchWithOpts(ctx context.Context, from string, to string, date time.Time, opts *ConnOpts) (*ConnectionResult, error) {
	d, t, err := s.formatDate(date)
	if err != nil {
		return nil, fmt.Errorf("bad input parameter: %w", err)
	}

	path, err := s.buildUrlPath(from, to, d, t, opts)
	if err != nil {
		return nil, err
	}

	return s.query(ctx, path)
}

// Runs a connection query and returns a ConnectionResult struct
func (s *ConnectionService) query(ctx context.Context, path string) (*ConnectionResult, error) {
	if len(path) == 0 {
		return nil, errors.New("the request path can not be empty")
	}

	req, err := s.client.NewRequest(ctx, path)
	if err != nil {
		return nil, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	return s.parseResponse(res)
}

// Generates a formatted and url encoded path out of the provided parameters.
//
// Returns a full url path starting from base Url
func (s *ConnectionService) buildUrlPath(from string, to string, date connDate, time connTime, opts *ConnOpts) (string, error) {
	via, err := convListParam(opts.Via, "via")
	if err != nil {
		return "", err
	}

	transportations, err := convListParam(convSlice(opts.Transportations), "transportations")
	if err != nil {
		return "", err
	}

	// build the request url path
	path := fmt.Sprintf("connections?from=%s&to=%s&date=%s&time=%s&isArrivalTime=%d&direct=%d&bike=%d&sleeper=%d&couchette=%d%s%s&limit=%d",
		url.PathEscape(from),
		url.PathEscape(to),
		url.PathEscape(string(date)),
		url.PathEscape(string(time)),
		boolToInt(opts.IsArrival),
		boolToInt(opts.Direct),
		boolToInt(opts.Bike),
		boolToInt(opts.Sleeper),
		boolToInt(opts.Couchette),
		via,
		transportations,
		opts.Limit,
	)

	return path, nil
}

// Parse a json response to a connection response type
//
// Returns a connection response and an error if the parsing failed
func (s *ConnectionService) parseResponse(raw []byte) (*ConnectionResult, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("response buffer is empty")
	}

	var conResp ConnectionResult
	err := json.Unmarshal(raw, &conResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	s.client.debug.Printf("Parsed connection response with %d bytes to a structured type", len(raw))
	return &conResp, err
}

// Parse a time.Time Date / Time in a date string with a format, accepted by the API
//
// Returns a connection date and time of type time.Time and an error object
func (s *ConnectionService) formatDate(date time.Time) (connDate, connTime, error) {
	// check if the date is zero
	if date.IsZero() {
		return "", "", fmt.Errorf("provided date is zero: please provide a valid time.Time as date")
	}

	// parse date and time
	d := connDate(date.Format("2006-01-02"))
	t := connTime(date.Format("15:04"))

	return d, t, nil
}
