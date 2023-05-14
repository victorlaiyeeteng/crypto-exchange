build:
	go build -o bin/crypto-exchange

run: build
	./bin/crypto-exchange

test:
	go test -v ./...