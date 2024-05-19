build:
	go build -ldflags="-w -s" -o stack

build-amd:
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o stack

