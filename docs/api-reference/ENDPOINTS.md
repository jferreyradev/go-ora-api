# 📡 Referencia de Endpoints

Documentación completa de todos los endpoints disponibles

---

## 📋 Tabla de Contenidos

- [Health Check](#health-check-ping)
- [Consultas](#consultas-query--exec)
- [Procedimientos](#procedimientos-procedure)
- [Jobs Asíncronos](#jobs-asíncronos)
- [Upload/Download](#uploaddow nload)
- [Logs](#logs)

---

## Health Check (`/ping`)

Verifica si la API está activa y conectada a Oracle.

**Método:** `GET`

**URL:** `http://localhost:3000/ping`

**Headers:**
```
Authorization: Bearer <API_TOKEN>
```

**Ejemplo:**
```bash
curl -H "Authorization: Bearer test1" http://localhost:3000/ping
```

**Respuesta (200 OK):**
```json
{"status":"ok"}
```

---

## Consultas

### Query (`/query`)

Ejecuta **consultas SELECT** en Oracle.

**Método:** `POST`

**Content-Type:** `application/json`

**Headers:**
```
Authorization: Bearer <API_TOKEN>
Content-Type: application/json
```

**Body:**
```json
{
  "query": "SELECT * FROM tabla WHERE id = 1"
}
```

**Ejemplo:**
```bash
curl -X POST http://localhost:3000/query \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{"query":"SELECT sysdate FROM dual"}'
```

**Respuesta (200 OK):**
```json
{
  "results": [
    {"SYSDATE": "2026-04-24"}
  ]
}
```

#### Consultas Multilínea

```bash
curl -X POST http://localhost:3000/query \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT campo1, campo2\nFROM mi_tabla\nWHERE condicion = '\''valor'\''"
  }'
```

---

### Exec (`/exec`)

Ejecuta **INSERT, UPDATE, DELETE, DDL**.

**Método:** `POST`

**Content-Type:** `application/json`

**Body:**
```json
{
  "query": "INSERT INTO tabla (id, nombre) VALUES (1, 'test')"
}
```

**Ejemplo CREATE:**
```bash
curl -X POST http://localhost:3000/exec \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{"query":"CREATE TABLE test_tabla (id NUMBER)"}'
```

**Respuesta (200 OK):**
```json
{
  "status": "ok",
  "rows_affected": 1
}
```

---

## Procedimientos (`/procedure`)

Ejecuta procedimientos y funciones Oracle.

**Método:** `POST`

**Content-Type:** `application/json`

### Procedimiento Simple

```bash
curl -X POST http://localhost:3000/procedure \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mi_procedimiento",
    "params": [
      {"name": "p_input", "value": "test", "direction": "IN"},
      {"name": "p_output", "direction": "OUT", "type": "STRING"}
    ]
  }'
```

**Respuesta:**
```json
{
  "status": "ok",
  "out": {
    "p_output": "resultado"
  }
}
```

### Función de Paquete

```bash
curl -X POST http://localhost:3000/procedure \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "usuario.MI_PACKAGE.MI_FUNCION",
    "isFunction": true,
    "params": [
      {"name": "vDNI", "value": 26579673, "direction": "IN"},
      {"name": "resultado", "direction": "OUT", "type": "number"}
    ]
  }'
```

### Función en Otro Schema

```bash
curl -X POST http://localhost:3000/procedure \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{
    "schema": "WORKFLOW",
    "name": "MI_FUNCION",
    "isFunction": true,
    "params": [
      {"name": "resultado", "direction": "OUT", "type": "number"},
      {"name": "p_param1", "value": 100, "direction": "IN"}
    ]
  }'
```

### Múltiples OUT Parameters

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

### Parámetro con Fecha

```bash
curl -X POST http://localhost:3000/procedure \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mi_proc_con_fecha",
    "params": [
      {"name": "vFecha", "value": "2026-04-24", "direction": "IN"},
      {"name": "resultado", "direction": "OUT", "type": "date"}
    ]
  }'
```

**Respuesta:**
```json
{
  "status": "ok",
  "out": {
    "resultado": "24-04-2026"
  }
}
```

#### Tipos de Datos Soportados

| Type | Formato | Ejemplo |
|------|---------|---------|
| `NUMBER` | Numérico | `{"name": "p", "value": 123, "type": "NUMBER"}` |
| `STRING` | Texto | `{"name": "p", "value": "texto", "type": "STRING"}` |
| `DATE` | yyyy-mm-dd | `{"name": "p", "value": "2026-04-24", "type": "DATE"}` |

#### Detección Automática

- **IN parameters:** Se detectan por el valor JSON
- **OUT parameters:** Especificar `"type"` siempre, especialmente para DATE

---

## Jobs Asíncronos

### Crear Job (`POST /procedure/async`)

Ejecuta un procedimiento sin esperar respuesta.

```bash
curl -X POST http://localhost:3000/procedure/async \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "PROC_TEST_DEMORA",
    "params": [
      {"name": "segundos", "value": 5, "direction": "IN", "type": "NUMBER"}
    ]
  }'
```

**Respuesta (202 Accepted):**
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "pending"
}
```

### Ver Estado del Job (`GET /jobs/{job_id}`)

```bash
curl http://localhost:3000/jobs/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer test1"
```

**Respuesta:**
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "PROC_TEST_DEMORA",
  "status": "running",
  "progress": 50,
  "result": null,
  "error": null,
  "created_at": "2026-04-24T10:30:00Z",
  "updated_at": "2026-04-24T10:30:02Z"
}
```

#### Estados Posibles

| Estado | Significado |
|--------|------------|
| `pending` | Creado, esperando ejecución |
| `running` | Ejecutándose |
| `completed` | Finalizado exitosamente |
| `failed` | Terminó con error |

### Listar Todos los Jobs (`GET /jobs`)

```bash
curl http://localhost:3000/jobs \
  -H "Authorization: Bearer test1"
```

**Respuesta:**
```json
{
  "jobs": [
    {
      "job_id": "...",
      "name": "PROC_TEST_DEMORA",
      "status": "completed",
      "progress": 100,
      "result": {...},
      "error": null
    }
  ],
  "total": 1
}
```

### Filtrar por Estado

```bash
curl "http://localhost:3000/jobs?status=completed" \
  -H "Authorization: Bearer test1"
```

### Eliminar Job (`DELETE /jobs/{job_id}`)

```bash
curl -X DELETE http://localhost:3000/jobs/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer test1"
```

**Respuesta:**
```json
{"status": "ok", "message": "Job deleted"}
```

### Eliminar Múltiples Jobs

```bash
# Jobs completados
curl -X DELETE "http://localhost:3000/jobs?status=completed" \
  -H "Authorization: Bearer test1"

# Jobs más antiguos que N días
curl -X DELETE "http://localhost:3000/jobs?older_than=7" \
  -H "Authorization: Bearer test1"
```

---

## Upload/Download

### Subir Archivo (`POST /upload`)

**Método:** `POST` (multipart/form-data)

```bash
curl -X POST http://localhost:3000/upload \
  -H "Authorization: Bearer test1" \
  -F "file=@mi_archivo.txt" \
  -F "descripcion=Archivo de prueba"
```

**Respuesta:**
```json
{
  "status": "ok",
  "blob_id": "12345",
  "file_name": "mi_archivo.txt",
  "file_size": 1024
}
```

### Descargar Archivo (`GET /download`)

```bash
curl -O "http://localhost:3000/download?blob_id=12345" \
  -H "Authorization: Bearer test1"
```

---

## Logs

### Ver Logs (`GET /logs`)

```bash
curl http://localhost:3000/logs \
  -H "Authorization: Bearer test1"
```

**Respuesta:**
```json
{
  "logs": [
    "2026-04-24 10:30:00 INFO: Conectado a Oracle",
    "2026-04-24 10:30:01 INFO: Request /procedure recibido",
    "..."
  ]
}
```

---

## Códigos HTTP

| Código | Significado |
|--------|------------|
| `200 OK` | Éxito |
| `202 Accepted` | Job asíncrono creado |
| `400 Bad Request` | Datos inválidos |
| `401 Unauthorized` | Token inválido o falta |
| `403 Forbidden` | IP no permitida |
| `500 Internal Error` | Error en Oracle o servidor |

---

## Autenticación

Todos los endpoints requieren autenticación:

```
Authorization: Bearer <API_TOKEN>
```

El token se configura en `.env` como `API_TOKEN`.

---

Para más detalles, ver [docs/INDEX.md](../INDEX.md)
