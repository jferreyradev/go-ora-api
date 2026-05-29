-- ==============================================================================
-- Migración: soporte de concurrencia para ASYNC_JOBS
-- Agrega execution_mode + lock_key y nuevos índices
-- ==============================================================================

ALTER TABLE ASYNC_JOBS ADD (
    EXECUTION_MODE VARCHAR2(20) DEFAULT 'parallel' NOT NULL,
    LOCK_KEY VARCHAR2(200)
);

UPDATE ASYNC_JOBS
SET LOCK_KEY = PROCEDURE_NAME
WHERE LOCK_KEY IS NULL;

ALTER TABLE ASYNC_JOBS MODIFY (LOCK_KEY NOT NULL);

ALTER TABLE ASYNC_JOBS ADD CONSTRAINT CHK_ASYNC_JOBS_EXECUTION_MODE
CHECK (EXECUTION_MODE IN ('parallel', 'sequential', 'exclusive'));

CREATE INDEX IDX_ASYNC_JOBS_STATUS_NAME ON ASYNC_JOBS(STATUS, PROCEDURE_NAME);
CREATE INDEX IDX_ASYNC_JOBS_LOCK_KEY_STATUS ON ASYNC_JOBS(LOCK_KEY, STATUS, CREATED_AT);

COMMIT;
