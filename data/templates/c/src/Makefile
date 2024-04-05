BUILD_PATH ?= /app/build
RM          = rm -rf
BIN_PATH    = $(BUILD_PATH)/bin
PATHS       = $(BIN_PATH)

$(BIN_PATH)/{{.app.name}}: | $(PATHS)
	$(CC) $(CFLAGS) -o $@ {{.app.name}}.c

$(PATHS):
	mkdir -p $@

clean:
	$(RM) $(BIN_PATH)

.PHONY: clean
