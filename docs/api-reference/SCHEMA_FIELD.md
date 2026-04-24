# 🏗️ Schema y Nomenclatura Oracle

Cómo especificar correctamente nombres de procedimientos y funciones

---

## 📋 Tabla de Contenidos

- [Campo schema](#campo-schema)
- [Conflictos Comunes](#conflictos-comunes)
- [Soluciones](#soluciones)
- [Ejemplos](#ejemplos)

---

## 🆔 Campo `schema`

Campo **opcional** para especificar el esquema del objeto.

### Cuándo Usar

**✅ Usar cuando:**
- Función/procedimiento **standalone** en otro esquema
- Mayor claridad y separación de responsabilidades
- Evitar ambigüedad con packages

**❌ No usar cuando:**
- Objeto está en un **package**
- Usando notación completa en `name`

### Ejemplo

```json
{
  "schema": "WORKFLOW",
  "name": "MI_FUNCION",
  "isFunction": true,
  "params": [
    {"name": "resultado", "direction": "OUT", "type": "number"}
  ]
}
```

Genera SQL: `BEGIN :1 := "WORKFLOW"."MI_FUNCION"(...); END;`

---

## Notación Oracle

### Nomenclatura Estándar

| Tipo | Notación | Ejemplo |
|------|----------|---------|
| Standalone en esquema actual | `NOMBRE` | `MI_PROC` |
| En package | `PACKAGE.NOMBRE` | `PKG_UTILS.SUMA` |
| Standalone en otro esquema | `SCHEMA.NOMBRE` | `WORKFLOW.MI_PROC` |
| En package de otro esquema | `SCHEMA.PACKAGE.NOMBRE` | `WORKFLOW.PKG_UTILS.SUMA` |

### Cómo Especificar en la API

#### Opción 1: Campo `schema` + `name` (Recomendado)

```json
{
  "schema": "WORKFLOW",
  "name": "EXISTE_PROC_CAB",
  "isFunction": true,
  "params": [...]
}
```

#### Opción 2: Notación Completa en `name`

```json
{
  "name": "WORKFLOW.EXISTE_PROC_CAB",
  "isFunction": true,
  "params": [...]
}
```

#### Opción 3: Package + Nombre

```json
{
  "name": "PKG_UTILS.SUMA",
  "params": [
    {"name": "a", "value": 5, "direction": "IN"},
    {"name": "b", "value": 3, "direction": "IN"},
    {"name": "resultado", "direction": "OUT", "type": "number"}
  ]
}
```

---

## ⚠️ Conflictos Comunes

### Problema: Package vs Schema

**Escenario:**
- Existe `WORKFLOW` como **PACKAGE** (en user USUARIO)
- Existe `WORKFLOW` como **SCHEMA** (separado)
- Función `EXISTE_PROC_CAB` en schema WORKFLOW (standalone)

**Resultado:**
```
WORKFLOW.EXISTE_PROC_CAB → Oracle busca en PACKAGE ❌
```

Error: `ORA-06550: PLS-00302: component must be declared`

### Problema: Ambigüedad de Nomenclatura

Cuando Oracle no sabe si es package o schema, prioriza:
1. Package en esquema actual
2. Package en otros esquemas
3. Schema/objeto standalone

---

## ✅ Soluciones

### Solución 1: Usar Campo `schema` (RECOMENDADO)

```json
{
  "schema": "WORKFLOW",
  "name": "EXISTE_PROC_CAB",
  "isFunction": true,
  "params": [...]
}
```

**Ventajas:**
- Más claro
- Evita ambigüedad
- Genera SQL correcto

### Solución 2: Crear Sinónimo

En SQL*Plus:
```sql
CREATE SYNONYM EXISTE_PROC_CAB FOR WORKFLOW.EXISTE_PROC_CAB;
```

Luego en API:
```json
{
  "name": "EXISTE_PROC_CAB",
  "isFunction": true,
  "params": [...]
}
```

**Ventajas:**
- Simplifica llamadas
- Funciona para ambos casos

### Solución 3: Renombrar Package

Si es posible, renombrar el package conflictivo:
```sql
ALTER PACKAGE WORKFLOW RENAME TO WORKFLOW_PKG;
```

Luego usar:
```json
{
  "name": "WORKFLOW_PACKAGE.EXISTE_PROC_CAB",
  "isFunction": true,
  "params": [...]
}
```

### ❌ Soluciones que NO Funcionan

```json
// ❌ Comillas no resuelven conflicto
{"name": "\"WORKFLOW\".EXISTE_PROC_CAB"}

// ❌ Notación expandida no ayuda
{"name": "WORKFLOW.WORKFLOW.EXISTE_PROC_CAB"}

// ❌ Grant de execute tampoco
// GRANT EXECUTE ON WORKFLOW.EXISTE_PROC_CAB TO USUARIO;
```

---

## 📚 Ejemplos

### Ejemplo 1: Función Standalone

```json
{
  "schema": "WORKFLOW",
  "name": "BUSCA_PERSONA",
  "isFunction": true,
  "params": [
    {"name": "vDNI", "value": 123456789, "direction": "IN"},
    {"name": "resultado", "direction": "OUT", "type": "number"}
  ]
}
```

### Ejemplo 2: Función en Package

```json
{
  "name": "TRANSFORMADOR.BUSCA_PERSONA",
  "isFunction": true,
  "params": [
    {"name": "vDNI", "value": 123456789, "direction": "IN"},
    {"name": "resultado", "direction": "OUT", "type": "number"}
  ]
}
```

### Ejemplo 3: Procedimiento en Package de Otro Schema

```json
{
  "name": "WORKFLOW.CONTROLES.PROCESAR",
  "params": [
    {"name": "vPERIODO", "value": "2026-04", "direction": "IN"},
    {"name": "vRESULT", "direction": "OUT", "type": "number"}
  ]
}
```

### Ejemplo 4: Evitando Conflicto

```json
// ✅ Correcto: Usa campo schema
{
  "schema": "WORKFLOW",
  "name": "EXISTE_PROC_CAB",
  "isFunction": true,
  "params": [
    {"name": "vCOUNT", "direction": "OUT", "type": "number"},
    {"name": "vIDGRUPOREP", "value": 1, "direction": "IN"}
  ]
}

// ❌ Evitar: Ambiguo
{
  "name": "WORKFLOW.EXISTE_PROC_CAB",
  "isFunction": true,
  "params": [...]
}
```

---

## 🔍 Detección Automática

El backend detecta automáticamente:

```go
func formatObjectName(schema, name string) string {
    if schema != "" {
        // Si hay schema explícito, usarlo
        return fmt.Sprintf("%s.%s", 
            strings.ToUpper(schema), 
            strings.ToUpper(name))
    } else if strings.Contains(name, ".") && !strings.Contains(name, "\"") {
        // Si hay punto sin comillas, agregar comillas
        parts := strings.Split(name, ".")
        for i, part := range parts {
            parts[i] = fmt.Sprintf("\"%s\"", 
                strings.ToUpper(part))
        }
        return strings.Join(parts, ".")
    }
    // Sin punto, nombre simple
    return strings.ToUpper(name)
}
```

---

## 💡 Mejores Prácticas

1. **Prefiere campo `schema`** para objetos standalone
2. **Usa notación completa `PACKAGE.NOMBRE`** para packages
3. **Crea sinónimos** si hay conflictos frecuentes
4. **Documenta la nomenclatura** en tu BD

---

Para más detalles, ver [docs/INDEX.md](../INDEX.md)
