default: all
all: generate build run

generate:
	godep go generate

build:
	godep go build

run: 
	PORT=8888 ./hod-test-app

deploy: generate build
	cf push

clean:
	rm -f bindata_* hod-test-app
