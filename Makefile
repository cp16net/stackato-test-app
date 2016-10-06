default: all
all: generate build run

generate:
	godep go generate

build:
	godep go build

run: 
	./hod-test-app
