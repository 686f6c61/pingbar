# Changelog

Todos los cambios notables de este proyecto seran documentados en este archivo.

El formato esta basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [0.1.0] - 2026-07-22

### Seguridad

- Toolchain fijado a Go 1.26.5 en go.mod: elimina 13 vulnerabilidades de la
  biblioteca estandar presentes al compilar con Go 1.26.0 (govulncheck limpio)
- Dependencias actualizadas: cobra v1.8.0 -> v1.10.2, fatih/color v1.16.0 -> v1.19.0,
  go-colorable, go-isatty, pflag y golang.org/x/sys a sus ultimas versiones

### Corregido

- La cache local no se usaba: estaba implementada y documentada, pero las
  busquedas siempre llamaban a la API. Ahora los resultados se cachean 24 horas
  y el estado abierto/cerrado se recalcula con la hora actual al servir desde cache
- La clave de cache incluye ahora el limite de resultados para evitar servir
  resultados truncados o incompletos entre busquedas con distinto --limit
- Escritura determinista del archivo de configuracion (claves ordenadas);
  antes el orden cambiaba en cada guardado

### Agregado

- Pipeline de CI en GitHub Actions: formato, vet, build, tests con detector de
  carreras, govulncheck y compilacion multiplataforma en cada push y PR
- Tests unitarios para extraccion de horarios, calculo abierto/cerrado,
  configuracion y cache (antes no habia ningun test)

### Mejorado

- Las busquedas de horarios de los 3 primeros resultados se ejecutan en paralelo
  (antes eran secuenciales: hasta 3 llamadas HTTP encadenadas)
- El filtrado por ciudad ignora tildes: "Málaga" y "malaga" ahora coinciden
  (normalizacion Unicode con golang.org/x/text)
- La extraccion de horarios prioriza el dia actual cuando el snippet lista
  varios dias con horas distintas, y reconoce nombres de dia con y sin tilde
- Cliente HTTP compartido con keep-alive en lugar de crear uno por peticion
- strconv.Atoi en lugar de fmt.Sscanf para parsear enteros
- errors.As en lugar de asercion de tipo directa para errores de API
- interface{} reemplazado por any

### Eliminado

- Flag --week (documentado pero sin implementacion, como el antiguo --tomorrow)
- Mensajes i18n sin uso (Tomorrow, ClosesIn, ClosedAgo, Holiday, SpecialHours,
  MoreResults, ViewAll, Yes, No) y funcion GetDay muerta

## [0.0.2] - 2026-03-04

### Seguridad

- Permisos de directorios de configuracion y cache cambiados de 0755 a 0700
- Permisos de archivos de configuracion y cache cambiados de 0644 a 0600
- Limitacion de respuestas HTTP con io.LimitReader (maximo 10MB)
- Validacion de longitud de entrada (maximo 200 caracteres por campo)
- Verificacion de integridad SHA-256 en scripts de instalacion (install.sh e install.ps1)
- Fallback seguro para os.UserHomeDir() con variables de entorno HOME/USERPROFILE
- Eliminado fallback silencioso a version hardcodeada en scripts de instalacion

### Corregido

- Errores ignorados en llamadas HTTP y serializacion JSON que podian causar panics
- Error de config.Load() no manejado en el comando raiz (causaba nil pointer panic)
- Resultados vacios de la API al usar parametro location innecesario
- Parametro num excesivo que causaba respuestas vacias en el tier gratuito de Serper
- Funcion maskAPIKey duplicada entre cmd/config.go e internal/config/config.go
- Version hardcodeada en pantalla "about" (ahora usa la version del binario)
- Error de encoder.Encode() no manejado en salida JSON
- Comparacion incorrecta de PATH en install.ps1

### Mejorado

- Regex precompiladas como variables de paquete (mejor rendimiento)
- install.sh: trap de limpieza, soporte para jq, verificacion de tipo de binario
- install.ps1: usa Invoke-WebRequest, descarga temporal antes de mover
- Makefile: install -m 755, .PHONY completo, checksums portables, version desde git
- Requisito minimo de Go actualizado a 1.22

### Eliminado

- Flag --tomorrow (no tenia implementacion)
- Funciones muertas: GetRawResponse, ParseCachedResponse
- Variable showTomorrow sin uso

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

