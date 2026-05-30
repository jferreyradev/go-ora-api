-- ==============================================================================
-- Script de creación de paquete de prueba
-- Propósito: Crear paquete PKG_TEST con procedimientos almacenados para testing de la API
-- ==============================================================================

-- ==============================================================================
-- ESPECIFICACIÓN DEL PAQUETE
-- ==============================================================================
CREATE OR REPLACE PACKAGE PKG_TEST AS
    
    -- Procedimiento simple de prueba
    PROCEDURE TEST (
        p_input IN VARCHAR2,
        p_output OUT VARCHAR2
    );
    
    -- Procedimiento con múltiples parámetros
    PROCEDURE TEST_PARAMS (
        p_number IN NUMBER,
        p_varchar IN VARCHAR2,
        p_date IN DATE,
        p_result OUT VARCHAR2
    );
    
    -- Procedimiento que retorna cursor
    PROCEDURE TEST_CURSOR (
        p_limit IN NUMBER,
        p_cursor OUT SYS_REFCURSOR
    );
    
    -- Procedimiento con manejo de errores
    PROCEDURE TEST_ERROR (
        p_should_fail IN NUMBER
    );
    
    -- Procedimiento con operaciones DML
    PROCEDURE TEST_DML (
        p_table_name IN VARCHAR2,
        p_rows_created OUT NUMBER
    );
    
    -- Procedimiento con demora - Opción 1: Loop con timestamp
    PROCEDURE TEST_DEMORA_LOOP (
        segundos IN NUMBER
    );
    
    -- Procedimiento con demora - Opción 3: Loop con consulta
    PROCEDURE TEST_DEMORA_QUERY (
        segundos IN NUMBER
    );
    
END PKG_TEST;
/

-- ==============================================================================
-- CUERPO DEL PAQUETE
-- ==============================================================================
CREATE OR REPLACE PACKAGE BODY PKG_TEST AS

    -- Procedimiento simple de prueba
    PROCEDURE TEST (
        p_input IN VARCHAR2,
        p_output OUT VARCHAR2
    ) AS
    BEGIN
        p_output := 'Procesado: ' || p_input;
    END TEST;
    
    -- Procedimiento con múltiples parámetros
    PROCEDURE TEST_PARAMS (
        p_number IN NUMBER,
        p_varchar IN VARCHAR2,
        p_date IN DATE,
        p_result OUT VARCHAR2
    ) AS
    BEGIN
        p_result := 'Number: ' || p_number || 
                    ', String: ' || p_varchar || 
                    ', Date: ' || TO_CHAR(p_date, 'DD/MM/YYYY');
    END TEST_PARAMS;
    
    -- Procedimiento que retorna cursor
    PROCEDURE TEST_CURSOR (
        p_limit IN NUMBER,
        p_cursor OUT SYS_REFCURSOR
    ) AS
    BEGIN
        OPEN p_cursor FOR
            SELECT LEVEL as ID, 
                   'Item ' || LEVEL as NOMBRE,
                   SYSDATE as FECHA
            FROM DUAL
            CONNECT BY LEVEL <= p_limit;
    END TEST_CURSOR;
    
    -- Procedimiento con manejo de errores
    PROCEDURE TEST_ERROR (
        p_should_fail IN NUMBER
    ) AS
    BEGIN
        IF p_should_fail = 1 THEN
            RAISE_APPLICATION_ERROR(-20001, 'Error intencional de prueba');
        END IF;
        
        DBMS_OUTPUT.PUT_LINE('Ejecución exitosa');
    END TEST_ERROR;
    
    -- Procedimiento con operaciones DML
    PROCEDURE TEST_DML (
        p_table_name IN VARCHAR2,
        p_rows_created OUT NUMBER
    ) AS
        v_sql VARCHAR2(4000);
    BEGIN
        -- Este procedimiento requiere que tengas una tabla temporal
        -- Por simplicidad, solo simula la operación
        p_rows_created := 100;
        
        DBMS_OUTPUT.PUT_LINE('Operación DML simulada en tabla: ' || p_table_name);
    END TEST_DML;
    
    -- Procedimiento con demora - Opción 1: Loop con timestamp
    -- No requiere permisos especiales, usa SYSTIMESTAMP
    PROCEDURE TEST_DEMORA_LOOP (
        segundos IN NUMBER
    ) AS
        v_start_time TIMESTAMP;
        v_current_time TIMESTAMP;
        v_elapsed NUMBER;
    BEGIN
        v_start_time := SYSTIMESTAMP;
        
        LOOP
            v_current_time := SYSTIMESTAMP;
            v_elapsed := EXTRACT(SECOND FROM (v_current_time - v_start_time)) +
                         EXTRACT(MINUTE FROM (v_current_time - v_start_time)) * 60 +
                         EXTRACT(HOUR FROM (v_current_time - v_start_time)) * 3600;
            
            EXIT WHEN v_elapsed >= segundos;
            
            -- Pequeña pausa para no saturar CPU
            NULL;
        END LOOP;
        
        DBMS_OUTPUT.PUT_LINE('Procesamiento completado después de ' || segundos || ' segundos (LOOP)');
    END TEST_DEMORA_LOOP;
    
    -- Procedimiento con demora - Opción 3: Loop con consulta
    -- Consume menos CPU que la opción 1
    PROCEDURE TEST_DEMORA_QUERY (
        segundos IN NUMBER
    ) AS
        v_start_time DATE;
        v_dummy NUMBER;
    BEGIN
        v_start_time := SYSDATE;
        
        LOOP
            EXIT WHEN (SYSDATE - v_start_time) * 86400 >= segundos;
            
            -- Consulta ligera para dar tiempo al sistema
            SELECT 1 INTO v_dummy FROM DUAL WHERE ROWNUM = 1;
        END LOOP;
        
        DBMS_OUTPUT.PUT_LINE('Procesamiento completado después de ' || segundos || ' segundos (QUERY)');
    END TEST_DEMORA_QUERY;

END PKG_TEST;
/

-- ==============================================================================
-- VERIFICACIÓN
-- ==============================================================================

-- Verificar paquete creado
PROMPT ========================================
PROMPT Verificando paquete PKG_TEST
PROMPT ========================================

SELECT object_name, object_type, status
FROM user_objects
WHERE object_name = 'PKG_TEST'
ORDER BY object_type;

-- Listar procedimientos del paquete
PROMPT 
PROMPT Procedimientos en el paquete:
PROMPT ========================================

SELECT object_name, procedure_name
FROM user_procedures
WHERE object_name = 'PKG_TEST'
ORDER BY procedure_name;

-- Mostrar resultados
PROMPT 
PROMPT ========================================
PROMPT Paquete PKG_TEST creado con éxito:
PROMPT - TEST (simple)
PROMPT - TEST_PARAMS (múltiples params)
PROMPT - TEST_CURSOR (con cursor)
PROMPT - TEST_ERROR (manejo errores)
PROMPT - TEST_DML (operaciones DML)
PROMPT - TEST_DEMORA_LOOP (demora con timestamp)
PROMPT - TEST_DEMORA_QUERY (demora con consulta)
PROMPT ========================================
PROMPT 
PROMPT Ejemplo de uso:
PROMPT   PKG_TEST.TEST('valor', :v_out);
PROMPT   PKG_TEST.TEST_PARAMS(123, 'texto', SYSDATE, :v_result);
PROMPT ========================================
