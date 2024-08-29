default: build

deps:
	go mod tidy
	go mod vendor

build: build-client build-server
.PHONY: build

build-client: deps
	go build -o bin/client ./client
.PHONY: build-client

build-server: deps
	go build -o bin/server ./server
.PHONY: build-server

run-client: build-client
	cd ./client && ../bin/client
.PHONY: run-client

run-server: build-server
	cd ./server && ../bin/server
.PHONY: run-server

docker-image:
	docker build -f ./server/Dockerfile -t "server:latest" .
	docker build -f ./client/Dockerfile -t "client:latest" .
.PHONY: docker-image

docker-compose-up: docker-image
	docker compose -f docker-compose-dev.yaml up
.PHONY: docker-compose-up

docker-compose-down:
	docker compose -f docker-compose-dev.yaml stop -t 1
	docker compose -f docker-compose-dev.yaml down
.PHONY: docker-compose-down

docker-compose-logs:
	docker compose -f docker-compose-dev.yaml logs -f
.PHONY: docker-compose-logs

data:
	unzip -u client/.data/dataset -d client/.data
