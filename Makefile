default: build

deps:
	go mod tidy
	go mod vendor

build: deps
	go build -o bin/client ./client
.PHONY: build

run-client: build
	cd ./client && ../bin/client

docker-image:
	docker build -f ./server/Dockerfile -t "server:latest" .
	docker build -f ./client/Dockerfile -t "client:latest" .
.PHONY: docker-image

docker-compose-up:
	docker compose -f docker-compose-dev.yaml up --build
.PHONY: docker-compose-up

docker-compose-down:
	docker compose -f docker-compose-dev.yaml stop -t 1
	docker compose -f docker-compose-dev.yaml down
.PHONY: docker-compose-down

docker-compose-logs:
	docker compose -f docker-compose-dev.yaml logs -f
.PHONY: docker-compose-logs
