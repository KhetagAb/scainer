package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scainer/internal/contests"
	"scainer/internal/domain"
)

type FindingsRepository struct{ col *mongo.Collection }

func NewFindingsRepository(db *mongo.Database) *FindingsRepository {
	return &FindingsRepository{col: db.Collection("findings")}
}

func (r *FindingsRepository) Put(ctx context.Context, snapshot contests.FindingsSnapshot) error {
	_, err := r.col.ReplaceOne(ctx,
		bson.M{"_id": string(snapshot.ContestID)}, snapshot, options.Replace().SetUpsert(true))
	return err
}

func (r *FindingsRepository) Get(ctx context.Context, id domain.ContestID) (contests.FindingsSnapshot, bool, error) {
	var snapshot contests.FindingsSnapshot
	err := r.col.FindOne(ctx, bson.M{"_id": string(id)}).Decode(&snapshot)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return contests.FindingsSnapshot{}, false, nil
	}
	if err != nil {
		return contests.FindingsSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func (r *FindingsRepository) Delete(ctx context.Context, id domain.ContestID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": string(id)})
	return err
}

var _ contests.FindingsStore = (*FindingsRepository)(nil)
