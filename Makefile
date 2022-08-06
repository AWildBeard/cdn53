BUILD=go build
LDFLAGS=-X main.buildType=debug
DATE=$(shell date '+%s')
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)
GOARM=
ifeq "$(strip $(GOOS))" "darwin"
NPROC=$(shell sysctl -n hw.ncpu)
else
NPROC=$(shell nproc)
endif
ifndef OUTSUFFIX
OUTSUFFIX=${GOOS}-${GOARCH}
endif
ifndef OUT
OUT=$(shell pwd)/build/$(shell basename $(PWD))_${OUTSUFFIX}
endif

# Text coloring & styling
BOLD=\033[1m
UNDERLINE=\033[4m
HEADER=${BOLD}${UNDERLINE}

GREEN=\033[38;5;118m
RED=\033[38;5;196m
GREY=\033[38;5;250m

RESET=\033[m

# Targets
all: build

l: lint
lint:
	@printf "${GREEN}${HEADER}Linting${RESET}\n"
	go vet ./...

d: deps
deps: dependencies
dependencies:
	@printf "${GREEN}${HEADER}Downloading dependencies${RESET}\n"
	go get ./...

f: format
format:
	@printf "${GREEN}${HEADER}Formatting${RESET}\n"
	go fmt ./...

t: test
test:
	@printf "${GREEN}${HEADER}Starting test suite${RESET}\n"
	go test -parallel ${NPROC} ./...

release:
	$(eval LDFLAGS=-w -s -X main.buildType=release)
	@:

b: build 
build: clean
	$(eval LDFLAGS=${LDFLAGS} -X main.buildVersion=${DATE})
	@printf "${GREEN}${HEADER}Compiling for ${GOARCH}-${GOOS} to '${OUT}'${RESET}\n"
	mkdir -p $(shell dirname ${OUT})
	GOARM=${GOARM} GOARCH=${GOARCH} GOOS=${GOOS} ${BUILD} -p ${NPROC} -ldflags="${LDFLAGS}" -o ${OUT}

clean:
	@printf "${GREEN}${HEADER}Cleaning previous build${RESET}\n"
	rm -rf ${OUT}

# OS presets
darwin:
	$(eval GOOS=darwin)
	@:
dragonfly:
	$(eval GOOS=dragonfly)
	@:
freebsd:
	$(eval GOOS=freebsd)
	@:
linux:
	$(eval GOOS=linux)
	@:
nacl:
	$(eval GOOS=nacl)
	@:
netbsd:
	$(eval GOOS=netbsd)
	@:
openbsd:
	$(eval GOOS=openbsd)
	@:
plan9:
	$(eval GOOS=plan9)
	@:
solaris:
	$(eval GOOS=solaris)
	@:
windows:
	$(eval GOOS=windows)
	@:

# Architectures
ppc64:
	$(eval GOARCH=ppc64)
	@:
ppc64le:
	$(eval GOARCH=ppc64le)
	@:
mips:
	$(eval GOARCH=mips)
	@:
mipsle:
	$(eval GOARCH=mipsle)
	@:
mips64:
	$(eval GOARCH=mips64)
	@:
mips64le:
	$(eval GOARCH=mips64le)
	@:
386:
	$(eval GOARCH=386)
	@:
amd64:
	$(eval GOARCH=amd64)
	@:
amd64p32:
	$(eval GOARCH=amd64p32)
	@:
arm:
	$(eval GOARCH=arm)
	@:
7:
	$(eval GOARM=7)
	@:
6:
	$(eval GOARM=6)
	@:
5:
	$(eval GOARM=5)
	@:
arm64:
	$(eval GOARCH=arm64)
	@:
s390x:
	$(eval GOARCH=s390x)
	@:
