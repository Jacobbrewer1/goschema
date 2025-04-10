# Define variables
hash = $(shell git rev-parse --short HEAD)
DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

linux: clean
	@echo "Building for linux"
	GOOS=linux GOARCH=amd64 go build -o bin/goschema -ldflags '-X main.Commit=$(hash) -X main.Date=$(DATE)' .
windows: clean
	@echo "Building for windows"
	GOOS=windows GOARCH=amd64 go build -o bin/goschema -ldflags '-X main.Commit=$(hash) -X main.Date=$(DATE)' .
mac: clean
	@echo "Building for mac"
	GOOS=darwin GOARCH=amd64 go build -o bin/goschema -ldflags '-X main.Commit=$(hash) -X main.Date=$(DATE)' .
clean:
	@echo "Cleaning up"
	# Remove the bin directory
	rm -rf bin
pr-approval:
	@echo "Running PR CI"
	go build ./...
	go vet ./...
	go test ./...
