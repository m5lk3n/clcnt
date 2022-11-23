# default target
.PHONY: help
help:
	@echo "usage: make <target>"
	@echo
	@echo "  where <target> is one of the following"
	@echo
	@echo "    clean       to delete the go module"
	@echo "    init        to initialize the module"
	@echo "    get         to fetch all package dependencies"
	@echo "    build       to compile binary (for linux amd64 architecture)"
	@echo "    all         to run all targets"
	@echo
	@echo "    help        to show this text"

.PHONY: clean
clean:
	rm -f go.mod

.PHONY: init
init:
	go mod init lttl.dev/clcnt

.PHONY: get
get:
	# go get github.com/gin-gonic/contrib/static
	go get github.com/gin-gonic/gin
	go get github.com/sirupsen/logrus
	go get -u github.com/mattn/go-sqlite3
	
.PHONY: build
build:
	# GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build
	go build

.PHONY: all
all: clean init get build