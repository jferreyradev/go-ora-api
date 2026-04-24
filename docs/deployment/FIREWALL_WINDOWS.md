# 🔒 Firewall Windows - Configuración

Cómo permitir acceso a Go Oracle API a través de firewall de Windows

---

## ⚡ Acceso desde Cualquier IP

Abre **PowerShell como administrador** y ejecuta:

```powershell
New-NetFirewallRule `
  -DisplayName "Go Oracle API" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 3000
```

> Cambia `3000` si usas otro puerto

---

## 🔐 Acceso solo desde IP Específica

Por ejemplo, solo desde `192.168.1.50`:

```powershell
New-NetFirewallRule `
  -DisplayName "Go Oracle API - IP 192.168.1.50" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 3000 `
  -RemoteAddress 192.168.1.50
```

### Agregar Múltiples IPs

Crear regla separada para cada IP:

```powershell
# IP 1
New-NetFirewallRule `
  -DisplayName "Go Oracle API - IP 192.168.1.50" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 3000 `
  -RemoteAddress 192.168.1.50

# IP 2
New-NetFirewallRule `
  -DisplayName "Go Oracle API - IP 192.168.1.51" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 3000 `
  -RemoteAddress 192.168.1.51
```

---

## 🌐 Acceso desde Rango (CIDR)

Permite un rango de IPs, ej: `192.168.1.0/24`:

```powershell
New-NetFirewallRule `
  -DisplayName "Go Oracle API - Red 192.168.1.0/24" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 3000 `
  -RemoteAddress 192.168.1.0/24
```

---

## 🧹 Eliminar Regla

```powershell
Remove-NetFirewallRule -DisplayName "Go Oracle API"
```

> Usa el nombre exacto de la regla que creaste

---

## 📋 Ver Todas las Reglas

```powershell
Get-NetFirewallRule | Select-Object DisplayName, Enabled, Direction, Action, LocalPort, RemoteAddress | Format-Table -AutoSize
```

### Ver solo nuestras reglas
```powershell
Get-NetFirewallRule -DisplayName "*Oracle API*" | Format-Table -AutoSize
```

---

## ✅ Verificar que Funciona

Desde otra máquina en la red:

```bash
# Test de conectividad
curl http://IP_DEL_SERVIDOR:3000/ping \
  -H "Authorization: Bearer tu_token"

# Debería responder:
# {"status":"ok"}
```

---

## ❌ Solucionar Problemas

### "Access is denied"
- Ejecutar PowerShell como **administrador**
- Usar `Run as administrator`

### Regla no funciona
- Verificar nombre de regla es correcto
- Usar `Get-NetFirewallRule` para listar
- Verificar port es el mismo (3000 en ejemplos)

### Windows Firewall está desactivado
- Control Panel → Windows Defender Firewall
- O usar: `Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled True`

### Necesito IPv6
```powershell
# IPv6 - desde cualquier IP
New-NetFirewallRule `
  -DisplayName "Go Oracle API IPv6" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 3000 `
  -RemoteAddress "::" 

# IPv6 - rango específico
New-NetFirewallRule `
  -DisplayName "Go Oracle API IPv6 Rango" `
  -Direction Inbound `
  -Action Allow `
  -Protocol TCP `
  -LocalPort 3000 `
  -RemoteAddress "2001:db8::/32"
```

---

## 📝 Checklist

- [ ] PowerShell abierto como **Administrador**
- [ ] Comando ejecutado sin errores
- [ ] Regla visible en `Get-NetFirewallRule`
- [ ] Ping desde cliente funciona
- [ ] API responde en `/ping`

---

Para más ayuda, ver [docs/INDEX.md](../INDEX.md)
