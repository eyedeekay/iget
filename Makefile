
OUTFOLDER = $(PWD)/bin

GOPATH = $(PWD)/.go

CGO_ENABLED=0

GO_COMPILER_OPTS = -a -tags netgo -ldflags '-w -extldflags "-static"'

build: $(OUTFOLDER) fmt lint
	GOOS=linux GOARCH=amd64 go build \
		$(GO_COMPILER_OPTS) \
		-o $(OUTFOLDER)/iget \
		./cmd/main.go
	@echo 'built'

fmt:
	find . -path ./.go -prune -o -name '*.go' -exec gofmt -w {} \;

lint:
	find . -path ./.go -prune -o -name '*.go' -exec golint {} \;

$(OUTFOLDER):
	mkdir -p $(OUTFOLDER)

test:
	go test

clean:
	rm -rf bin

deps:
	go get -u github.com/eyedeekay/gosam
	go get -u github.com/eyedeekay/iget

README.md:
	@echo "# iget" | tee README.md
	@echo "i2p terminal http client." | tee -a README.md
	@echo ""
	@echo "## to build:"
	@echo ""
	@echo "        make deps build"
	@echo ""
	@echo "## to use:"
	@echo ""
	./bin/iget --help
	@echo ""

rcl:
	rm -f README.md

re: rcl README.md
