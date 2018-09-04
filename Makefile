
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
	@echo "# iget" | tee $(PWD)/README.md
	@echo "i2p terminal http client." | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "## to build:" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "        make deps build" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "## to use:" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo '```' | tee -a $(PWD)/README.md
	./bin/iget -h 2>&1 | tee -a $(PWD)/README.md
	@echo '```' | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md

rcl:
	rm -f $(PWD)/README.md

re: rcl README.md

fire:
	./bin/iget -url http://i2p-projekt.i2p
