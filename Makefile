UNAME_S := $(shell uname -s)
BINDIR   := $(HOME)/.local/bin

.PHONY: build install uninstall clean

build:
	go build -o bin/tempod ./cmd/tempod
	go build -o bin/tempo  ./cmd/tempo

install: build
	mkdir -p $(BINDIR)
	cp bin/tempod $(BINDIR)/tempod
	cp bin/tempo  $(BINDIR)/tempo
ifeq ($(UNAME_S),Linux)
	-systemctl --user stop tempod
	mkdir -p $(HOME)/.config/systemd/user
	cp systemd/tempod.service $(HOME)/.config/systemd/user/tempod.service
	systemctl --user daemon-reload
	systemctl --user enable --now tempod
	@echo "Installed. Daemon running as systemd user service."
else ifeq ($(UNAME_S),Darwin)
	mkdir -p $(HOME)/Library/LaunchAgents
	sed 's|/Users/Shared/placeholder|$(HOME)/.local/bin/tempod|' \
		launchd/com.tempod.plist \
		> $(HOME)/Library/LaunchAgents/com.tempod.plist
	-launchctl unload $(HOME)/Library/LaunchAgents/com.tempod.plist 2>/dev/null; true
	launchctl load -w $(HOME)/Library/LaunchAgents/com.tempod.plist
	@echo "Installed. Daemon running as launchd user agent."
else
	@echo "Installed binaries to $(BINDIR). Start tempod manually."
endif

uninstall:
	rm -f $(BINDIR)/tempod $(BINDIR)/tempo
ifeq ($(UNAME_S),Linux)
	-systemctl --user disable --now tempod
	rm -f $(HOME)/.config/systemd/user/tempod.service
	systemctl --user daemon-reload
else ifeq ($(UNAME_S),Darwin)
	-launchctl unload $(HOME)/Library/LaunchAgents/com.tempod.plist
	rm -f $(HOME)/Library/LaunchAgents/com.tempod.plist
endif
	@echo "Uninstalled."

clean:
	rm -rf bin/
