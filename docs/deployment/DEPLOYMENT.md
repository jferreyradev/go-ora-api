# 🌐 Despliegue del Microservicio

Guía completa para desplegar en producción

---

## 📋 Tabla de Contenidos

- [Compilación](#compilación)
- [Transferencia](#transferencia)
- [Servicios Windows](#servicios-windows)
- [Servicios Linux](#servicios-linux)
- [Verificación](#verificación)
- [Troubleshooting](#troubleshooting)

---

## 🔨 Compilación

### Para Windows:
```bash
GOOS=windows GOARCH=amd64 go build -o oracle-api.exe
```

### Para Linux:
```bash
GOOS=linux GOARCH=amd64 go build -o oracle-api
```

### Para macOS:
```bash
GOOS=darwin GOARCH=amd64 go build -o oracle-api
```

---

## 📤 Transferencia de Archivos

Copiar al servidor:
- Binario compilado (`oracle-api.exe` o `oracle-api`)
- Archivo `.env` con configuración
- Script de inicialización (si aplica)

---

## 🪟 Servicios Windows

### Usando NSSM (Recomendado)

#### 1. Descargar NSSM
```
https://nssm.cc/download
Descargar versión apropiada (32-bit o 64-bit)
```

#### 2. Instalar Servicio
Abre PowerShell como **administrador**:

```powershell
# Navega a carpeta NSSM
cd C:\path\to\nssm\win64

# Instalar servicio
.\nssm.exe install GoOracleAPI "C:\ruta\a\oracle-api.exe"
```

#### 3. Configurar Entorno (Opcional)
```powershell
# Si prefieres, configura variables de entorno
# En la ventana de configuración, ve a la pestaña "Environment"
.\nssm.exe edit GoOracleAPI
```

#### 4. Iniciar Servicio
```powershell
.\nssm.exe start GoOracleAPI
```

### Verificar Estado
```powershell
# Ver estado
.\nssm.exe status GoOracleAPI

# Ver logs
.\nssm.exe info GoOracleAPI

# Desde Administrador de Servicios:
services.msc
```

### Detener/Reiniciar
```powershell
.\nssm.exe stop GoOracleAPI
.\nssm.exe restart GoOracleAPI
.\nssm.exe remove GoOracleAPI
```

---

## 🐧 Servicios Linux

### Usando systemd (Recomendado)

#### 1. Preparar Directorios
```bash
sudo mkdir -p /opt/oracle-api
sudo cp oracle-api /opt/oracle-api/
sudo cp .env /opt/oracle-api/
sudo chmod +x /opt/oracle-api/oracle-api
```

#### 2. Crear Usuario Servicio
```bash
sudo useradd -r -s /bin/false oracleapi
sudo chown -R oracleapi:oracleapi /opt/oracle-api
```

#### 3. Crear Archivo de Servicio
```bash
sudo nano /etc/systemd/system/oracle-api.service
```

Contenido:
```ini
[Unit]
Description=Go Oracle API Microservice
After=network.target

[Service]
Type=simple
User=oracleapi
WorkingDirectory=/opt/oracle-api
ExecStart=/opt/oracle-api/oracle-api
EnvironmentFile=/opt/oracle-api/.env
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

#### 4. Habilitar y Iniciar
```bash
sudo systemctl daemon-reload
sudo systemctl enable oracle-api
sudo systemctl start oracle-api
```

### Verificar Estado
```bash
# Estado del servicio
sudo systemctl status oracle-api

# Ver logs en tiempo real
sudo journalctl -u oracle-api -f

# Ver últimas 50 líneas
sudo journalctl -u oracle-api -n 50
```

### Controlar Servicio
```bash
# Restart
sudo systemctl restart oracle-api

# Stop
sudo systemctl stop oracle-api

# Ver si está habilitado
sudo systemctl is-enabled oracle-api
```

---

## ✅ Verificación

### Test de Conectividad
```bash
curl http://localhost:3000/ping \
  -H "Authorization: Bearer tu_token"
```

**Respuesta esperada:**
```json
{"status":"ok"}
```

### Verificar Logs
**Windows (NSSM):**
```powershell
Get-EventLog -LogName Application -Source GoOracleAPI -Newest 50
```

**Linux (systemd):**
```bash
sudo journalctl -u oracle-api -n 50 -f
```

### Verificar Conexión Oracle
```bash
# Desde dentro del contenedor/servidor
sqlplus usuario/password@host:puerto/servicio
```

---

## 🆘 Troubleshooting

### Error: "connection refused"
```
✓ Verificar Oracle está corriendo
✓ Verificar ORACLE_HOST y ORACLE_PORT en .env
✓ Ping al servidor: ping 192.168.1.100
✓ Verificar firewall permite puerto 1521
```

### Error: "ORA-12514: listener does not know of service name"
```
✓ Verificar ORACLE_SERVICE es correcto
✓ En servidor Oracle: lsnrctl status
✓ Listar listeners configurados
```

### Error: "Invalid username/password"
```
✓ Verificar ORACLE_USER y ORACLE_PASSWORD
✓ Probar credenciales en SQL*Plus
✓ Verificar usuario está activo en BD
```

### Servicio no inicia (Windows)
```
✓ Verificar permisos del binario
✓ Verificar ruta es correcta
✓ Ver logs en Event Viewer
✓ Ejecutar como administrador
```

### Servicio no inicia (Linux)
```
✓ Verificar permisos: chmod +x /opt/oracle-api/oracle-api
✓ Verificar .env existe y es legible
✓ Ver logs: sudo journalctl -u oracle-api
✓ Verificar usuario oracleapi tiene permisos
```

### Puerto ya en uso
```bash
# Windows - Ver qué usa el puerto 3000
netstat -ano | findstr :3000

# Linux - Ver qué usa el puerto 3000
sudo lsof -i :3000
# O
sudo netstat -tlnp | grep :3000
```

### API no responde
```
✓ Verificar servicio está corriendo
✓ Verificar puerto en .env
✓ Verificar firewall permite acceso
✓ Ver logs para mensajes de error
✓ Verificar conectividad a Oracle
```

---

## 📊 Monitoreo en Producción

### Endpoint Health Check
```bash
# Script de monitoreo (cada 5 min)
curl -f http://localhost:3000/ping \
  -H "Authorization: Bearer token" \
  || systemctl restart oracle-api
```

### Ver Logs Constantemente
```bash
# Terminal 1: Logs en tiempo real
sudo journalctl -u oracle-api -f

# Terminal 2: Ejecutar tests/queries
# ...
```

### Límites de Recursos
```bash
# Ver uso de memoria/CPU del servicio
ps aux | grep oracle-api

# Monitor continuo (Linux)
top -p $(pidof oracle-api)
```

---

## 🔒 Seguridad en Producción

- [x] Cambiar token en `API_TOKEN` (.env)
- [x] Restricción de IPs en `API_ALLOWED_IPS`
- [x] Firewall configurado correctamente
- [x] Archivo .env no versionado en Git
- [x] Credenciales Oracle con permisos mínimos
- [x] Logs rotados regularmente

---

Para más detalles, ver [INDEX.md](../INDEX.md)
