CURDIR=$(shell pwd)
TARGET=deepinid-data-migrate

build: oauth user

rebuild: clean build

clean: 
	rm -rfv $(CURDIR)/cmd/user/deepin-migrate-users; rm -rfv $(CURDIR)/cmd/oauth/deepin-migrate-oauth

user:
	go build -mod vendor -o $(CURDIR)/cmd/user/deepin-migrate-users $(CURDIR)/cmd/user/main.go

oauth:
	go build -mod vendor -o $(CURDIR)/cmd/oauth/deepin-migrate-oauth $(CURDIR)/cmd/oauth/main.go

deal:
	go build -mod vendor -o $(CURDIR)/cmd/deal/deepin-migrate-deal $(CURDIR)/cmd/deal/main.go