# Luis Figueiredo (luisf@casemyway.com)
# 

#GOPATH=$(CURDIR)/deps:$(CURDIR)
DIST=./bin
ENV=

# Source in packages names
# What packages to build

# CLI Packages
PKG=github.com/gohxs/hsterm/example

BIN=example
# Windows build
ifeq ($(GOOS),windows)
	ENV+=CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH)
	BIN:=$(addsuffix .exe, $(BIN))
endif

.PHONY: all deps clean dist-clean 

all:
	@$(ENV) echo -e "\e[01mBuilt for OS: `go env GOOS`, ARCH: `go env GOARCH`\e[0m"
	$(ENV) GOPATH="$(GOPATH)" go build -o $(DIST)/$(BIN)
	

# generate
generate:
	GOPATH="$(GOPATH)" go generate $(PKG)/...
	#folder2go -handler ../frontend/hmi/dist data src/mt3hmi/cli/server/data

test:
	$(ENV) go test $(PKG)/...

#$(BIN): $(addprefix src/,$(SOURCE))
#	echo $<

deps:
	$(ENV) GOPATH="$(GOPATH)" go get -v $(PKG)/... # everything from source
	# test package 
	$(ENV) GOPATH="$(GOPATH)" go get -v -t $(PKG)/...

clean:
	rm -rf $(DIST)

dist-clean: clean
	#rm -rf bin
	#rm -rf pkg
	#rm -rf deps

$(DIST):
	mkdir -p $(DIST)


