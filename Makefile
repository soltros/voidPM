# Makefile for voidPM (vpm)

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
PKGNAME = voidpm
VERSION = 0.1.0

.PHONY: all build install uninstall package clean test

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
	@echo "==> Packaging voidPM for void-packages (xbps-src)..."
	@mkdir -p ~/void-packages/srcpkgs/$(PKGNAME)/files
	@cp -a template ~/void-packages/srcpkgs/$(PKGNAME)/template
	@cp -a . ~/void-packages/srcpkgs/$(PKGNAME)/files/
	@echo "==> Triggering xbps-src build..."
	vpm src build $(PKGNAME)

clean:
	@echo "==> Cleaning build artifacts..."
	rm -f vpm
