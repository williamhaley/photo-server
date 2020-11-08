build:
	rm bin/photo-server* || true
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o bin/photo-server.darwin.amd64
