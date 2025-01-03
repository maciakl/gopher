BINARY_NAME := "gopher"

[linux]
[macos]
[unix]
clean:
    go clean:w

    rm {{BINARY_NAME}}_*.tar.gz
    rm -rf test 

[windows]
clean:
	go clean
	rm {{BINARY_NAME}}_*.zip
	rm test -Force -Recurse -Confirm:$false



tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
	go mod verify
