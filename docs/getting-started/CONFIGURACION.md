# ⚙️ Configuración - Go Oracle API

Guía completa para configurar el archivo `.env`

---

## Template Básico

Copiar a `.env` en la raíz del proyecto:

```env
# --- Conexión a Oracle ---
ORACLE_USER=usuario
ORACLE_PASSWORD=contraseña
ORACLE_HOST=localhost
ORACLE_PORT=1521
ORACLE_SERVICE=servicio_o_sid

# --- Seguridad API ---
API_TOKEN=tu_token_seguro

# --- IPs permitidas ---
# Puedes poner IPs exactas, rangos CIDR, o 'localhost'.
# Ejemplo: 192.168.1.10,192.168.1.0/24,127.0.0.1,::1,localhost
API_ALLOWED_IPS=127.0.0.1,::1,192.168.1.0/24,localhost

# --- Puerto de escucha ---
PORT=8080

# --- Desactivar autenticación y restricción de IPs (solo para pruebas) ---
# Si es 1, desactiva autenticación y restricción de IPs (NO usar en producción)
API_NO_AUTH=0
```

---

## Variables de Entorno

### Conexión a Oracle

#### ORACLE_USER
- **Tipo:** String
- **Requerido:** ✅ Sí
- **Descripción:** Usuario de la base de datos Oracle
- **Ejemplo:** `USUARIO` o `SYS`

#### ORACLE_PASSWORD
- **Tipo:** String  
- **Requerido:** ✅ Sí
- **Descripción:** Contraseña del usuario Oracle
- **Ejemplo:** `miContraseña123`

#### ORACLE_HOST
- **Tipo:** String (IP o hostname)
- **Requerido:** ✅ Sí
- **Descripción:** Host/IP del servidor Oracle
- **Ejemplo:** `192.168.1.100` o `oracle.example.com`

#### ORACLE_PORT
- **Tipo:** Número
- **Requerido:** ❌ No (por defecto: 1521)
- **Descripción:** Puerto de escucha de Oracle
- **Ejemplo:** `1521`

#### ORACLE_SERVICE
- **Tipo:** String
- **Requerido:** ✅ Sí
- **Descripción:** Nombre del servicio o SID de Oracle
- **Ejemplo:** `HTEST01` o `XE`

---

### Seguridad de API

#### API_TOKEN
- **Tipo:** String
- **Requerido:** ✅ Sí (a menos que API_NO_AUTH=1)
- **Descripción:** Token Bearer para autenticar requests
- **Ejemplo:** `abc123def456ghi789`
- **Recomendación:** Usar token fuerte (mínimo 16 caracteres)

#### API_ALLOWED_IPS
- **Tipo:** String (lista CSV)
- **Requerido:** ❌ No
- **Descripción:** IPs permitidas para acceder a la API
- **Ejemplo:** `192.168.1.10,192.168.1.0/24,127.0.0.1,::1`
- **Formatos soportados:**
  - IP exacta: `192.168.1.10`
  - Rango CIDR: `192.168.1.0/24` (de 192.168.1.0 a 192.168.1.255)
  - IPv6: `::1` o `2001:db8::/32`
  - Hostname: `localhost`, `example.com`
  - Vacío: Permitir todas las IPs

#### API_NO_AUTH
- **Tipo:** Número (0 o 1)
- **Requerido:** ❌ No (por defecto: 0)
- **Descripción:** Desactiva autenticación y restricción de IPs
- **Valores:**
  - `0` = Autenticación activa (recomendado)
  - `1` = Sin autenticación (SOLO TESTING)
- **⚠️ ADVERTENCIA:** Nunca usar en producción

---

### Configuración de Puerto

#### PORT
- **Tipo:** Número
- **Requerido:** ❌ No (por defecto: 8080)
- **Descripción:** Puerto donde escucha la API
- **Ejemplo:** `3000`, `8080`, `9000`
- **Rango:** 1-65535
- **Nota:** Puertos < 1024 requieren privilegios especiales

---

### Configuración de Conexiones (Avanzado)

#### MAX_IDLE_CONNECTIONS
- **Tipo:** Número
- **Requerido:** ❌ No (por defecto: 10)
- **Descripción:** Máximo de conexiones ociosas en el pool
- **Ejemplo:** `10`, `20`, `50`

