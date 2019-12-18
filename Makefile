.PHONY: build raspi-build

GOCMD=go
GOBUILD=$(GOCMD) build

build:
	$(GOBUILD) -o hankyu_service_bot main.go
raspi-build:
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o hankyu_service_bot main.go
