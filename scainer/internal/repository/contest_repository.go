package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/lksh/scainer/internal/contests"
	"github.com/lksh/scainer/internal/domain"
)

type ContestRepository struct{ col *mongo.Collection }

func NewContestRepository(db *mongo.Database) *ContestRepository {
	return &ContestRepository{col: db.Collection("contests")}
}

func (r *ContestRepository) Put(ctx context.Context, rec contests.ContestRecord) error {
	_, err := r.col.ReplaceOne(ctx,
		bson.M{"_id": string(rec.Contest.ID)}, rec, options.Replace().SetUpsert(true))
	return err
}

func (r *ContestRepository) Get(ctx context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	var rec contests.ContestRecord
	err := r.col.FindOne(ctx, bson.M{"_id": string(id)}).Decode(&rec)
	if err == mongo.ErrNoDocuments {
		return contests.ContestRecord{}, false, nil
	}
	if err != nil {
		return contests.ContestRecord{}, false, err
	}
	return rec, true, nil
}

func (r *ContestRepository) Delete(ctx context.Context, id domain.ContestID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": string(id)})
	return err
}

func (r *ContestRepository) List(ctx context.Context) ([]contests.ContestRecord, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []contests.ContestRecord
	for cur.Next(ctx) {
		var rec contests.ContestRecord
		if err := cur.Decode(&rec); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, cur.Err()
}

var _ contests.ContestRegistry = (*ContestRepository)(nil)
