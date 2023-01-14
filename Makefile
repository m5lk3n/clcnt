# default target
.PHONY: help
help:
	@echo "usage: make <target>"
	@echo
	@echo "  where <target> is one of the following"
	@echo
	@echo "    clean       to delete the go module and the binary"
	@echo "    init        to initialize the module"
	@echo "    get         to fetch all package dependencies"
	@echo "    test        to run all tests"
	@echo "    run         to run the code without binary compilation"
	@echo "    build       to compile a self-contained binary (for the local architecture)"
	@echo "    all         to run all targets but test and run"
	@echo
	@echo "    help        to show this text"

.PHONY: clean
clean:
	rm -f go.mod
	rm -f clcnt

.PHONY: init
init:
	go mod init lttl.dev/clcnt

.PHONY: get
get:
	# go mod tidy
	go get github.com/gin-gonic/contrib/static
	go get github.com/gin-gonic/gin
	go get github.com/sirupsen/logrus
	go get github.com/mattn/go-sqlite3
	go get github.com/swaggo/gin-swagger

.PHONY: test
test:
	mv clcnt.db clcnt.bak 2>/dev/null || true
	rm -f clcnt-test.db
	go test
	mv clcnt.bak clcnt.db 2>/dev/null || rm -f clcnt.db

.PHONY: needs_swag # checks existence of required tool, fails if not available
needs_swag:
	swag > /dev/null

.PHONY: init_swag
init_swag: needs_swag
	swag init

.PHONY: run
run: init_swag
	go run main.go -debug

.PHONY: build
build: init_swag
	CGO_ENABLED=1 go build

.PHONY: all
all: clean init get build
