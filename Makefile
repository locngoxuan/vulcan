.PHONY: clean

PWD=$(shell pwd)

default: clean toolchains vlocal

clean:
	@rm -rf output

toolchains: 
	@mkdir -p output/vulcan/toolchains
	docker run -it --rm \
	-v $(PWD)/builtin:/app/builtin \
	-v $(PWD)/cmd:/app/cmd \
	-v $(PWD)/core:/app/core \
	-v $(PWD)/go.mod:/app/go.mod \
	-v $(PWD)/output:/app/output \
	-v $(PWD)/toolchains.sh:/app/toolchains.sh \
	--workdir=/app \
	golang:1.16.5-alpine3.13 /bin/sh -c ./toolchains.sh

vlocal:
	@mkdir -p output/vulcan/bin	
	go mod tidy
	go get -v ./cmd/vlocal
	go build -ldflags="-s -w" -o ./output/vulcan/bin/vlocal ./cmd/vlocal

install:
	mkdir -p ~/.vulcan
	cp -r ./output/vulcan/* ~/.vulcan/