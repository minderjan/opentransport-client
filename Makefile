test:
	go test -v -race -coverprofile bin/coverage.out -covermode atomic ./...

integration:
	go test -v -tags=integration ./opentransport

doc:
	godoc -http=:6060