default: all
all: generate build run

generate:
	godep go generate

build:
	godep go build

run: 
	PORT=8888 ./hod-test-app
