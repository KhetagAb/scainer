package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
)

type AnalysisRepository struct{ col *mongo.Collection }

func NewAnalysisRepository(db *mongo.Database) *AnalysisRepository {
	return &AnalysisRepository{col: db.Collection("analysis")}
}

func (r *AnalysisRepository) Put(ctx context.Context, snapshot contests.AnalysisSnapshot) error {
	_, err := r.col.ReplaceOne(ctx,
		bson.M{"_id": string(snapshot.ContestID)}, snapshot, options.Replace().SetUpsert(true))
	return err
}

func (r *AnalysisRepository) Get(ctx context.Context, id domain.ContestID) (contests.AnalysisSnapshot, bool, error) {
	var snapshot contests.AnalysisSnapshot
	err := r.col.FindOne(ctx, bson.M{"_id": string(id)}).Decode(&snapshot)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return contests.AnalysisSnapshot{}, false, nil
	}
	if err != nil {
		return contests.AnalysisSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func (r *AnalysisRepository) Delete(ctx context.Context, id domain.ContestID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": string(id)})
	return err
}

var _ contests.AnalysisRepository = (*AnalysisRepository)(nil)
