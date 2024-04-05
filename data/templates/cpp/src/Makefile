BUILD_PATH ?= /app/build
RM          = rm -rf
BIN_PATH    = $(BUILD_PATH)/bin
PATHS       = $(BIN_PATH)

$(BIN_PATH)/{{.app.name}}: | $(PATHS)
	$(CXX) $(CXXFLAGS) -o $@ {{.app.name}}.cpp

$(PATHS):
	mkdir -p $@

clean:
	$(RM) $(BIN_PATH)

.PHONY: clean
