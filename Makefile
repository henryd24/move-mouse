# move-mouse - tareas de construccion
# Uso: make help

APP      := move-mouse
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DIST     := dist
LDFLAGS  := -s -w -X main.version=$(VERSION)

EXE      := $(shell go env GOEXE 2>/dev/null)
GOOS     := $(shell go env GOOS 2>/dev/null)
GOARCH   := $(shell go env GOARCH 2>/dev/null)
PREFIX   ?= $(HOME)/.local

# La GUI usa webkit2gtk 4.1 en Linux
GUI_TAGS := production$(if $(filter linux,$(GOOS)),$(comma)webkit2_41)
comma    := ,
GUI_LDFLAGS := $(LDFLAGS)$(if $(filter windows,$(GOOS)), -H=windowsgui)

.DEFAULT_GOAL := help
.PHONY: help deps deps-linux deps-gui-linux build run test vet fmt lint package \
        gui-build gui-package install uninstall deb rpm clean

help: ## Muestra esta ayuda
	@awk 'BEGIN{FS=":.*## "} /^[a-zA-Z_-]+:.*## /{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)

deps: ## Descarga dependencias de Go
	go mod download

deps-linux: ## Librerias de sistema para la CLI en Debian/Ubuntu (requiere sudo)
	sudo apt-get update
	sudo apt-get install -y gcc libc6-dev libx11-dev xorg-dev libxtst-dev libpng-dev \
		libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev libxkbcommon-dev xsel xclip

deps-gui-linux: deps-linux ## Ademas, librerias para la GUI (GTK + WebKit)
	sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev

build: ## Compila la CLI en dist/ para el SO actual
	mkdir -p $(DIST)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)$(EXE) ./cmd/move-mouse

run: ## Ejecuta la CLI (ARGS="--mode jiggle")
	go run ./cmd/move-mouse $(ARGS)

test: ## Ejecuta tests
	go test ./...

vet: ## go vet
	go vet ./...

fmt: ## Formatea el codigo Go
	gofmt -w .

lint: ## Verifica formato y vet (falla si hay codigo sin formatear)
	@test -z "$$(gofmt -l .)" || (echo "Ejecuta 'make fmt':"; gofmt -l .; exit 1)
	$(MAKE) vet

gui-build: ## Compila la app grafica (Wails) en dist/
	mkdir -p $(DIST)
	cd app && go build -trimpath -tags "$(GUI_TAGS)" -ldflags "$(GUI_LDFLAGS)" -o ../$(DIST)/$(APP)-gui$(EXE) .

# $(call pack,<binario>,<nombre>)
define pack
	cd $(DIST) && \
	if [ "$(GOOS)" = "windows" ]; then \
		NAME=$(2)-$(VERSION)-$(GOOS)-$(GOARCH).zip; \
		cp ../LICENSE ../README.md . && zip -q $$NAME $(1)$(EXE) LICENSE README.md && rm LICENSE README.md; \
	else \
		NAME=$(2)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz; \
		cp ../LICENSE ../README.md . && tar -czf $$NAME $(1)$(EXE) LICENSE README.md && rm LICENSE README.md; \
	fi; \
	(sha256sum $$NAME 2>/dev/null || shasum -a 256 $$NAME) > $$NAME.sha256; \
	echo "Generado: $(DIST)/$$NAME"
endef

package: build ## Empaqueta la CLI (.tar.gz / .zip + sha256)
	$(call pack,$(APP),$(APP))

gui-package: gui-build ## Empaqueta la GUI (.tar.gz / .zip + sha256)
	$(call pack,$(APP)-gui,$(APP)-gui)

deb: build ## Genera .deb (requiere nfpm)
	VERSION=$(patsubst v%,%,$(VERSION)) ARCH=$(GOARCH) nfpm package -p deb -f packaging/nfpm.yaml -t $(DIST)/

rpm: build ## Genera .rpm (requiere nfpm)
	VERSION=$(patsubst v%,%,$(VERSION)) ARCH=$(GOARCH) nfpm package -p rpm -f packaging/nfpm.yaml -t $(DIST)/

install: build ## Instala la CLI en $(PREFIX)/bin (por defecto ~/.local/bin)
	install -d $(PREFIX)/bin
	install -m 755 $(DIST)/$(APP)$(EXE) $(PREFIX)/bin/$(APP)$(EXE)

uninstall: ## Elimina la CLI instalada
	rm -f $(PREFIX)/bin/$(APP)$(EXE)

clean: ## Borra artefactos de compilacion
	rm -rf $(DIST) build
