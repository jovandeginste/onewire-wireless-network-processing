BINARY=receive

ALL_ARCHS=lnx64 rpi rpi2 win64

all:
	-$(MAKE) deps
	$(MAKE) build-all

deps:
	go get

build-all: $(foreach arch,$(ALL_ARCHS),build-$(arch))

build:
	go build -o "bin/$(BINARY).$(SUFFIX)"
	file "bin/$(BINARY).$(SUFFIX)"

build-lnx64:
	GOARCH=amd64 GOOS=linux SUFFIX=$(subst build-,,$(@)) $(MAKE) build

build-rpi:
	GOARCH=arm GOARM=6 SUFFIX=$(subst build-,,$(@)) $(MAKE) build

build-rpi2:
	GOARCH=arm GOARM=7 SUFFIX=$(subst build-,,$(@)) $(MAKE) build

build-win64:
	GOOS=windows GOARCH=amd64 SUFFIX=$(subst build-,,$(@)).exe $(MAKE) build
