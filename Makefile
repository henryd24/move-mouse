# move-mouse - tareas de construccion
# Uso: make help

APP      := move-mouse
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DIST     := dist
GO_DIR   := go-version
PY_DIR   := py-version
LDFLAGS  := -s -w -X main.version=$(VERSION)

EXE      := $(shell go env GOEXE 2>/dev/null)
GOOS     := $(shell go env GOOS 2>/dev/null)
GOARCH   := $(shell go env GOARCH 2>/dev/null)
PREFIX   ?= $(HOME)/.local

.DEFAULT_GOAL := help
.PHONY: help deps deps-linux build run test vet fmt lint package install uninstall \
        py-setup py-run py-build clean

help: ## Muestra esta ayuda
	@awk 'BEGIN{FS=":.*## "} /^[a-zA-Z_-]+:.*## /{printf "  \033[36m%-12s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)

deps: ## Descarga dependencias de Go
	cd $(GO_DIR) && go mod download

deps-linux: ## Instala librerias de sistema necesarias en Debian/Ubuntu (requiere sudo)
	sudo apt-get update
	sudo apt-get install -y gcc libc6-dev libx11-dev xorg-dev libxtst-dev libpng-dev \
		libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev libxkbcommon-dev xsel xclip

build: ## Compila el binario Go en dist/ para el SO actual
	mkdir -p $(DIST)
	cd $(GO_DIR) && go build -trimpath -ldflags "$(LDFLAGS)" -o ../$(DIST)/$(APP)$(EXE) .

run: ## Ejecuta la version Go
	cd $(GO_DIR) && go run .

test: ## Ejecuta tests de Go
	cd $(GO_DIR) && go test ./...

vet: ## go vet
	cd $(GO_DIR) && go vet ./...

fmt: ## Formatea el codigo Go
	cd $(GO_DIR) && gofmt -w .

lint: ## Verifica formato y vet (falla si hay codigo sin formatear)
	@test -z "$$(cd $(GO_DIR) && gofmt -l .)" || (echo "Ejecuta 'make fmt'"; cd $(GO_DIR) && gofmt -l .; exit 1)
	$(MAKE) vet

package: build ## Empaqueta el binario en dist/ (.tar.gz, o .zip en Windows) + sha256
	cd $(DIST) && \
	if [ "$(GOOS)" = "windows" ]; then \
		NAME=$(APP)-$(VERSION)-$(GOOS)-$(GOARCH).zip; \
		cp ../LICENSE ../README.md . && zip -q $$NAME $(APP)$(EXE) LICENSE README.md && rm LICENSE README.md; \
	else \
		NAME=$(APP)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz; \
		cp ../LICENSE ../README.md . && tar -czf $$NAME $(APP)$(EXE) LICENSE README.md && rm LICENSE README.md; \
	fi; \
	(sha256sum $$NAME 2>/dev/null || shasum -a 256 $$NAME) > $$NAME.sha256; \
	echo "Generado: $(DIST)/$$NAME"

install: build ## Instala en $(PREFIX)/bin (por defecto ~/.local/bin)
	install -d $(PREFIX)/bin
	install -m 755 $(DIST)/$(APP)$(EXE) $(PREFIX)/bin/$(APP)$(EXE)

uninstall: ## Elimina el binario instalado
	rm -f $(PREFIX)/bin/$(APP)$(EXE)

py-setup: ## Crea el venv de Python e instala dependencias
	cd $(PY_DIR) && python3 -m venv .venv && .venv/bin/pip install -U pip && .venv/bin/pip install -r requirements.txt

py-run: ## Ejecuta la version Python
	cd $(PY_DIR) && .venv/bin/python main.py

py-build: ## Genera ejecutable de Python con PyInstaller en dist/
	cd $(PY_DIR) && .venv/bin/pip install pyinstaller && \
	.venv/bin/pyinstaller --onefile --name $(APP)-py --distpath ../$(DIST) --workpath ../build/py --specpath ../build/py main.py

clean: ## Borra artefactos de compilacion
	rm -rf $(DIST) build
