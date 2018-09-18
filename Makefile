GOOS=linux
GOARCH=amd64
ROOT_PKG=.
OUT=container-exporter-$(GOOS)-$(GOARCH)

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	  -v \
	  -installsuffix cgo \
	  -ldflags "-s -w" \
	  -o out/$(OUT) \
	  $(ROOT_PKG)
	@cd out && \
	rm -f $(OUT).tar.gz && \
	tar czf $(OUT).tar.gz --owner=0 --group=0 $(OUT)
