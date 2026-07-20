package repository_test

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/lksh/scainer/internal/repository"
)

func mustDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri, err := repository.URIFromEnv()
	if err != nil {
		t.Skipf("%v — пропускаем интеграционный тест с реальной MongoDB", err)
	}

	ctx := context.Background()
	client, db, err := repository.Connect(ctx, uri, "scainer_repository_test")
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})
	return db
}
