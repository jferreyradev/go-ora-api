# Release Notes - Go Oracle API v1.0.0

## 📋 Contenido del Release

### Archivos Incluidos
```
go-oracle-api/
├── go-oracle-api.exe          # Ejecutable compilado (Windows)
├── main.go                     # Código fuente Go
├── go.mod, go.sum              # Dependencias
├── .env.example                # Template de configuración
├── README.md                   # Documentación principal
├── GUIA_RAPIDA.md             # Guía rápida de uso
├── CHANGELOG.md               # Este archivo
├── LICENSE                     # MIT License
├── docs/                       # Documentación detallada
│   ├── ASYNC_JOBS.md
│   ├── CONFIGURACION_ENV.md
│   ├── DEPLOYMENT.md
│   ├── FIREWALL_WINDOWS.md
│   ├── SCHEMA_FIELD.md
│   └── USO_Y_PRUEBAS.md
├── examples/                   # Ejemplos de código
│   └── ejemplo_completo.js
└── sql/                        # Scripts de base de datos
    ├── create_async_jobs_table.sql
    ├── create_query_log_table.sql
    └── create_test_procedures.sql
```

## 🚀 Guía de Inicio Rápido

### 1. Configuración Inicial

```bash
# Clonar o descargar el release
cd go-oracle-api

# Copiar template de configuración
cp .env.example .env

# Editar .env con tus credenciales Oracle
# ORACLE_USER=tu_usuario
# ORACLE_PASSWORD=tu_password
# ORACLE_HOST=ip_oracle
# ORACLE_PORT=1521
# ORACLE_SERVICE=nombre_servicio
```

### 2. Setup de Base de Datos (Primera Vez)

```bash
# Usando sqlplus
sqlplus usuario/password@host:puerto/servicio
  @sql/create_async_jobs_table.sql
  @sql/create_query_log_table.sql
  @sql/create_test_procedures.sql
```

### 3. Ejecutar API

```bash
# Método 1: Usar el ejecutable directamente
./go-oracle-api.exe

# Método 2: Con archivo .env personalizado
./go-oracle-api.exe otro.env 3000

# La API estará disponible en:
# http://localhost:3000
```

### 4. Verificar Instalación

```bash
# Test ping (requiere token en Authorization header)
curl http://localhost:3000/ping \
  -H "Authorization: Bearer test1"

# Respuesta esperada:
# {"status":"ok"}
```

## 🐛 Bug Fixes en v1.0.0

### CRITICAL: Múltiples parámetros OUT

**Problema Solucionado:**
- API devolvía solo el último parámetro OUT en procedimientos Oracle
- Los demás parámetros OUT se perdían en la ejecución

**Causa:**
- Variables de buffer para OUT parameters se creaban como locales dentro de loops
- Perdían scope después del loop, dejando punteros inválidos

**Solución:**
- Pre-asignar todos los buffers FUERA del loop
- Mantener referencias persistentes durante toda la ejecución
- Indexación secuencial correcta

**Impacto:**
✅ Ahora funciona correctamente con ANY número de OUT parameters
✅ Soporta NUMBER, VARCHAR2, DATE como tipos de salida
✅ Se devuelven TODOS los valores en la respuesta

**Ejemplo de Uso:**
```bash
curl -X POST http://localhost:3000/procedure \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
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

## 📊 Requisitos del Sistema

### Hardware
- CPU: 2 cores mínimo
- RAM: 512 MB mínimo
- Disco: 25 MB (binario + logs)

### Software
- **Windows**: Windows 7 SP1 o superior
- **Linux**: Cualquier distribución con glibc 2.17+
- **macOS**: OS X 10.12 o superior

### Base de Datos
- Oracle 11g o superior
- Acceso a puerto 1521 (configurable)
- Usuario con permisos para crear tablas

## 🔧 Variables de Entorno Disponibles

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `ORACLE_USER` | Usuario de conexión | `USUARIO` |
| `ORACLE_PASSWORD` | Contraseña | `password123` |
| `ORACLE_HOST` | Host de Oracle | `192.168.1.100` |
| `ORACLE_PORT` | Puerto | `1521` |
| `ORACLE_SERVICE` | Service name | `HTEST01` |
| `API_TOKEN` | Token autenticación | `test1` |
| `PORT` | Puerto API | `3000` |
| `MAX_IDLE_CONNECTIONS` | Conexiones ociosas | `10` |
| `MAX_OPEN_CONNECTIONS` | Máx conexiones | `100` |

Ver `docs/CONFIGURACION_ENV.md` para lista completa.

## 🧪 Testing

### Ejemplo Completo (Node.js)
```bash
node examples/ejemplo_completo.js
```

### Tests Unitarios
```bash
node tests/test.js
```

## 📚 Documentación Completa

- **[README.md](README.md)** - Documentación general
- **[GUIA_RAPIDA.md](GUIA_RAPIDA.md)** - Quick reference con ejemplos
- **[docs/USO_Y_PRUEBAS.md](docs/USO_Y_PRUEBAS.md)** - Ejemplos de uso completos
- **[docs/ASYNC_JOBS.md](docs/ASYNC_JOBS.md)** - Sistema de jobs asíncronos
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Despliegue en producción

## 🆘 Support & Troubleshooting

### Error: "connection refused"
- Verificar que Oracle está corriendo
- Verificar credenciales en `.env`
- Verificar host y puerto de Oracle

### Error: "ORA-12514: TNS:listener does not know of service name"
- Verificar que el SERVICE_NAME es correcto
- Ejecutar: `lsnrctl status` en servidor Oracle

### Error: "declared and not used" en compilación
- Asegurarse de usar Go 1.20+
- Ejecutar: `go mod tidy`

Ver `docs/` para troubleshooting detallado.

## 📝 Licencia
MIT License - Ver archivo LICENSE

## 🤝 Contribuciones
Este proyecto fue desarrollado con colaboración de GitHub Copilot.

---

**Versión:** 1.0.0  
**Fecha:** 2026-04-24  
**Status:** Stable ✅
