deps:
	@echo "Installing dependencies"
	go get -d -t ./...

updatedeps:
	@echo "Updating all dependencies"
	@go get -d -u -f -fix -t ./...

test: deps
	@echo "Testing"
	go test -v ./... -tags=unit

# Run the integration test against OpenWhisk
integration_test:
	@echo "Launch the integration tests."
	go test -v ./... -tags=integration
