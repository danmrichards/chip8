GOARCH=amd64
BINARY=chip8

build:
	GOOS=linux go build -o ./out/${BINARY}-linux-${GOARCH} ./cmd/chip8

test:
	go test -count=1 -failfast -cover ./...

.PHONY: build