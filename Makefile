# Compilers and stuff
GO?=go
GOFMT?=gofmt
STRIP?=strip --strip-all
UPX?=upx --lzma -9

# Vars
GLUE=$(wildcard app/*)

all: carbon-app
carbon-app: glue/generated_glue.go fmt
	$(GO) build -o $@ -v ./carbon-app.go

glue/generated_glue.go: $(GLUE)
	$(GO) generate

fmt:
	$(GOFMT) -w -s .

test:
	$(GO) test -v ./...

clean:
	rm -f carbon-app

.PHONY: carbon-app test

# Convenience stuff
dist: carbon
	$(STRIP) carbon-app
	$(UPX) carbon-app
