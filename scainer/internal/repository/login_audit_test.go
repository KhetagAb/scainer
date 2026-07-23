package repository_test

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"scainer/internal/repository"
)

func TestLoginAuditRepository_LogLoginFailed(t *testing.T) {
	db := mustDB(t)
	repo := repository.NewLoginAuditRepository(db)

	ctx := context.Background()
	if err := repo.LogLoginFailed(ctx, "alice", "wrong-pass"); err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Login    string    `bson:"login"`
		Password string    `bson:"password"`
		At       time.Time `bson:"at"`
	}
	if err := db.Collection("login_audit").FindOne(ctx, bson.M{"login": "alice"}).Decode(&doc); err != nil {
		t.Fatal(err)
	}
	if doc.Password != "wrong-pass" {
		t.Fatalf("password=%q", doc.Password)
	}
	if doc.At.IsZero() {
		t.Fatal("at is zero")
	}
}
