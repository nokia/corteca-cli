BUILD_PATH ?= /app/build
RM           = rm -rf
BIN_PATH     = $(BUILD_PATH)/bin
LIB_PATH     = $(BUILD_PATH)/lib
PATHS        = $(BIN_PATH) $(LIB_PATH)
GO_LDFLAG    = -s -w
LIBRARIES    = -lubus -lubox -lblobmsg_json
CGO_ENABLED = 0

$(BIN_PATH)/{{.app.name}}: | $(PATHS)
	env && go env
	CGO_ENABLED=$(CGO_ENABLED) go build -buildvcs=false -ldflags="$(GO_LDFLAGS)" -o $@ .

$(PATHS):
	@echo "Create target path $@"
	mkdir -p $@

clean:
	$(RM) $(BIN_PATH) $(LIB_PATH)

.PHONY: clean 
