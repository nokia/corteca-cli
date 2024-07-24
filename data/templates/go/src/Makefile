BUILD_PATH    ?= /app/build
RM             = rm -rf
DEBUG_MODE    ?= false
SYSROOT       ?= $(abspath /)
GO_HLAPI_PATH ?= $(abspath /usr/src/go/hlapi)
BIN_PATH       = $(BUILD_PATH)/bin
LIB_PATH       = $(BUILD_PATH)/lib
PATHS          = $(BIN_PATH) $(LIB_PATH)
GO_LDFLAGS     = -s -w
INCLUDE_DIR    = $(SYSROOT)usr/include
LIBRARIES_DIR  = $(SYSROOT)usr/lib
CGO_LDFLAGS    = "-L${LIBRARIES_DIR} $(LIBRARIES)"
CGO_CFLAGS     = "-I$(INCLUDE_DIR)"
CGO_ENABLED    = 1

ifeq ($(DEBUG_MODE),true)
	GO_EXTRA_FLAGS = 
else 
	GO_EXTRA_FLAGS = -ldflags="$(GO_LDFLAGS)"  
endif

$(BIN_PATH)/{{ .app.name }}: | $(PATHS)
	env && go env
	CGO_ENABLED=$(CGO_ENABLED) CGO_CFLAGS=$(CGO_CFLAGS) go build -buildvcs=false ${GO_EXTRA_FLAGS} -o $@ .

$(PATHS):
	@echo "Create target path $@"
	mkdir -p $@


clean:
	$(RM) $(BIN_PATH) $(LIB_PATH)

.PHONY: clean
