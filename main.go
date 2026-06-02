package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/sijms/go-ora/v2"
	"gopkg.in/yaml.v3"
)

var db *sql.DB
var logFileName string  // Nombre del log de la instancia
var instanceName string // Nombre/etiqueta de la instancia

// JobStatus representa el estado de un job as├¡ncrono
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type ExecutionMode string

const (
	ExecutionModeParallel   ExecutionMode = "parallel"
	ExecutionModeSequential ExecutionMode = "sequential"
	ExecutionModeExclusive  ExecutionMode = "exclusive"
)

// AsyncJob representa un job de procedimiento en ejecuci├│n
type AsyncJob struct {
	ID            string                 `json:"id"`
	Status        JobStatus              `json:"status"`
	ProcName      string                 `json:"procedure_name"`
	ExecutionMode ExecutionMode          `json:"execution_mode,omitempty"`
	LockKey       string                 `json:"lock_key,omitempty"`
	Params        map[string]interface{} `json:"params,omitempty"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      string                 `json:"duration,omitempty"`
	Result        map[string]interface{} `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Progress      int                    `json:"progress"` // 0-100
}

// QueryLog representa un registro de consulta ejecutada
type QueryLog struct {
	ID            string    `json:"id"`
	QueryType     string    `json:"query_type"` // QUERY, EXEC, PROCEDURE
	QueryText     string    `json:"query_text"`
	Params        string    `json:"params,omitempty"`
	ExecutionTime time.Time `json:"execution_time"`
	Duration      string    `json:"duration"`
	RowsAffected  int64     `json:"rows_affected"`
	Success       bool      `json:"success"`
	ErrorMsg      string    `json:"error_msg,omitempty"`
	UserIP        string    `json:"user_ip,omitempty"`
}

// JobManager gestiona los jobs as├¡ncronos
type JobManager struct {
	jobs map[string]*AsyncJob
	mu   sync.RWMutex
}

var jobManager = &JobManager{
	jobs: make(map[string]*AsyncJob),
}

var errExclusiveJobConflict = fmt.Errorf("ya existe un job activo para esta lock_key")

// generateJobID genera un ID ├║nico para el job
func generateJobID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Error generando ID: %v", err)
		// Fallback: usar timestamp
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func normalizeExecutionMode(mode string) (ExecutionMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return ExecutionModeParallel, nil
	}
	switch ExecutionMode(normalized) {
	case ExecutionModeParallel, ExecutionModeSequential, ExecutionModeExclusive:
		return ExecutionMode(normalized), nil
	default:
		return "", fmt.Errorf("execution_mode inválido: %s", mode)
	}
}

func normalizeLockKey(lockKey, procName string) string {
	normalized := strings.TrimSpace(lockKey)
	if normalized == "" {
		return procName
	}
	return normalized
}

// CreateJob crea un nuevo job y lo registra (en memoria y BD)
func (jm *JobManager) CreateJob(procName string, params map[string]interface{}, executionMode ExecutionMode, lockKey string) (*AsyncJob, error) {
	jm.mu.Lock()

	job := &AsyncJob{
		ID:            generateJobID(),
		Status:        JobStatusPending,
		ProcName:      procName,
		ExecutionMode: executionMode,
		LockKey:       lockKey,
		Params:        params,
		StartTime:     time.Now(),
		Progress:      0,
	}
	jm.jobs[job.ID] = job
	jm.mu.Unlock()

	// Guardar en base de datos
	if err := jm.saveJobToDB(job); err != nil {
		jm.mu.Lock()
		delete(jm.jobs, job.ID)
		jm.mu.Unlock()
		return nil, err
	}

	return job, nil
}

// GetJob obtiene un job por su ID
func (jm *JobManager) GetJob(id string) (*AsyncJob, bool) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	job, exists := jm.jobs[id]
	return job, exists
}

// GetAllJobs retorna todos los jobs
func (jm *JobManager) GetAllJobs() []*AsyncJob {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jobs := make([]*AsyncJob, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// UpdateJob actualiza el estado de un job (en memoria y BD)
func (jm *JobManager) UpdateJob(id string, updateFn func(*AsyncJob)) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if job, exists := jm.jobs[id]; exists {
		updateFn(job)
		// Actualizar en base de datos
		go jm.updateJobInDB(job)
	}
}

func (jm *JobManager) markJobRunningInMemory(id string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	if job, exists := jm.jobs[id]; exists {
		job.Status = JobStatusRunning
		job.Progress = 10
	}
}

func (jm *JobManager) TryStartJob(id string, executionMode ExecutionMode, lockKey string) (bool, error) {
	if db == nil {
		jm.markJobRunningInMemory(id)
		return true, nil
	}

	if executionMode == ExecutionModeParallel {
		result, err := db.Exec(`
			UPDATE ASYNC_JOBS
			SET STATUS = :1, PROGRESS = :2
			WHERE JOB_ID = :3
			  AND STATUS = :4`,
			string(JobStatusRunning), 10, id, string(JobStatusPending),
		)
		if err != nil {
			return false, err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return false, err
		}
		if rowsAffected == 0 {
			return false, nil
		}
		jm.markJobRunningInMemory(id)
		return true, nil
	}

	result, err := db.Exec(`
		UPDATE ASYNC_JOBS j
		SET j.STATUS = :1,
		    j.PROGRESS = :2
		WHERE j.JOB_ID = :3
		  AND j.STATUS = :4
		  AND NOT EXISTS (
		    SELECT 1
		    FROM ASYNC_JOBS r
		    WHERE r.LOCK_KEY = :5
		      AND r.STATUS = :6
		      AND r.JOB_ID <> j.JOB_ID
		  )
		  AND NOT EXISTS (
		    SELECT 1
		    FROM ASYNC_JOBS p
		    WHERE p.LOCK_KEY = :5
		      AND p.STATUS = :4
		      AND (p.CREATED_AT < j.CREATED_AT OR (p.CREATED_AT = j.CREATED_AT AND p.JOB_ID < j.JOB_ID))
		  )`,
		string(JobStatusRunning), 10, id, string(JobStatusPending), lockKey, string(JobStatusRunning),
	)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}
	jm.markJobRunningInMemory(id)
	return true, nil
}

// CleanupOldJobs elimina jobs completados hace m├ís de 24 horas
func (jm *JobManager) CleanupOldJobs() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for id, job := range jm.jobs {
		if job.EndTime != nil && job.EndTime.Before(cutoff) {
			delete(jm.jobs, id)
		}
	}
}

// formatObjectName formatea el nombre de un procedimiento/funci├│n considerando esquema
// Retorna el nombre formateado seg├║n las reglas:
// - Si schema especificado: SCHEMA.NAME (sin comillas)
// - Si name contiene punto: "PARTE1"."PARTE2" (con comillas)
// - Si nombre simple: NAME (sin comillas, en may├║sculas)
func formatObjectName(schema, name string) string {
	if schema != "" {
		// Si se especifica el esquema por separado, usar SCHEMA.NAME sin comillas
		return fmt.Sprintf("%s.%s", strings.ToUpper(schema), strings.ToUpper(name))
	} else if strings.Contains(name, ".") && !strings.Contains(name, "\"") {
		// Si contiene punto y no tiene comillas, agregar comillas dobles a cada parte
		parts := strings.Split(name, ".")
		for i, part := range parts {
			parts[i] = fmt.Sprintf("\"%s\"", strings.ToUpper(part))
		}
		return strings.Join(parts, ".")
	}
	// Nombre simple, sin esquema
	return strings.ToUpper(name)
}

// parseDateParam intenta parsear un valor como fecha
func parseDateParam(value interface{}) (time.Time, error) {
	var t time.Time
	if timeVal, ok := value.(time.Time); ok {
		return timeVal, nil
	}
	if s, ok := value.(string); ok {
		for _, layout := range getDateInputFormats() {
			if parsedTime, err := time.Parse(layout, s); err == nil {
				return parsedTime, nil
			}
		}
	}
	return t, fmt.Errorf("no se pudo parsear fecha")
}

// setupLogFileName genera nombre de archivo de log con estructura: log/{instanceName}/{YYYY-MM-DD}/{instanceName}_{port}_{timestamp}.log
// Crea las carpetas necesarias automáticamente
func setupLogFileName(instanceName, port string) string {
	now := time.Now()
	timestamp := now.Format("2006-01-02_15-04-05")
	dateFolder := now.Format("2006-01-02")

	var baseName string
	if strings.ToLower(instanceName) == "auto" {
		baseName = "auto"
	} else {
		baseName = instanceName
	}

	// Crear estructura de carpetas: log/{baseName}/{dateFolder}/
	logDir := fmt.Sprintf("log/%s/%s", baseName, dateFolder)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Usar fmt.Println porque log aún no está configurado
		fmt.Fprintf(os.Stderr, "⚠️  Error creando directorio de logs %s: %v\n", logDir, err)
	}

	// Nombre representativo: {INSTANCIA}_{PUERTO}_{timestamp}.log
	return fmt.Sprintf("%s/%s_%s_%s.log", logDir, baseName, port, timestamp)
}

// formatDateOutput formatea fecha para salida (DD/MM/YYYY)
func formatDateOutput(t time.Time) string {
	return t.Format("02/01/2006")
}

// getDateInputFormats retorna formatos soportados para entrada
func getDateInputFormats() []string {
	return []string{
		"2006-01-02",
		"02/01/2006",
		"02-01-2006",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
	}
}

// DeleteJob elimina un job espec├¡fico por ID (memoria y BD)
func (jm *JobManager) DeleteJob(id string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if _, exists := jm.jobs[id]; !exists {
		return fmt.Errorf("job no encontrado")
	}

	// Eliminar de memoria
	delete(jm.jobs, id)

	// Eliminar de BD
	if db != nil {
		_, err := db.Exec("DELETE FROM ASYNC_JOBS WHERE JOB_ID = :1", id)
		if err != nil {
			log.Printf("Error eliminando job %s de BD: %v", id, err)
			return err
		}
	}

	return nil
}

// DeleteJobs elimina m├║ltiples jobs seg├║n criterios
func (jm *JobManager) DeleteJobs(status []string, olderThanDays int) (int, error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	var deleted []string
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	for id, job := range jm.jobs {
		shouldDelete := false

		// Filtro por status
		if len(status) > 0 {
			for _, s := range status {
				if string(job.Status) == s {
					shouldDelete = true
					break
				}
			}
		} else if olderThanDays > 0 {
			// Filtro por fecha
			if job.StartTime.Before(cutoff) {
				shouldDelete = true
			}
		}

		if shouldDelete {
			deleted = append(deleted, id)
		}
	}

	// Eliminar de memoria
	for _, id := range deleted {
		delete(jm.jobs, id)
	}

	// Eliminar de BD
	if db != nil && len(deleted) > 0 {
		for _, id := range deleted {
			_, err := db.Exec("DELETE FROM ASYNC_JOBS WHERE JOB_ID = :1", id)
			if err != nil {
				log.Printf("Error eliminando job %s de BD: %v", id, err)
			}
		}
	}

	return len(deleted), nil
}

