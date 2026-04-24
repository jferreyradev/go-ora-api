# Go Oracle API Microservicio

Puente HTTP seguro y ligero entre Oracle y tus aplicaciones.

## 🚀 Inicio Rápido

**Documentación completa en:** [📚 docs/INDEX.md](docs/INDEX.md) ⭐

### 5 Minutos para Empezar

```bash
# 1. Configurar
cp .env.example .env
# Editar .env con credenciales Oracle

# 2. Iniciar
go run main.go
# o usar: ./go-oracle-api.exe

# 3. Probar
curl http://localhost:3000/ping \
  -H "Authorization: Bearer test1"
```

## 📚 Documentación

**👉 [Abre docs/INDEX.md](docs/INDEX.md)** para:
- Guías por rol (Principiante, Desarrollador, DevOps, Release Manager)
- Búsqueda rápida de información
- Rutas recomendadas de lectura
- Todos los archivos de documentación

## ✨ Características

✅ API RESTful para Oracle  
✅ Procedimientos y Funciones con IN/OUT  
✅ Múltiples OUT parameters (FIXED en v1.0.0)  
✅ Jobs Asíncronos  
✅ Autenticación + Restricción de IPs  
✅ Manejo automático de tipos (NUMBER, VARCHAR2, DATE)  

## 📦 Contenido

- `main.go` - Código fuente (~2000 líneas)
- `go-oracle-api.exe` - Ejecutable compilado
- `examples/` - Ejemplos de uso
- `sql/` - Scripts de setup
- `docs/` - Documentación completa

## 📋 Requisitos

- Go 1.20+ o ejecutable precompilado
- Oracle 11g o superior
- .env configurado

## 🎯 Búsqueda Rápida

| Necesito | Ver |
|----------|-----|
| Instalar rápido | [docs/getting-started/QUICKSTART.md](docs/getting-started/QUICKSTART.md) |
| Ejemplos de código | [docs/getting-started/GUIA_RAPIDA.md](docs/getting-started/GUIA_RAPIDA.md) |
| Configurar .env | [docs/getting-started/CONFIGURACION.md](docs/getting-started/CONFIGURACION.md) |
| API Reference | [docs/api-reference/ENDPOINTS.md](docs/api-reference/ENDPOINTS.md) |
| Desplegar producción | [docs/deployment/DEPLOYMENT.md](docs/deployment/DEPLOYMENT.md) |
| Publicar en GitHub | [docs/release/PUBLISH_GUIDE.md](docs/release/PUBLISH_GUIDE.md) |
| **Todo lo anterior** | **[docs/INDEX.md](docs/INDEX.md)** ⭐ |

## 📝 Licencia

MIT License - Ver [LICENSE](LICENSE)

---

**Comienza aquí:** [📚 docs/INDEX.md](docs/INDEX.md)

#### Script automatizado:
```bash
# Dar permisos de ejecución (primera vez)
chmod +x scripts/*.sh

# Ejecutar scripts
./scripts/run_multiple_instances.sh
./scripts/monitor_instances.sh
```

### Identificación de instancias

Cada instancia se identifica de las siguientes maneras:

1. **Título de ventana**: `Go Oracle API - [Nombre] (Puerto XXXX)`
   - ✅ **Windows**: Título en barra de tareas y ventana CMD
   - ✅ **Linux/macOS**: Título en terminal (terminales compatibles)
   - ✅ **Multiplataforma**: Funciona en todas las plataformas
2. **Log individual**: `log/[Nombre]_YYYY-MM-DD_HH-MM-SS.log`
3. **Mensaje de inicio**: Muestra el nombre de la instancia en consola
4. **Puerto único**: Cada instancia escucha en un puerto diferente

### Ventajas del sistema de instancias

- **Logs separados**: Cada instancia tiene su propio archivo de log
- **Identificación visual**: Títulos de ventana personalizados
- **Configuración independiente**: Cada instancia usa su propio .env
- **Monitoreo centralizado**: Scripts para verificar estado y logs
- **Gestión simplificada**: Detener/iniciar instancias específicas

## Endpoints disponibles

- **`/ping`** - Verificación de estado y conectividad con Oracle
- **`/query`** - Ejecutar consultas SELECT (soporta multilínea)
- **`/exec`** - Ejecutar sentencias de modificación (INSERT, UPDATE, DELETE, DDL)
- **`/procedure`** - Ejecutar procedimientos y funciones de paquetes Oracle (síncrono)
- **`/procedure/async`** - Ejecutar procedimientos de larga duración en segundo plano
- **`/jobs/{id}`** - Consultar estado de un job asíncrono específico
- **`/jobs`** - Listar y gestionar jobs asíncronos (GET, DELETE)
- **`/upload`** - Subir archivos como BLOB a la base de datos
- **`/logs`** - Consultar logs de consultas ejecutadas
- **`/docs`** - Documentación integrada

