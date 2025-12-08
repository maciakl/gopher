BINARY_NAME := "gopher"

[linux]
[macos]
[unix]
clean:
    go clean

    rm {{BINARY_NAME}}_*.zip
    rm -rf test 
    rm -rf dist 

[windows]
clean:
	go clean
	rm {{BINARY_NAME}}_*.zip
	rm test -Force -Recurse -Confirm:$false
	rm dist -Force -Recurse -Confirm:$false



tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
	go mod verify
