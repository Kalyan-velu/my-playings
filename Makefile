BINARY_NAME=main
PORT?=8080

export PORT

clean:
	rm -f ./bin/$(BINARY_NAME)*

.PHONY: dev debug

build:
	go build -C cmd/server -gcflags="all=-N -l" -o ../../bin .
dev:
	air -c .air.toml

debug:build
	dlv exec ./bin/main.exe --listen=:2345 --headless=true --api-version=2 --accept-multiclient --continue --log