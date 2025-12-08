# Changelog

Todos los cambios notables de este proyecto seran documentados en este archivo.

El formato esta basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [0.0.1] - 2025-12-08

### Agregado

- Busqueda de negocios por nombre y ciudad
- Deteccion automatica de estado abierto/cerrado
- Extraccion de horarios desde snippets de Google (via Serper API)
- Soporte bilingue (espanol e ingles)
- Salida con colores en terminal (verde=abierto, rojo=cerrado, amarillo=sin horario)
- Salida JSON con flag `--json`
- Sistema de cache local con TTL de 24 horas
- Configuracion persistente (API key, idioma, ciudad por defecto, etc.)
- Subcomandos: `config`, `cache`, `about`, `uninstall`, `version`
- Scripts de instalacion para Linux, macOS y Windows
- Compilacion multiplataforma (Linux amd64/arm64, macOS amd64/arm64, Windows amd64)
- Filtrado de resultados por ciudad en la direccion
- Informacion adicional: rating, opiniones, categoria, telefono

### Limitaciones conocidas

- La API de Serper Places no proporciona horarios estructurados
- Los horarios se extraen de snippets, pueden no estar disponibles para todos los negocios
- Solo se buscan horarios para los primeros 3 resultados (ahorro de creditos API)

