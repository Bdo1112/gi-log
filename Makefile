BINARY     = gi-log
INSTALL_PATH = $(HOME)/.local/bin/$(BINARY)
SETTINGS   = $(HOME)/.claude/settings.json

.PHONY: build install clean

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(HOME)/.local/bin
	mv $(BINARY) $(INSTALL_PATH)
	@echo "Wiring hooks into $(SETTINGS)..."
	gi-log install
	@echo ""
	@echo "Done. gi-log installed."
	@echo "Next: add your OpenAI key to ~/.gi-log/config.json"

clean:
	rm -f $(BINARY)
