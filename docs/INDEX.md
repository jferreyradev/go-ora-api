# 📚 Documentación - Go Oracle API

**Tabla de Contenidos de la documentación completa**

---

## 🚀 Comenzar

**Para nuevos usuarios - Empieza aquí:**

1. [📖 Inicio Rápido](getting-started/QUICKSTART.md) - Instalación y primer uso en 5 minutos
2. [⚙️ Configuración](getting-started/CONFIGURACION.md) - Variables de entorno y setup

---

## 📡 Referencia de API

- [🔗 Endpoints Disponibles](api-reference/ENDPOINTS.md) - Listado completo de operaciones
- [� Nomenclatura Oracle](api-reference/ORACLE_NAMING.md) - Esquemas, packages y resolución de nombres
- [⏱️ Jobs Asíncronos](api-reference/ASYNC_JOBS.md) - Sistema de jobs en background con políticas `parallel`, `sequential` y `exclusive`

---

## 🚀 Despliegue en Producción

- [🌐 Deployment](deployment/DEPLOYMENT.md) - Guía de despliegue
- [🔒 Firewall Windows](deployment/FIREWALL_WINDOWS.md) - Configurar firewall

---

## 📦 Release & Versiones

- [📋 Changelog](release/CHANGELOG.md) - Historial de cambios y versiones

---

## 🎯 Búsqueda Rápida

| Necesito... | Ver... |
|----------|--------|
| Instalar y ejecutar | [Inicio Rápido](getting-started/QUICKSTART.md) |
| Configurar .env | [Configuración](getting-started/CONFIGURACION.md) |
| Llamar un endpoint | [Endpoints](api-reference/ENDPOINTS.md) |
| Resolver conflictos de esquemas/packages | [Nomenclatura Oracle](api-reference/ORACLE_NAMING.md) |
| Usar jobs asíncronos | [Async Jobs](api-reference/ASYNC_JOBS.md) |
| Desplegar en producción | [Deployment](deployment/DEPLOYMENT.md) |
| Abrir puerto en Windows | [Firewall Windows](deployment/FIREWALL_WINDOWS.md) |
| Ver qué cambió | [Changelog](release/CHANGELOG.md) |
| Publicar release | [Publish Guide](release/PUBLISH_GUIDE.md) |

---

## 📞 Ayuda Rápida

### Estructura de Carpetas
```
docs/
├── INDEX.md                          ← Estás aquí
│
├── getting-started/                  # Para nuevos usuarios
│   ├── QUICKSTART.md
│   ├── CONFIGURACION.md
│   └── GUIA_RAPIDA.md
│
├── api-reference/                    # Referencia técnica
│   ├── ENDPOINTS.md
│   ├── SCHEMA_FIELD.md
│   └── ASYNC_JOBS.md                 # Incluye control de concurrencia
│
├── deployment/                       # Producción
│   ├── DEPLOYMENT.md
│   └── FIREWALL_WINDOWS.md
│
├── release/                          # Versiones
│   ├── CHANGELOG.md
│   ├── RELEASE_NOTES.md
│   ├── RELEASE_SUMMARY.md
│   └── PUBLISH_GUIDE.md
│
└── auxiliar/                         # Referencia
    ├── ESTRUCTURA.md
    └── USO_Y_PRUEBAS.md
```

### Archivos en Raíz
| Archivo | Propósito |
|---------|----------|
| [README.md](../README.md) | Inicio del proyecto (overview) |
| [.env.example](../.env.example) | Template de variables |
| [go.mod](../go.mod) | Dependencias Go |
| [main.go](../main.go) | Código fuente |
| [LICENSE](../LICENSE) | MIT License |

---

## 🔄 Navegación

- **Primero**: Lee [README.md](../README.md) para contexto general
- **Luego**: Sigue [Inicio Rápido](getting-started/QUICKSTART.md)
- **Desarrollo**: Consulta [API Reference](api-reference/)
- **Jobs y concurrencia**: Revisa [Async Jobs](api-reference/ASYNC_JOBS.md)
- **Producción**: Lee [Deployment](deployment/)
- **Versiones**: Mira [Release](release/)

---

**Última actualización:** 29/05/2026  
**Versión Actual:** 1.0.0  
**Licencia:** MIT
