# Development for OpenTransport API Library

## API Mock during development

Some of the Request of transport.opendata.ch are mocked with [Mockoon](https://mockoon.com/). You can use them by [importing](https://mockoon.com/tutorial/import-export-environments-routes/) the file into a running Mockoon instance.

__available requests are:__

- https://localhost:3001/v1/
- https://localhost:3001/v1/connections
- https://localhost:3001/v1/connections?from=8503000&to=8507000
- https://localhost:3001/v1/connections?from=Bern&to=Z%C3%BCrich&limit=1
- https://localhost:3001/v1/connections?from=Waldgarten&to=Z%C3%BCrich,HB&transportations[]=bus
- https://localhost:3001/v1/connections?from=Bellevue&to=Liebefeld&via[]=Olten&via[]=Aarau
- https://localhost:3001/v1/stationboard
- https://localhost:3001/v1/stationboard?station=8591420
- https://localhost:3001/v1/stationboard?station=Waldgarten
- https://localhost:3001/v1/locations
- https://localhost:3001/v1/locations?query=Z%C3%BCrich
- https://localhost:3001/v1/locations?query=Bern
- https://localhost:3001/v1/locations?x=47.403718&y=8.557201