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
        return $release.tag_name
    } catch {
        return "v0.0.1"
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
    
    if ($currentPath -notlike "*$dir*") {
        $newPath = $currentPath + ";" + $dir
        [Environment]::SetEnvironmentVariable("Path", $newPath, [EnvironmentVariableTarget]::User)
        Write-Host "Directorio agregado al PATH del usuario" -ForegroundColor Green
    }
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
    
    Write-Host "Descargando pingbar..." -ForegroundColor Yellow
    
    try {
        # Descargar
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadFile($downloadUrl, $binaryPath)
        
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
        Write-Host "Error durante la instalacion: $_" -ForegroundColor Red
        exit 1
    }
}

# Ejecutar instalacion
Install-Pingbar
