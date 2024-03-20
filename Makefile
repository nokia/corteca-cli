VERSION := $(shell git describe --tags)
BUILD_FLAGS := -x -v -ldflags "-X corteca/cmd.appVersion=$(VERSION)"
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

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

install: $(GOOS)-target

uninstall:
	@if [ -e $(DESTDIR)/usr/bin/$(BINARY_NAME) ]; then unlink $(DESTDIR)/usr/bin/$(BINARY_NAME) && echo "removed '$(DESTDIR)/usr/bin/$(BINARY_NAME)' symlink"; fi
	@if [ -e $(DESTDIR)/opt/corteca/$(BINARY_NAME) ]; then rm -v $(DESTDIR)/opt/corteca/$(BINARY_NAME) && rmdir -v $(DESTDIR)/opt/corteca; fi
	rm -rfv $(DESTDIR)/etc/corteca

$(PACKAGES):
	mkdir -p $(PACKAGES)

deb: GOOS := linux
deb: DESTDIR := $(TMP)/$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)
deb: PKGNAME := $(BINARY_NAME)_$(VERSION)_$(GOARCH).deb
deb: $(GOOS)-target | $(PACKAGES)
	fpm -f -s dir \
		-t deb \
		-C "$(DESTDIR)" \
		--name corteca-cli \
		--version $(VERSION) \
		--iteration 1 \
		--description "Corteca Developer Toolkit cli" \
		--no-deb-generate-changes \
		--package $(PACKAGES)/$(PKGNAME) \
		--depends "docker.io | docker-ce | podman-docker" \
		--architecture $(GOARCH) \
		--maintainer "ContainerApplicationsTeamAthens@groups.nokia.com" \
		--url "https://nokia.com"

rpm: GOOS := linux
rpm: DESTDIR := $(TMP)/$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)
rpm: PKGNAME := $(BINARY_NAME)_$(VERSION)_$(GOARCH).rpm
rpm: $(GOOS)-target | $(PACKAGES)
	fpm -f -s dir \
		-t rpm \
		-C "$(DESTDIR)" \
		--name corteca-cli \
		--version $(VERSION) \
		--iteration 1 \
		--description "Corteca Developer Toolkit cli" \
		--package $(PACKAGES)/$(PKGNAME) \
		--architecture $(GOARCH) \
		--maintainer "ContainerApplicationsTeamAthens@groups.nokia.com" \
		--url "https://nokia.com"

osx: GOOS := darwin
osx: DESTDIR := $(TMP)/$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)
osx: PKGNAME := $(BINARY_NAME)_$(VERSION)_$(GOARCH).osxpkg
osx: $(GOOS)-target | $(PACKAGES)
	(cd $(TMP) && zip -r $(PACKAGES)/$(BINARY_NAME)_$(VERSION)_$(GOARCH).zip "$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)")
# 	mkdir -p $(PACKAGES)
#	@echo This can only run on OSX
#	 fpm -f -s dir \
#		-t osxpkg \
#		-C "$(DESTDIR)" \
#		--name corteca-cli \
#		--version $(VERSION) \
#		--iteration 1 \
#		--description "Corteca Developer Toolkit cli" \
#		--package $(PACKAGES)/$(PKGNAME) \
#		--architecture $(GOARCH) \
#		--maintainer "ContainerApplicationsTeamAthens@groups.nokia.com" \
#		--url "https://nokia.com" .

msi: GOOS := windows
msi: DESTDIR := $(TMP)/$(BINARY_NAME)_$(GOOS)_$(VERSION)_$(GOARCH)
msi: PKGNAME := $(BINARY_NAME)_$(VERSION)_$(GOARCH).msi
msi: GUID := $(shell uuidgen)
msi: INSTALLER_XML := corteca.wxs
msi: DEST_REL_PATH := $(shell realpath --relative-to=$(CURRDIR) $(DESTDIR))
msi: $(GOOS)-target | $(PACKAGES)
	mv $(DESTDIR)/opt/corteca/$(BINARY_NAME) $(DESTDIR)/opt/corteca/$(BINARY_NAME).exe

	@echo Generating XML files for corteca components...
	find $(DESTDIR)/opt/corteca | wixl-heat -p $(DESTDIR)/opt/corteca/ --component-group CortecaExeComponentGroup --var var.SourceDir \
	--directory-ref=INSTALLFOLDER > $(DESTDIR)/exec.wxs
	find $(DESTDIR)/etc/corteca | wixl-heat -p $(DESTDIR)/etc/corteca/ --component-group CortecaConfigComponentGroup --var var.SourceConfigDir \
	--directory-ref=PROGRAMDATADIR > $(DESTDIR)/config.wxs

	@echo Building MSI...
	wixl -v -o $(PACKAGES)/$(PKGNAME) $(INSTALLER_XML) $(DESTDIR)/exec.wxs $(DESTDIR)/config.wxs -D SourceDir="$(DEST_REL_PATH)/opt/corteca" -D SourceConfigDir="$(DEST_REL_PATH)/etc/corteca" -D Guid="$(GUID)" -D Version="$(VERSION)" --arch x64

all-packages: msi deb rpm osx 

.PHONY: all-packages osx rpm deb msi win-binary osx-binary linux-binary uninstall install distclean clean test linux-target windows-target darwin-target
