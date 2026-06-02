package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeExecutionMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ExecutionMode
		wantErr bool
	}{
		{name: "default parallel when empty", input: "", want: ExecutionModeParallel},
		{name: "parallel", input: "parallel", want: ExecutionModeParallel},
		{name: "sequential uppercase", input: "SEQUENTIAL", want: ExecutionModeSequential},
		{name: "exclusive with spaces", input: " exclusive ", want: ExecutionModeExclusive},
		{name: "invalid mode", input: "batch", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeExecutionMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeLockKey(t *testing.T) {
	if got := normalizeLockKey("", "PROC_X"); got != "PROC_X" {
		t.Fatalf("expected default lock key to use procedure name, got %q", got)
	}
	if got := normalizeLockKey("  CUSTOMER:1  ", "PROC_X"); got != "CUSTOMER:1" {
		t.Fatalf("expected trimmed lock key, got %q", got)
	}
}

func TestLoadConfigUsesYAMLAsFallback(t *testing.T) {
	withCleanConfigEnv(t)
	withWorkingDir(t, t.TempDir())
	writeFile(t, "config.yaml", `oracle:
  user: yaml_user
  password: yaml_password
  host: yaml-host
  port: 1521
  service: yaml_service
api:
  token: yaml_token
  allowed_ips:
    - 127.0.0.1
    - localhost
  no_auth: true
server:
  port: 8088
`)
	withArgs(t, []string{"cmd"})

	cfg := loadConfig()

	if cfg.OracleUser != "yaml_user" {
		t.Fatalf("expected Oracle user from YAML, got %q", cfg.OracleUser)
	}
	if cfg.OraclePassword != "yaml_password" {
		t.Fatalf("expected Oracle password from YAML, got %q", cfg.OraclePassword)
	}
	if cfg.OracleHost != "yaml-host" {
		t.Fatalf("expected Oracle host from YAML, got %q", cfg.OracleHost)
	}
	if cfg.OraclePort != "1521" {
		t.Fatalf("expected Oracle port from YAML, got %q", cfg.OraclePort)
	}
	if cfg.OracleService != "yaml_service" {
		t.Fatalf("expected Oracle service from YAML, got %q", cfg.OracleService)
	}
	if cfg.ListenPort != "8088" {
		t.Fatalf("expected listen port from YAML, got %q", cfg.ListenPort)
	}
	if got := os.Getenv("API_TOKEN"); got != "yaml_token" {
		t.Fatalf("expected API token from YAML, got %q", got)
	}
	if got := os.Getenv("API_ALLOWED_IPS"); got != "127.0.0.1,localhost" {
		t.Fatalf("expected allowed IPs CSV from YAML, got %q", got)
	}
	if got := os.Getenv("API_NO_AUTH"); got != "1" {
		t.Fatalf("expected API_NO_AUTH=1 from YAML, got %q", got)
	}
}

func TestLoadConfigPrecedenceProcessEnvDotEnvYAML(t *testing.T) {
	withCleanConfigEnv(t)
	withWorkingDir(t, t.TempDir())
	writeFile(t, ".env", `ORACLE_USER=dotenv_user
ORACLE_PASSWORD=dotenv_password
ORACLE_HOST=dotenv-host
ORACLE_PORT=1522
API_TOKEN=dotenv_token
API_ALLOWED_IPS=10.0.0.1
`)
	writeFile(t, "config.yaml", `oracle:
  user: yaml_user
  password: yaml_password
  host: yaml-host
  port: 1521
  service: yaml_service
api:
  token: yaml_token
  allowed_ips:
    - 127.0.0.1
  no_auth: false
server:
  port: 8088
`)
	withArgs(t, []string{"cmd"})
	if err := os.Setenv("ORACLE_USER", "process_user"); err != nil {
		t.Fatalf("set ORACLE_USER: %v", err)
	}

	cfg := loadConfig()

	if cfg.OracleUser != "process_user" {
		t.Fatalf("expected process env to win for ORACLE_USER, got %q", cfg.OracleUser)
	}
	if cfg.OraclePassword != "dotenv_password" {
		t.Fatalf("expected .env to win for ORACLE_PASSWORD, got %q", cfg.OraclePassword)
	}
	if cfg.OracleHost != "dotenv-host" {
		t.Fatalf("expected .env to win for ORACLE_HOST, got %q", cfg.OracleHost)
	}
	if cfg.OraclePort != "1522" {
		t.Fatalf("expected .env to win for ORACLE_PORT, got %q", cfg.OraclePort)
	}
	if cfg.OracleService != "yaml_service" {
		t.Fatalf("expected YAML fallback for ORACLE_SERVICE, got %q", cfg.OracleService)
	}
	if cfg.ListenPort != "8088" {
		t.Fatalf("expected YAML fallback for PORT, got %q", cfg.ListenPort)
	}
	if got := os.Getenv("API_TOKEN"); got != "dotenv_token" {
		t.Fatalf("expected .env to win for API_TOKEN, got %q", got)
	}
	if got := os.Getenv("API_ALLOWED_IPS"); got != "10.0.0.1" {
		t.Fatalf("expected .env to win for API_ALLOWED_IPS, got %q", got)
	}
	if got := os.Getenv("API_NO_AUTH"); got != "0" {
		t.Fatalf("expected YAML fallback for API_NO_AUTH, got %q", got)
	}
}

func TestLoadConfigKeepsCLIListenPortFallback(t *testing.T) {
	withCleanConfigEnv(t)
	withWorkingDir(t, t.TempDir())
	writeFile(t, "config.yaml", `oracle:
  user: yaml_user
  password: yaml_password
  host: yaml-host
  port: 1521
  service: yaml_service
`)
	withArgs(t, []string{"cmd", "custom.env", "9091"})

	cfg := loadConfig()

	if cfg.ListenPort != "9091" {
		t.Fatalf("expected CLI listen port fallback, got %q", cfg.ListenPort)
	}
}

func withCleanConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"ENV_FILE",
		"ORACLE_USER",
		"ORACLE_PASSWORD",
		"ORACLE_HOST",
		"ORACLE_PORT",
		"ORACLE_SERVICE",
		"API_TOKEN",
		"API_ALLOWED_IPS",
		"API_NO_AUTH",
		"PORT",
	}

	previous := make(map[string]*string, len(keys))
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			valueCopy := value
			previous[key] = &valueCopy
		} else {
			previous[key] = nil
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for _, key := range keys {
			if previous[key] == nil {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, *previous[key])
		}
	})
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(previous)
	})
}

func withArgs(t *testing.T, args []string) {
	t.Helper()

	previous := os.Args
	os.Args = args

	t.Cleanup(func() {
		os.Args = previous
	})
}

func writeFile(t *testing.T, name, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(".", name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
