package opentransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Provides access to stationboards
type StationboardService struct {
	client *Client
}

// Request options for stationboard queries
type StbOpts struct {
	// An optional filter of returned transport types. Possible values are Train, Bus, Tram, Ship or Cableway.
	Transportations []Transportation

	// Date and time for the departure or arrival of the connections at the defined location.
	DateTime       time.Time

	// Indicates if the returned connections are departure or arrival.
	// When the value is true, arrival connections are returned.
	Arrival        bool

	// Number of departing or arrival connections to return. This is not a hard limit.
	// If multiple connections leave at the same time it'll return any connections
	// that leave at the same time as the last connection within the limit.
	Limit          int
}

type StationboardResult struct {
	// Used station for the returned journeys.
	Station     Location     `json:"station"`

	// A list of transportation with the stop of the line leaving or arriving from/to that station.
	Journeys []StationBoardJourney `json:"stationboard"`
}

// The actual transportation of a connection, e.g. a bus or a train with the stop of the line leaving or arriving from/to that station.
type StationBoardJourney struct {
	// The stop of the line leaving or arriving from/to that station.
	Stop Stop `json:"stop"`

	// The actual transportation, e.g. a bus or a train between two stations.
	Journey
}

func newStationboardService(client *Client) *StationboardService {
	return &StationboardService{client: client}
}

// Search for the next connections leaving from a specific location now.
// The location can be searched by its name or location id. The result is limited to 15 connections.
//
// Returns a stationboard result
func (s *StationboardService) Search(ctx context.Context, name string) (*StationboardResult, error) {
	opts := StbOpts{
		DateTime:        time.Now(),
		Arrival:         false,
		Limit:           15,
	}
	return s.SearchWithOpts(ctx, name, opts)
}

// Search for the next connections leaving from a specific location at a specific time.
// The location can be searched by its name or location id. The date has to be non Zero.
// The result is limited to 15 connections.
//
// Returns a stationboard result
func (s *StationboardService) SearchWithDate(ctx context.Context, name string, date time.Time) (*StationboardResult, error) {
	opts := StbOpts{
		DateTime:        date,
		Arrival:         false,
		Limit:           15,
	}
	return s.SearchWithOpts(ctx, name, opts)
}

// Search for the next connections leaving from a specific location at a specific time.
// The location can be searched by its name or location id. The date has to be non Zero.
// Possible transportation filters are Train, Bus, Tram, Ship or Cableway. These types are available as constants.
// The result is limited to 15 connections.
//
// Returns a stationboard result
func (s *StationboardService) SearchWithType(ctx context.Context, name string, date time.Time, transportations []Transportation) (*StationboardResult, error) {
	if len(transportations) == 0 {
		return nil, fmt.Errorf("transportation filter is empty (use SearchWithDate() instead)")
	}
	opts := StbOpts{
		DateTime:        date,
		Arrival:         false,
		Limit:           15,
		Transportations: transportations,
	}
	return s.SearchWithOpts(ctx, name, opts)
}

// Search for the next connections leaving or arriving from a specific location.
// The location can be searched by its name or location id. The date has to be non Zero.
// The limit can be disabled, if the value is 0.
// The transportation filter an be a Train, Bus, Tram, Ship or Cableway. These types are available as constants.
//
// Returns a stationboard result
func (s *StationboardService) SearchWithOpts(ctx context.Context, name string, opts StbOpts) (*StationboardResult, error) {
	path, err := s.buildUrlPath(name, opts)
	if err != nil {
		return nil, err
	}

	return s.query(ctx, path)
}

func (s *StationboardService) query(ctx context.Context, path string) (*StationboardResult, error) {
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
func (s *StationboardService) buildUrlPath(name string, opts StbOpts) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("no location name or id to search for")
	}

	transportations, err := convListParam(convSlice(opts.Transportations), "transportations")
	if err != nil {
		return "", err
	}

	// If the name is a valid location id, the url path looks different
	stationAttr := "station"
	if isId(name) {
		stationAttr = "id"
	}

	// Default direction type is departure
	directionType := "departure"
	if opts.Arrival {
		directionType = "arrival"
	}

	date, err := s.formatDate(opts.DateTime)
	if err != nil {
		return "", fmt.Errorf("failed to build url path: %w", err)
	}

	path := fmt.Sprintf("stationboard?%s=%s&limit=%d&type=%s&datetime=%s%s",
		stationAttr,
		url.PathEscape(name),
		opts.Limit,
		url.PathEscape(directionType),
		url.PathEscape(date),
		transportations)

	return path, nil
}

// Parse a time.Time Date / Time in a date string with a format, accepted by the API
//
// Returns a connection date and time of type time.Time and an error object
func (s *StationboardService) formatDate(date time.Time) (string, error) {
	// check if the date is zero
	if date.IsZero() {
		return "", fmt.Errorf("provided date is zero: please provide a valid time.Time as date")
	}
	return date.Format("2006-01-02 15:04"), nil
}

// Parse a json response to a connection response type
//
// Returns a connection response and an error if the parsing failed
func (s *StationboardService) parseResponse(raw []byte) (*StationboardResult, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("response buffer is empty")
	}

	var stbResp StationboardResult
	err := json.Unmarshal(raw, &stbResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	s.client.debug.Printf("Parsed stationboard response with %d bytes to a structured type", len(raw))
	return &stbResp, err
}
