package main

import "testing"

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
