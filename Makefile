# Compilers and stuff
GO?=go
GO_BINDATA?=go-bindata
GOFMT?=gofmt
STRIP?=strip --strip-all
UPX?=upx --lzma -9

# Vars
OUTPUT?=carbon-app
GLUE_DIRS=$(shell find ./app -type d | grep -v ".git")
GLUE_FILES=$(shell find ./app -type f | grep -v ".git")
GLUE_OUTPUT=glue/generated_glue.go

all: $(OUTPUT)
$(OUTPUT): $(GLUE_OUTPUT) fmt
	$(GO) build -o $@ -v ./carbon-app.go

$(GLUE_OUTPUT): $(GLUE_FILES)
	$(GO_BINDATA) -nomemcopy -o $(GLUE_OUTPUT) -pkg=glue -prefix ./app $(GLUE_DIRS)
	$(GOFMT) -w -s glue

fmt:
	$(GOFMT) -w -s .

test:
	$(GO) test -v ./...

clean:
	rm -f $(OUTPUT)

.PHONY: $(OUTPUT) test

# Convenience stuff
dist: $(OUTPUT)
	$(STRIP) $(OUTPUT)
	$(UPX) $(OUTPUT)
