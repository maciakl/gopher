BINARY_NAME := "gopher"

# clean up build artifacts
[group('clean')]
clean:
    go clean
    -rm -rf test 
    -rm -rf dist 
    -rm coverage
    -rm coverage.*

# remove the test folder only
[group('clean')]
rmtest:
    -rm -rf test

# tidy up
[group('build')]
tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
	go mod verify

# run the tests in verbose mode
[group('test')]
test:
    go test -v ./... | clrz

# generate code coverage report
[group('test')]
coverage:
    go test -coverprofile=coverage ./...
    go tool cover -html=coverage -o coverage.html

# open the coverage report in the default browser
[windows]
[group('test')]
check: coverage
   pwsh -c Start-Process coverage.html 

# open the coverage report in the default browser
[macos]
[group('test')]
check: coverage
    open coverage.html

# release the project and generate a scoop file
[group('release')]
release: clean test
    gopher release
    gopher scoop
