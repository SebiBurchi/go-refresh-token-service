PKGS := $(shell go list ./...)

redis: 
	docker run --name my-redis -p 6379:6379 -d redis -e REDIS_PASSWORD=redis

gomod: 
	go mod tidy

build: 
	go build -o jwt-service

run:
	./jwt-service

test: gomod
	mkdir -p out
	go test -v ./... -race -short -coverprofile out/cover.out
	go tool cover -html=out/cover.out -o out/cover.html
	go tool cover -func=out/cover.out

all: gomod build run