all:
	go build -o gourami && ./gourami

test:
	go test ./...
