GOBIN := $(GOPATH)/bin
LIBNAME := redic

build:
	CGO_ENABLED=1 go build -tags "json1 fts5 foreign_keys math_functions" -ldflags "-X main.BuildHash=`git rev-parse HEAD`" -o $(LIBNAME)

install: build
	cp $(LIBNAME) $(GOBIN)/$(LIBNAME)

run: install
	redic server --verbose

init: install
	redic init --time

reset: install
	redic init --repopulate --reset-tables --time