// saveJobToDB guarda un job en la base de datos
func (jm *JobManager) saveJobToDB(job *AsyncJob) error {
	if db == nil {
		return nil
	}

	// Convertir par├ímetros a JSON
	var paramsJSON string
	if job.Params != nil {
		if jsonBytes, err := json.Marshal(job.Params); err == nil {
			paramsJSON = string(jsonBytes)
		}
	}

	insertSQL := `
		INSERT INTO ASYNC_JOBS (
			JOB_ID, STATUS, PROCEDURE_NAME, EXECUTION_MODE, LOCK_KEY, PARAMS, START_TIME, 
			END_TIME, DURATION, RESULT, ERROR_MSG, PROGRESS
		) VALUES (
			:1, :2, :3, :4, :5, :6, :7, :8, :9, :10, :11, :12
		)`

	args := []interface{}{
		job.ID,
		string(job.Status),
		job.ProcName,
		string(job.ExecutionMode),
		job.LockKey,
		paramsJSON,
		job.StartTime,
		job.EndTime,
		job.Duration,
		nil, // RESULT ser├í actualizado despu├®s
		job.Error,
		job.Progress,
	}

	if job.ExecutionMode == ExecutionModeExclusive {
		result, err := db.Exec(`
			INSERT INTO ASYNC_JOBS (
				JOB_ID, STATUS, PROCEDURE_NAME, EXECUTION_MODE, LOCK_KEY, PARAMS, START_TIME, 
				END_TIME, DURATION, RESULT, ERROR_MSG, PROGRESS
			)
			SELECT :1, :2, :3, :4, :5, :6, :7, :8, :9, :10, :11, :12
			FROM DUAL
			WHERE NOT EXISTS (
				SELECT 1
				FROM ASYNC_JOBS
				WHERE LOCK_KEY = :13
				  AND STATUS IN ('pending', 'running')
			)`,
			args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10], args[11], job.LockKey,
		)
		if err != nil {
			log.Printf("Error guardando job %s en BD: %v", job.ID, err)
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return errExclusiveJobConflict
		}
		return nil
	}

	if _, err := db.Exec(insertSQL, args...); err != nil {
		log.Printf("Error guardando job %s en BD: %v", job.ID, err)
		return err
	}
	return nil
}

// updateJobInDB actualiza un job en la base de datos
func (jm *JobManager) updateJobInDB(job *AsyncJob) {
	if db == nil {
		return
	}

	// Convertir resultado a JSON
	var resultJSON string
	if job.Result != nil {
		if jsonBytes, err := json.Marshal(job.Result); err == nil {
			resultJSON = string(jsonBytes)
		}
	}

	_, err := db.Exec(`
		UPDATE ASYNC_JOBS SET
			STATUS = :1,
			END_TIME = :2,
			DURATION = :3,
			RESULT = :4,
			ERROR_MSG = :5,
			PROGRESS = :6
		WHERE JOB_ID = :7`,
		string(job.Status),
		job.EndTime,
		job.Duration,
		resultJSON,
		job.Error,
		job.Progress,
		job.ID,
	)

	if err != nil {
		log.Printf("Error actualizando job %s en BD: %v", job.ID, err)
	}
}

// LoadJobsFromDB carga jobs desde la base de datos al iniciar
func (jm *JobManager) LoadJobsFromDB() {
	if db == nil {
		return
	}

	rows, err := db.Query(`
		SELECT JOB_ID, STATUS, PROCEDURE_NAME, EXECUTION_MODE, LOCK_KEY, PARAMS, START_TIME,
		       END_TIME, DURATION, RESULT, ERROR_MSG, PROGRESS
		FROM ASYNC_JOBS
		WHERE START_TIME >= SYSDATE - 1
		ORDER BY START_TIME DESC
	`)

	if err != nil {
		if strings.Contains(err.Error(), "ORA-00942") {
			log.Println("ÔÜá´©Å  Tabla ASYNC_JOBS no existe. Ejecuta: sqlplus @sql/create_async_jobs_table.sql")
		}
		return
	}
	defer rows.Close()

	jm.mu.Lock()
	defer jm.mu.Unlock()

	count := 0
	for rows.Next() {
		var job AsyncJob
		var endTime sql.NullTime
		var duration, paramsJSON, resultJSON, errorMsg, executionMode, lockKey sql.NullString

		err := rows.Scan(&job.ID, &job.Status, &job.ProcName, &executionMode, &lockKey, &paramsJSON, &job.StartTime,
			&endTime, &duration, &resultJSON, &errorMsg, &job.Progress)

		if err == nil {
			if executionMode.Valid {
				job.ExecutionMode = ExecutionMode(strings.ToLower(executionMode.String))
			} else {
				job.ExecutionMode = ExecutionModeParallel
			}
			if lockKey.Valid {
				job.LockKey = lockKey.String
			}
			if job.LockKey == "" {
				job.LockKey = job.ProcName
			}
			if endTime.Valid {
				job.EndTime = &endTime.Time
			}
			if duration.Valid {
				job.Duration = duration.String
			}
			if paramsJSON.Valid && paramsJSON.String != "" {
				if err := json.Unmarshal([]byte(paramsJSON.String), &job.Params); err != nil {
					log.Printf("Error deserializando par├ímetros del job %s: %v", job.ID, err)
				}
			}
			if resultJSON.Valid && resultJSON.String != "" {
				if err := json.Unmarshal([]byte(resultJSON.String), &job.Result); err != nil {
					log.Printf("Error deserializando resultado del job %s: %v", job.ID, err)
				}
			}
			if errorMsg.Valid {
				job.Error = errorMsg.String
			}
			jm.jobs[job.ID] = &job
			count++
		}
	}

	if count > 0 {
		log.Printf("Ô£à Cargados %d jobs desde Oracle", count)
	}
}

// createTableIfNotExists crea la tabla ASYNC_JOBS si no existe
func createTableIfNotExists() error {
	if db == nil {
		return fmt.Errorf("base de datos no disponible")
	}

	// Verificar si la tabla existe
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM USER_TABLES WHERE TABLE_NAME = 'ASYNC_JOBS'").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Ô£à Tabla ASYNC_JOBS ya existe")
		return nil
	}

	// Crear la tabla
	log.Println("­ƒôØ Creando tabla ASYNC_JOBS...")

	createTableSQL := `
		CREATE TABLE ASYNC_JOBS (
			JOB_ID VARCHAR2(32) PRIMARY KEY,
			STATUS VARCHAR2(20) NOT NULL,
			PROCEDURE_NAME VARCHAR2(200) NOT NULL,
			EXECUTION_MODE VARCHAR2(20) DEFAULT 'parallel' NOT NULL CHECK (EXECUTION_MODE IN ('parallel', 'sequential', 'exclusive')),
			LOCK_KEY VARCHAR2(200) NOT NULL,
			PARAMS CLOB,
			START_TIME TIMESTAMP NOT NULL,
			END_TIME TIMESTAMP,
			DURATION VARCHAR2(50),
			RESULT CLOB,
			ERROR_MSG CLOB,
			PROGRESS NUMBER DEFAULT 0,
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creando tabla: %v", err)
	}

	// Crear ├¡ndices
	if _, err := db.Exec("CREATE INDEX IDX_ASYNC_JOBS_STATUS ON ASYNC_JOBS(STATUS)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_ASYNC_JOBS_STATUS: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX IDX_ASYNC_JOBS_STATUS_NAME ON ASYNC_JOBS(STATUS, PROCEDURE_NAME)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_ASYNC_JOBS_STATUS_NAME: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX IDX_ASYNC_JOBS_LOCK_KEY_STATUS ON ASYNC_JOBS(LOCK_KEY, STATUS, CREATED_AT)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_ASYNC_JOBS_LOCK_KEY_STATUS: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX IDX_ASYNC_JOBS_START_TIME ON ASYNC_JOBS(START_TIME)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_ASYNC_JOBS_START_TIME: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX IDX_ASYNC_JOBS_CREATED_AT ON ASYNC_JOBS(CREATED_AT)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_ASYNC_JOBS_CREATED_AT: %v", err)
	}

	log.Println("Ô£à Tabla ASYNC_JOBS creada exitosamente")
	return nil
}

// createQueryLogTable crea la tabla QUERY_LOG si no existe
func createQueryLogTable() error {
	// Verificar si la tabla existe
	var tableName string
	err := db.QueryRow("SELECT table_name FROM user_tables WHERE table_name = 'QUERY_LOG'").Scan(&tableName)
	if err == nil {
		log.Println("Ô£à Tabla QUERY_LOG ya existe")
		return nil
	}

	log.Println("­ƒôØ Creando tabla QUERY_LOG...")

	createTableSQL := `
		CREATE TABLE QUERY_LOG (
			LOG_ID VARCHAR2(32) PRIMARY KEY,
			QUERY_TYPE VARCHAR2(20) NOT NULL,
			QUERY_TEXT CLOB NOT NULL,
			PARAMS CLOB,
			EXECUTION_TIME TIMESTAMP NOT NULL,
			DURATION VARCHAR2(50),
			ROWS_AFFECTED NUMBER,
			SUCCESS NUMBER(1) DEFAULT 1,
			ERROR_MSG CLOB,
			USER_IP VARCHAR2(50),
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creando tabla QUERY_LOG: %v", err)
	}

	// Crear ├¡ndices
	if _, err := db.Exec("CREATE INDEX IDX_QUERY_LOG_TYPE ON QUERY_LOG(QUERY_TYPE)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_QUERY_LOG_TYPE: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX IDX_QUERY_LOG_TIME ON QUERY_LOG(EXECUTION_TIME)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_QUERY_LOG_TIME: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX IDX_QUERY_LOG_SUCCESS ON QUERY_LOG(SUCCESS)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_QUERY_LOG_SUCCESS: %v", err)
	}
	if _, err := db.Exec("CREATE INDEX IDX_QUERY_LOG_CREATED ON QUERY_LOG(CREATED_AT)"); err != nil {
		log.Printf("ÔÜá´©Å  Error creando ├¡ndice IDX_QUERY_LOG_CREATED: %v", err)
	}

	log.Println("Ô£à Tabla QUERY_LOG creada exitosamente")
	return nil
}

