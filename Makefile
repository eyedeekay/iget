
OUTFOLDER = $(PWD)/iget

GOPATH = $(PWD)/.go

CGO_ENABLED=0

GO_COMPILER_OPTS = -a -tags netgo -ldflags '-w -extldflags "-static"'

build: $(OUTFOLDER) fmt lint
	GOOS=linux GOARCH=amd64 go build \
		$(GO_COMPILER_OPTS) \
		-o $(OUTFOLDER)/iget \
		$(OUTFOLDER)/main.go
	@echo 'built'

install:
	install -m755 bin/iget /usr/bin/iget

install-wrapper:
	ls $(shell which eepget).bak || which eepget && mv $(shell which eepget) $(shell which eepget).bak; true
	install -m755 bin/eepget-wrapper.sh /usr/bin/eepget-wrapper.sh
	ln -sf /usr/bin/eepget-wrapper.sh /usr/bin/eepget

fmt:
	find . -path ./.go -prune -o -name '*.go' -exec gofmt -w {} \;

lint:
	find . -path ./.go -prune -o -name '*.go' -exec golint {} \;

# This is just to make sure that I don't leave unnecessary crap behind in the code.
checkuses:
	find . -path ./.go -prune -o -name '*.go' -exec grep Do {} \;

$(OUTFOLDER):
	mkdir -p $(OUTFOLDER)

test:
	go test

clean:
	rm -f bin/iget

deps:
	go get -u github.com/eyedeekay/goSam
	go get -u github.com/eyedeekay/iget

README.md:
	@echo "# iget" | tee $(PWD)/README.md
	@echo "i2p terminal http client." | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "## description:" | tee -a $(PWD)/README.md
	@echo "This is a highly-configurable curl/wget like client which exclusively works" | tee -a $(PWD)/README.md
	@echo "over i2p. It works via the SAM API which means it has some advantages and" | tee -a $(PWD)/README.md
	@echo "some disadvantages, as follows:" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "### Advantages:" | tee -a $(PWD)/README.md
	@echo "These advantages motivated development. More may emerge as it continues." | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "  - uses the SAM API to prevent destination-reuse for different sites" | tee -a $(PWD)/README.md
	@echo "  - uses the SAM API directly(not forwarding) so it can't leak information" | tee -a $(PWD)/README.md
	@echo "    to clearnet services" | tee -a $(PWD)/README.md
	@echo "  - inline options to configure i2cp, so for example we can have 8 tunnels" | tee -a $(PWD)/README.md
	@echo "    in and 2 tunnels out" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "### Disadvantages:" | tee -a $(PWD)/README.md
	@echo "Only two I know of so far." | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "  - marginally slower, due to tunnel-creation at runtime." | tee -a $(PWD)/README.md
	@echo "  - a few missing options compared to eepget(These are being handled by" | tee -a $(PWD)/README.md
	@echo "    including a wrapper, which will be fully compatible with eepget)" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "Wherever possible, short arguments will mirror their curl equivalents." | tee -a $(PWD)/README.md
	@echo "However, I'm not trying to implement every single curl option, and if" | tee -a $(PWD)/README.md
	@echo "there are arguments that are labeled differently between curl and eepget," | tee -a $(PWD)/README.md
	@echo "eepget options will be used instead. I haven't decided if I want it to be" | tee -a $(PWD)/README.md
	@echo "able to spider eepsites on it's own, but I'm leaning toward no. That's what" | tee -a $(PWD)/README.md
	@echo "lynx and grep are for." | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "## to build:" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "        make deps build" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo "## to use:" | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md
	@echo '```' | tee -a $(PWD)/README.md
	$(OUTFOLDER)/iget -help 2>&1 | tee -a $(PWD)/README.md
	@echo '```' | tee -a $(PWD)/README.md
	@echo "" | tee -a $(PWD)/README.md

rcl:
	rm -f $(PWD)/README.md

re: rcl README.md

fire:
	./bin/iget -url http://i2p-projekt.i2p

fire2:
	./bin/iget http://566niximlxdzpanmn4qouucvua3k7neniwss47li5r6ugoertzuq.b32.i2p

fire3:
	eepget -l 80 "http://566niximlxdzpanmn4qouucvua3k7neniwss47li5r6ugoertzuq.b32.i2p"

fire4:
	eepget -p 127.0.0.1:4444 -l 80 "http://566niximlxdzpanmn4qouucvua3k7neniwss47li5r6ugoertzuq.b32.i2p"
