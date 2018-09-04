
OUTFOLDER = $(PWD)/bin

GOPATH = $(PWD)/.go

CGO_ENABLED=0

GO_COMPILER_OPTS = -a -tags netgo -ldflags '-w -extldflags "-static"'

build: $(OUTFOLDER) fmt
	GOOS=linux GOARCH=amd64 go build \
		$(GO_COMPILER_OPTS) \
		-o $(OUTFOLDER)/iget \
		./cmd/main.go
	@echo 'built'

fmt:
	find . -name '*.go' -exec gofmt -w {} \;

$(OUTFOLDER):
	mkdir -p $(OUTFOLDER)

test:
	go test

clean:
	rm -rf bin

deps:
	go get -u github.com/eyedeekay/gosam
	go get -u github.com/eyedeekay/iget