### 📋 Sistema de Jobs Asíncronos

El sistema de jobs permite ejecutar procedimientos en segundo plano con monitoreo en tiempo real:

```javascript
// Crear job
const res = await fetch('/procedure/async', {
  method: 'POST',
  body: JSON.stringify({
    name: "PROC_LARGO",
    params: [{ name: "p1", value: 100 }]
  })
});
const { job_id } = await res.json();

// Monitorear progreso
const job = await fetch(`/jobs/${job_id}`).then(r => r.json());
console.log(`Estado: ${job.status} (${job.progress}%)`);
```

**Características:**
- ✅ Ejecución no bloqueante
- ✅ Progreso en tiempo real (0-100%)
- ✅ Persistencia en Oracle (sobrevive a reinicios)
- ✅ Limpieza automática de jobs antiguos
- ✅ Mensajes de error mejorados

**Documentación completa:** [docs/ASYNC_JOBS.md](docs/ASYNC_JOBS.md)

## Funcionalidades destacadas

### 🔧 Procedimientos y Funciones de Paquetes

El backend maneja automáticamente la nomenclatura de objetos Oracle mediante la función helper `formatObjectName()`, que centraliza la lógica de formateo en un solo lugar.

**Uso con campo `schema` (recomendado para claridad):**
```json
{
  "schema": "WORKFLOW",
  "name": "MI_FUNCION",
  "isFunction": true,
  "params": [
    { "name": "result", "direction": "OUT", "type": "number" },
    { "name": "input_param", "value": 123 }
  ]
}
```

**Procedimiento con múltiples parámetros OUT:**
```json
{
  "name": "PRUEBA1",
  "params": [
    { "name": "vIDPERS", "value": 123, "direction": "IN", "type": "NUMBER" },
    { "name": "vDNI", "value": 45678901, "direction": "IN", "type": "NUMBER" },
    { "name": "vSALIDA", "direction": "OUT", "type": "NUMBER" },
    { "name": "vError", "direction": "OUT", "type": "NUMBER" },
    { "name": "vErrorMsg", "direction": "OUT", "type": "STRING" }
  ]
}
```

**Respuesta (todos los OUT parameters se devuelven):**
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

**Uso tradicional (esquema.paquete.función):**
```json
{
  "name": "SCHEMA.PACKAGE.FUNCTION_NAME",
  "isFunction": true,
  "params": [
    { "name": "input_param", "value": 123 },
    { "name": "result", "direction": "OUT", "type": "number" }
  ]
}
```

**⚠️ Nota sobre conflictos de nomenclatura:** Si existe un PACKAGE con el mismo nombre que un SCHEMA/USER, Oracle interpretará `SCHEMA.FUNCION` como `PACKAGE.FUNCION`. En estos casos, usa sinónimos:
```sql
CREATE SYNONYM EXISTE_PROC_CAB FOR WORKFLOW.EXISTE_PROC_CAB;
```

### 📅 Manejo Automático de Fechas
```json
{
  "name": "MY_PROCEDURE", 
  "params": [
    { "name": "fecha_param", "value": "2025-10-21" },
    { "name": "periodo", "value": "21/10/2025" }
  ]
}
```

### 📝 Consultas Multilínea
```json
{
  "query": "SELECT campo1, campo2\nFROM mi_tabla\nWHERE condicion = 'valor'"
}
```

## 📚 Documentación

### Guías Principales
- **[GUIA_RAPIDA.md](GUIA_RAPIDA.md)** - ⭐ Guía de inicio rápido y referencia

### Documentación Detallada
- **[ASYNC_JOBS.md](docs/ASYNC_JOBS.md)** - Sistema de jobs asíncronos
- **[SCHEMA_FIELD.md](docs/SCHEMA_FIELD.md)** - Campo schema y nomenclatura Oracle
- **[USO_Y_PRUEBAS.md](docs/USO_Y_PRUEBAS.md)** - Ejemplos de uso completos
- **[CONFIGURACION_ENV.md](docs/CONFIGURACION_ENV.md)** - Variables de entorno
- **[DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Despliegue en producción
- **[FIREWALL_WINDOWS.md](docs/FIREWALL_WINDOWS.md)** - Configuración de firewall

### 🧪 Ejemplo y Tests

```bash
# Ejecutar ejemplo completo (demuestra todas las funcionalidades)
node examples/ejemplo_completo.js

# Ejecutar suite de tests completa (7 tests)
node tests/test.js

# Ejecutar test específico
node tests/test.js ping
node tests/test.js query
node tests/test.js procedure
```

---

## Créditos y autoría

Este proyecto fue desarrollado en colaboración entre [jferreyradev](https://github.com/jferreyradev/jferreyradev) y GitHub Copilot, combinando experiencia humana y asistencia de IA para lograr una solución robusta y documentada.

## Licencia
MIT

