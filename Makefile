.PHONY: build install clean run test

BINARY=delvop

build:
	go build -o $(BINARY) .

install: build
	cp $(BINARY) $(GOPATH)/bin/$(BINARY) 2>/dev/null || cp $(BINARY) ~/go/bin/$(BINARY)

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY)

test:
	go test ./... -v -coverprofile=coverage.txt -covermode=atomic
