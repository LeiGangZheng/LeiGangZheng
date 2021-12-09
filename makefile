CURDIR=$(shell pwd)
TARGET=deepinid-data-migrate

build:
	go build -mod vendor -o $(TARGET) $(CURDIR)/cmd/main.go