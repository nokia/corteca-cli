VERSION := $(shell git describe --tags)
BUILD_FLAGS := -x -v -ldflags "-X corteca/cmd.appVersion=$(VERSION)"
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_ENV_GOOS ?= linux
BUILD_ENV_GOARCH ?= amd64

CURRDIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
DIST := $(CURRDIR)dist
BIN := $(DIST)/bin
TMP := $(shell mktemp -d)
PACKAGES := $(DIST)/packages
BINARY_NAME = corteca
FULL_BINARY_NAME = $(BINARY_NAME)-$(GOOS)-$(GOARCH)-$(VERSION)

# build main binary
$(BIN)/$(FULL_BINARY_NAME): $(TMP)
	go env
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $(BIN)/$(FULL_BINARY_NAME) main.go

default-corteca-binary:
	CGO_ENABLED=0 GOOS=$(BUILD_ENV_GOOS) GOARCH=$(BUILD_ENV_GOARCH) go build $(BUILD_FLAGS) -o $(BIN)/default-$(BINARY_NAME)-$(BUILD_ENV_GOOS)-$(BUILD_ENV_GOARCH)-$(VERSION) main.go

$(TMP):
	mkdir -p $(TMP)

# run all available tests
test:
	go test ./... -v

clean:
	rm -rfv $(BIN)
	go clean -modcache

distclean: clean
	rm -rfv $(DIST)

