package opentransport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// The location represents a station, address or poi.
type Location struct {
	Id         string     `json:"id"`         // The id of the station
	Name       string     `json:"name"`       // The location name
	Score      float32    `json:"score"`      // The accuracy of the result
	Coordinate Coordinate `json:"coordinate"` // The location coordinates
	Distance   int        `json:"distance"`   // If search has been with coordinates, distance to original point in meters
	Icon       string     `json:"icon"`       // Indicates if the location is a train, tram, bus, ship or cableway station
}

// The location coordinates.
type Coordinate struct {
	Type string  `json:"type"` // The type of the given coordinate
	X    float64 `json:"x"`    // Latitude
	Y    float64 `json:"y"`    // Longitude
}

// The result returned by the API.
type LocationResult struct {
	Stations []Location `json:"stations"`
}

// Used to declare a specific location type like (station, address, poi)
type LocationType string

const (
	TypeAll     LocationType = "all"
	TypeStation LocationType = "station"
	TypePoi     LocationType = "poi"
	TypeAddress LocationType = "address"
)

// Provides access to query locations
type LocationService struct {
	client *Client
}

// Create a new LocationService.
//
// Returns a pointer to a LocationService
func newLocationService(client *Client) *LocationService {
	ls := &LocationService{client: client}
	return ls
}

// Search for a specific address, poi or station by a name.
// The API will try to autocomplete the keyword based on the given name.
//
// Returns an array with locations and an error.
func (s *LocationService) Search(ctx context.Context, name string) ([]Location, error) {
	return s.SearchWithType(ctx, name, TypeAll)
}

// Search for a specific address, poi or station by a name.
// The API will try to autocomplete the keyword based on the name parameter.
// The response contains only, data which are matching the given location type.
// Use the constants TypeAll, TypeStation, TypeAddress or TypePoi as LocationType.
//
// This filter is currently not supported by the transport.opendata.ch api: https://github.com/OpendataCH/Transport/issues/187
//
// Returns an array with locations and an error.
func (s *LocationService) SearchWithType(ctx context.Context, name string, locationType LocationType) ([]Location, error) {
	path := fmt.Sprintf("locations?query=%s&type=%s", url.PathEscape(name), locationType)
	return s.query(ctx, path)
}

// Search for a specific address, poi or station by lat / long coordinates.
// The API will try to autocomplete the keyword based on the name parameter.
//
// Returns an array with locations and an error.
func (s *LocationService) SearchWithCoordinates(ctx context.Context, lat float64, long float64) ([]Location, error) {
	path := fmt.Sprintf("locations?x=%f&y=%f", lat, long)
	return s.query(ctx, path)
}

// Runs a location query and returns a list of locations
func (s *LocationService) query(ctx context.Context, path string) ([]Location, error) {
	if len(path) == 0 {
		return nil, errors.New("the request path can not be empty")
	}

	req, err := s.client.NewRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create location request: %w", err)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to proceed request: %w", err)
	}

	locResult, err := s.parseResponse(res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse location response: %w", err)
	}

	return locResult.Stations, nil
}

// Parse a json raw response to a location response type.
//
// Returns a location response and an error if the parsing failed.
func (s *LocationService) parseResponse(raw []byte) (*LocationResult, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("response buffer is empty")
	}

	var locResp LocationResult
	err := json.Unmarshal(raw, &locResp)

	s.client.debug.Printf("Parse location response to a typed object")

	return &locResp, err
}

// Returns true if the location
func (l *Location) Station() bool {
	if len(l.Id) > 0 {
		return true
	}
	return false
}