// saveQueryLog guarda un registro de consulta en la base de datos
func saveQueryLog(qlog *QueryLog) {
	startTime := time.Now()

	// Preparar valores para INSERT
	successInt := 0
	if qlog.Success {
		successInt = 1
	}

	query := `
		INSERT INTO QUERY_LOG (
			LOG_ID, QUERY_TYPE, QUERY_TEXT, PARAMS,
			EXECUTION_TIME, DURATION, ROWS_AFFECTED,
			SUCCESS, ERROR_MSG, USER_IP, CREATED_AT
		) VALUES (
			:1, :2, :3, :4, :5, :6, :7, :8, :9, :10, CURRENT_TIMESTAMP
		)`

	_, err := db.Exec(query,
		qlog.ID,
		qlog.QueryType,
		qlog.QueryText,
		qlog.Params,
		qlog.ExecutionTime,
		qlog.Duration,
		qlog.RowsAffected,
		successInt,
		qlog.ErrorMsg,
		qlog.UserIP,
	)

	if err != nil {
		log.Printf("Error guardando query log %s en BD: %v", qlog.ID, err)
	} else {
		elapsed := time.Since(startTime)
		if elapsed > 100*time.Millisecond {
			log.Printf("ÔÜá´©Å  saveQueryLog tard├│ %v para log %s", elapsed, qlog.ID)
		}
	}
}

// generateID genera un ID ├║nico para los logs
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Error generando ID: %v", err)
		// Fallback: usar timestamp
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// setWindowTitle cambia el t├¡tulo de la ventana seg├║n la plataforma
func setWindowTitle(title string) {
	switch runtime.GOOS {
	case "windows":
		// Windows: usar cmd title
		exec.Command("cmd", "/C", "title", title).Run()
	case "linux":
		// Linux: escape sequence para terminales compatibles
		fmt.Printf("\033]0;%s\007", title)
	case "darwin":
		// macOS: escape sequence para Terminal.app
		fmt.Printf("\033]0;%s\007", title)
	default:
		// Otras plataformas: no hacer nada o usar escape sequence gen├®rico
		fmt.Printf("\033]0;%s\007", title)
	}
}

// Configuraci├│n de la aplicaci├│n
type AppConfig struct {
	OracleUser     string
	OraclePassword string
	OracleHost     string
	OraclePort     string
	OracleService  string
	ListenPort     string
}

type yamlString string

func (s *yamlString) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err == nil {
		*s = yamlString(str)
		return nil
	}

	var number int
	if err := value.Decode(&number); err == nil {
		*s = yamlString(strconv.Itoa(number))
		return nil
	}

	return fmt.Errorf("valor YAML inválido para texto: %q", value.Value)
}

type YAMLConfig struct {
	Oracle struct {
		User     string     `yaml:"user"`
		Password string     `yaml:"password"`
		Host     string     `yaml:"host"`
		Port     yamlString `yaml:"port"`
		Service  string     `yaml:"service"`
	} `yaml:"oracle"`
	API struct {
		Token      string   `yaml:"token"`
		AllowedIPs []string `yaml:"allowed_ips"`
		NoAuth     *bool    `yaml:"no_auth"`
	} `yaml:"api"`
	Server struct {
		Port yamlString `yaml:"port"`
	} `yaml:"server"`
}

func setEnvIfEmpty(key, value string) {
	if os.Getenv(key) == "" && value != "" {
		_ = os.Setenv(key, value)
	}
}

func loadYAMLConfig(configFile string) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	var cfg YAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️ No se pudo leer %s: %v\n", configFile, err)
		return
	}

	setEnvIfEmpty("ORACLE_USER", cfg.Oracle.User)
	setEnvIfEmpty("ORACLE_PASSWORD", cfg.Oracle.Password)
	setEnvIfEmpty("ORACLE_HOST", cfg.Oracle.Host)
	setEnvIfEmpty("ORACLE_PORT", string(cfg.Oracle.Port))
	setEnvIfEmpty("ORACLE_SERVICE", cfg.Oracle.Service)
	setEnvIfEmpty("API_TOKEN", cfg.API.Token)
	setEnvIfEmpty("PORT", string(cfg.Server.Port))

	if os.Getenv("API_ALLOWED_IPS") == "" && len(cfg.API.AllowedIPs) > 0 {
		_ = os.Setenv("API_ALLOWED_IPS", strings.Join(cfg.API.AllowedIPs, ","))
	}
	if os.Getenv("API_NO_AUTH") == "" && cfg.API.NoAuth != nil {
		if *cfg.API.NoAuth {
			_ = os.Setenv("API_NO_AUTH", "1")
		} else {
			_ = os.Setenv("API_NO_AUTH", "0")
		}
	}
}

// Carga la configuraci├│n desde variables de entorno y argumentos, valida obligatorias
// Carga la configuración desde variables de entorno y argumentos, valida obligatorias
func loadConfig() AppConfig {
	envFile := ".env"
	// Ignorar flags especiales al buscar archivo .env
	if len(os.Args) > 1 && os.Args[1] != "" {
		arg := strings.ToLower(os.Args[1])
		// Solo usar como archivo .env si NO es un flag especial
		if arg != "--check" && arg != "-c" && arg != "check" &&
			arg != "--help" && arg != "-h" && arg != "help" {
			envFile = os.Args[1]
		}
	}
	if envFile == ".env" {
		if customEnv := os.Getenv("ENV_FILE"); customEnv != "" {
			envFile = customEnv
		}
	}
	_ = godotenv.Load(envFile)
	loadYAMLConfig("config.yaml")

	user := os.Getenv("ORACLE_USER")
	password := os.Getenv("ORACLE_PASSWORD")
	host := os.Getenv("ORACLE_HOST")
	port := os.Getenv("ORACLE_PORT")
	service := os.Getenv("ORACLE_SERVICE")

	// Limpiar comillas simples o dobles si existen
	// (godotenv las incluye literalmente, pero algunos sistemas las necesitan en el .env)
	password = strings.Trim(password, "'\"")
	user = strings.Trim(user, "'\"")
	host = strings.Trim(host, "'\"")
	service = strings.Trim(service, "'\"")

	missing := []string{}
	if user == "" {
		missing = append(missing, "ORACLE_USER")
	}
	if password == "" {
		missing = append(missing, "ORACLE_PASSWORD")
	}
	if host == "" {
		missing = append(missing, "ORACLE_HOST")
	}
	if port == "" {
		missing = append(missing, "ORACLE_PORT")
	}
	if service == "" {
		missing = append(missing, "ORACLE_SERVICE")
	}
	if len(missing) > 0 {
		msg := "Faltan variables obligatorias: " + strings.Join(missing, ", ") + "\nRevisa tu archivo .env o variables de entorno."
		fmt.Fprintln(os.Stderr, msg)
		_ = os.WriteFile("log/last_error.txt", []byte(msg+"\n"), 0644)
		os.Exit(2)
	}

	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		if len(os.Args) > 2 && os.Args[2] != "" {
			listenPort = os.Args[2]
		} else {
			listenPort = "8080"
		}
	}

	return AppConfig{
		OracleUser:     user,
		OraclePassword: password,
		OracleHost:     host,
		OraclePort:     port,
		OracleService:  service,
		ListenPort:     listenPort,
	}
}

// Abre la conexi├│n a Oracle y la retorna
func openOracleConnection(cfg AppConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"oracle://%s:%s@%s:%s/%s",
		url.QueryEscape(cfg.OracleUser),
		url.QueryEscape(cfg.OraclePassword),
		cfg.OracleHost,
		cfg.OraclePort,
		cfg.OracleService,
	)

	database, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir driver Oracle: %w", err)
	}

	// Configurar pool de conexiones para procedimientos largos
	database.SetMaxOpenConns(25)                  // M├íximo de conexiones abiertas
	database.SetMaxIdleConns(5)                   // Conexiones idle
	database.SetConnMaxLifetime(0)                // Sin l├¡mite de tiempo de vida (0 = infinito)
	database.SetConnMaxIdleTime(10 * time.Minute) // Cerrar idle despu├®s de 10 min

	// Verificar que la conexi├│n sea v├ílida con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("no se pudo conectar a la base de datos Oracle (host=%s:%s servicio=%s usuario=%s): %w",
			cfg.OracleHost, cfg.OraclePort, cfg.OracleService, cfg.OracleUser, err)
	}

	return database, nil
}

