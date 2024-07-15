BUILD_PATH   ?= /app/build
BUILD_PATH 	 ?= /app/build
RM            = rm -rf
SYSROOT      ?= $(abspath /)
BIN_PATH      = $(BUILD_PATH)/bin
PATHS         = $(BIN_PATH)
INCLUDE_DIR   = $(SYSROOT)usr/include
LIBRARIES_DIR = $(SYSROOT)usr/lib
LDFLAGS       = -L$(LIBRARIES_DIR) ${LIBRARIES}
CXXFLAGS        = -I$(INCLUDE_DIR)

$(BIN_PATH)/{{.app.name}}: | $(PATHS)
	$(CXX) $(CXXFLAGS) {{.app.name}}.cpp -o $@

$(PATHS):
	mkdir -p $@

clean:
	$(RM) $(BIN_PATH)

.PHONY: clean
