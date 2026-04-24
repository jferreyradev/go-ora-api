# 📦 Release Notes - v1.0.0

## Contenido del Release

```
go-oracle-api/
├── go-oracle-api.exe          # Ejecutable compilado (Windows)
├── main.go                     # Código fuente Go
├── go.mod, go.sum              # Dependencias
├── .env.example                # Template de configuración
├── README.md                   # Documentación principal
├── LICENSE                     # MIT License
├── docs/                       # Documentación detallada
└── sql/                        # Scripts de base de datos
```

## 🎉 Características Principales

✨ **API RESTful** - Acceso HTTP a Oracle sin drivers en clientes  
✨ **Procedimientos y Funciones** - Soporte completo para PL/SQL  
✨ **Jobs Asíncronos** - Ejecución no bloqueante con monitoreo  
✨ **Tipos Automáticos** - Detección de NUMBER, VARCHAR2, DATE  
✨ **Seguridad** - Autenticación Bearer Token + restricción de IPs  

## 🐛 Bug Fixes Críticos

### ✅ CRITICAL: Múltiples parámetros OUT (SOLUCIONADO)

**Problema:**
- API devolvía solo el último parámetro OUT en procedimientos
- Parámetros de salida se perdían

**Causa:**
- Buffers de parámetros OUT creados como variables locales en loop
- Perdían scope después del loop → punteros inválidos

**Solución:**
- Pre-asignar todos los buffers FUERA del loop
- Usar referencias persistentes durante la ejecución
- Indexación secuencial correcta

**Impacto:**
✅ TODOS los OUT parameters ahora se devuelven correctamente  
✅ Soporta ANY número de parámetros (NUMBER, VARCHAR2, DATE)  

**Ejemplo:**
```bash
curl -X POST http://localhost:3000/procedure \
  -d '{
    "name": "PRUEBA1",
    "params": [
      {"name": "vIDPERS", "value": 123, "direction": "IN", "type": "NUMBER"},
      {"name": "vDNI", "value": 45678901, "direction": "IN", "type": "NUMBER"},
      {"name": "vSALIDA", "direction": "OUT", "type": "NUMBER"},
      {"name": "vError", "direction": "OUT", "type": "NUMBER"},
      {"name": "vErrorMsg", "direction": "OUT", "type": "STRING"}
    ]
  }'
```

**Respuesta:**
```json
{
  "status": "ok",
  "out": {
    "vSALIDA": 10,
    "vError": -999,
    "vErrorMsg": "Error generado"
  }
}
```

## 📥 Descarga

**Binario Compilado:**
- `go-oracle-api.exe` (20.36 MB) - Windows x86_64

## 🚀 Inicio Rápido

1. **Configurar**
   ```bash
   cp .env.example .env
   # Editar .env con credenciales Oracle
   ```

2. **Setup de BD (primera vez)**
   ```bash
   sqlplus user/pass@db @sql/create_async_jobs_table.sql
   sqlplus user/pass@db @sql/create_query_log_table.sql
   ```

3. **Ejecutar**
   ```bash
   ./go-oracle-api.exe
   # API en http://localhost:3000
   ```

4. **Probar**
   ```bash
   curl http://localhost:3000/ping \
     -H "Authorization: Bearer test1"
   ```

## 📊 Requisitos

| Componente | Requisito |
|-----------|----------|
| OS | Windows 7+ / Linux / macOS |
| Arquitectura | x86_64 |
| Oracle | 11g o superior |
| RAM | 512 MB mínimo |
| Go (opcional) | 1.20+ para compilar |

## 🔧 Variables de Entorno

| Variable | Descripción |
|----------|-------------|
| `ORACLE_USER` | Usuario Oracle |
| `ORACLE_PASSWORD` | Contraseña |
| `ORACLE_HOST` | Host/IP |
| `ORACLE_PORT` | Puerto (default: 1521) |
| `ORACLE_SERVICE` | Service name |
| `API_TOKEN` | Token de autenticación |
| `PORT` | Puerto API (default: 8080) |
| `API_ALLOWED_IPS` | IPs permitidas |

Ver `docs/getting-started/CONFIGURACION.md` para detalles.

## 📡 Endpoints Disponibles

- `GET /ping` - Health check
- `POST /query` - Consultas SELECT
- `POST /exec` - INSERT, UPDATE, DELETE, DDL
- `POST /procedure` - Procedimientos síncrono
- `POST /procedure/async` - Procedimientos asíncrono
- `GET /jobs` - Listar jobs
- `GET /jobs/{id}` - Estado del job
- `DELETE /jobs/{id}` - Eliminar job
- `GET /logs` - Ver logs

Ver `docs/api-reference/ENDPOINTS.md` para referencia completa.

## 📚 Documentación

- [Inicio Rápido](../getting-started/QUICKSTART.md)
- [API Reference](../api-reference/ENDPOINTS.md)
- [Deployment](../deployment/DEPLOYMENT.md)
- [Troubleshooting](../deployment/DEPLOYMENT.md#troubleshooting)

## 🆘 Soporte

**Documentación:**
- Ver [docs/INDEX.md](../INDEX.md) para índice completo
- [Guía de Configuración](../getting-started/CONFIGURACION.md)
- [API Reference](../api-reference/ENDPOINTS.md)

**Problemas Comunes:**
- Verificar credenciales en `.env`
- Verificar conectividad a Oracle
- Ver [Deployment Troubleshooting](../deployment/DEPLOYMENT.md#troubleshooting)

## 📝 Licencia

**MIT License** - Libre para usar, modificar y distribuir

## 🚀 Roadmap (Próximas Versiones)

**v1.1.0 (Planned)**
- WebSocket support
- Caché de resultados
- Métricas Prometheus

**v1.2.0 (Planned)**
- GraphQL endpoint
- Rate limiting
- Audit trail completo

---

**Status:** ✅ Stable  
**Release Date:** 2026-04-24  
**Version:** 1.0.0
