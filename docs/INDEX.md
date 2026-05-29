# 📚 Documentación - Go Oracle API

**Tabla de Contenidos de la documentación completa**

---

## 🚀 Comenzar

**Para nuevos usuarios - Empieza aquí:**

1. [📖 Inicio Rápido](getting-started/QUICKSTART.md) - Instalación y primer uso en 5 minutos
2. [⚙️ Configuración](getting-started/CONFIGURACION.md) - Variables de entorno y setup
3. [🏃 Guía Rápida](getting-started/GUIA_RAPIDA.md) - Ejemplos básicos de uso

---

## 📖 Guías Completas

### 🔧 Uso y Desarrollo
- [📝 Uso y Pruebas](../USO_Y_PRUEBAS.md) - Guía detallada de endpoints y ejemplos
- [🏗️ Estructura del Proyecto](../ESTRUCTURA.md) - Organización del código
- [📊 Trabajar con Parámetros](../SCHEMA_FIELD.md) - Tipos de datos y parámetros

### 🚀 Despliegue en Producción
- [🌐 Deployment](deployment/DEPLOYMENT.md) - Guía de despliegue
- [🔒 Firewall Windows](deployment/FIREWALL_WINDOWS.md) - Configurar firewall
- [⚡ Trabajos Asíncronos](api-reference/ASYNC_JOBS.md) - Jobs en background, cola secuencial y ejecución exclusiva

### 📡 Referencia de API
- [🔗 Endpoints Disponibles](api-reference/ENDPOINTS.md) - Listado completo de operaciones
- [📋 Schema y Campos](api-reference/SCHEMA_FIELD.md) - Tipos de datos soportados
- [⏱️ Jobs Asíncronos](api-reference/ASYNC_JOBS.md) - Procesos en background, políticas `parallel`, `sequential` y `exclusive`

---

## 📦 Release & Versiones

**Información sobre versiones y cómo publicar:**

- [📋 Changelog](release/CHANGELOG.md) - Historial de cambios
- [📝 Release Notes v1.0.0](release/RELEASE_NOTES.md) - Características v1.0.0
- [🎯 Release Summary](release/RELEASE_SUMMARY.md) - Resumen ejecutivo
- [🚀 Publish Guide](release/PUBLISH_GUIDE.md) - Cómo publicar en GitHub

---

## 💡 Ejemplos & Casos de Uso

**Código funcional de ejemplo:**

- [💻 Ejemplo Completo](../examples/ejemplo_completo.js) - Node.js con todas las operaciones
- [📚 Base de Datos Setup](../sql/create_test_procedures.sql) - Scripts SQL para testing

---

## 📋 Índice por Categoría

### Configuration (⚙️)
| Archivo | Descripción |
|---------|------------|
| [CONFIGURACION.md](getting-started/CONFIGURACION.md) | Variables de entorno y secretos |
| [.env.example](../.env.example) | Template de configuración |

### API Reference (📡)
| Archivo | Descripción |
|---------|------------|
| [ENDPOINTS.md](api-reference/ENDPOINTS.md) | Todos los endpoints disponibles |
| [SCHEMA_FIELD.md](api-reference/SCHEMA_FIELD.md) | Tipos de datos y parámetros |
| [ASYNC_JOBS.md](api-reference/ASYNC_JOBS.md) | Jobs en background, estado y control de concurrencia |

### Operations (🚀)
| Archivo | Descripción |
|---------|------------|
| [DEPLOYMENT.md](deployment/DEPLOYMENT.md) | Despliegue a producción |
| [FIREWALL_WINDOWS.md](deployment/FIREWALL_WINDOWS.md) | Configuración de firewall |
| [USO_Y_PRUEBAS.md](../USO_Y_PRUEBAS.md) | Testing y pruebas |

### Release (📦)
| Archivo | Descripción |
|---------|------------|
| [CHANGELOG.md](release/CHANGELOG.md) | Historial de todas las versiones |
| [RELEASE_NOTES.md](release/RELEASE_NOTES.md) | Detalles de v1.0.0 |
| [PUBLISH_GUIDE.md](release/PUBLISH_GUIDE.md) | Publicar release en GitHub |

---

## 🎯 Búsqueda Rápida

**¿Qué necesito?** Encuentra la sección:

| Necesito... | Ver... |
|----------|--------|
| Instalar y ejecutar | [Inicio Rápido](getting-started/QUICKSTART.md) |
| Configurar .env | [Configuración](getting-started/CONFIGURACION.md) |
| Ver ejemplos | [Guía Rápida](getting-started/GUIA_RAPIDA.md) |
| Llamar un endpoint | [Endpoints](api-reference/ENDPOINTS.md) |
| Trabajar con parámetros OUT | [Schema & Fields](api-reference/SCHEMA_FIELD.md) |
| Usar jobs asíncronos | [Async Jobs](api-reference/ASYNC_JOBS.md) |
| Configurar ejecución secuencial o exclusiva | [Async Jobs](api-reference/ASYNC_JOBS.md) |
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
