VERSION := $(shell cat main.go | grep "const VERSION" | awk -F "\"" '{print $$2}')

.PHONY: build

build:
	go build -o build/emftoimg

deps:
	go get -t -u -v ./...

clean:
	rm -rf build/
