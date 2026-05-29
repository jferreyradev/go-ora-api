# ⏱️ Jobs Asíncronos

Sistema para ejecutar procedimientos de larga duración sin bloquear la API

---

## 📑 Tabla de Contenidos

- [Descripción General](#descripción-general)
- [Modos de Ejecución](#modos-de-ejecución)
- [Configuración](#configuración)
- [Uso Básico](#uso-básico)
- [API Reference](#api-reference)
- [Monitoreo](#monitoreo)
- [Limpieza](#limpieza-de-jobs)
- [Troubleshooting](#troubleshooting)

---

## 🎯 Descripción General

Los jobs asíncronos permiten ejecutar procedimientos sin esperar la respuesta.

### Casos de Uso

- **Procedimientos lentos:** Operaciones de minutos/horas
- **Procesamiento en lote:** Grandes volúmenes de datos
- **Tareas programadas:** Ejecución diferida
- **Mejor UX:** API responde inmediatamente con `job_id`
- **Control de concurrencia:** Evita ejecutar el mismo procedimiento en paralelo cuando no corresponde

### Estados del Job

| Estado | Descripción |
|--------|------------|
| `pending` | Creado, esperando ejecución o turno |
| `running` | Ejecutándose actualmente |
| `completed` | Finalizado exitosamente |
| `failed` | Terminó con error |

### Progreso

Cada job tiene `progress` (0-100):
- 0% - Creado
- 30% - Parámetros procesados
- 50% - Statement preparado
- 80% - Ejecución completa
- 100% - Finalizado

---

## 🔒 Modos de Ejecución

Para procedimientos que no deben ejecutarse en paralelo, la API soporta políticas de concurrencia por job.

### Modos soportados

| `execution_mode` | Comportamiento |
|------------------|----------------|
| `parallel` | Comportamiento por defecto. Permite múltiples ejecuciones simultáneas |
| `sequential` | Permite múltiples jobs, pero se ejecutan uno por uno en orden FIFO |
| `exclusive` | No permite crear un nuevo job si ya existe otro `pending` o `running` para la misma clave |

### ¿Cuándo usar cada modo?

- Usa **`parallel`** para procedimientos independientes.
- Usa **`sequential`** cuando quieras **encolar** ejecuciones del mismo procedimiento.
- Usa **`exclusive`** cuando quieras **rechazar** nuevas solicitudes mientras exista una ejecución activa.

### Clave de bloqueo (`lock_key`)

Opcionalmente, puedes agrupar la exclusión o secuencialidad por una clave lógica en vez de hacerlo solo por nombre de procedimiento.

Ejemplos:
- `CIERRE_MENSUAL`
- `REPORTE:cliente_123`
- `IMPORTACION:lote_2026_05_29`

Si no se envía `lock_key`, la API puede usar el `name` del procedimiento como clave por defecto.

### Reglas de negocio

#### `sequential`

- El job siempre se crea con estado `pending`.
- Solo un job con la misma `lock_key` puede pasar a `running` al mismo tiempo.
- Los siguientes jobs quedan en cola y se ejecutan por orden de creación.

#### `exclusive`

- Antes de crear el job, la API verifica si existe otro job con la misma `lock_key` en estado `pending` o `running`.
- Si existe, la API responde `409 Conflict`.
- Si no existe, el job se crea normalmente.

> **Recomendación:** para producción, define el `execution_mode` por configuración del backend o catálogo de procedimientos, no solo desde el cliente.

---

## ⚙️ Configuración

### 1. Crear Tabla

Ejecuta el script SQL:

```bash
sqlplus usuario/password@db @sql/create_async_jobs_table.sql
```

### 2. Extender tabla para control de concurrencia

Si vas a usar ejecución secuencial o exclusiva, añade columnas como:

```sql
ALTER TABLE ASYNC_JOBS ADD (
  execution_mode VARCHAR2(20) DEFAULT 'parallel' NOT NULL,
  lock_key       VARCHAR2(200)
);
```

### 3. Índices recomendados

```sql
CREATE INDEX IDX_ASYNC_JOBS_STATUS_NAME ON ASYNC_JOBS(status, name);
CREATE INDEX IDX_ASYNC_JOBS_LOCK_KEY_STATUS ON ASYNC_JOBS(lock_key, status, created_at);
```

### 4. Crear Procedimientos de Prueba (Opcional)

```bash
sqlplus usuario/password@db @sql/create_test_procedures.sql
```

Procedimientos creados:
- `PROC_TEST` - Procedimiento simple
- `PROC_TEST_DEMORA` - Simula operación lenta
- `PROC_TEST_PARAMS` - Múltiples parámetros
- `PROC_TEST_ERROR` - Manejo de errores

---

## 🚀 Uso Básico

### Crear Job Asíncrono

```javascript
const response = await fetch('http://localhost:3000/procedure/async', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer test1'
  },
  body: JSON.stringify({
    name: "PROC_TEST_DEMORA",
    execution_mode: "parallel",
    params: [
      { 
        name: "segundos", 
        value: 5, 
        direction: "IN", 
        type: "NUMBER" 
      }
    ]
  })
});

const data = await response.json();
console.log('Job ID:', data.job_id);
console.log('Status:', data.status); // "pending"
```

### Crear Job Secuencial

```javascript
const response = await fetch('http://localhost:3000/procedure/async', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer test1'
  },
  body: JSON.stringify({
    name: "PROC_CIERRE_MENSUAL",
    execution_mode: "sequential",
    lock_key: "CIERRE_MENSUAL",
    params: []
  })
});

const data = await response.json();
console.log('Job ID:', data.job_id);
console.log('Status:', data.status); // "pending"
console.log('Queue mode:', data.execution_mode); // "sequential"
```

### Crear Job Exclusivo

```javascript
const response = await fetch('http://localhost:3000/procedure/async', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer test1'
  },
  body: JSON.stringify({
    name: "PROC_REINDEXAR",
    execution_mode: "exclusive",
    lock_key: "PROC_REINDEXAR",
    params: []
  })
});

if (response.status === 409) {
  const err = await response.json();
  console.error(err.error);
}
```

### Monitorear Progreso

```javascript
const checkJob = async (jobId) => {
  const response = await fetch(`http://localhost:3000/jobs/${jobId}`, {
    headers: { 'Authorization': 'Bearer test1' }
  });
  
  const job = await response.json();
  return job;
};

// Polling cada 2 segundos
const jobId = data.job_id;
const interval = setInterval(async () => {
  const job = await checkJob(jobId);
  console.log(`[${job.status}] ${job.progress}%`);
  
  if (job.status === 'completed' || job.status === 'failed') {
    clearInterval(interval);
    console.log('Resultado:', job.result);
  }
}, 2000);
```

---

## 📡 API Reference

### Crear Job (`POST /procedure/async`)

```bash
curl -X POST http://localhost:3000/procedure/async \
  -H "Authorization: Bearer test1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MI_PROCEDIMIENTO",
    "execution_mode": "sequential",
    "lock_key": "MI_PROCEDIMIENTO",
    "params": [
      {"name": "param1", "value": "valor", "direction": "IN"}
    ]
  }'
```

### Campos soportados en el request

| Campo | Tipo | Requerido | Descripción |
|------|------|-----------|-------------|
| `name` | string | Sí | Nombre del procedimiento Oracle |
| `params` | array | No | Lista de parámetros |
| `execution_mode` | string | No | `parallel`, `sequential` o `exclusive`. Default: `parallel` |
| `lock_key` | string | No | Clave lógica para serializar o bloquear ejecuciones |

**Respuesta exitosa (202 Accepted):**
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "pending",
  "execution_mode": "sequential",
  "lock_key": "MI_PROCEDIMIENTO"
}
```

**Respuesta si el modo es `exclusive` y ya existe un job activo (409 Conflict):**
```json
{
  "error": "Ya existe un job activo para la clave MI_PROCEDIMIENTO",
  "code": "JOB_ALREADY_RUNNING"
}
```

### Ver Estado (`GET /jobs/{job_id}`)

```bash
curl http://localhost:3000/jobs/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer test1"
```

**Respuesta:**
```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "MI_PROCEDIMIENTO",
  "status": "running",
  "progress": 50,
  "execution_mode": "sequential",
  "lock_key": "MI_PROCEDIMIENTO",
  "result": null,
  "error": null,
  "created_at": "2026-04-24T10:30:00Z",
  "updated_at": "2026-04-24T10:30:02Z"
}
```

### Listar Jobs (`GET /jobs`)

```bash
curl http://localhost:3000/jobs \
  -H "Authorization: Bearer test1"
```

**Opciones:**
- `?status=completed` - Solo completados
- `?status=failed` - Solo con error
- `?limit=10` - Máximo 10 resultados
- `?execution_mode=sequential` - Filtrar por política de ejecución
- `?lock_key=MI_PROCEDIMIENTO` - Filtrar por clave lógica

### Eliminar Job (`DELETE /jobs/{job_id}`)

```bash
curl -X DELETE http://localhost:3000/jobs/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer test1"
```

**Opciones:**
- `?status=completed` - Eliminar todos los completados
- `?older_than=7` - Eliminar más antiguos que 7 días

### Lógica recomendada del worker

Para respetar `sequential` y `exclusive`, el worker debe validar la política antes de mover un job de `pending` a `running`.

#### Verificar job activo por `lock_key`

```sql
SELECT COUNT(*) AS active_jobs
FROM ASYNC_JOBS
WHERE NVL(lock_key, name) = NVL(:lock_key, :name)
  AND status IN ('pending', 'running')
  AND job_id <> :job_id;
```

#### Obtener siguiente job secuencial

```sql
SELECT *
FROM ASYNC_JOBS
WHERE NVL(lock_key, name) = NVL(:lock_key, :name)
  AND status = 'pending'
ORDER BY created_at ASC;
```

> **Importante:** el control de concurrencia debe hacerse dentro de una transacción o usando un mecanismo de locking en base de datos para evitar race conditions.

---

## 📊 Monitoreo

### Ver Logs de Job

```bash
curl http://localhost:3000/logs \
  -H "Authorization: Bearer test1"
```

### Estadísticas de Jobs

Consultar tabla `ASYNC_JOBS`:

```sql
SELECT 
  COUNT(*) as total,
  SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completados,
  SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as en_ejecucion,
  SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as fallidos,
  SUM(CASE WHEN execution_mode = 'sequential' THEN 1 ELSE 0 END) as secuenciales,
  SUM(CASE WHEN execution_mode = 'exclusive' THEN 1 ELSE 0 END) as exclusivos
FROM ASYNC_JOBS;
```

### Ver colas por clave lógica

```sql
SELECT 
  NVL(lock_key, name) AS queue_key,
  status,
  COUNT(*) AS total
FROM ASYNC_JOBS
GROUP BY NVL(lock_key, name), status
ORDER BY queue_key, status;
```

---

## 🧹 Limpieza de Jobs

### Eliminar Completados

```bash
curl -X DELETE "http://localhost:3000/jobs?status=completed" \
  -H "Authorization: Bearer test1"
```

### Eliminar por Antigüedad

```bash
# Eliminar jobs más antiguos que 7 días
curl -X DELETE "http://localhost:3000/jobs?older_than=7" \
  -H "Authorization: Bearer test1"

# Eliminar más antiguos que 30 días
curl -X DELETE "http://localhost:3000/jobs?older_than=30" \
  -H "Authorization: Bearer test1"
```

### Limpiar Automáticamente

SQL para limpiar jobs antiguos:

```sql
-- Borrar jobs completados hace más de 30 días
DELETE FROM ASYNC_JOBS
WHERE status = 'completed'
AND created_at < SYSDATE - 30;

-- Borrar jobs fallidos hace más de 7 días
DELETE FROM ASYNC_JOBS
WHERE status = 'failed'
AND created_at < SYSDATE - 7;

COMMIT;
```

---

## 🆘 Troubleshooting

### Job nunca inicia

**Causa:** Tabla `ASYNC_JOBS` no existe  
**Solución:** Ejecutar `sql/create_async_jobs_table.sql`

### Jobs secuenciales quedan siempre en `pending`

**Causa posible:** Existe otro job `running` con la misma `lock_key`  
**Solución:**
1. Consultar jobs por `lock_key`
2. Verificar si hay jobs colgados en `running`
3. Revisar logs del worker
4. Corregir o limpiar jobs bloqueados

### Job exclusivo devuelve `409 Conflict`

**Causa:** Ya existe un job `pending` o `running` con la misma `lock_key`  
**Solución:**
1. Esperar a que finalice el job activo
2. Consultar `GET /jobs?lock_key=...`
3. Si el job quedó colgado, revisar logs o reiniciar el worker

### Job se queda en `running`

**Causa:** Procedimiento colgado en Oracle  
**Solución:**
1. Ver logs: `curl http://localhost:3000/logs`
2. Verificar en Oracle: `SELECT * FROM v$session`
3. Matar sesión si es necesario
4. Reiniciar API

### No puedo ver resultados

**Causa:** Campo `result` es NULL  
**Solución:**
1. Verificar procedimiento devuelve parámetros OUT
2. Especificar `"direction": "OUT"` en parámetros
3. Ver logs de error

### Job falla con error

**Síntomas:**
- `status: "failed"`
- `error: "ORA-..."`

**Solución:**
1. Leer mensaje de error completo
2. Ejecutar procedimiento directamente en SQL*Plus
3. Verificar permisos del usuario Oracle

---

Para más detalles, ver [docs/INDEX.md](../INDEX.md)
