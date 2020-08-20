CGO_ENABLED=0

.PHONY: all get

all : parser

get:
	@go get -v -d ./gethparser/...

parser: | get
	@go build -o parser ./gethparser