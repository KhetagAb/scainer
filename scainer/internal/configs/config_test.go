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
  username: "yaml-admin"
  password: "yaml-pass"
  jwt_secret: "yaml-secret"
  jwt_ttl: 1h
ejudge:
  base_url: "https://ejudge.example"
  api_key: ""
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

	t.Setenv("ADMIN_USERNAME", "env-admin")
	t.Setenv("ADMIN_PASSWORD", "env-pass")
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("JWT_TTL", "2h")
	t.Setenv("MONGO_INITDB_ROOT_USERNAME", "env-user")
	t.Setenv("MONGO_INITDB_ROOT_PASSWORD", "env-mongo")
	t.Setenv("MONGODB_HOST", "mongo:27017")
	t.Setenv("EJUDGE_API_KEY", "ej-key")
	t.Setenv("AIUSAGE_ENABLED", "1")

	cfg, err := configs.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Admin.Username != "env-admin" || cfg.Admin.JWTTTL != 2*time.Hour {
		t.Fatalf("admin: %+v", cfg.Admin)
	}
	if cfg.MongoDB.Username != "env-user" || cfg.MongoDB.Host != "mongo:27017" {
		t.Fatalf("mongo: %+v", cfg.MongoDB)
	}
	if !cfg.Ejudge.Enabled() || cfg.Ejudge.APIKey != "ej-key" {
		t.Fatalf("ejudge: %+v", cfg.Ejudge)
	}
	if !cfg.AIUsage.Enabled {
		t.Fatal("aiusage should be enabled via env")
	}
	uri, err := cfg.MongoDB.URI()
	if err != nil {
		t.Fatal(err)
	}
	if uri == "" {
		t.Fatal("empty uri")
	}
}

func TestLoadConfig_MissingAdmin(t *testing.T) {
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
  username: ""
  password: ""
  jwt_secret: ""
  jwt_ttl: 1h
ejudge:
  base_url: "https://ejudge.example"
  api_key: ""
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
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("JWT_SECRET", "")
	if _, err := configs.LoadConfig(path); err == nil {
		t.Fatal("expected error")
	}
}
