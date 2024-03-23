all: download vet lint test

download:
	go mod download

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race -v ./...

cover:
	go test -race -cover -coverprofile=coverage.out -v ./...
	go tool cover -html=coverage.out

clean:
	git clean -fxd -e .idea

start-db:
	docker run --name test-db -p 3671:5432 -d -e POSTGRES_USER=test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=test postgres:alpine

stop-db:
	docker rm --force test-db
