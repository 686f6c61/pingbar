# Makefile para pingbar
# https://github.com/686f6c61/pingbar

VERSION := 0.0.1
BINARY_NAME := pingbar
BUILD_DIR := build
LDFLAGS := -ldflags "-s -w -X github.com/686f6c61/pingbar/cmd.Version=$(VERSION)"

.PHONY: all build clean test install dev

# Compilación por defecto para la plataforma actual
all: build

# Compilar para la plataforma actual
build:
	@echo "Compilando $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Desarrollo: compilar y ejecutar
dev: build
	./$(BINARY_NAME)

# Instalar localmente
install: build
	@echo "Instalando $(BINARY_NAME) en /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/

# Compilar para todas las plataformas
release: clean
	@echo "Compilando releases para todas las plataformas..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 .
	
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 .
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "Generando checksums..."
	cd $(BUILD_DIR) && shasum -a 256 * > checksums.txt
	
	@echo "✓ Releases generados en $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Limpiar
clean:
	@echo "Limpiando..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Descargar dependencias
deps:
	go mod download
	go mod tidy

# Ejecutar tests
test:
	go test -v ./...

# Verificar código
lint:
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint no está instalado"; \
	fi

# Mostrar ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  make build    - Compilar para la plataforma actual"
	@echo "  make install  - Instalar en /usr/local/bin"
	@echo "  make release  - Compilar para todas las plataformas"
	@echo "  make clean    - Limpiar archivos generados"
	@echo "  make deps     - Descargar dependencias"
	@echo "  make test     - Ejecutar tests"
	@echo "  make help     - Mostrar esta ayuda"

