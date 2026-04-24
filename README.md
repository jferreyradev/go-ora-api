# Go Oracle API Microservicio

Este microservicio en Go expone endpoints HTTP para consultar y modificar una base de datos Oracle, pensado como puente entre Oracle y otras APIs.

## Resumen

En muchos entornos de desarrollo, diferentes aplicaciones necesitan acceder a datos almacenados en bases de datos Oracle. Sin embargo, integrar directamente con Oracle suele requerir la instalación de drivers o librerías específicos en cada entorno, lo que complica la interoperabilidad y el despliegue.

Este microservicio resuelve ese problema actuando como un puente seguro y ligero entre una base de datos Oracle y otras aplicaciones, exponiendo endpoints HTTP para consultas y modificaciones. Así, cualquier sistema capaz de realizar peticiones HTTP/JSON puede interactuar con Oracle sin necesidad de instalar librerías, drivers ni configuraciones adicionales de Oracle en el cliente.

## Ventajas principales
- Acceso centralizado a Oracle mediante HTTP.
- No requiere que los sistemas consumidores instalen librerías de Oracle.
- Permite la integración de APIs y servicios hechos en cualquier lenguaje o framework.
- Permite operaciones de consulta y modificación (SELECT, INSERT, UPDATE, DELETE) a través de una API REST.
- **Soporte completo para procedimientos y funciones de paquetes Oracle**.
- **Soporte para múltiples parámetros OUT** - Devuelve correctamente todos los parámetros de salida (NUMBER, VARCHAR2, DATE).
- **Campo `schema` separado** para especificar el esquema sin ambigüedad.
- **Detección automática de tipos de datos** para parámetros OUT (NUMBER, VARCHAR2).
- **Manejo inteligente de fechas** con conversión automática desde formatos estándar.
- **Consultas multilínea** con normalización automática de saltos de línea.
- Facilita la integración de sistemas modernos (microservicios, aplicaciones web/móviles, otros servicios) con bases de datos Oracle.
- Seguridad mediante autenticación de token y restricción opcional por IP.
- **CORS configurado** para integración desde aplicaciones web frontend.
- Reduce el riesgo de exposición de credenciales o la base de datos a múltiples sistemas.

## Configuración del archivo `.env`

Consulta la guía completa para crear y configurar el archivo de entorno en [`docs/CONFIGURACION_ENV.md`](docs/CONFIGURACION_ENV.md).

## Ejecución

Puedes ejecutar el microservicio de dos formas:

### 1. Desde Go (modo desarrollo)

```sh
go run main.go [archivo_env] [puerto]
```
- `archivo_env` (opcional): Archivo de variables de entorno (por defecto `.env`).
- `puerto` (opcional): Puerto donde escuchará la API (por defecto `8080`).

Ejemplos:
```sh
go run main.go
# o con archivo y puerto personalizados
go run main.go otro.env 9090
```

### 2. Como ejecutable compilado

Primero compila el binario:
```sh
go build -o go-oracle-api.exe main.go
```
Luego ejecútalo:
```sh
./go-oracle-api.exe [archivo_env] [puerto]
```

También puedes usar variables de entorno:
- `ENV_FILE` para el archivo de configuración
- `PORT` para el puerto

Ejemplo:
```sh
set ENV_FILE=otro.env
set PORT=9090
./go-oracle-api.exe
```

## Opciones de ejecución
- Si no se especifica archivo de entorno ni puerto, se usan `.env` y `8080` por defecto.
- Puedes combinar argumentos y variables de entorno según tu preferencia.

### Ejecutar varias instancias con diferentes configuraciones

Puedes tener varios archivos `.env` (por ejemplo, `.env1`, `.env2`, etc.) y ejecutar varias instancias de la app, cada una con su propio archivo, puerto y nombre identificativo:

#### Método manual:
```sh
# Ventana 1 - Producción
go run main.go .env1 8081 "Produccion"

# Ventana 2 - Testing  
go run main.go .env2 8082 "Testing"

# Ventana 3 - Desarrollo
go run main.go .env3 8083 "Desarrollo"
```

#### Con ejecutable compilado:
```sh
go build -o go-oracle-api.exe main.go

start go-oracle-api.exe .env1 8081 "Produccion"
start go-oracle-api.exe .env2 8082 "Testing" 
start go-oracle-api.exe .env3 8083 "Desarrollo"
```

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

