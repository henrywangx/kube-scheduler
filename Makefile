all: build docker

build: 
	@echo "start to go build"
	@go build -o kube-scheduler
docker:
	@echo "start to build docker image"
	@docker build -t scheduler:v0.1 .