// checkRequirements verifica que todos los requisitos estén configurados correctamente
func checkRequirements(cfg AppConfig) int {
	fmt.Println("======================================")
	fmt.Println("🔍 VERIFICACIÓN DE REQUISITOS")
	fmt.Println("======================================")
	fmt.Println()

	successCount := 0
	warningCount := 0
	errorCount := 0

	// [1/6] Configuración (.env / config.yaml)
	fmt.Println("[1/6] Configuración (.env / config.yaml)")
	envFile := ".env"
	if len(os.Args) > 1 && os.Args[1] != "" && os.Args[1] != "--check" && os.Args[1] != "-c" && os.Args[1] != "check" {
		envFile = os.Args[1]
	}
	yamlFile := "config.yaml"
	yamlFound := false
	if _, err := os.Stat(yamlFile); err == nil {
		fmt.Printf("  ✅ Archivo %s encontrado\n", yamlFile)
		successCount++
		yamlFound = true
	}
	if _, err := os.Stat(envFile); err == nil {
		fmt.Printf("  ✅ Archivo %s encontrado\n", envFile)
		successCount++
	} else {
		if yamlFound {
			fmt.Printf("  ℹ️  Archivo %s no encontrado (usando config.yaml y/o variables de entorno del sistema)\n", envFile)
		} else {
			fmt.Printf("  ⚠️  Archivo %s no encontrado (usando variables de entorno del sistema)\n", envFile)
			warningCount++
		}
	}

	// Verificar variables requeridas
	requiredVars := []string{"ORACLE_USER", "ORACLE_PASSWORD", "ORACLE_HOST", "ORACLE_SERVICE"}
	for _, v := range requiredVars {
		if val := os.Getenv(v); val != "" {
			fmt.Printf("  ✅ %s configurado\n", v)
			successCount++
		} else {
			fmt.Printf("  ❌ %s NO configurado\n", v)
			errorCount++
		}
	}

	// Variables opcionales
	if os.Getenv("API_ALLOWED_IPS") == "" {
		fmt.Println("  ⚠️  API_ALLOWED_IPS no configurado (usará 0.0.0.0 - acepta todas las IPs)")
		warningCount++
	} else {
		fmt.Println("  ✅ API_ALLOWED_IPS configurado")
		successCount++
	}

	if os.Getenv("API_TOKEN") != "" {
		fmt.Println("  ✅ API_TOKEN configurado")
		successCount++
	} else {
		fmt.Println("  ⚠️  API_TOKEN no configurado (autenticación deshabilitada)")
		warningCount++
	}
	fmt.Println()

	// [2/6] Conectividad de Red
	fmt.Println("[2/6] Conectividad de Red")
	// Resolver hostname
	if ips, err := net.LookupHost(cfg.OracleHost); err == nil {
		fmt.Printf("  ✅ Hostname '%s' resuelve a %v\n", cfg.OracleHost, ips)
		successCount++
	} else {
		fmt.Printf("  ❌ Hostname '%s' no se puede resolver: %v\n", cfg.OracleHost, err)
		errorCount++
	}

	// Verificar puerto Oracle accesible
	address := net.JoinHostPort(cfg.OracleHost, cfg.OraclePort)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err == nil {
		conn.Close()
		fmt.Printf("  ✅ Puerto %s accesible\n", cfg.OraclePort)
		successCount++
	} else {
		fmt.Printf("  ❌ Puerto %s no accesible: %v\n", cfg.OraclePort, err)
		errorCount++
	}

	// Verificar puerto API disponible
	listener, err := net.Listen("tcp", ":"+cfg.ListenPort)
	if err == nil {
		listener.Close()
		fmt.Printf("  ✅ Puerto API %s disponible\n", cfg.ListenPort)
		successCount++
	} else {
		fmt.Printf("  ⚠️  Puerto API %s ya en uso\n", cfg.ListenPort)
		warningCount++
	}
	fmt.Println()

	// [3/6] Conexión a Oracle
	fmt.Println("[3/6] Conexión a Oracle")
	testDB, err := openOracleConnection(cfg)
	if err != nil {
		fmt.Printf("  ❌ No se pudo conectar a Oracle: %v\n", err)
		errorCount++
		fmt.Println()
	} else {
		defer testDB.Close()
		fmt.Println("  ✅ Conectado a Oracle exitosamente")
		successCount++

		// Obtener versión de Oracle
		var version string
		err := testDB.QueryRow("SELECT banner FROM v$version WHERE ROWNUM = 1").Scan(&version)
		if err == nil {
			fmt.Printf("  ℹ️  Versión Oracle: %s\n", version)
		}
		fmt.Println()

		// [4/6] Estructura de Base de Datos
		fmt.Println("[4/6] Estructura de Base de Datos")

		// Verificar tabla ASYNC_JOBS
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM user_tables WHERE table_name = 'ASYNC_JOBS'").Scan(&count)
		if err == nil && count > 0 {
			fmt.Println("  ✅ Tabla ASYNC_JOBS existe")
			successCount++
		} else {
			fmt.Println("  ⚠️  Tabla ASYNC_JOBS no encontrada (ejecutar: sql/create_async_jobs_table.sql)")
			warningCount++
		}

		// Verificar tabla QUERY_LOG
		err = testDB.QueryRow("SELECT COUNT(*) FROM user_tables WHERE table_name = 'QUERY_LOG'").Scan(&count)
		if err == nil && count > 0 {
			fmt.Println("  ✅ Tabla QUERY_LOG existe")
			successCount++
		} else {
			fmt.Println("  ⚠️  Tabla QUERY_LOG no encontrada (ejecutar: sql/create_query_log_table.sql)")
			warningCount++
		}

		// Verificar paquete PKG_TEST (opcional)
		err = testDB.QueryRow("SELECT COUNT(*) FROM user_objects WHERE object_name = 'PKG_TEST' AND object_type = 'PACKAGE'").Scan(&count)
		if err == nil && count > 0 {
			fmt.Println("  ✅ Paquete PKG_TEST existe")
			successCount++
		} else {
			fmt.Println("  ⚠️  Paquete PKG_TEST no encontrado (opcional, ejecutar: sql/create_test_procedures.sql)")
			warningCount++
		}
		fmt.Println()
	}

	// [5/6] Sistema Operativo
	fmt.Println("[5/6] Sistema Operativo")
	// Verificar directorio log
	logDir := "log"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err == nil {
			fmt.Printf("  ✅ Directorio /%s creado\n", logDir)
			successCount++
		} else {
			fmt.Printf("  ❌ No se pudo crear directorio /%s: %v\n", logDir, err)
			errorCount++
		}
	} else {
		// Verificar permisos de escritura
		testFile := logDir + "/test_write.tmp"
		if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
			os.Remove(testFile)
			fmt.Printf("  ✅ Directorio /%s existe y es escribible\n", logDir)
			successCount++
		} else {
			fmt.Printf("  ❌ Directorio /%s no es escribible: %v\n", logDir, err)
			errorCount++
		}
	}

	// Información del sistema
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("  ℹ️  Sistema: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// [6/6] Resumen
	fmt.Println("[6/6] Resumen")
	fmt.Printf("  ✅ %d verificaciones exitosas\n", successCount)
	if warningCount > 0 {
		fmt.Printf("  ⚠️  %d advertencias (no críticas)\n", warningCount)
	}
	if errorCount > 0 {
		fmt.Printf("  ❌ %d errores críticos\n", errorCount)
	}
	fmt.Println()

	fmt.Println("======================================")
	if errorCount > 0 {
		fmt.Println("❌ ERRORES CRÍTICOS ENCONTRADOS")
		fmt.Println("======================================")
		fmt.Println()
		fmt.Println("Corrige los errores antes de iniciar el servidor.")
		return 1
	} else if warningCount > 0 {
		fmt.Println("⚠️  SISTEMA FUNCIONAL CON ADVERTENCIAS")
		fmt.Println("======================================")
		fmt.Println()
		fmt.Println("El sistema puede iniciarse, pero revisa las advertencias.")
		fmt.Println()
		fmt.Println("Iniciar servidor:")
		fmt.Println("  go run main.go")
		fmt.Println("  ./go-oracle-api.exe")
		return 2
	} else {
		fmt.Println("✅ SISTEMA LISTO PARA EJECUTAR")
		fmt.Println("======================================")
		fmt.Println()
		fmt.Println("Todos los requisitos están configurados correctamente.")
		fmt.Println()
		fmt.Println("Iniciar servidor:")
		fmt.Println("  go run main.go")
		fmt.Println("  ./go-oracle-api.exe")
		return 0
	}
}

func main() {
	// ===============================
	// 1. Mostrar ayuda si se solicita
	// ===============================
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		if arg == "-h" || arg == "--help" || arg == "help" {
			fmt.Print(`
Go Oracle API - Opciones de ejecución

USO:
  go run main.go [archivo_env] [puerto] [nombre_instancia]
  go-oracle-api.exe [archivo_env] [puerto] [nombre_instancia]

Argumentos opcionales:
  archivo_env       Archivo de variables de entorno (por defecto .env)
  puerto            Puerto donde escuchará la API (por defecto 8080)
  nombre_instancia  Nombre para identificar esta instancia (por defecto auto)

Comandos especiales:
  --check, -c, check    Verificar requisitos y configuración sin iniciar servidor
  --help, -h, help      Mostrar esta ayuda

También puedes usar variables de entorno:
  ENV_FILE          Archivo de configuración
  PORT              Puerto de escucha
  INSTANCE_NAME     Nombre de la instancia

Ejemplos:
  go run main.go --check              # Verificar configuración
  go run main.go .env1 8081 "Produccion"
  go run main.go .env2 8082 "Testing"
  
  set INSTANCE_NAME=Desarrollo
  go run main.go .env3 8083

Para más información consulta:
  - README.md
  - docs/getting-started/QUICKSTART.md
  - Endpoint /docs
  - Endpoint /health (monitoreo)
`)
			os.Exit(0)
		}

		// Verificar si se solicitó verificación de requisitos
		if arg == "--check" || arg == "-c" || arg == "check" {
			cfg := loadConfig()
			exitCode := checkRequirements(cfg)
			os.Exit(exitCode)
		}
	}

	// ===============================
	// 2. Carga de configuración (primero para determinar puerto correcto)
	// ===============================
	cfg := loadConfig()

	// Determinar nombre de instancia
	instanceName = "auto"
	if len(os.Args) > 3 && os.Args[3] != "" {
		instanceName = os.Args[3]
	} else if customInstance := os.Getenv("INSTANCE_NAME"); customInstance != "" {
		instanceName = customInstance
	}

	// Generar nombre de log con estructura jerárquica (usando puerto correcto de config)
	logFileName = setupLogFileName(instanceName, cfg.ListenPort)
	fmt.Printf("📝 Archivo de log: %s\n", logFileName)

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		defer logFile.Close()  // Cerrar el archivo al finalizar
		log.SetOutput(logFile) // Solo archivo, no consola
		fmt.Printf("✅ Log configurado correctamente\n\n")
	} else {
		log.SetOutput(os.Stdout)
		fmt.Fprintf(os.Stderr, "❌ No se pudo abrir %s para logging: %v\n", logFileName, err)
		fmt.Fprintf(os.Stderr, "Los logs se escribirán en consola\n\n")
	}

	// ===============================
	// 3. Registro de todos los endpoints
	// ===============================
	http.HandleFunc("/docs", docsHandler)
	http.HandleFunc("/health", healthHandler) // Sin autenticación para monitoreo
	http.HandleFunc("/logs", logRequest(authMiddleware(logsHandler)))
	http.HandleFunc("/upload", logRequest(authMiddleware(uploadHandler)))
	http.HandleFunc("/download", logRequest(authMiddleware(downloadHandler)))
	http.HandleFunc("/ping", logRequest(authMiddleware(pingHandler)))
	http.HandleFunc("/query", logRequest(authMiddleware(queryHandler)))
	http.HandleFunc("/exec", logRequest(authMiddleware(execHandler)))
	http.HandleFunc("/procedure", logRequest(authMiddleware(procedureHandler)))
	http.HandleFunc("/procedure/async", logRequest(authMiddleware(asyncProcedureHandler)))
	http.HandleFunc("/jobs/", logRequest(authMiddleware(jobsHandler))) // /jobs/{id} y /jobs

	// ===============================
	// 4. Conexión a Oracle
	// ===============================
	fmt.Printf("Host: %s:%s\n", cfg.OracleHost, cfg.OraclePort)
	fmt.Printf("Servicio: %s\n", cfg.OracleService)
	fmt.Printf("Usuario: %s\n", cfg.OracleUser)
	fmt.Println("==============================")

	db, err = openOracleConnection(cfg)
	if err != nil {
		errorMsg := fmt.Sprintf("\nÔØî ERROR FATAL: No se pudo establecer conexi├│n con la base de datos\n\n%v\n\nVerifica:\n1. Que Oracle est├® ejecut├índose\n2. Los datos de conexi├│n en el archivo .env\n3. La conectividad de red al servidor\n4. El firewall y puertos abiertos\n\n", err)
		fmt.Fprint(os.Stderr, errorMsg)
		_ = os.WriteFile("log/last_error.txt", []byte(errorMsg), 0644)
		log.Fatal(errorMsg)
	}
	defer db.Close()

	fmt.Println("Ô£à Conexi├│n a Oracle establecida correctamente")
	fmt.Println()

	port := cfg.ListenPort
	user := cfg.OracleUser
	host := cfg.OracleHost
	service := cfg.OracleService

	// ===============================
	// 5. Inicializar tabla y cargar jobs as├¡ncronos
	// ===============================
	if err := createTableIfNotExists(); err != nil {
		log.Printf("ÔÜá´©Å  No se pudo crear/verificar tabla ASYNC_JOBS: %v", err)
	}
	if err := createQueryLogTable(); err != nil {
		log.Printf("ÔÜá´©Å  No se pudo crear/verificar tabla QUERY_LOG: %v", err)
	}
	jobManager.LoadJobsFromDB()

	// ===============================
	// 7. Detecci├│n de IPs locales
	// ===============================
	ips := []string{"0.0.0.0"}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		ips = []string{}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
		if len(ips) == 0 {
			ips = []string{"0.0.0.0"}
		}
	}

	// ===============================
	// 8. Resumen de arranque y estado
	// ===============================
	fmt.Println("==============================")
	if instanceName != "auto" {
		fmt.Printf("INSTANCIA: %s\n", instanceName)
		// Cambiar t├¡tulo de la ventana seg├║n la plataforma
		setWindowTitle(fmt.Sprintf("Go Oracle API - %s (Puerto %s)", instanceName, port))
	} else {
		// T├¡tulo por defecto
		setWindowTitle(fmt.Sprintf("Go Oracle API - Puerto %s", port))
	}
	fmt.Println("API escuchando en el puerto:", port)
	fmt.Printf("Conectado a Oracle: usuario=%s host=%s puerto=%s servicio=%s\n", user, host, cfg.OraclePort, service)
	fmt.Printf("Log de esta instancia: %s\n", logFileName)
	fmt.Println("==============================")
	log.Println("==============================")
	log.Println("Estado de la API al iniciar:")
	for _, ip := range ips {
		log.Printf("- Endpoint disponible: http://%s:%s", ip, port)
	}
	log.Println("- Endpoint de logs: /logs")
	log.Println("- Endpoint de ping: /ping")
	log.Println("- Endpoint de query: /query")
	log.Println("- Endpoint de exec: /exec")
	log.Println("- Endpoint de procedure: /procedure")
	log.Println("- Endpoint de upload: /upload")
	log.Println("- Endpoint de download: /download")
	log.Printf("- Conectado a Oracle: usuario=%s host=%s puerto=%s servicio=%s", user, host, port, service)
	// Estado de conexi├│n a Oracle
	if err := db.Ping(); err == nil {
		log.Printf("- Conexi├│n a Oracle: OK")
	} else {
		log.Printf("- Conexi├│n a Oracle: ERROR: %v", err)
		// Registrar error de ping en archivo especial
		_ = os.WriteFile("log/last_error.txt", []byte(fmt.Sprintf("Error de conexi├│n a Oracle (ping): %v\n", err)), 0644)
	}
	log.Println("==============================")
	fmt.Println("Endpoints disponibles:")
	for _, ip := range ips {
		fmt.Printf("- http://%s:%s\n", ip, port)
	}
	fmt.Println("  /logs      - Consulta el log actual de la instancia")
	fmt.Println("  /ping      - Prueba de vida de la API (GET)")
	fmt.Println("  /query     - Ejecuta una consulta SQL (GET)")
	fmt.Println("  /procedure - Ejecuta un procedimiento almacenado (POST)")
	fmt.Println("  /procedure/async - Ejecuta un procedimiento en segundo plano (POST)")
	fmt.Println("  /jobs                - Lista todos los jobs as├¡ncronos (GET)")
	fmt.Println("  /jobs?status=...     - Elimina jobs por status: completed,failed (DELETE)")
	fmt.Println("  /jobs?older_than=7   - Elimina jobs m├ís antiguos que N d├¡as (DELETE)")
	fmt.Println("  /jobs/{id}           - Consulta el estado de un job espec├¡fico (GET)")
	fmt.Println("  /jobs/{id}           - Elimina un job espec├¡fico (DELETE)")
	fmt.Println("  /upload    - Sube un archivo como BLOB (POST)")
	fmt.Println("  /download  - Descarga un archivo BLOB por ID (GET)")
	fmt.Println("              Params: id (requerido), table (opcional, default: archivos)")
	fmt.Println("              Ejemplo: /download?id=123 o /download?id=123&table=documentos")
	fmt.Println("\nPara detalles de uso y ejemplos, consulta la documentaci├│n en:")
	fmt.Println("  - /docs (endpoint)")
	fmt.Println("  - docs/USO_Y_PRUEBAS.md (archivo)")

	// ===============================
	// 9. Iniciar limpieza peri├│dica de jobs antiguos
	// ===============================
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			jobManager.CleanupOldJobs()
			log.Println("Limpieza de jobs antiguos completada")
		}
	}()

	// ===============================
	// 10. Iniciar servidor HTTP con graceful shutdown
	// ===============================
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: http.DefaultServeMux,
	}

	// Canal para se├▒ales del sistema
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor escuchando en 0.0.0.0:%s (Ctrl+C para detener)", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar el servidor: %v", err)
		}
	}()

	<-quit
	log.Println("\nSe├▒al de apagado recibida, cerrando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error en shutdown: %v", err)
	}
	log.Println("Servidor cerrado correctamente.")
}

