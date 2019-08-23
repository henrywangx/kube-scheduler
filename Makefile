all: build docker

build: 
	@echo "start to go build"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler
docker:
	@echo "start to build docker image"
	@docker build -t scheduler:v0.1 .