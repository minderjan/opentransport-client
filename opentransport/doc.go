// Use of this source code is governed by a MIT License.
// License that can be found in the LICENSE file.

// The OpenTransport Client provides a simplified access to the
// Swiss public transport API https://transport.opendata.ch.
//
// The Transport API allows interested developers to build their own applications
// using public timetable data, whether they're on the web, the desktop or mobile devices.
// The aim of this inofficial API is to cover public transport within Switzerland.
// The API uses a web service provided by search.ch.
//
// If you are looking for an officially supported source or need to
// download all data e.g in GTFS format, please check opendata.swiss.
//
// In order to be kept update on the future development of this API,
// please subscribe to our low-traffic https://groups.google.com/a/opendata.ch/forum/#!forum/transport-wg
//
// Rate Limit
//
// The number of HTTP requests you can send is constraint by the rate limit of https://timetable.search.ch/api/help
//
// If these limits are reached, you can contact search.ch to find a solution.
//
// Basic Usage
//
// The basic functions can be used as follows:
//
//	client := opentransport.NewClient()
//
//	// Search for Locations by name
//	result, err := client.Location.Search(context.Background(), "Zürich, HB")
//
//	// Search for connections between two locations
//	result, err := client.Connection.Search(context.Background(), "Zürich, HB", "Bern", time.Now())
//
//	// Search for connections departing from a location
//	result, err := client.Stationboard.Search(context.Background(), "Zürich, HB")
//
// Logging
//
// The library does not produce log messages by default. However, this can be adjusted. You can either
// enable logging with an empty io.Writer to write to os.stdout.You can also define your own output,
// for example to write the logs to a file.
//
//	// Write to stdout and stderr
//	client.EnableLogs(nil)
//
//	// Create a logfile
//	f, _ := os.OpenFile("opentransport.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
//	// Create new multi writer to write the logs to the file and to stdout in parallel
//	multi := io.MultiWriter(f, os.Stdout)
//	client.EnableLogs(multi)
//
package opentransport
