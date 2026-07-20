package repository_test

import (
	"os"
	"strings"
	"testing"

	"scainer/internal/repository"
)

func TestURIFromEnv(t *testing.T) {
	t.Setenv("MONGO_INITDB_ROOT_USERNAME", "u")
	t.Setenv("MONGO_INITDB_ROOT_PASSWORD", "p@ss:word")
	t.Setenv("MONGODB_HOST", "localhost")

	uri, err := repository.URIFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(uri, "mongodb://u:") {
		t.Fatalf("uri prefix: %q", uri)
	}
	if !strings.Contains(uri, "@localhost:27017") || !strings.Contains(uri, "authSource=admin") {
		t.Fatalf("uri host/query: %q", uri)
	}
	// спецсимволы пароля должны быть percent-encoded
	if strings.Contains(uri, "p@ss:word") {
		t.Fatalf("password not encoded: %q", uri)
	}
}

func TestURIFromEnvMissing(t *testing.T) {
	for _, k := range []string{"MONGO_INITDB_ROOT_USERNAME", "MONGO_INITDB_ROOT_PASSWORD", "MONGODB_HOST"} {
		_ = os.Unsetenv(k)
	}
	if _, err := repository.URIFromEnv(); err == nil {
		t.Fatal("expected error")
	}
}
