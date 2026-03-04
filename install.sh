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

# Directorio temporal global para limpieza con trap
TMP_DIR=""
cleanup() {
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}
trap cleanup EXIT INT TERM

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
    local api_response
    api_response=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null) || true

    if command -v jq &> /dev/null; then
        VERSION=$(echo "$api_response" | jq -r '.tag_name // empty' 2>/dev/null)
    else
        VERSION=$(echo "$api_response" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [ -z "$VERSION" ]; then
        echo -e "${RED}Error: No se pudo obtener la ultima version de GitHub.${NC}"
        echo "Comprueba tu conexion a internet o descarga manualmente desde:"
        echo "  https://github.com/${REPO}/releases"
        exit 1
    fi

    echo -e "${GREEN}Version: ${VERSION}${NC}"
}

# Verificar checksum del binario descargado
verify_checksum() {
    local file="$1"
    local checksums_url="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
    local checksums_file="${TMP_DIR}/checksums.txt"

    echo -e "${YELLOW}Verificando integridad...${NC}"

    if command -v curl &> /dev/null; then
        curl -sL "$checksums_url" -o "$checksums_file" 2>/dev/null || true
    elif command -v wget &> /dev/null; then
        wget -q "$checksums_url" -O "$checksums_file" 2>/dev/null || true
    fi

    if [ ! -s "$checksums_file" ]; then
        echo -e "${YELLOW}Aviso: No se pudo descargar checksums.txt para verificar integridad.${NC}"
        return 0
    fi

    local expected_hash
    expected_hash=$(grep "pingbar-${PLATFORM}" "$checksums_file" | awk '{print $1}')

    if [ -z "$expected_hash" ]; then
        echo -e "${YELLOW}Aviso: No se encontro hash para pingbar-${PLATFORM} en checksums.txt.${NC}"
        return 0
    fi

    local actual_hash
    if command -v sha256sum &> /dev/null; then
        actual_hash=$(sha256sum "$file" | awk '{print $1}')
    elif command -v shasum &> /dev/null; then
        actual_hash=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        echo -e "${YELLOW}Aviso: No se encontro sha256sum ni shasum para verificar integridad.${NC}"
        return 0
    fi

    if [ "$expected_hash" != "$actual_hash" ]; then
        echo -e "${RED}Error: El checksum del binario no coincide.${NC}"
        echo "  Esperado: $expected_hash"
        echo "  Obtenido: $actual_hash"
        echo "La descarga puede estar corrupta o haber sido manipulada."
        exit 1
    fi

    echo -e "${GREEN}Checksum verificado correctamente.${NC}"
}

# Descargar e instalar
install() {
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/pingbar-${PLATFORM}"

    echo -e "${YELLOW}Descargando pingbar...${NC}"

    # Crear directorio temporal
    TMP_DIR=$(mktemp -d)
    local tmp_file="${TMP_DIR}/${BINARY_NAME}"

    # Descargar
    if command -v curl &> /dev/null; then
        curl -sL "$download_url" -o "$tmp_file" || {
            echo -e "${RED}Error: Fallo en la descarga${NC}"
            exit 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "$download_url" -O "$tmp_file" || {
            echo -e "${RED}Error: Fallo en la descarga${NC}"
            exit 1
        }
    else
        echo -e "${RED}Error: Se requiere curl o wget${NC}"
        exit 1
    fi

    # Verificar que el archivo existe y no esta vacio
    if [ ! -f "$tmp_file" ] || [ ! -s "$tmp_file" ]; then
        echo -e "${RED}Error: Archivo descargado vacio o no existe${NC}"
        exit 1
    fi

    # Verificar que es un binario valido (no una pagina HTML de error)
    local file_type
    file_type=$(file -b "$tmp_file" 2>/dev/null || echo "unknown")
    case "$file_type" in
        *ELF*|*Mach-O*|*executable*)
            ;;
        *)
            echo -e "${RED}Error: El archivo descargado no es un binario valido (tipo: ${file_type})${NC}"
            echo "Puede que la version ${VERSION} no tenga binario para ${PLATFORM}."
            exit 1
            ;;
    esac

    # Verificar checksum
    verify_checksum "$tmp_file"

    # Dar permisos de ejecucion (solo propietario + lectura/ejecucion para otros)
    chmod 755 "$tmp_file"

    # Instalar
    echo -e "${YELLOW}Instalando en ${INSTALL_DIR}...${NC}"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo "Se requieren permisos de administrador para instalar en ${INSTALL_DIR}"
        sudo mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

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
