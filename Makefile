GOBIN := $(GOPATH)/bin
LIBNAME := redic

build:
	CGO_ENABLED=1 go build -tags "json1 fts5 foreign_keys math_functions" -o $(LIBNAME)

install: build
	cp $(LIBNAME) $(GOBIN)/$(LIBNAME)

rebuild: install
	redic create-tables --adapter sqlite3 -n dictionary.sqlite --time
	redic reindex --adapter sqlite3 -n dictionary.sqlite -d ./wordnet/english --time
