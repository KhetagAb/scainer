package configs_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"scainer/internal/configs"
)

func TestLoadConfig_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`
http:
  addr: ":9999"
  shutdown_timeout: 5s
store:
  dir: "./data"
mongodb:
  username: "from-yaml"
  password: "yaml-pass"
  host: "yaml-host"
  database: "scainer"
admin:
  jwt_secret: "yaml-secret"
  jwt_ttl: 1h
ejudge:
  base_url: "https://ejudge.example"
  timeout: 10s
jplag:
  jar_path: "bin/jplag.jar"
analyze:
  jobs_max_concurrent: 2
  analyze_concurrency: 3
aiusage:
  enabled: false
openai:
  base_url: "https://example/v1"
  model: ""
`), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("JWT_TTL", "2h")
	t.Setenv("MONGO_INITDB_ROOT_USERNAME", "env-user")
	t.Setenv("MONGO_INITDB_ROOT_PASSWORD", "env-mongo")
	t.Setenv("MONGODB_HOST", "mongo:27017")
	t.Setenv("EJUDGE_BASE_URL", "https://ejudge.env")
	t.Setenv("TEACHERS_LOGINS", "alice;bob")
	t.Setenv("AIUSAGE_ENABLED", "1")

	cfg, err := configs.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Admin.JWTSecret != "env-secret" || cfg.Admin.JWTTTL != 2*time.Hour {
		t.Fatalf("admin: %+v", cfg.Admin)
	}
	if cfg.MongoDB.Username != "env-user" || cfg.MongoDB.Host != "mongo:27017" {
		t.Fatalf("mongo: %+v", cfg.MongoDB)
	}
	if !cfg.Ejudge.Enabled() || cfg.Ejudge.BaseURL != "https://ejudge.env" {
		t.Fatalf("ejudge: %+v", cfg.Ejudge)
	}
	if !cfg.AIUsage.Enabled {
		t.Fatal("aiusage should be enabled via env")
	}
	if len(cfg.TeachersLogins) != 2 || cfg.TeachersLogins[0] != "alice" {
		t.Fatalf("teachers logins: %+v", cfg.TeachersLogins)
	}
	uri, err := cfg.MongoDB.URI()
	if err != nil {
		t.Fatal(err)
	}
	if uri == "" {
		t.Fatal("empty uri")
	}
}

func TestLoadConfig_MissingJWTSecret(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`
http:
  addr: ":8080"
  shutdown_timeout: 5s
store:
  dir: "./data"
mongodb:
  username: "u"
  password: "p"
  host: "localhost"
  database: "scainer"
admin:
  jwt_secret: ""
  jwt_ttl: 1h
ejudge:
  base_url: "https://ejudge.example"
  timeout: 10s
jplag:
  jar_path: "bin/jplag.jar"
analyze:
  jobs_max_concurrent: 1
  analyze_concurrency: 1
aiusage:
  enabled: false
openai:
  base_url: "https://example/v1"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JWT_SECRET", "")
	t.Setenv("TEACHERS_LOGINS", "alice")
	if _, err := configs.LoadConfig(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfig_MissingEjudgeBaseURL(t *testing.T) {
	path := writeMinimalConfig(t, "")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("EJUDGE_BASE_URL", "")
	t.Setenv("TEACHERS_LOGINS", "alice")
	if _, err := configs.LoadConfig(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfig_MissingTeachersLogins(t *testing.T) {
	path := writeMinimalConfig(t, "https://ejudge.example")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("EJUDGE_BASE_URL", "https://ejudge.example")
	t.Setenv("TEACHERS_LOGINS", "")
	if _, err := configs.LoadConfig(path); err == nil {
		t.Fatal("expected error")
	}
}

func writeMinimalConfig(t *testing.T, ejudgeBaseURL string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
http:
  addr: ":8080"
  shutdown_timeout: 5s
store:
  dir: "./data"
mongodb:
  username: "u"
  password: "p"
  host: "localhost"
  database: "scainer"
admin:
  jwt_secret: "yaml-secret"
  jwt_ttl: 1h
ejudge:
  base_url: "` + ejudgeBaseURL + `"
  timeout: 10s
jplag:
  jar_path: "bin/jplag.jar"
analyze:
  jobs_max_concurrent: 1
  analyze_concurrency: 1
aiusage:
  enabled: false
openai:
  base_url: "https://example/v1"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
