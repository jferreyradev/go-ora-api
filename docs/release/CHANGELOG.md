# 📋 Changelog

Historial de todas las versiones de Go Oracle API

---

## [1.0.0] - 2026-04-24

### 🎉 Initial Release

#### ✨ Características Principales
- **API RESTful para Oracle** - Acceso HTTP a base de datos Oracle sin necesidad de drivers en clientes
- **Soporte completo para Procedimientos y Funciones** - Ejecución de PL/SQL con parámetros IN/OUT/IN OUT
- **Jobs Asíncronos** - Ejecución de procedimientos de larga duración con monitoreo de progreso
  - Políticas de concurrencia: `parallel`, `sequential`, `exclusive`
  - Sistema de cola para ejecución secuencial
- **Manejo Inteligente de Tipos de Datos** - Detección automática de NUMBER, VARCHAR2, DATE
- **CORS y Autenticación** - Seguridad mediante tokens Bearer + restricción por IPs

#### 🐛 Bug Fixes
- **CRITICAL: Múltiples parámetros OUT** - Corregido error donde solo se devolvía el último parámetro OUT
  - Problema: Buffers de parámetros OUT creados como variables locales perdían scope
  - Solución: Pre-asignar todos los buffers fuera del loop con referencias persistentes
  - Impacto: Ahora se devuelven correctamente TODOS los parámetros de salida (NUMBER, VARCHAR2, DATE)

#### 🚀 Endpoints Disponibles
- `GET /ping` - Health check básico
- `GET /health` - Estado detallado del sistema (para monitoreo)
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

#### 📦 Incluido
- Código fuente completo
- Binario compilado Windows (go-oracle-api.exe - 20.36 MB)
- Scripts SQL para setup (async_jobs, query_log, procedimientos de prueba)
- Ejemplos completos en Node.js
- Documentación completa

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

### v1.3.0 (Planned)
- [ ] Soporte para múltiples conexiones Oracle
- [ ] Failover automático
- [ ] Query caching distribuido

---

**Licencia:** MIT  
**Status:** ✅ Stable
