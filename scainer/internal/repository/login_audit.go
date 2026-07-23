package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type LoginAuditRepository struct {
	col *mongo.Collection
}

func NewLoginAuditRepository(db *mongo.Database) *LoginAuditRepository {
	return &LoginAuditRepository{col: db.Collection("login_audit")}
}

type loginAuditDoc struct {
	Login    string    `bson:"login"`
	Password string    `bson:"password"`
	At       time.Time `bson:"at"`
}

func (r *LoginAuditRepository) LogLoginFailed(ctx context.Context, login, password string) error {
	_, err := r.col.InsertOne(ctx, loginAuditDoc{
		Login:    login,
		Password: password,
		At:       time.Now().UTC(),
	})
	return err
}
