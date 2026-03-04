#
# Script de instalacion de pingbar para Windows
# https://github.com/686f6c61/pingbar
#

$ErrorActionPreference = "Stop"

# Configuracion
$Repo = "686f6c61/pingbar"
$BinaryName = "pingbar.exe"

Write-Host ""
Write-Host "=================================" -ForegroundColor Cyan
Write-Host "    Instalador de pingbar" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# Detectar arquitectura
function Get-Platform {
    if ([System.Environment]::Is64BitOperatingSystem) {
        return "windows-amd64"
    } else {
        Write-Host "Error: Solo se soporta Windows de 64 bits" -ForegroundColor Red
        exit 1
    }
}

# Obtener ultima version
function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        if ($release.tag_name) {
            return $release.tag_name
        }
        throw "tag_name no encontrado"
    } catch {
        Write-Host "Error: No se pudo obtener la ultima version de GitHub." -ForegroundColor Red
        Write-Host "Comprueba tu conexion a internet o descarga manualmente desde:" -ForegroundColor Yellow
        Write-Host "  https://github.com/$Repo/releases" -ForegroundColor Yellow
        exit 1
    }
}

# Obtener directorio de instalacion
function Get-InstallDir {
    $installDir = Join-Path $env:ProgramFiles "pingbar"

    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    return $installDir
}

# Agregar al PATH si es necesario
function Add-ToPath($dir) {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
    $paths = $currentPath -split ';'

    if ($dir -notin $paths) {
        $newPath = $currentPath + ";" + $dir
        [Environment]::SetEnvironmentVariable("Path", $newPath, [EnvironmentVariableTarget]::User)
        Write-Host "Directorio agregado al PATH del usuario" -ForegroundColor Green
    }
}

# Verificar checksum
function Verify-Checksum($filePath, $version, $platform) {
    Write-Host "Verificando integridad..." -ForegroundColor Yellow

    $checksumsUrl = "https://github.com/$Repo/releases/download/$version/checksums.txt"
    $checksumsFile = Join-Path $env:TEMP "pingbar-checksums.txt"

    try {
        Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsFile -UseBasicParsing
    } catch {
        Write-Host "Aviso: No se pudo descargar checksums.txt para verificar integridad." -ForegroundColor Yellow
        return
    }

    $checksumContent = Get-Content $checksumsFile -ErrorAction SilentlyContinue
    if (-not $checksumContent) {
        Write-Host "Aviso: checksums.txt vacio o no legible." -ForegroundColor Yellow
        Remove-Item $checksumsFile -ErrorAction SilentlyContinue
        return
    }

    $expectedLine = $checksumContent | Where-Object { $_ -match "pingbar-$platform" }
    if (-not $expectedLine) {
        Write-Host "Aviso: No se encontro hash para pingbar-$platform." -ForegroundColor Yellow
        Remove-Item $checksumsFile -ErrorAction SilentlyContinue
        return
    }

    $expectedHash = ($expectedLine -split '\s+')[0]
    $actualHash = (Get-FileHash -Path $filePath -Algorithm SHA256).Hash.ToLower()

    Remove-Item $checksumsFile -ErrorAction SilentlyContinue

    if ($expectedHash -ne $actualHash) {
        Write-Host "Error: El checksum del binario no coincide." -ForegroundColor Red
        Write-Host "  Esperado: $expectedHash" -ForegroundColor Red
        Write-Host "  Obtenido: $actualHash" -ForegroundColor Red
        Write-Host "La descarga puede estar corrupta o haber sido manipulada." -ForegroundColor Red
        Remove-Item $filePath -ErrorAction SilentlyContinue
        exit 1
    }

    Write-Host "Checksum verificado correctamente." -ForegroundColor Green
}

# Instalacion principal
function Install-Pingbar {
    # Detectar plataforma
    $platform = Get-Platform
    Write-Host "Plataforma: $platform" -ForegroundColor Green

    # Obtener version
    $version = Get-LatestVersion
    Write-Host "Version: $version" -ForegroundColor Green

    # Construir URL de descarga
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/pingbar-$platform.exe"

    # Directorio de instalacion
    $installDir = Get-InstallDir
    $binaryPath = Join-Path $installDir $BinaryName

    # Descargar a directorio temporal primero
    $tmpPath = Join-Path $env:TEMP "pingbar-download.exe"

    Write-Host "Descargando pingbar..." -ForegroundColor Yellow

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpPath -UseBasicParsing

        # Verificar que el archivo existe y no esta vacio
        if (-not (Test-Path $tmpPath) -or (Get-Item $tmpPath).Length -eq 0) {
            Write-Host "Error: Archivo descargado vacio o no existe" -ForegroundColor Red
            exit 1
        }

        # Verificar checksum
        Verify-Checksum $tmpPath $version $platform

        # Mover al destino final
        Move-Item -Path $tmpPath -Destination $binaryPath -Force

        # Agregar al PATH
        Add-ToPath $installDir

        Write-Host ""
        Write-Host "pingbar instalado correctamente" -ForegroundColor Green
        Write-Host ""
        Write-Host "Para empezar:" -ForegroundColor White
        Write-Host "  1. Abre una nueva terminal (para cargar el PATH actualizado)" -ForegroundColor Gray
        Write-Host "  2. Obten una API Key gratuita en https://serper.dev" -ForegroundColor Gray
        Write-Host "  3. Configura tu API Key: pingbar config set apikey TU_API_KEY" -ForegroundColor Gray
        Write-Host "  4. Prueba: pingbar 'farmacia' madrid" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Mas informacion: pingbar --help" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Instalado en: $binaryPath" -ForegroundColor Gray

    } catch {
        # Limpiar archivo temporal si existe
        Remove-Item $tmpPath -ErrorAction SilentlyContinue
        Write-Host "Error durante la instalacion: $_" -ForegroundColor Red
        exit 1
    }
}

# Ejecutar instalacion
Install-Pingbar
