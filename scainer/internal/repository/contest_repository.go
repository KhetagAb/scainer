package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"scainer/internal/services/contests"
	"scainer/internal/domain"
)

type ContestRepository struct{ col *mongo.Collection }

func NewContestRepository(db *mongo.Database) *ContestRepository {
	return &ContestRepository{col: db.Collection("contests")}
}

func (r *ContestRepository) Put(ctx context.Context, record contests.ContestRecord) error {
	_, err := r.col.ReplaceOne(ctx,
		bson.M{"_id": string(record.Contest.ID)}, record, options.Replace().SetUpsert(true))
	return err
}

func (r *ContestRepository) Get(ctx context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	var record contests.ContestRecord
	err := r.col.FindOne(ctx, bson.M{"_id": string(id)}).Decode(&record)
	if err == mongo.ErrNoDocuments {
		return contests.ContestRecord{}, false, nil
	}
	if err != nil {
		return contests.ContestRecord{}, false, err
	}
	return record, true, nil
}

func (r *ContestRepository) Delete(ctx context.Context, id domain.ContestID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": string(id)})
	return err
}

func (r *ContestRepository) List(ctx context.Context) ([]contests.ContestRecord, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var out []contests.ContestRecord
	for cursor.Next(ctx) {
		var record contests.ContestRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, cursor.Err()
}

var _ contests.ContestRegistry = (*ContestRepository)(nil)
