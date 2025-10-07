APP_NAME := todox
BUILD_DIR := ./bin
# change to /usr/local/bin for system-wide
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: all build install clean

all: build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) main.go

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)
	chmod +x $(INSTALL_DIR)/$(APP_NAME)
	@echo "Installed $(APP_NAME) to $(INSTALL_DIR)"

clean:
	rm -rf $(BUILD_DIR)

remove:
	rm -rf $(INSTALL_DIR)/$(APP_NAME)
	@echo "Removed $(APP_NAME) from $(INSTALL_DIR)"
