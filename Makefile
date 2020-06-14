.PHONY: test testcover

test:
	go test -v github.com/han-so1omon/graphtools/structures

testcover:
	go test -coverprofile graphtools-structures-coverage.html -v github.com/han-so1omon/graphtools/structures
	go tool cover -html=graphtools-structures-coverage.html
