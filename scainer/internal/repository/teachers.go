package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"scainer/internal/services/teachers"
)

type TeachersRepository struct {
	col *mongo.Collection
}

func NewTeachersRepository(db *mongo.Database) *TeachersRepository {
	return &TeachersRepository{col: db.Collection("teachers")}
}

type teacherDoc struct {
	Password string `bson:"password"`
}

func (r *TeachersRepository) Get(ctx context.Context, login string) (teachers.Record, bool, error) {
	var doc teacherDoc
	err := r.col.FindOne(ctx, bson.M{"_id": login}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return teachers.Record{}, false, nil
	}
	if err != nil {
		return teachers.Record{}, false, err
	}
	return teachers.Record{Login: login, Password: doc.Password}, true, nil
}