// docsHandler sirve el contenido del README.md como documentaci├│n b├ísica
func docsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	data, err := os.ReadFile("README.md")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error leyendo la documentaci├│n: " + err.Error()))
		return
	}
	w.Write(data)
}

// logsHandler sirve el contenido del archivo de log de la instancia actual
func logsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if logFileName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No hay log de instancia disponible"))
		return
	}
	data, err := os.ReadFile(logFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error leyendo el log: " + err.Error()))
		return
	}
	w.Write(data)
}

// logRequest es un middleware que registra cada petici├│n HTTP entrante
func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if colon := strings.LastIndex(ip, ":"); colon != -1 {
			ip = ip[:colon]
		}
		log.Printf("%s %s desde %s", r.Method, r.URL.Path, ip)
		next(w, r)
	}
}

// uploadHandler recibe archivos v├¡a multipart/form-data y los guarda en una tabla BLOB
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solo se permite POST"})
		return
	}
	err := r.ParseMultipartForm(100 << 20) // 100 MB m├íximo
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al leer el archivo: " + err.Error()})
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Archivo no recibido: " + err.Error()})
		return
	}
	defer file.Close()

	// Leer el archivo en memoria (para archivos muy grandes, usar streaming a la BD)
	data := make([]byte, handler.Size)
	_, err = file.Read(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error leyendo archivo: " + err.Error()})
		return
	}

	// Metadatos opcionales
	nombre := handler.Filename
	descripcion := r.FormValue("descripcion")

	// Insertar en la tabla (ejemplo: archivos(id, nombre, descripcion, contenido BLOB))
	_, err = db.Exec("INSERT INTO archivos (nombre, descripcion, contenido) VALUES (:1, :2, :3)", nombre, descripcion, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error guardando en BD: " + err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "nombre": nombre})
}

// downloadHandler descarga archivos BLOB desde la base de datos
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solo se permite GET"})
		return
	}

	// Obtener par├ímetros de la query
	id := r.URL.Query().Get("id")
	table := r.URL.Query().Get("table")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Falta el par├ímetro 'id'"})
		return
	}

	// Tabla por defecto
	if table == "" {
		table = "archivos"
	}

	// Validar nombre de tabla para prevenir SQL injection
	// Solo permitir letras, n├║meros y guiones bajos, m├íximo 30 caracteres
	validTableName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validTableName.MatchString(table) || len(table) > 30 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nombre de tabla inv├ílido"})
		return
	}

	// Query para obtener el archivo
	query := fmt.Sprintf("SELECT nombre, contenido FROM %s WHERE id = :1", table)

	var nombre string
	var contenido []byte

	err := db.QueryRow(query, id).Scan(&nombre, &contenido)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Archivo no encontrado"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error consultando BD: " + err.Error()})
		return
	}

	// Detectar tipo MIME basado en la extensi├│n del archivo
	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(nombre), ".pdf") {
		contentType = "application/pdf"
	} else if strings.HasSuffix(strings.ToLower(nombre), ".jpg") || strings.HasSuffix(strings.ToLower(nombre), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(nombre), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(nombre), ".txt") {
		contentType = "text/plain"
	} else if strings.HasSuffix(strings.ToLower(nombre), ".json") {
		contentType = "application/json"
	}

	// Configurar headers para la descarga
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", nombre))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(contenido)))

	// Enviar el contenido del archivo
	w.WriteHeader(http.StatusOK)
	w.Write(contenido)
}

