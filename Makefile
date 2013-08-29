all:
	go build -o gourami && ./gourami

test:
	go test ./...

test_storage:
	go test storage/* -v

test_types:
	go test types/* -v
