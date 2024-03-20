export BUILD_PATH ?= /app/build
RM                := rm -rf
TESTS              = tests/functional tests/product

build:
	@echo "Building application"
	$(MAKE) -C src/

$(TESTS):
	cd $@ && ./tests.sh

distclean:
	$(RM) build/ dist/

clean:
	$(MAKE) -C src/ clean
	
precompiled:
	@echo "You can add here the commands to download precompiled binaries"
	@echo "Put architecture agnostic files in './target/noarch' (scripts, configuration, etc)"
	@echo "put architecture specific in the respective folder eg:'./target/armv8' (precompiled binaries ot libraries, etc)"
	@echo "in order for them to be included in the target root fs"

docs:
	@echo "You can add here the commands to generate the documentation if needed"

.PHONY: build $(TESTS) clean precompiled docs distclean
