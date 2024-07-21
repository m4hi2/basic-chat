all: build_client

build_client:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/client_darwin_arm64_$(VERSION) cmd/client/client.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/client_darwin_amd64_$(VERSION) cmd/client/client.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o bin/client_windows_arm64_$(VERSION).exe cmd/client/client.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/client_windows_amd64_$(VERSION).exe cmd/client/client.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/client_linux_arm64_$(VERSION) cmd/client/client.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/client_linux_amd64_$(VERSION) cmd/client/client.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/chat_server_linux_amd64_$(VERSION) cmd/server/server.go

clean:
	rm -rf bin/