// procedureHandler ejecuta un procedimiento almacenado con par├ímetros IN y OUT
func procedureHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solo se permite POST"})
		return
	}

	var req struct {
		Name   string `json:"name"`
		Schema string `json:"schema,omitempty"` // Esquema del procedimiento/funci├│n
		Params []struct {
			Name      string      `json:"name"`
			Value     interface{} `json:"value,omitempty"`
			Direction string      `json:"direction,omitempty"`
			Type      string      `json:"type,omitempty"` // "number", "string", "date"
		} `json:"params"`
		IsFunction bool `json:"isFunction,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inv├ílido"})
		return
	}

	// Log para debug: mostrar el JSON recibido
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[PROCEDURE] JSON recibido:\n%s", string(reqJSON))

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Falta el campo 'name'"})
		return
	}

	// Validar que parámetros IN tengan valores
	var missingParams []string
	for _, p := range req.Params {
		if strings.ToUpper(p.Direction) != "OUT" {
			if p.Value == nil {
				missingParams = append(missingParams, p.Name)
			}
		}
	}
	if len(missingParams) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":          "Parámetros requeridos sin valor",
			"missing_params": missingParams,
		})
		return
	}

	if req.Schema != "" {
		log.Printf("[PROCEDURE] Ejecutando: %s.%s con %d par├ímetros", req.Schema, req.Name, len(req.Params))
	} else {
		log.Printf("[PROCEDURE] Ejecutando: %s con %d par├ímetros", req.Name, len(req.Params))
	}

	placeholders := []string{}
	args := []interface{}{}
	outIndexes := make(map[int]string)
	outBuffers := make(map[int]*string)
	outNumMap := make(map[int]*sql.NullFloat64)
	outDateMap := make(map[int]*sql.NullTime)

	// Si es función, el primer OUT es el valor de retorno
	if req.IsFunction {
		// Contar parámetros OUT para funciones
		outParamCount := 0
		for _, p := range req.Params {
			if strings.ToUpper(p.Direction) == "OUT" {
				outParamCount++
			}
		}

		// Preasignar buffers para el RETURN value + todos los OUT params
		preallocDates := make([]sql.NullTime, outParamCount+1)
		preallocNums := make([]sql.NullFloat64, outParamCount+1)
		preallocStrs := make([]string, outParamCount+1)
		for j := 0; j < outParamCount+1; j++ {
			preallocStrs[j] = strings.Repeat(" ", 4000)
		}

		// :1 es siempre el RETURN value de la función
		placeholders = append(placeholders, ":1")

		// El retorno de la función va en índice 0
		// Asumir que es NUMBER a menos que se especifique otro tipo
		args = append(args, sql.Out{Dest: &preallocNums[0], In: false})
		outNumMap[0] = &preallocNums[0]
		outIndexes[0] = "return_value"

		// Procesar parámetros en orden (IN y OUT)
		paramPos := 2
		outIdx := 1 // Índice para OUT params (después del RETURN)
		for _, p := range req.Params {
			placeholders = append(placeholders, fmt.Sprintf(":%d", paramPos))

			if strings.ToUpper(p.Direction) == "OUT" {
				lowerName := strings.ToLower(p.Name)
				isDate := strings.ToLower(p.Type) == "date"
				isNum := strings.ToLower(p.Type) == "number" ||
					strings.Contains(lowerName, "resultado") || strings.Contains(lowerName, "result") ||
					strings.Contains(lowerName, "total") || strings.Contains(lowerName, "count") ||
					strings.Contains(lowerName, "suma") || strings.Contains(lowerName, "num") ||
					strings.Contains(lowerName, "int") || strings.Contains(lowerName, "id")

				outIndexes[outIdx] = p.Name

				if isDate {
					args = append(args, sql.Out{Dest: &preallocDates[outIdx], In: false})
					outDateMap[outIdx] = &preallocDates[outIdx]
				} else if isNum {
					args = append(args, sql.Out{Dest: &preallocNums[outIdx], In: false})
					outNumMap[outIdx] = &preallocNums[outIdx]
				} else {
					args = append(args, sql.Out{Dest: &preallocStrs[outIdx], In: false})
					outBuffers[outIdx] = &preallocStrs[outIdx]
				}
				outIdx++
			} else {
				// Parámetro IN: verificar si es fecha
				pTypeLower := strings.ToLower(p.Type)
				pNameLower := strings.ToLower(p.Name)
				isDateType := pTypeLower == "date"
				isDateName := strings.Contains(pNameLower, "fecha") || strings.Contains(pNameLower, "periodo")

				if isDateType || isDateName {
					// Intentar parsear como fecha
					if parsedTime, err := parseDateParam(p.Value); err == nil {
						args = append(args, parsedTime)
					} else {
						log.Printf("[PROCEDURE] Advertencia al parsear fecha en parámetro '%s': %v", p.Name, err)
						args = append(args, p.Value)
					}
				} else {
					args = append(args, p.Value)
				}
			}
			paramPos++
		}

		// Formatear el nombre para manejar esquema.función correctamente
		functionName := formatObjectName(req.Schema, req.Name)

		call := fmt.Sprintf("BEGIN :1 := %s(%s); END;", functionName, strings.Join(placeholders[1:], ", "))
		log.Printf("[PROCEDURE] SQL generado para funci├│n: %s", call)

		stmt, err := db.Prepare(call)
		if err != nil {
			errorMsg := err.Error()

			// Mejorar el mensaje de error
			if strings.Contains(errorMsg, "PLS-00201") {
				errorMsg = fmt.Sprintf("Funci├│n '%s' no encontrada. Verifica que existe en la base de datos.", req.Name)
			} else if strings.Contains(errorMsg, "PLS-00306") {
				errorMsg = fmt.Sprintf("Par├ímetros incorrectos para '%s'. Verifica tipos y cantidad de par├ímetros.", req.Name)
			}

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
			return
		}
		defer stmt.Close()

		if _, err := stmt.Exec(args...); err != nil {
			errorMsg := err.Error()

			// Mejorar mensajes de error comunes
			if strings.Contains(errorMsg, "ORA-06502") {
				errorMsg = "Error de conversi├│n de tipos. Verifica que los tipos de datos sean correctos."
			} else if strings.Contains(errorMsg, "ORA-01403") {
				errorMsg = "No se encontraron datos. La funci├│n no retorn├│ resultados."
			}

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
			return
		}

		out := make(map[string]interface{})
		for i, name := range outIndexes {
			if datePtr, ok := outDateMap[i]; ok && datePtr != nil {
				if datePtr.Valid {
					out[name] = formatDateOutput(datePtr.Time)
				} else {
					out[name] = nil
				}
				continue
			}
			if numPtr, ok := outNumMap[i]; ok && numPtr != nil {
				if numPtr.Valid {
					out[name] = numPtr.Float64
				} else {
					out[name] = nil
				}
				continue
			}
			if ptr, ok := outBuffers[i]; ok && ptr != nil {
				out[name] = *ptr
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "out": out})
		return
	}

	// Procedimiento normal (no función)
	// Contar parámetros OUT para preasignar buffers
	outParamCount := 0
	for _, p := range req.Params {
		if strings.ToUpper(p.Direction) == "OUT" {
			outParamCount++
		}
	}

	// Preasignar buffers para todos los OUT params
	preallocDates := make([]sql.NullTime, outParamCount)
	preallocNums := make([]sql.NullFloat64, outParamCount)
	preallocStrs := make([]string, outParamCount)
	for j := 0; j < outParamCount; j++ {
		preallocStrs[j] = strings.Repeat(" ", 4000)
	}

	// Procesar parámetros
	outIdx := 0
	for i, p := range req.Params {
		placeholders = append(placeholders, fmt.Sprintf(":%d", i+1))

		if strings.ToUpper(p.Direction) == "OUT" {
			lowerName := strings.ToLower(p.Name)
			isDate := strings.ToLower(p.Type) == "date"
			isNum := strings.ToLower(p.Type) == "number" ||
				strings.Contains(lowerName, "resultado") || strings.Contains(lowerName, "result") ||
				strings.Contains(lowerName, "total") || strings.Contains(lowerName, "count") ||
				strings.Contains(lowerName, "suma") || strings.Contains(lowerName, "num") ||
				strings.Contains(lowerName, "int") || strings.Contains(lowerName, "id")

			outIndexes[i] = p.Name // Guardar nombre con índice global

			if isDate {
				args = append(args, sql.Out{Dest: &preallocDates[outIdx], In: false})
				outDateMap[i] = &preallocDates[outIdx]
			} else if isNum {
				args = append(args, sql.Out{Dest: &preallocNums[outIdx], In: false})
				outNumMap[i] = &preallocNums[outIdx]
			} else {
				args = append(args, sql.Out{Dest: &preallocStrs[outIdx], In: false})
				outBuffers[i] = &preallocStrs[outIdx]
			}
			outIdx++
		} else {
			// Parámetro IN
			pTypeLower := strings.ToLower(p.Type)
			pNameLower := strings.ToLower(p.Name)
			isDateType := pTypeLower == "date"
			isDateName := strings.Contains(pNameLower, "fecha") || strings.Contains(pNameLower, "periodo")

			if isDateType || isDateName {
				if parsedTime, err := parseDateParam(p.Value); err == nil {
					args = append(args, parsedTime)
				} else {
					log.Printf("[PROCEDURE] Advertencia al parsear fecha en parámetro '%s': %v", p.Name, err)
					args = append(args, p.Value)
				}
			} else {
				args = append(args, p.Value)
			}
		}
	}

	// Formatear el nombre para manejar esquema.procedimiento correctamente
	procedureName := formatObjectName(req.Schema, req.Name)

	call := fmt.Sprintf("BEGIN %s(%s); END;", procedureName, strings.Join(placeholders, ", "))
	log.Printf("[PROCEDURE] SQL generado para procedimiento: %s", call)

	// Crear log
	startExec := time.Now()
	paramsJSON, _ := json.Marshal(req.Params)
	qlog := &QueryLog{
		ID:            generateID(),
		QueryType:     "PROCEDURE",
		QueryText:     req.Name,
		Params:        string(paramsJSON),
		ExecutionTime: startExec,
		UserIP:        r.RemoteAddr,
	}

	stmt, err := db.Prepare(call)
	if err != nil {
		errorMsg := err.Error()

		// Mejorar el mensaje de error
		if strings.Contains(errorMsg, "PLS-00201") {
			errorMsg = fmt.Sprintf("Procedimiento '%s' no encontrado. Verifica que existe en la base de datos.", req.Name)
		} else if strings.Contains(errorMsg, "PLS-00306") {
			errorMsg = fmt.Sprintf("Par├ímetros incorrectos para '%s'. Verifica tipos y cantidad de par├ímetros.", req.Name)
		}

		qlog.Success = false
		qlog.ErrorMsg = errorMsg
		qlog.Duration = time.Since(startExec).String()
		go saveQueryLog(qlog)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(args...); err != nil {
		errorMsg := err.Error()

		// Mejorar mensajes de error comunes
		if strings.Contains(errorMsg, "ORA-06502") {
			errorMsg = "Error de conversi├│n de tipos. Verifica que los tipos de datos sean correctos."
		} else if strings.Contains(errorMsg, "ORA-01403") {
			errorMsg = "No se encontraron datos. El procedimiento no retorn├│ resultados."
		}

		qlog.Success = false
		qlog.ErrorMsg = errorMsg
		qlog.Duration = time.Since(startExec).String()
		go saveQueryLog(qlog)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": errorMsg})
		return
	}

	out := make(map[string]interface{})
	for i, name := range outIndexes {
		if datePtr, ok := outDateMap[i]; ok && datePtr != nil {
			if datePtr.Valid {
				out[name] = formatDateOutput(datePtr.Time)
			} else {
				out[name] = nil
			}
			continue
		}
		if numPtr, ok := outNumMap[i]; ok && numPtr != nil {
			if numPtr.Valid {
				out[name] = numPtr.Float64
			} else {
				out[name] = nil
			}
			continue
		}
		if ptr, ok := outBuffers[i]; ok && ptr != nil {
			out[name] = *ptr
		}
	}

	qlog.Success = true
	qlog.RowsAffected = int64(len(out))
	qlog.Duration = time.Since(startExec).String()
	go saveQueryLog(qlog)

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "out": out})
}

// asyncProcedureHandler ejecuta un procedimiento de forma as├¡ncrona
func asyncProcedureHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solo se permite POST"})
		return
	}

	var req struct {
		Name          string `json:"name"`
		Schema        string `json:"schema,omitempty"` // Esquema del procedimiento/funci├│n
		IsFunction    bool   `json:"isFunction"`
		ExecutionMode string `json:"execution_mode,omitempty"`
		LockKey       string `json:"lock_key,omitempty"`
		Params        []struct {
			Name      string      `json:"name"`
			Value     interface{} `json:"value,omitempty"`
			Direction string      `json:"direction"`
			Type      string      `json:"type,omitempty"`
		} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inv├ílido"})
		return
	}

	// Validar parámetros requeridos antes de crear el job
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Falta el campo 'name'"})
		return
	}
	executionMode, err := normalizeExecutionMode(req.ExecutionMode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "execution_mode debe ser: parallel, sequential o exclusive"})
		return
	}
	lockKey := normalizeLockKey(req.LockKey, req.Name)

	var missingParams []string
	for _, p := range req.Params {
		if strings.ToUpper(p.Direction) != "OUT" {
			if p.Value == nil {
				missingParams = append(missingParams, p.Name)
			}
		}
	}
	if len(missingParams) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":          "Parámetros requeridos sin valor",
			"missing_params": missingParams,
		})
		return
	}

	// Preparar par├ímetros para guardar en el job
	paramsMap := make(map[string]interface{})
	paramsMap["name"] = req.Name
	paramsMap["isFunction"] = req.IsFunction
	paramsMap["execution_mode"] = executionMode
	paramsMap["lock_key"] = lockKey
	paramsArray := []map[string]interface{}{}
	for _, p := range req.Params {
		paramObj := map[string]interface{}{
			"name":      p.Name,
			"direction": p.Direction,
		}
		if p.Value != nil {
			paramObj["value"] = p.Value
		}
		if p.Type != "" {
			paramObj["type"] = p.Type
		}
		paramsArray = append(paramsArray, paramObj)
	}
	paramsMap["params"] = paramsArray

	// Crear el job con los par├ímetros
	job, err := jobManager.CreateJob(req.Name, paramsMap, executionMode, lockKey)
	if err != nil {
		if err == errExclusiveJobConflict {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   fmt.Sprintf("Ya existe un job activo para lock_key '%s'", lockKey),
				"code":    "JOB_ALREADY_RUNNING",
				"job_key": lockKey,
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "No se pudo crear el job asíncrono"})
		return
	}

	// Responder inmediatamente con el ID del job
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "accepted",
		"job_id":           job.ID,
		"execution_mode":   job.ExecutionMode,
		"lock_key":         job.LockKey,
		"message":          "Procedimiento ejecut├índose en segundo plano",
		"check_status_url": fmt.Sprintf("/jobs/%s", job.ID),
	})

	// Ejecutar el procedimiento en una goroutine
	go func() {
		// Capturar panics para evitar que el job quede colgado
		defer func() {
			if r := recover(); r != nil {
				endTime := time.Now()
				jobManager.UpdateJob(job.ID, func(j *AsyncJob) {
					j.Status = JobStatusFailed
					j.Error = fmt.Sprintf("Panic recuperado: %v", r)
					j.EndTime = &endTime
					j.Duration = endTime.Sub(j.StartTime).String()
					j.Progress = 100
				})
				log.Printf("ÔØî Panic en job %s: %v", job.ID, r)
			}
		}()

		for {
			started, err := jobManager.TryStartJob(job.ID, executionMode, lockKey)
			if err != nil {
				endTime := time.Now()
				jobManager.UpdateJob(job.ID, func(j *AsyncJob) {
					j.Status = JobStatusFailed
					j.Error = fmt.Sprintf("No se pudo iniciar el job: %v", err)
					j.EndTime = &endTime
					j.Duration = endTime.Sub(j.StartTime).String()
					j.Progress = 100
				})
				return
			}
			if started {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}

		// Preparar par├ímetros igual que en procedureHandler
		placeholders := []string{}
		args := []interface{}{}
		outIndexes := []string{}
		outBuffers := make(map[int]*string)
		outNumMap := make(map[int]*sql.NullFloat64)
		outDateMap := make(map[int]*sql.NullTime)

		if req.IsFunction {
			// Contar parámetros OUT para funciones
			outParamCount := 0
			for _, p := range req.Params {
				if strings.ToUpper(p.Direction) == "OUT" {
					outParamCount++
				}
			}

			// Preasignar buffers para el RETURN value + todos los OUT params
			preallocDates := make([]sql.NullTime, outParamCount+1)
			preallocNums := make([]sql.NullFloat64, outParamCount+1)
			preallocStrs := make([]string, outParamCount+1)
			for j := 0; j < outParamCount+1; j++ {
				preallocStrs[j] = strings.Repeat(" ", 4000)
			}

			// :1 es siempre el RETURN value de la función
			placeholders = append(placeholders, ":1")

			// El retorno de la función va en índice 0
			// Asumir que es NUMBER a menos que se especifique otro tipo
			args = append(args, sql.Out{Dest: &preallocNums[0], In: false})
			outNumMap[0] = &preallocNums[0]
			outIndexes = append(outIndexes, "return_value")

			// Procesar parámetros en orden (IN y OUT)
			paramPos := 2
			outIdx := 1 // Índice para OUT params (después del RETURN)
			for _, p := range req.Params {
				placeholders = append(placeholders, fmt.Sprintf(":%d", paramPos))

				if strings.ToUpper(p.Direction) == "OUT" {
					lowerName := strings.ToLower(p.Name)
					isDate := strings.ToLower(p.Type) == "date"
					isNum := strings.ToLower(p.Type) == "number" ||
						strings.Contains(lowerName, "resultado") || strings.Contains(lowerName, "result") ||
						strings.Contains(lowerName, "total") || strings.Contains(lowerName, "count") ||
						strings.Contains(lowerName, "suma") || strings.Contains(lowerName, "num") ||
						strings.Contains(lowerName, "int") || strings.Contains(lowerName, "id")

					outIndexes = append(outIndexes, p.Name)

					if isDate {
						args = append(args, sql.Out{Dest: &preallocDates[outIdx], In: false})
						outDateMap[outIdx] = &preallocDates[outIdx]
					} else if isNum {
						args = append(args, sql.Out{Dest: &preallocNums[outIdx], In: false})
						outNumMap[outIdx] = &preallocNums[outIdx]
					} else {
						args = append(args, sql.Out{Dest: &preallocStrs[outIdx], In: false})
						outBuffers[outIdx] = &preallocStrs[outIdx]
					}
					outIdx++
				} else {
					// Parámetro IN
					pTypeLower := strings.ToLower(p.Type)
					pNameLower := strings.ToLower(p.Name)
					isDateType := pTypeLower == "date"
					isDateName := strings.Contains(pNameLower, "fecha") || strings.Contains(pNameLower, "periodo")

					if isDateType || isDateName {
						if parsedTime, err := parseDateParam(p.Value); err == nil {
							args = append(args, parsedTime)
						} else {
							log.Printf("[ASYNC_PROCEDURE] Advertencia al parsear fecha en parámetro '%s': %v", p.Name, err)
							args = append(args, p.Value)
						}
					} else {
						args = append(args, p.Value)
					}
				}
				paramPos++
			}
		} else {
			// Procedimiento normal (no función)
			for i, p := range req.Params {
				placeholders = append(placeholders, fmt.Sprintf(":%d", i+1))

				if strings.ToUpper(p.Direction) == "OUT" {
					outIdx := len(outIndexes)
					lowerName := strings.ToLower(p.Name)
					isDate := strings.ToLower(p.Type) == "date"
					isNum := strings.ToLower(p.Type) == "number" ||
						strings.Contains(lowerName, "resultado") || strings.Contains(lowerName, "result") ||
						strings.Contains(lowerName, "total") || strings.Contains(lowerName, "count") ||
						strings.Contains(lowerName, "suma") || strings.Contains(lowerName, "num") ||
						strings.Contains(lowerName, "int") || strings.Contains(lowerName, "id")

					outIndexes = append(outIndexes, p.Name)

					if isDate {
						datePtr := new(sql.NullTime)
						outDateMap[outIdx] = datePtr
						args = append(args, sql.Out{Dest: datePtr, In: false})
					} else if isNum {
						numPtr := new(sql.NullFloat64)
						outNumMap[outIdx] = numPtr
						args = append(args, sql.Out{Dest: numPtr, In: false})
					} else {
						ptr := new(string)
						*ptr = strings.Repeat(" ", 4000)
						outBuffers[outIdx] = ptr
						args = append(args, sql.Out{Dest: ptr, In: false})
					}
				} else {
					// Parámetro IN
					pTypeLower := strings.ToLower(p.Type)
					pNameLower := strings.ToLower(p.Name)
					isDateType := pTypeLower == "date"
					isDateName := strings.Contains(pNameLower, "fecha") || strings.Contains(pNameLower, "periodo")

					if isDateType || isDateName {
						if parsedTime, err := parseDateParam(p.Value); err == nil {
							args = append(args, parsedTime)
						} else {
							log.Printf("[ASYNC_PROCEDURE] Advertencia al parsear fecha en parámetro '%s': %v", p.Name, err)
							args = append(args, p.Value)
						}
					} else {
						args = append(args, p.Value)
					}
				}
			}
		}

		// Formatear el nombre para manejar esquema.procedimiento correctamente
		procedureName := formatObjectName(req.Schema, req.Name)

		// Construir la llamada diferenciando entre función y procedimiento
		var call string
		if req.IsFunction {
			// Para funciones: BEGIN :1 := function_name(params); END;
			call = fmt.Sprintf("BEGIN :1 := %s(%s); END;", procedureName, strings.Join(placeholders[1:], ", "))
		} else {
			// Para procedimientos: BEGIN procedure_name(params); END;
			call = fmt.Sprintf("BEGIN %s(%s); END;", procedureName, strings.Join(placeholders, ", "))
		}

		stmt, err := db.Prepare(call)
		if err != nil {
			endTime := time.Now()
			errorMsg := err.Error()

			// Mejorar el mensaje de error para procedimientos no encontrados
			if strings.Contains(errorMsg, "PLS-00201") {
				errorMsg = fmt.Sprintf("Procedimiento '%s' no encontrado. Verifica que existe en la base de datos.", req.Name)
			} else if strings.Contains(errorMsg, "PLS-00306") {
				errorMsg = fmt.Sprintf("Par├ímetros incorrectos para '%s'. Verifica tipos y cantidad de par├ímetros.", req.Name)
			}

			jobManager.UpdateJob(job.ID, func(j *AsyncJob) {
				j.Status = JobStatusFailed
				j.Error = errorMsg
				j.EndTime = &endTime
				j.Duration = endTime.Sub(j.StartTime).String()
				j.Progress = 100
			})
			return
		}
		defer stmt.Close()

		jobManager.UpdateJob(job.ID, func(j *AsyncJob) {
			j.Progress = 50
		})

		if _, err := stmt.Exec(args...); err != nil {
			endTime := time.Now()
			errorMsg := err.Error()

			// Mejorar mensajes de error comunes
			if strings.Contains(errorMsg, "ORA-06502") {
				errorMsg = "Error de conversi├│n de tipos. Verifica que los tipos de datos sean correctos."
			} else if strings.Contains(errorMsg, "ORA-01403") {
				errorMsg = "No se encontraron datos. El procedimiento no retorn├│ resultados."
			}

			jobManager.UpdateJob(job.ID, func(j *AsyncJob) {
				j.Status = JobStatusFailed
				j.Error = errorMsg
				j.EndTime = &endTime
				j.Duration = endTime.Sub(j.StartTime).String()
				j.Progress = 100
			})
			return
		}

		jobManager.UpdateJob(job.ID, func(j *AsyncJob) {
			j.Progress = 80
		})

		// Recopilar resultados OUT
		out := make(map[string]interface{})
		for i, name := range outIndexes {
			if datePtr, ok := outDateMap[i]; ok && datePtr != nil {
				if datePtr.Valid {
					out[name] = formatDateOutput(datePtr.Time)
				} else {
					out[name] = nil
				}
				continue
			}
			if numPtr, ok := outNumMap[i]; ok && numPtr != nil {
				if numPtr.Valid {
					out[name] = numPtr.Float64
				} else {
					out[name] = nil
				}
				continue
			}
			if ptr, ok := outBuffers[i]; ok && ptr != nil {
				out[name] = *ptr
			}
		}

		// Completado exitosamente
		endTime := time.Now()
		jobManager.UpdateJob(job.ID, func(j *AsyncJob) {
			j.Status = JobStatusCompleted
			j.Result = out
			j.EndTime = &endTime
			j.Duration = endTime.Sub(j.StartTime).String()
			j.Progress = 100
		})
	}()
}

// jobsHandler maneja consultas de estado de jobs
func jobsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/jobs")
	path = strings.TrimPrefix(path, "/")

	// Si hay un ID, buscar/eliminar ese job espec├¡fico
	if path != "" {
		if r.Method == http.MethodDelete {
			err := jobManager.DeleteJob(path)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Job eliminado correctamente",
				"job_id":  path,
			})
			return
		}

		// GET de job espec├¡fico
		job, exists := jobManager.GetJob(path)
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Job no encontrado"})
			return
		}
		json.NewEncoder(w).Encode(job)
		return
	}

	// Sin ID - Listar o eliminar m├║ltiples jobs
	if r.Method == http.MethodDelete {
		// DELETE /jobs?status=completed,failed&older_than=7
		queryParams := r.URL.Query()
		statusParam := queryParams.Get("status")
		olderThanParam := queryParams.Get("older_than")

		var statusFilter []string
		if statusParam != "" {
			statusFilter = strings.Split(statusParam, ",")
		}

		olderThan := 0
		if olderThanParam != "" {
			if days, err := strconv.Atoi(olderThanParam); err == nil {
				olderThan = days
			}
		}

		// Validar que al menos un filtro est├® presente
		if len(statusFilter) == 0 && olderThan == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Debes especificar al menos un filtro: ?status=completed,failed o ?older_than=7",
			})
			return
		}

		count, err := jobManager.DeleteJobs(statusFilter, olderThan)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Jobs eliminados correctamente",
			"deleted": count,
		})
		return
	}

	// GET - Listar todos los jobs
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solo se permite GET o DELETE"})
		return
	}

	jobs := jobManager.GetAllJobs()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total": len(jobs),
		"jobs":  jobs,
	})
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// healthHandler proporciona información detallada del estado del sistema
// No requiere autenticación para permitir monitoreo externo
func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Verificar conexión a Oracle
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		health["status"] = "error"
		health["oracle_connection"] = "failed"
		health["error"] = err.Error()
		json.NewEncoder(w).Encode(health)
		return
	}
	health["oracle_connection"] = "ok"

	// Obtener versión de Oracle
	var version string
	if err := db.QueryRow("SELECT banner FROM v$version WHERE ROWNUM = 1").Scan(&version); err == nil {
		health["database_version"] = version
	}

	// Estadísticas de jobs
	jobManager.mu.RLock()
	pendingJobs := 0
	runningJobs := 0
	for _, job := range jobManager.jobs {
		if job.Status == JobStatusPending {
			pendingJobs++
		} else if job.Status == JobStatusRunning {
			runningJobs++
		}
	}
	jobManager.mu.RUnlock()

	health["async_jobs"] = map[string]interface{}{
		"pending": pendingJobs,
		"running": runningJobs,
		"total":   len(jobManager.jobs),
	}

	// Información del sistema
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	health["system"] = map[string]interface{}{
		"go_version":   runtime.Version(),
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solo se permite POST"})
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inv├ílido"})
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Falta el campo 'query'"})
		return
	}

	// Normalizar saltos de l├¡nea: reemplazar \r\n y \n por salto de l├¡nea real
	normalizedQuery := strings.ReplaceAll(req.Query, "\r\n", "\n")
	normalizedQuery = strings.ReplaceAll(normalizedQuery, "\\n", "\n")

	// Detectar si es una consulta de log (evitar recursi├│n)
	upperQuery := strings.ToUpper(normalizedQuery)
	isLogQuery := strings.Contains(upperQuery, "FROM QUERY_LOG") ||
		strings.Contains(upperQuery, "FROM ASYNC_JOBS") ||
		strings.Contains(upperQuery, "USER_TABLES")

	// Crear log solo si no es una consulta de log
	var qlog *QueryLog
	var startExec time.Time
	if !isLogQuery {
		startExec = time.Now()
		qlog = &QueryLog{
			ID:            generateID(),
			QueryType:     "QUERY",
			QueryText:     normalizedQuery,
			ExecutionTime: startExec,
			UserIP:        r.RemoteAddr,
		}
	}

	log.Printf("[QUERY] Ejecutando: %s", normalizedQuery)
	rows, err := db.Query(normalizedQuery)
	if err != nil {
		if qlog != nil {
			qlog.Success = false
			qlog.ErrorMsg = err.Error()
			qlog.Duration = time.Since(startExec).String()
			go saveQueryLog(qlog)
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		if qlog != nil {
			qlog.Success = false
			qlog.ErrorMsg = err.Error()
			qlog.Duration = time.Since(startExec).String()
			go saveQueryLog(qlog)
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	results := []map[string]interface{}{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			if qlog != nil {
				qlog.Success = false
				qlog.ErrorMsg = err.Error()
				qlog.Duration = time.Since(startExec).String()
				go saveQueryLog(qlog)
			}

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columns[i]
			b, ok := val.([]byte)
			if ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = val
			}
		}
		results = append(results, rowMap)
	}

	// Registro exitoso
	if qlog != nil {
		qlog.Success = true
		qlog.RowsAffected = int64(len(results))
		qlog.Duration = time.Since(startExec).String()
		go saveQueryLog(qlog)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
}

func ipAllowed(remoteIP string, allowedIPs []string) bool {
	parsedRemote := net.ParseIP(remoteIP)
	if parsedRemote == nil {
		return false
	}
	for _, ip := range allowedIPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if ip == "localhost" && (remoteIP == "127.0.0.1" || remoteIP == "::1") {
			return true
		}
		if strings.Contains(ip, "/") {
			// Rango CIDR
			_, cidrNet, err := net.ParseCIDR(ip)
			if err == nil && cidrNet.Contains(parsedRemote) {
				return true
			}
		} else {
			// IP exacta
			if ip == remoteIP {
				return true
			}
		}
	}
	return false
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(&w, r)

		// Permitir peticiones OPTIONS (preflight) sin autenticaci├│n
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if os.Getenv("API_NO_AUTH") == "1" {
			next(w, r)
			return
		}
		token := os.Getenv("API_TOKEN")
		authHeader := r.Header.Get("Authorization")
		expected := "Bearer " + token
		if token == "" || authHeader != expected {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "No autorizado"})
			return
		}
		allowedIPs := os.Getenv("API_ALLOWED_IPS")
		if allowedIPs != "" {
			ipList := strings.Split(allowedIPs, ",")
			remoteIP := r.RemoteAddr
			if colon := strings.LastIndex(remoteIP, ":"); colon != -1 {
				remoteIP = remoteIP[:colon]
			}
			remoteIP = strings.Trim(remoteIP, "[]")
			log.Printf("Debug IP: remoteIP=%s, allowedIPs=%v", remoteIP, ipList)
			if !ipAllowed(remoteIP, ipList) {
				log.Printf("IP rechazada: %s", remoteIP)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "IP no permitida", "ip": remoteIP})
				return
			}
		}
		next(w, r)
	}
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w, r)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Solo se permite POST"})
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inv├ílido"})
		return
	}
	if req.Query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Falta el campo 'query'"})
		return
	}

	log.Printf("[EXEC] Ejecutando: %s", req.Query)

	// Crear log
	startExec := time.Now()
	qlog := &QueryLog{
		ID:            generateID(),
		QueryType:     "EXEC",
		QueryText:     req.Query,
		ExecutionTime: startExec,
		UserIP:        r.RemoteAddr,
	}

	// Detectar si es un comando de modificaci├│n
	q := req.Query
	qType := ""
	if len(q) > 0 {
		for i := 0; i < len(q) && (q[i] == ' ' || q[i] == '\t' || q[i] == '\n'); i++ {
			q = q[1:]
		}
		if len(q) >= 6 && (q[:6] == "INSERT" || q[:6] == "insert") {
			qType = "exec"
		} else if len(q) >= 6 && (q[:6] == "UPDATE" || q[:6] == "update") {
			qType = "exec"
		} else if len(q) >= 6 && (q[:6] == "DELETE" || q[:6] == "delete") {
			qType = "exec"
		}
	}

	if qType == "exec" {
		res, err := db.Exec(req.Query)
		if err != nil {
			qlog.Success = false
			qlog.ErrorMsg = err.Error()
			qlog.Duration = time.Since(startExec).String()
			go saveQueryLog(qlog)

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			log.Printf("ÔÜá´©Å  No se pudo obtener rows affected: %v", err)
			rowsAffected = 0
		}

		qlog.Success = true
		qlog.RowsAffected = rowsAffected
		qlog.Duration = time.Since(startExec).String()
		go saveQueryLog(qlog)

		json.NewEncoder(w).Encode(map[string]interface{}{"rows_affected": rowsAffected})
		return
	}

	rows, err := db.Query(req.Query)
	if err != nil {
		qlog.Success = false
		qlog.ErrorMsg = err.Error()
		qlog.Duration = time.Since(startExec).String()
		go saveQueryLog(qlog)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		qlog.Success = false
		qlog.ErrorMsg = err.Error()
		qlog.Duration = time.Since(startExec).String()
		go saveQueryLog(qlog)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	results := []map[string]interface{}{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			qlog.Success = false
			qlog.ErrorMsg = err.Error()
			qlog.Duration = time.Since(startExec).String()
			go saveQueryLog(qlog)

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			rowMap[col] = v
		}
		results = append(results, rowMap)
	}

	qlog.Success = true
	qlog.RowsAffected = int64(len(results))
	qlog.Duration = time.Since(startExec).String()
	go saveQueryLog(qlog)

	json.NewEncoder(w).Encode(results)
}

// enableCORS agrega los headers necesarios para CORS
func enableCORS(w *http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	(*w).Header().Set("Access-Control-Allow-Origin", origin)
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Max-Age", "3600")
}
