default: all
all: generate build run

generate:
	go generate

build:
	go build

run: 
	./hod-test-app
