linux: clean
	@echo "Building for linux"
	GOOS=linux GOARCH=amd64 go build -o bin/linux ./cmd/schema
windows: clean
	@echo "Building for windows"
	GOOS=windows GOARCH=amd64 go build -o bin/windows ./cmd/schema
mac: clean
	@echo "Building for mac"
	GOOS=darwin GOARCH=amd64 go build -o bin/mac ./cmd/schema
clean:
	@echo "Cleaning up"
	# Remove the bin directory
	rm -rf bin
pr-approval:
	@echo "Running PR CI"
	go build ./...
	go vet ./...
	go test ./...
