GOBIN := $(GOPATH)/bin
LIBNAME := redic

build:
	CGO_ENABLED=1 go build -tags "json1 fts5 foreign_keys math_functions" -o $(LIBNAME)

install: build
	cp $(LIBNAME) $(GOBIN)/$(LIBNAME)

server: install
	redic server --verbose

init: install
	redic init --time

reset: install
	redic init --repopulate --reset-tables --time
