# Changelog

Todos los cambios notables en este proyecto serán documentados en este archivo.

## [1.0.0] - 2026-04-24

### 🎉 Initial Release

#### ✨ Características Principales
- **API RESTful para Oracle** - Acceso HTTP a base de datos Oracle sin necesidad de drivers en clientes
- **Soporte completo para Procedimientos y Funciones** - Ejecución de PL/SQL con parámetros IN/OUT
- **Jobs Asíncronos** - Ejecución de procedimientos de larga duración con monitoreo de progreso
- **Manejo Inteligente de Tipos de Datos** - Detección automática de NUMBER, VARCHAR2, DATE
- **CORS y Autenticación** - Seguridad mediante tokens Bearer

#### 🐛 Bug Fixes
- **CRITICAL: Múltiples parámetros OUT** - Fixed issue donde solo se devolvía el último parámetro OUT en procedimientos
  - Cambio: Preasignar buffers fuera del loop en lugar de como variables locales
  - Impacto: Ahora se devuelven correctamente TODOS los parámetros de salida (NUMBER, VARCHAR2, DATE)
  - Ejemplo: Procedimiento con 3 OUT parameters ahora devuelve los 3 valores correctamente

#### 📚 Documentación Completa
- README.md - Guía principal y características
- GUIA_RAPIDA.md - Referencia rápida con ejemplos
- docs/ASYNC_JOBS.md - Sistema de jobs asíncronos
- docs/CONFIGURACION_ENV.md - Variables de entorno
- docs/DEPLOYMENT.md - Guía de despliegue
- docs/FIREWALL_WINDOWS.md - Configuración de firewall
- docs/SCHEMA_FIELD.md - Nomenclatura Oracle
- docs/USO_Y_PRUEBAS.md - Ejemplos de uso

#### 📦 Incluido
- `main.go` - Código fuente completo (~2000 líneas)
- `go-oracle-api.exe` - Binario compilado (20.36 MB)
- `sql/` - Scripts de setup (tablas async_jobs, query_log, procedimientos de prueba)
- `examples/` - Ejemplo completo en Node.js
- Soporte para múltiples instancias en paralelo

#### 🚀 Endpoints Disponibles
- `GET /ping` - Health check
- `POST /query` - Consultas SELECT
- `POST /exec` - INSERT, UPDATE, DELETE, DDL
- `POST /procedure` - Procedimientos síncrono
- `POST /procedure/async` - Procedimientos asíncrono
- `GET /jobs` - Listar jobs asíncronos
- `GET /jobs/{id}` - Estado del job
- `DELETE /jobs/{id}` - Eliminar job
- `POST /upload` - Subir archivos BLOB
- `GET /download` - Descargar BLOB
- `GET /logs` - Consultar logs

#### 🛠️ Requisitos
- Go 1.20 o superior
- Oracle 11g o superior
- .env configurado con credenciales

#### 📝 Notas de Instalación
1. Configurar `.env` con credenciales Oracle
2. Ejecutar scripts SQL: `create_async_jobs_table.sql`, `create_query_log_table.sql`
3. Ejecutar binario: `./go-oracle-api.exe`

---

## Próximas Versiones (Roadmap)

### v1.1.0 (Planned)
- [ ] WebSocket support para actualizaciones en tiempo real
- [ ] Caché de resultados de queries
- [ ] Métricas de performance (Prometheus)
- [ ] Soporte para Oracle CLOB/BLOB mejorado

### v1.2.0 (Planned)
- [ ] GraphQL endpoint alternativo
- [ ] Rate limiting por IP
- [ ] Audit trail completo
- [ ] Encriptación de credenciales en tránsito
