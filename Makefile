GOARCH=arm
GOARM=6

all:
	$(MAKE) build
	$(MAKE) copy

deps:
	go get

build:
	GOARCH=$(GOARCH) GOARM=$(GOARM) go build receive.go
