# Makefile for voidPM (vpm)

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
PKGNAME = voidpm
VERSION = 0.1.0

.PHONY: all build install uninstall package publish clean test

all: build

build:
	@echo "==> Building voidPM binary..."
	go build -ldflags="-s -w" -o vpm .

install: build
	@echo "==> Installing vpm binary to $(DESTDIR)$(BINDIR)..."
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 vpm $(DESTDIR)$(BINDIR)/vpm

uninstall:
	@echo "==> Removing vpm binary from $(DESTDIR)$(BINDIR)..."
	rm -f $(DESTDIR)$(BINDIR)/vpm

package:
	@./scripts/build-package.sh

publish:
	@./scripts/build-package.sh --publish

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf vpm dist/

