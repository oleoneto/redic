build:
	CGO_ENABLED=1 go build -tags "json1 fts5 foreign_keys math_functions" 

run: build
	./redic

search: build
	./redic search brother spouse -o yaml | yq

define: build
	./redic define mother -o yaml | yq

load: build
	./redic build -n redic.db -d ./wordnet/english
