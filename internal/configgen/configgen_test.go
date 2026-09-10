package configgen

import (
	"errors"
	"strings"
	"testing"

	"secrets-vault/internal/secrets"
)

func TestGenerateTypeScript(t *testing.T) {
	items := []secrets.Secret{
		{Key: "db.host", Value: "localhost"},
		{Key: "db.port", Value: "5432"},
		{Key: "db.password", Value: "s3cret", IsSecret: true},
		{Key: "log.level", Value: "debug"},
		{Key: "featureX.enabled", Value: "true"},
	}

	want := `export interface MyServiceConfig {
  db: {
    host: string;
    password: string;
    port: number;
  };
  featureX: {
    enabled: boolean;
  };
  log: {
    level: string;
  };
}
`

	result, err := Generate("my-service", items, LangTS)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if result.Code != want {
		t.Errorf("code mismatch:\n--- got ---\n%s\n--- want ---\n%s", result.Code, want)
	}
	if result.Filename != "my-service.config.ts" {
		t.Errorf("filename = %q, want %q", result.Filename, "my-service.config.ts")
	}
	if result.Language != LangTS {
		t.Errorf("language = %q, want %q", result.Language, LangTS)
	}
}

func TestGenerateGo(t *testing.T) {
	items := []secrets.Secret{
		{Key: "db.host", Value: "localhost"},
		{Key: "db.port", Value: "5432"},
		{Key: "db.password", Value: "s3cret", IsSecret: true},
		{Key: "log.level", Value: "debug"},
		{Key: "featureX.enabled", Value: "true"},
	}

	want := `type MyServiceConfig struct {
	Db struct {
		Host     string ` + "`json:\"host\"`" + `
		Password string ` + "`json:\"password\"`" + `
		Port     int    ` + "`json:\"port\"`" + `
	} ` + "`json:\"db\"`" + `
	FeatureX struct {
		Enabled bool ` + "`json:\"enabled\"`" + `
	} ` + "`json:\"featureX\"`" + `
	Log struct {
		Level string ` + "`json:\"level\"`" + `
	} ` + "`json:\"log\"`" + `
}
`

	result, err := Generate("my-service", items, LangGo)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if result.Code != want {
		t.Errorf("code mismatch:\n--- got ---\n%s\n--- want ---\n%s", result.Code, want)
	}
	if result.Filename != "my_service_config.go" {
		t.Errorf("filename = %q, want %q", result.Filename, "my_service_config.go")
	}
}

func TestGenerateInferTypes(t *testing.T) {
	items := []secrets.Secret{
		{Key: "enabled", Value: "false"},
		{Key: "ratio", Value: "1.5"},
		{Key: "retries", Value: "-3"},
		{Key: "name", Value: "vault"},
	}

	result, err := Generate("provider", items, LangTS)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	for _, want := range []string{
		"enabled: boolean;",
		"ratio: number;",
		"retries: number;",
		"name: string;",
	} {
		if !strings.Contains(result.Code, want) {
			t.Errorf("code missing %q:\n%s", want, result.Code)
		}
	}
}

func TestGenerateSanitizesIdentifiers(t *testing.T) {
	items := []secrets.Secret{
		{Key: "api-key", Value: "abc"},
		{Key: "db.read_only", Value: "true"},
		{Key: "2fa.enabled", Value: "true"},
	}

	ts, err := Generate("my-app", items, LangTS)
	if err != nil {
		t.Fatalf("generate ts: %v", err)
	}
	for _, want := range []string{"apiKey: string;", "readOnly: boolean;", "_2fa: {"} {
		if !strings.Contains(ts.Code, want) {
			t.Errorf("ts code missing %q:\n%s", want, ts.Code)
		}
	}

	goCode, err := Generate("my-app", items, LangGo)
	if err != nil {
		t.Fatalf("generate go: %v", err)
	}
	for _, want := range []string{"ApiKey string", "ReadOnly bool", "X2fa struct"} {
		if !strings.Contains(goCode.Code, want) {
			t.Errorf("go code missing %q:\n%s", want, goCode.Code)
		}
	}
}

func TestGenerateCollisionPrefersNested(t *testing.T) {
	items := []secrets.Secret{
		{Key: "db", Value: "scalar"},
		{Key: "db.host", Value: "localhost"},
	}

	result, err := Generate("svc", items, LangTS)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if strings.Contains(result.Code, "db: string;") {
		t.Errorf("scalar leaf should be dropped when nested children exist:\n%s", result.Code)
	}
	if !strings.Contains(result.Code, "host: string;") {
		t.Errorf("nested child missing:\n%s", result.Code)
	}
}

func TestGenerateEmpty(t *testing.T) {
	result, err := Generate("empty", nil, LangTS)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	want := `export interface EmptyConfig {
}
`
	if result.Code != want {
		t.Errorf("code mismatch:\n--- got ---\n%s\n--- want ---\n%s", result.Code, want)
	}
}

func TestGenerateFallbackTypeName(t *testing.T) {
	result, err := Generate("", nil, LangGo)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.Contains(result.Code, "type ProviderConfig struct") {
		t.Errorf("expected fallback type name:\n%s", result.Code)
	}
	if result.Filename != "provider_config.go" {
		t.Errorf("filename = %q, want %q", result.Filename, "provider_config.go")
	}
}

func TestGenerateUnsupportedLanguage(t *testing.T) {
	_, err := Generate("svc", nil, Language("rust"))
	if !errors.Is(err, ErrUnsupportedLanguage) {
		t.Fatalf("err = %v, want ErrUnsupportedLanguage", err)
	}
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		raw     string
		want    Language
		wantErr bool
	}{
		{raw: "", want: LangTS},
		{raw: "ts", want: LangTS},
		{raw: "TS", want: LangTS},
		{raw: " go ", want: LangGo},
		{raw: "python", wantErr: true},
	}

	for _, tt := range tests {
		got, err := ParseLanguage(tt.raw)
		if tt.wantErr {
			if !errors.Is(err, ErrUnsupportedLanguage) {
				t.Errorf("ParseLanguage(%q) err = %v, want ErrUnsupportedLanguage", tt.raw, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseLanguage(%q) unexpected err: %v", tt.raw, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseLanguage(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}
