# 🚀 Guía de Publicación de Release

Cómo publicar v1.0.0 en GitHub

---

## 📋 Pre-Checklist

- [x] Código compilado: `go-oracle-api.exe`
- [x] CHANGELOG.md actualizado
- [x] RELEASE_NOTES.md completo
- [x] README.md con características v1.0.0
- [x] Documentación en docs/
- [x] LICENSE incluido (MIT)
- [x] .gitignore configurado
- [x] SQL scripts listos

---

## 📤 Publicar en GitHub

### Opción 1: CLI de Git (Recomendado)

```bash
# 1. Crear tag
git tag -a v1.0.0 -m "Release v1.0.0 - Multiple OUT parameters bug fix"

# 2. Hacer push del tag
git push origin v1.0.0

# 3. Crear release en GitHub UI:
#    - Ir a: https://github.com/tu-usuario/go-oracle-api/releases
#    - Click en "Draft a new release"
#    - Seleccionar tag v1.0.0
#    - Título: "Release v1.0.0 - Stable"
#    - Descripción: Copiar contenido de RELEASE_NOTES.md
#    - Upload: go-oracle-api.exe
#    - Click "Publish Release"
```

### Opción 2: GitHub CLI

```bash
# Instalar: https://cli.github.com/

gh release create v1.0.0 \
  --title "Release v1.0.0 - Stable" \
  --notes "$(cat RELEASE_NOTES.md)" \
  ./go-oracle-api.exe
```

### Opción 3: GitHub Web UI

1. Ir a: https://github.com/tu-usuario/go-oracle-api/releases
2. Click "Draft a new release"
3. Rellenar:
   - **Choose a tag**: v1.0.0
   - **Release title**: "Release v1.0.0 - Stable"
   - **Description**: Copiar contenido de RELEASE_NOTES.md
4. **Upload binaries**: go-oracle-api.exe
5. Click "Publish Release"

---

## 📝 Template para Descripción

```markdown
# Go Oracle API v1.0.0 - Stable Release

## 🎉 Características Principales

✨ API RESTful completa para Oracle
✨ Soporte para Procedimientos y Funciones
✨ Sistema de Jobs Asíncronos
✨ Manejo Automático de Tipos de Datos

## 🐛 Bug Fixes Críticos

**Fixed: Múltiples parámetros OUT**
- Devuelve correctamente TODOS los OUT parameters (NUMBER, VARCHAR2, DATE)
- Antes: solo devolvía el último parámetro
- Afectaba a procedimientos con múltiples salidas

## 📥 Descarga

- **go-oracle-api.exe** - Ejecutable Windows (20.36 MB)

## 🚀 Inicio Rápido

1. Descargar go-oracle-api.exe
2. Crear .env con credenciales Oracle
3. Ejecutar: `./go-oracle-api.exe`
4. Probar: `curl http://localhost:3000/ping`

## 📚 Documentación

Ver [docs/INDEX.md](../../docs/INDEX.md) para acceso a:
- Guía de Inicio Rápido
- API Reference
- Deployment
- Troubleshooting

## 📋 Requisitos

- Windows 7+ / Linux / macOS (x86_64)
- Oracle 11g o superior
- Go 1.20+ (para compilar desde fuente)

## 📝 Licencia

MIT License - Libre para usar, modificar y distribuir

---

**Status:** ✅ Stable
**Release Date:** 2026-04-24
```

---

## ✅ Verificación Post-Release

Después de publicar:

- [ ] Release visible en GitHub
- [ ] Binario descargable
- [ ] Links en RELEASE_NOTES funcionan
- [ ] Tag v1.0.0 presente

---

## 🔄 Próximas Versiones

**v1.1.0**: WebSocket, Caché, Prometheus  
**v1.2.0**: GraphQL, Rate limiting, Audit trail

---

**¿Listo?** Ejecuta los comandos de Opción 1 o Opción 2 arriba 🚀
