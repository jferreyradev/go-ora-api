# ⚡ Inicio Rápido - Go Oracle API

**5 minutos para tener la API funcionando**

---

## 📥 Instalación (Opción 1: Ejecutable)

### Windows
```bash
# 1. Descargar go-oracle-api.exe
# 2. Crear archivo .env en la misma carpeta
# 3. Copiar y editar:
cp .env.example .env

# 4. Ejecutar
.\go-oracle-api.exe
```

**API disponible en:** http://localhost:3000

### Linux/macOS
```bash
# Descargar el binario precompilado
chmod +x go-oracle-api
./go-oracle-api
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

# 2. Copiar y configurar
cp .env.example .env
# → Editar .env con tus credenciales Oracle

# 3. Compilar (opcional)
go build -o go-oracle-api.exe

# 4. Ejecutar
go run main.go
# o usar el ejecutable compilado:
./go-oracle-api.exe
```

---

## ⚙️ Configuración (.env)

Crear archivo `.env` en la carpeta del proyecto:

```env
# Base de datos Oracle
ORACLE_USER=tu_usuario
ORACLE_PASSWORD=tu_contraseña
ORACLE_HOST=localhost
ORACLE_PORT=1521
ORACLE_SERVICE=tu_servicio

# Seguridad
API_TOKEN=token_seguro_123
API_ALLOWED_IPS=127.0.0.1,::1,localhost

# Puerto de escucha
PORT=3000
```

**Más detalles:** Ver [Configuración Completa](CONFIGURACION.md)

---

## 🧪 Pruebas Básicas

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
- [ ] Archivo `.env` creado y configurado
- [ ] Ping endpoint responde OK
- [ ] Puedo ejecutar una consulta SELECT

---

**¿Necesitas ayuda?** Ver [Documentación Completa](../INDEX.md)