$(GOOS)-target: $(BIN)/$(FULL_BINARY_NAME) 
	mkdir -p $(DESTDIR)/opt/corteca $(DESTDIR)/etc/corteca $(DESTDIR)/usr/bin 
	cp -v $(BIN)/$(FULL_BINARY_NAME) $(DESTDIR)/opt/corteca/$(BINARY_NAME)
	ln -sf /opt/corteca/$(BINARY_NAME) $(DESTDIR)/usr/bin/
	cp -rv data/* $(DESTDIR)/etc/corteca/

unix-completions: default-corteca-binary
	mkdir -p $(BASH_COMPLETION_DIR) $(ZSH_COMPLETION_DIR) $(FISH_COMPLETION_DIR)
	$(BIN)/default-$(BINARY_NAME)-$(BUILD_ENV_GOOS)-$(BUILD_ENV_GOARCH)-$(VERSION) -r data/ completion bash > $(BASH_COMPLETION_DIR)/$(BINARY_NAME).bash ; \
	$(BIN)/default-$(BINARY_NAME)-$(BUILD_ENV_GOOS)-$(BUILD_ENV_GOARCH)-$(VERSION) -r data/ completion zsh >  $(ZSH_COMPLETION_DIR)/_$(BINARY_NAME); \
	$(BIN)/default-$(BINARY_NAME)-$(BUILD_ENV_GOOS)-$(BUILD_ENV_GOARCH)-$(VERSION) -r data/ completion fish > $(FISH_COMPLETION_DIR)/$(BINARY_NAME).fish; \


windows-completions: default-corteca-binary
	mkdir -p $(PS1_COMPLETION_MODULE_DIR)
	$(BIN)/default-$(BINARY_NAME)-$(BUILD_ENV_GOOS)-$(BUILD_ENV_GOARCH)-$(VERSION) -r data/ completion powershell > $(PS1_COMPLETION_MODULE_DIR)/$(BINARY_NAME_CAPITALIZED).psm1; \

install: $(GOOS)-target

uninstall:
	@if [ -e $(DESTDIR)/usr/bin/$(BINARY_NAME) ]; then unlink $(DESTDIR)/usr/bin/$(BINARY_NAME) && echo "removed '$(DESTDIR)/usr/bin/$(BINARY_NAME)' symlink"; fi
	@if [ -e $(DESTDIR)/opt/corteca/$(BINARY_NAME) ]; then rm -v $(DESTDIR)/opt/corteca/$(BINARY_NAME) && rmdir -v $(DESTDIR)/opt/corteca; fi
	rm -rfv $(DESTDIR)/etc/corteca
	rm -fv $(BASH_COMPLETION_DIR)/$(BINARY_NAME).bash
	rm -fv $(ZSH_COMPLETION_DIR)/_$(BINARY_NAME)
	rm -fv $(FISH_COMPLETION_DIR)/$(BINARY_NAME).fish

$(PACKAGES):
	mkdir -p $(PACKAGES)

deb rpm: GOOS := linux
deb rpm: DESTDIR := $(TMP)/$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)
deb rpm: BASH_COMPLETION_DIR := ${DESTDIR}/etc/bash_completion.d
deb rpm: ZSH_COMPLETION_DIR := ${DESTDIR}/usr/share/zsh/site-functions
deb rpm: FISH_COMPLETION_DIR := ${DESTDIR}/usr/share/fish/completions
deb rpm: PKGNAME := $(BINARY_NAME)_$(VERSION)_$(GOARCH)
deb rpm: $(GOOS)-target | $(PACKAGES)
deb rpm: unix-completions
	VERSION=$(VERSION) \
	GOARCH=$(GOARCH) \
	DESTDIR=$(DESTDIR) \
	envsubst < nfpm.yaml.template > nfpm.yaml
	nfpm pkg --config nfpm.yaml --packager $@ --target $(PACKAGES)/$(PKGNAME).$@

osx: GOOS := darwin
osx: DESTDIR := $(TMP)/$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)
osx: BASH_COMPLETION_DIR := ${DESTDIR}/etc/bash_completion.d
osx: ZSH_COMPLETION_DIR := ${DESTDIR}/usr/share/zsh/site-functions
osx: FISH_COMPLETION_DIR := ${DESTDIR}/usr/share/fish/completions
osx: PKGNAME := $(BINARY_NAME)_$(VERSION)_$(GOARCH).osxpkg
osx: $(GOOS)-target | $(PACKAGES)
osx: unix-completions
	(cd $(TMP) && zip -r $(PACKAGES)/$(BINARY_NAME)_$(VERSION)_$(GOARCH).zip "$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)")

msi: GOOS := windows
msi: DESTDIR := $(TMP)/$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)
msi: PKGNAME := $(BINARY_NAME)_$(VERSION)_$(GOARCH).msi
msi: GUID := $(shell uuidgen)
msi: INSTALLER_XML := corteca.wxs
msi: DEST_REL_PATH := $(shell realpath --relative-to=$(CURRDIR) $(DESTDIR))
msi: BINARY_NAME_CAPITALIZED := $(shell echo $(BINARY_NAME) | awk '{print toupper(substr($$0,1,1)) tolower(substr($$0,2))}')
msi: PS1_COMPLETION_DIR := ${DESTDIR}/completions
msi: PS1_COMPLETION_MODULE_DIR := $(PS1_COMPLETION_DIR)/$(BINARY_NAME_CAPITALIZED)
msi: $(GOOS)-target | $(PACKAGES)
msi: windows-completions
	mv $(DESTDIR)/opt/corteca/$(BINARY_NAME) $(DESTDIR)/opt/corteca/$(BINARY_NAME).exe

	@echo Generating XML files for corteca components...
	find $(DESTDIR)/opt/corteca | wixl-heat -p $(DESTDIR)/opt/corteca/ --component-group CortecaExeComponentGroup --var var.SourceDir \
	--directory-ref=INSTALLFOLDER > $(DESTDIR)/exec.wxs
	find $(PS1_COMPLETION_DIR) | wixl-heat -p $(PS1_COMPLETION_DIR)/ --component-group CortecaPSCompletionComponentGroup --var var.SourceModuleDir \
	--directory-ref=PSMODULECOMPLETIONS > $(DESTDIR)/autocomplete.wxs
	find $(DESTDIR)/etc/corteca | wixl-heat -p $(DESTDIR)/etc/corteca/ --component-group CortecaConfigComponentGroup --var var.SourceConfigDir \
	--directory-ref=PROGRAMDATADIR > $(DESTDIR)/config.wxs

	@echo Building MSI...
	wixl -v -o $(PACKAGES)/$(PKGNAME) $(INSTALLER_XML) $(DESTDIR)/exec.wxs $(DESTDIR)/config.wxs $(DESTDIR)/autocomplete.wxs -D SourceDir="$(DEST_REL_PATH)/opt/corteca" -D SourceModuleDir="$(DEST_REL_PATH)/completions" -D SourceConfigDir="$(DEST_REL_PATH)/etc/corteca" -D Guid="$(GUID)" -D Version="$(VERSION)" --arch x64

all-packages: msi deb rpm osx

.PHONY: all-packages osx rpm deb msi win-binary osx-binary linux-binary uninstall install distclean clean test linux-target windows-target darwin-target
