#!/bin/bash
#
# Script de instalacion de pingbar
# https://github.com/686f6c61/pingbar
#

set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuracion
REPO="686f6c61/pingbar"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="pingbar"

echo ""
echo "================================="
echo "    Instalador de pingbar"
echo "================================="
echo ""

# Detectar arquitectura y SO
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    case "$os" in
        linux)
            os="linux"
            ;;
        darwin)
            os="macos"
            ;;
        *)
            echo -e "${RED}Error: Sistema operativo no soportado: $os${NC}"
            exit 1
            ;;
    esac

    case "$arch" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        *)
            echo -e "${RED}Error: Arquitectura no soportada: $arch${NC}"
            exit 1
            ;;
    esac

    PLATFORM="${os}-${arch}"
    echo -e "${GREEN}Plataforma detectada: ${PLATFORM}${NC}"
}

# Obtener ultima version
get_latest_version() {
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        VERSION="v0.0.1"
    fi
    echo -e "${GREEN}Version: ${VERSION}${NC}"
}

# Descargar e instalar
install() {
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/pingbar-${PLATFORM}"
    
    echo -e "${YELLOW}Descargando pingbar...${NC}"
    
    # Crear directorio temporal
    local tmp_dir=$(mktemp -d)
    local tmp_file="${tmp_dir}/${BINARY_NAME}"
    
    # Descargar
    if command -v curl &> /dev/null; then
        curl -sL "$download_url" -o "$tmp_file" || {
            echo -e "${RED}Error: Fallo en la descarga${NC}"
            rm -rf "$tmp_dir"
            exit 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "$download_url" -O "$tmp_file" || {
            echo -e "${RED}Error: Fallo en la descarga${NC}"
            rm -rf "$tmp_dir"
            exit 1
        }
    else
        echo -e "${RED}Error: Se requiere curl o wget${NC}"
        exit 1
    fi
    
    # Verificar descarga
    if [ ! -f "$tmp_file" ] || [ ! -s "$tmp_file" ]; then
        echo -e "${RED}Error: Archivo descargado vacio o no existe${NC}"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    # Dar permisos de ejecucion
    chmod +x "$tmp_file"
    
    # Instalar
    echo -e "${YELLOW}Instalando en ${INSTALL_DIR}...${NC}"
    
    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Limpiar
    rm -rf "$tmp_dir"
    
    echo ""
    echo -e "${GREEN}pingbar instalado correctamente${NC}"
    echo ""
    echo "Para empezar:"
    echo "  1. Obten una API Key gratuita en https://serper.dev"
    echo "  2. Configura tu API Key: pingbar config set apikey TU_API_KEY"
    echo "  3. Prueba: pingbar 'farmacia' madrid"
    echo ""
    echo "Mas informacion: pingbar --help"
}

# Main
detect_platform
get_latest_version
install
