.PHONY: run build test docker-build docker-run lint clean

run:
	go run main.go

build:
	go build -o bin/api-server main.go

test:
	go test ./... -v

docker-build:
	docker build -t wombocombo-backend .

docker-run:
	docker run --rm -p 3000:3000 --env-file .env wombocombo-backend

lint:
	golangci-lint run

clean:
	rm -rf bin/

tidy:
	go mod tidy
