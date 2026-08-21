VERSION := $(shell git describe --tags)
BUILD_FLAGS := -x -v -ldflags "-X corteca/cmd.appVersion=$(VERSION) -s"
HOSTOS ?= $(shell go env GOOS)
HOSTARCH ?= $(shell go env GOARCH)
DESTOS ?= $(HOSTOS)
DESTARCH ?= $(HOSTARCH)

CURRDIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
DIST := $(CURRDIR)dist
BIN := $(DIST)/bin
TMP := $(shell mktemp -d)
PACKAGES := $(DIST)/packages
FULL_BINARY_NAME = corteca-$(DESTOS)-$(DESTARCH)-$(VERSION)
BASH_COMPLETION_DIR = etc/bash_completion.d
ZSH_COMPLETION_DIR = usr/share/zsh/site-functions
FISH_COMPLETION_DIR = usr/share/fish/completions
PS1_COMPLETION_DIR = completions

# build main binary (for destination os/architecture)
$(BIN)/$(FULL_BINARY_NAME): $(TMP)
	mkdir -p $(BIN)
	go env
	CGO_ENABLED=0 GOOS=$(DESTOS) GOARCH=$(DESTARCH) go build $(BUILD_FLAGS) -o $(BIN)/$(FULL_BINARY_NAME) main.go

# build host binary (to generate completions for destination filesystems)
host-corteca-binary:
	CGO_ENABLED=0 GOOS=$(HOSTOS) GOARCH=$(HOSTARCH) go build $(BUILD_FLAGS) -o $(BIN)/host-corteca main.go

$(TMP):
	mkdir -p $(TMP)

# run all available tests
test:
	go test ./... -v -coverprofile=coverage.out -covermode=atomic && go tool cover -func=coverage.out

clean:
	rm -rfv $(BIN)
	go clean -modcache

distclean: clean
	rm -rfv $(DIST)

$(DESTOS)-target: host-corteca-binary $(BIN)/$(FULL_BINARY_NAME)
	mkdir -p $(DESTDIR)/opt/corteca $(DESTDIR)/etc/corteca $(DESTDIR)/usr/bin
	cp -v $(BIN)/$(FULL_BINARY_NAME) $(DESTDIR)/opt/corteca/corteca
	cp -rv data/* $(DESTDIR)/etc/corteca/
	mkdir -p $(DESTDIR)/$(BASH_COMPLETION_DIR) $(DESTDIR)/$(ZSH_COMPLETION_DIR) $(DESTDIR)/$(FISH_COMPLETION_DIR) $(DESTDIR)/$(PS1_COMPLETION_DIR)
	$(BIN)/host-corteca -r $(DESTDIR)/etc/corteca completion bash > $(DESTDIR)/$(BASH_COMPLETION_DIR)/corteca.bash
	$(BIN)/host-corteca -r $(DESTDIR)/etc/corteca completion zsh > $(DESTDIR)/$(ZSH_COMPLETION_DIR)/_corteca
	$(BIN)/host-corteca -r $(DESTDIR)/etc/corteca completion fish > $(DESTDIR)/$(FISH_COMPLETION_DIR)/corteca.bash
	$(BIN)/host-corteca -r $(DESTDIR)/etc/corteca completion powershell > $(DESTDIR)/$(PS1_COMPLETION_DIR)/corteca.psm1

install: $(DESTOS)-target

uninstall:
	@if [ -e $(DESTDIR)/usr/bin/corteca ]; then unlink $(DESTDIR)/usr/bin/corteca && echo "removed '$(DESTDIR)/usr/bin/corteca' symlink"; fi
	@if [ -e $(DESTDIR)/opt/corteca/corteca ]; then rm -v $(DESTDIR)/opt/corteca/corteca && rmdir -v $(DESTDIR)/opt/corteca; fi
	rm -rfv $(DESTDIR)/etc/corteca
	rm -fv $(BASH_COMPLETION_DIR)/corteca.bash
	rm -fv $(ZSH_COMPLETION_DIR)/_corteca
	rm -fv $(FISH_COMPLETION_DIR)/corteca.fish

$(PACKAGES):
	mkdir -p $(PACKAGES)

deb rpm: DESTOS := linux
deb rpm: DESTDIR := $(TMP)/corteca_$(DESTOS)_$(VERSION)_$(DESTARCH)
deb rpm: $(DESTOS)-target | $(PACKAGES)
deb rpm:
	VERSION=$(VERSION) \
	GOARCH=$(DESTARCH) \
	GOOS=$(DESTOS) \
	DESTDIR=$(DESTDIR) \
	envsubst < nfpm.yaml.template > nfpm.yaml
	nfpm pkg --config nfpm.yaml --packager $@ --target "$(PACKAGES)/corteca_$(VERSION)_$(DESTARCH).$@"
	rm nfpm.yaml

osx: DESTOS := darwin
osx: DESTDIR := $(TMP)/corteca_$(DESTOS)_$(VERSION)_$(DESTARCH)
osx: $(DESTOS)-target | $(PACKAGES)
osx:
	(cd $(TMP) && zip -r $(PACKAGES)/corteca_$(VERSION)_$(DESTOS)_$(DESTARCH).zip corteca_$(DESTOS)_$(VERSION)_$(DESTARCH))

msix: DESTOS := windows
msix: DESTDIR := $(TMP)/corteca_$(DESTOS)_$(VERSION)_$(DESTARCH)
msix: PKGNAME := corteca_$(VERSION)_$(DESTARCH).msix
msix: $(DESTOS)-target | $(PACKAGES)
msix:
	VERSION=$(VERSION) \
	GOARCH=$(DESTARCH) \
	GOOS=$(DESTOS) \
	DESTDIR=$(DESTDIR) \
	envsubst < nfpm.yaml.template > nfpm.yaml
	nfpm pkg --config nfpm.yaml --packager msix --target "$(PACKAGES)/corteca_$(VERSION)_$(DESTARCH).msix"
	rm nfpm.yaml

all-packages: msix deb rpm osx

.PHONY: all-packages osx rpm deb msix uninstall install distclean clean test linux-target windows-target darwin-target
