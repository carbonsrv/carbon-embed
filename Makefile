# Compilers and stuff
GO?=go
GO_BINDATA?=go-bindata
GOFMT?=gofmt
STRIP?=strip --strip-all
UPX?=upx --lzma -9

# Vars
APP_DIR?=./app
GLUE_DIRS=$(shell find $(APP_DIR) -type d | grep -v ".git")
GLUE_FILES=$(shell find $(APP_DIR) -type f | grep -v ".git")
GLUE_OUTPUT=glue/generated_glue.go

all: carbon-app
carbon-app: $(GLUE_OUTPUT) fmt
	$(GO) build -o $@ -v ./carbon-app.go

$(GLUE_OUTPUT): $(GLUE_FILES)
	$(GO_BINDATA) -nomemcopy -o $(GLUE_OUTPUT) -pkg=glue -prefix $(APP_DIR) $(GLUE_DIRS)
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
