package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scainer/internal/services/ejudge/gateway"
)

type EjudgeCredentialsRepository struct {
	col *mongo.Collection
}

func NewEjudgeCredentialsRepository(db *mongo.Database) *EjudgeCredentialsRepository {
	return &EjudgeCredentialsRepository{col: db.Collection("ejudge_credentials")}
}

type ejudgeCredDoc struct {
	APIKey string `bson:"api_key"`
}

func (r *EjudgeCredentialsRepository) Get(ctx context.Context, login string) (gateway.Credentials, bool, error) {
	var doc ejudgeCredDoc
	err := r.col.FindOne(ctx, bson.M{"_id": login}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return gateway.Credentials{}, false, nil
	}
	if err != nil {
		return gateway.Credentials{}, false, err
	}
	return gateway.Credentials{
		Login:  login,
		APIKey: doc.APIKey,
	}, true, nil
}

func (r *EjudgeCredentialsRepository) Upsert(ctx context.Context, cred gateway.Credentials) error {
	existing, found, err := r.Get(ctx, cred.Login)
	if err != nil {
		return err
	}

	set := bson.M{}
	if cred.APIKey != "" {
		set["api_key"] = cred.APIKey
	} else if found && existing.APIKey != "" {
		// Do not overwrite a stored api_key with empty.
	}

	if len(set) == 0 {
		return nil
	}

	_, err = r.col.UpdateOne(
		ctx,
		bson.M{"_id": cred.Login},
		bson.M{"$set": set},
		options.Update().SetUpsert(true),
	)
	return err
}

var _ gateway.CredentialsRepository = (*EjudgeCredentialsRepository)(nil)