#### MAX_OPEN_CONNECTIONS
- **Tipo:** Número
- **Requerido:** ❌ No (por defecto: 100)
- **Descripción:** Máximo de conexiones abiertas simultáneamente
- **Ejemplo:** `50`, `100`, `200`

---

## Ejemplos de Configuración

### Desarrollo Local

```env
ORACLE_USER=usuario_dev
ORACLE_PASSWORD=password_dev
ORACLE_HOST=localhost
ORACLE_PORT=1521
ORACLE_SERVICE=XE
API_TOKEN=dev_token_123
API_ALLOWED_IPS=127.0.0.1,::1,localhost
PORT=3000
API_NO_AUTH=1
```

### Staging

```env
ORACLE_USER=usuario_staging
ORACLE_PASSWORD=password_staging
ORACLE_HOST=10.0.1.50
ORACLE_PORT=1521
ORACLE_SERVICE=STAGING_DB
API_TOKEN=staging_token_abc123xyz
API_ALLOWED_IPS=10.0.1.0/24,10.0.2.0/24
PORT=8080
MAX_OPEN_CONNECTIONS=50
```

### Producción

```env
ORACLE_USER=usuario_prod
ORACLE_PASSWORD=password_prod_seguro
ORACLE_HOST=192.168.100.10
ORACLE_PORT=1521
ORACLE_SERVICE=PROD_SERVICE
API_TOKEN=prod_token_muy_seguro_abc123xyz789
API_ALLOWED_IPS=192.168.100.0/24,10.0.0.0/8
PORT=3000
MAX_OPEN_CONNECTIONS=200
API_NO_AUTH=0
```

---

## Cómo Aplicar la Configuración

### Paso 1: Crear .env
```bash
cp .env.example .env
```

### Paso 2: Editar con tus valores
```bash
# Windows
notepad .env

# Linux/macOS
nano .env
vim .env
```

### Paso 3: Guardar y ejecutar
```bash
# Linux/macOS
export $(cat .env | xargs)
go run main.go

# Windows PowerShell
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^=\s]+)\s*=\s*(.*)$') {
    [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
  }
}
.\go-oracle-api.exe
```

---

## Validación de Configuración

### Verificar conectividad
```bash
curl http://localhost:3000/ping \
  -H "Authorization: Bearer tu_token"
```

**Respuesta exitosa:**
```json
{"status":"ok"}
```

---

## Seguridad

### Recomendaciones

1. **Token Fuerte**
   - Mínimo 16 caracteres
   - Usar caracteres especiales: `!@#$%^&*`
   - Cambiar regularmente

2. **IPs Permitidas**
   - Restricción por red en producción
   - Usar rangos CIDR en lugar de IPs individuales
   - Documentar todas las IPs permitidas

3. **Credenciales Oracle**
   - Usar usuario con permisos mínimos
   - No compartir `.env` en repositorio
   - Usar `.env.example` como template

4. **Variables Sensibles**
   - Nunca loguear credenciales
   - Usar variables de entorno, no hardcoding
   - Rotar contraseñas regularmente

---

## Troubleshooting

### Error: "ORA-12514: TNS listener does not know of service name"
```
✓ Verificar ORACLE_SERVICE es correcto
✓ Ejecutar en servidor Oracle: lsnrctl status
✓ Verificar listeners configurados
```

### Error: "Connection refused"
```
✓ Verificar ORACLE_HOST y ORACLE_PORT
✓ Ping al servidor: ping 192.168.1.100
✓ Verificar firewall permite puerto 1521
```

### Error: "Unauthorized"
```
✓ Verificar API_TOKEN en .env y header Authorization
✓ Header debe ser: Authorization: Bearer tu_token
✓ Verificar API_NO_AUTH no está seteado a 1 en producción
```

### Error: "Invalid username/password"
```
✓ Verificar ORACLE_USER y ORACLE_PASSWORD
✓ Verificar credenciales en SQL*Plus
✓ Verificar usuario existe y está activo
```

---

Para más ayuda, ver [Índice de Documentación](../INDEX.md)
