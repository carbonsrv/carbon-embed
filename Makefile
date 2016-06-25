# Compilers and stuff
GO?=go
GO_BINDATA?=go-bindata
GOFMT?=gofmt
STRIP?=strip --strip-all
UPX?=upx --lzma -9

# Vars
GLUE_DIRS=$(shell find ./app -type d | grep -v ".git")
GLUE_FILES=$(shell find ./app -type f | grep -v ".git")
GLUE_OUTPUT=glue/generated_glue.go

all: carbon-app
carbon-app: $(GLUE_OUTPUT) fmt
	$(GO) build -o $@ -v ./carbon-app.go

$(GLUE_OUTPUT): $(GLUE_FILES)
	$(GO_BINDATA) -nomemcopy -o $(GLUE_OUTPUT) -pkg=glue -prefix "./builtin" $(GLUE_DIRS)
	$(GOFMT) -w -s modules/glue

fmt:
	$(GOFMT) -w -s .

test:
	$(GO) test -v ./...

clean:
	rm -f carbon-app

.PHONY: carbon-app test

# Convenience stuff
dist: carbon-app
	$(STRIP) carbon-app
	$(UPX) carbon-app
