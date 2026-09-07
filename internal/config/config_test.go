package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewAppConfig(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr error
	}{
		{"valid test environment", "test", nil},
		{"valid development environment", "development", nil},
		{"valid production environment", "production", nil},
		{"empty environment", "", ErrEnvironmentNotProvided},
		{"invalid environment", "staging", ErrInvalidEnvironment},
		{"invalid environment local", "local", ErrInvalidEnvironment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAppConfig(tt.env)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil AppConfig")
			}
			if got.Environment != tt.env {
				t.Errorf("expected environment %q, got %q", tt.env, got.Environment)
			}
		})
	}
}

func TestNewDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		dbURL   string
		dbName  string
		wantErr error
	}{
		{"valid config", "http://localhost:5432", "testdb", nil},
		{"valid postgres url", "postgres://user:pass@localhost:5432/db", "mydb", nil},
		{"empty dbURL", "", "testdb", ErrEmptyString},
		{"empty dbName", "http://localhost:5432", "", ErrEmptyString},
		{"both empty", "", "", ErrEmptyString},
		{"invalid URL", "://invalid", "testdb", ErrInvalidConnString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDatabaseConfig(tt.dbURL, tt.dbName)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !isExpectedErr(err, tt.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil DatabaseConfig")
			}
			if got.URL != tt.dbURL {
				t.Errorf("expected URL %q, got %q", tt.dbURL, got.URL)
			}
			if got.DBName != tt.dbName {
				t.Errorf("expected DBName %q, got %q", tt.dbName, got.DBName)
			}
		})
	}
}

func TestGetStringEnvVar(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     *string
		want         string
		wantErr      error
	}{
		{
			name:         "empty key returns error",
			key:          "",
			defaultValue: "default",
			wantErr:      ErrEmptyString,
		},
		{
			name:         "env not set returns default",
			key:          "TEST_STRING_VAR_NOT_SET",
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "env not set returns empty default",
			key:          "TEST_STRING_VAR_NOT_SET_2",
			defaultValue: "",
			want:         "",
		},
		{
			name:         "env set returns env value",
			key:          "TEST_STRING_VAR_SET",
			defaultValue: "fallback",
			envValue:     strPtr("from-env"),
			want:         "from-env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != nil {
				t.Setenv(tt.key, *tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			got, err := getStringEnvVar(tt.key, tt.defaultValue)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !isExpectedErr(err, tt.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestGetIntEnvVar(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		envVal  string
		want    int
		wantErr bool
	}{
		{"empty key returns error", "", "", 0, true},
		{"non-numeric value returns error", "TEST_INT_VAR", "abc", 0, true},
		{"empty string returns error", "TEST_INT_VAR_EMPTY", "", 0, true},
		{"valid int", "TEST_INT_VAR_VALID", "42", 42, false},
		{"negative int", "TEST_INT_VAR_NEG", "-10", -10, false},
		{"zero", "TEST_INT_VAR_ZERO", "0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != "" {
				t.Setenv(tt.key, tt.envVal)
			}

			got, err := getIntEnvVar(tt.key)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	cleanEnv(t, "ENV", "POSTGRES_URL", "POSTGRES_DB")

	t.Run("invalid path returns error", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/.env")
		if err == nil {
			t.Fatal("expected error for nonexistent path")
		}
		if !errors.Is(err, ErrLoadingConfig) {
			t.Fatalf("expected error wrapping ErrLoadingConfig, got %v", err)
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		cleanEnv(t, "ENV", "POSTGRES_URL", "POSTGRES_DB")
		content := "ENV=development\nPOSTGRES_URL=http://localhost:5432\nPOSTGRES_DB=mydb\n"
		path := writeTempEnv(t, content)

		got, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.AppConfig.Environment != "development" {
			t.Errorf("expected environment %q, got %q", "development", got.AppConfig.Environment)
		}
		if got.DBConfig.URL != "http://localhost:5432" {
			t.Errorf("expected URL %q, got %q", "http://localhost:5432", got.DBConfig.URL)
		}
		if got.DBConfig.DBName != "mydb" {
			t.Errorf("expected DBName %q, got %q", "mydb", got.DBConfig.DBName)
		}
	})

	t.Run("missing ENV uses empty string causing AppConfig error", func(t *testing.T) {
		cleanEnv(t, "ENV", "POSTGRES_URL", "POSTGRES_DB")
		content := "POSTGRES_URL=http://localhost:5432\nPOSTGRES_DB=test\n"
		path := writeTempEnv(t, content)

		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("expected error when ENV is missing")
		}
		if !errors.Is(err, ErrEnvironmentNotProvided) {
			t.Fatalf("expected ErrEnvironmentNotProvided, got %v", err)
		}
	})

	t.Run("invalid ENV value causes AppConfig error", func(t *testing.T) {
		cleanEnv(t, "ENV", "POSTGRES_URL", "POSTGRES_DB")
		content := "ENV=staging\nPOSTGRES_URL=http://localhost:5432\nPOSTGRES_DB=test\n"
		path := writeTempEnv(t, content)

		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("expected error for invalid ENV")
		}
		if !errors.Is(err, ErrInvalidEnvironment) {
			t.Fatalf("expected ErrInvalidEnvironment, got %v", err)
		}
	})

	t.Run("missing DBNAME uses default", func(t *testing.T) {
		cleanEnv(t, "ENV", "POSTGRES_URL", "POSTGRES_DB")
		content := "ENV=test\nPOSTGRES_URL=http://localhost:5432\n"
		path := writeTempEnv(t, content)

		got, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DBConfig.DBName != "test" {
			t.Errorf("expected default DBName %q, got %q", "test", got.DBConfig.DBName)
		}
	})

	t.Run("missing DBURL uses default", func(t *testing.T) {
		cleanEnv(t, "ENV", "POSTGRES_URL", "POSTGRES_DB")
		content := "ENV=test\nPOSTGRES_DB=mydb\n"
		path := writeTempEnv(t, content)

		got, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DBConfig.URL != "http://localhost:5432" {
			t.Errorf("expected default URL %q, got %q", "http://localhost:5432", got.DBConfig.URL)
		}
	})

	t.Run("all defaults", func(t *testing.T) {
		cleanEnv(t, "ENV", "POSTGRES_URL", "POSTGRES_DB")
		content := "ENV=production\n"
		path := writeTempEnv(t, content)

		got, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DBConfig.URL != "http://localhost:5432" {
			t.Errorf("expected default URL, got %q", got.DBConfig.URL)
		}
		if got.DBConfig.DBName != "test" {
			t.Errorf("expected default DBName, got %q", got.DBConfig.DBName)
		}
	})
}

func cleanEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp .env: %v", err)
	}
	return path
}

func strPtr(s string) *string { return &s }

func isExpectedErr(err, target error) bool {
	for unwrapped := err; unwrapped != nil; {
		if unwrapped == target {
			return true
		}
		u, ok := unwrapped.(interface{ Unwrap() error })
		if !ok {
			break
		}
		unwrapped = u.Unwrap()
	}
	return false
}
