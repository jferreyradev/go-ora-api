# ⚡ Inicio Rápido - Go Oracle API

**5 minutos para tener la API funcionando**

> ⚠️ **IMPORTANTE**: Debes especificar `--env` o `--config` (son mutuamente excluyentes)

---

## 📥 Instalación (Opción 1: Ejecutable)

### Windows - Opción A: Cargar desde .env
```bash
# 1. Descargar go-oracle-api.exe
# 2. Crear archivo .env con credenciales Oracle
cp .env.example .env

# 3. Ejecutar especificando --env
.\go-oracle-api.exe --env .env
# o forma corta: .\go-oracle-api.exe -e .env
```

### Windows - Opción B: Cargar desde config.yaml
```bash
# 1. Descargar go-oracle-api.exe
# 2. Crear config.yaml con credenciales Oracle

# 3. Ejecutar especificando --config
.\go-oracle-api.exe --config config.yaml
# o forma corta: .\go-oracle-api.exe -c config.yaml
```

**API disponible en:** http://localhost:8080 (o el puerto configurado)

### Linux/macOS
```bash
# Descargar el binario precompilado
chmod +x go-oracle-api

# Opción A: Usar .env
./go-oracle-api --env .env

# Opción B: Usar config.yaml
./go-oracle-api --config config.yaml
```

---

## 🔨 Instalación (Opción 2: Compilar desde código)

### Requisitos
- Go 1.20 o superior
- Oracle 11g o superior
- Git

### Pasos
```bash
# 1. Clonar repositorio
git clone https://github.com/tu-usuario/go-oracle-api.git
cd go-oracle-api

# 2. Copiar y configurar (elegir UNO)
cp .env.example .env        # O: crear config.yaml
# → Editar .env o config.yaml con tus credenciales Oracle

# 3. Compilar (opcional)
go build -o go-oracle-api.exe

# 4. Ejecutar (especificar --env O --config)
go run main.go --env .env
# o: go run main.go --config config.yaml
# o usar el ejecutable compilado:
# .\go-oracle-api.exe --env .env
# .\go-oracle-api.exe --config config.yaml
```

---

## ⚙️ Configuración (.env o config.yaml)

### Opción A: Usar .env
Crea un archivo `.env` en la carpeta del proyecto:

```env
# Base de datos Oracle
ORACLE_USER=tu_usuario
ORACLE_PASSWORD=tu_contraseña
ORACLE_HOST=localhost
ORACLE_PORT=1521
ORACLE_SERVICE=tu_servicio
ORACLE_CONNECTION_TIMEOUT=30
ORACLE_PROXY_SCHEMA=tu_schema_proxy

# Seguridad
API_TOKEN=token_seguro_123
API_ALLOWED_IPS=127.0.0.1,::1,localhost

# Puerto de escucha
PORT=8080
```

Luego ejecuta:
```bash
go run main.go --env .env
```

### Opción B: Usar config.yaml
Crea un archivo `config.yaml` en la carpeta del proyecto:

```yaml
oracle:
  user: tu_usuario
  password: tu_contraseña
  host: localhost
  port: 1521
  service: tu_servicio
  connection_timeout: "30"
  proxy_schema: tu_schema_proxy

api:
  token: token_seguro_123
  allowed_ips:
    - 127.0.0.1
    - ::1
    - localhost
  no_auth: false

server:
  port: 8080
```

Luego ejecuta:
```bash
go run main.go --config config.yaml
```

O usar `config.yaml` como alternativa:

```yaml
oracle:
  user: tu_usuario
  password: tu_contraseña
  host: localhost
  port: 1521
  service: tu_servicio

api:
  token: token_seguro_123
  allowed_ips:
    - 127.0.0.1
    - ::1
    - localhost
  no_auth: false

server:
  port: 3000
```

> Si existen variables de entorno, `.env` y `config.yaml`, la precedencia es: variables de entorno del proceso → `.env` → `config.yaml` → defaults.

> **Configuración completa:** Ver [CONFIGURACION.md](CONFIGURACION.md) para todas las variables disponibles, ejemplos y mejores prácticas.

---

## 🧪 Pruebas Básicas

### 0. Verificar requisitos (recomendado)

Antes de iniciar, verifica que todo esté configurado correctamente:

```bash
# Verificar configuración y requisitos
go run main.go --check
# o con el ejecutable:
./go-oracle-api.exe --check
```

**Salida esperada:**
```
======================================
🔍 VERIFICACIÓN DE REQUISITOS
======================================

[1/6] Configuración (.env)
  ✅ Archivo .env encontrado
  ✅ ORACLE_USER configurado
  ✅ ORACLE_PASSWORD configurado
  ...

✅ SISTEMA LISTO PARA EJECUTAR
```

Si hay errores, el comando te indicará qué corregir antes de iniciar.

### 1. Verificar conexión

```bash
curl http://localhost:3000/ping \
  -H "Authorization: Bearer token_seguro_123"
```

**Respuesta esperada:**
```json
{"status":"ok"}
```

### 2. Ejecutar consulta

```bash
curl -X POST http://localhost:3000/query \
  -H "Authorization: Bearer token_seguro_123" \
  -H "Content-Type: application/json" \
  -d '{"query":"SELECT sysdate FROM dual"}'
```

### 3. Ejecutar procedimiento

```bash
curl -X POST http://localhost:3000/procedure \
  -H "Authorization: Bearer token_seguro_123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MI_PROCEDIMIENTO",
    "params": [
      {"name": "entrada", "value": "test", "direction": "IN"},
      {"name": "salida", "direction": "OUT", "type": "STRING"}
    ]
  }'
```

---

## 📚 Siguientes Pasos

| Necesito... | Ver... |
|-----------|---------|
| Ejemplo funcional completo | [Guía Rápida](GUIA_RAPIDA.md) |
| Configuración avanzada | [Configuración](CONFIGURACION.md) |
| Todos los endpoints | [API Reference](../api-reference/ENDPOINTS.md) |
| Desplegar a producción | [Deployment](../deployment/DEPLOYMENT.md) |

---

## 🆘 Problemas Comunes

### ❌ "Connection refused"
- Verificar que Oracle está corriendo
- Verificar HOST, PORT y SERVICE en `.env`
- Ver [Troubleshooting](../deployment/DEPLOYMENT.md#troubleshooting)

### ❌ "Unauthorized"
- Verificar token en `.env` coincida con header `Authorization`
- Token debe ir como: `Authorization: Bearer tu_token`

### ❌ "ORA-01017: invalid username/password"
- Verificar credenciales en `.env`
- Usuario Oracle debe existir y tener permisos

### ❌ Procedimiento no encuentra parámetros OUT
- Verificar que no haya parámetros OUT sin especificar `"direction": "OUT"`
- Para fechas, incluir `"type": "date"`
- Ver [Parámetros OUT](../api-reference/SCHEMA_FIELD.md)

---

## ✅ Checklist de Inicio

- [ ] Go 1.20+ instalado (o ejecutable descargado)
- [ ] Oracle accesible y funcionando
- [ ] Archivo `.env` o `config.yaml` creado y configurado
- [ ] Verificación de requisitos exitosa (`--check`)
- [ ] Ping endpoint responde OK
- [ ] Puedo ejecutar una consulta SELECT

**Tip:** Usa `go run main.go --check` para verificar todos los requisitos automáticamente.

---

**¿Necesitas ayuda?** Ver [Documentación Completa](../INDEX.md)
