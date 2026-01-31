BINARY_NAME := "gopher"

[linux]
[macos]
[unix]
clean:
    go clean
    -rm -rf test 
    -rm -rf dist 
    -rm coverage
    -rm coverage.*

[windows]
clean:
	go clean
	-rm test -Force -Recurse -Confirm:$false
	-rm dist -Force -Recurse -Confirm:$false
	-rm coverage -Force -Confirm:$false
	-rm coverage.* -Force -Confirm:$false

tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
	go mod verify

test:
    go test ./...

coverage:
    go test -coverprofile=coverage ./...
    go tool cover -html=coverage -o coverage.html

[windows]
check: coverage
   pwsh -c Start-Process coverage.html 

[macos]
check: coverage
    open coverage.html
