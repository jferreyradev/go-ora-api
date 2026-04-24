# ⏱️ Jobs Asíncronos

Sistema para ejecutar procedimientos de larga duración sin bloquear la API

---

## 📑 Tabla de Contenidos

- [Descripción General](#descripción-general)
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

### Estados del Job

| Estado | Descripción |
|--------|------------|
| `pending` | Creado, esperando ejecución |
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

## ⚙️ Configuración

### 1. Crear Tabla

Ejecuta el script SQL:

```bash
sqlplus usuario/password@db @sql/create_async_jobs_table.sql
```

### 2. Crear Procedimientos de Prueba (Opcional)

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
    "params": [
      {"name": "param1", "value": "valor", "direction": "IN"}
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

### Eliminar Job (`DELETE /jobs/{job_id}`)

```bash
curl -X DELETE http://localhost:3000/jobs/a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
  -H "Authorization: Bearer test1"
```

**Opciones:**
- `?status=completed` - Eliminar todos los completados
- `?older_than=7` - Eliminar más antiguos que 7 días

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
  SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as fallidos
FROM ASYNC_JOBS;
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

### Job se queda en "running"

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
2. Ejecutar procedimiento diramente en SQL*Plus
3. Verificar permisos del usuario Oracle

---

Para más detalles, ver [docs/INDEX.md](../INDEX.md)
