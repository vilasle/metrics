# build-all: build-server,build-agent	
BUILD_VERSION=$(shell git describe --always --long)
BUILD_DATE=$(shell date +'%Y/%m/%d %H:%M:%S')
BUILD_COMMIT=$(shell git log --format='%H' -n 1)
build-server:
	go build -o bin/server -ldflags "-X 'main.buildVersion=$(BUILD_VERSION)' -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit=$(BUILD_COMMIT)'" cmd/server/*.go
build-agent:
	go build -o bin/agent -ldflags "-X 'main.buildVersion=$(BUILD_VERSION)' -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit=$(BUILD_COMMIT)'" cmd/agent/*.go
build-rsa:
	go build -o bin/rsa  cmd/rsa/*.go
rsa:
	./bin/rsa --name metric --length 4096
clear:
	rm -rf bin/*