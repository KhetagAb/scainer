package contests

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
)

type ContestRecord struct {
	Contest Contest    `bson:"info"` // ключ "info" — совместимость с уже лежащими в Mongo документами
	Source  SourceSpec `bson:"source"`
}

type ContestRegistry interface {
	Put(ctx context.Context, rec ContestRecord) error
	Get(ctx context.Context, id domain.ContestID) (ContestRecord, bool, error)
	Delete(ctx context.Context, id domain.ContestID) error
	List(ctx context.Context) ([]ContestRecord, error)
}

// yaml.Node напрямую в BSON не сериализуется — через generic Go-значение.
func (s SourceSpec) MarshalBSON() ([]byte, error) {
	var cfg any
	if s.Config.Kind != 0 {
		if err := s.Config.Decode(&cfg); err != nil {
			return nil, err
		}
	}
	return bson.Marshal(struct {
		Type   string `bson:"type"`
		Config any    `bson:"config"`
	}{Type: s.Type, Config: cfg})
}

func (s *SourceSpec) UnmarshalBSON(data []byte) error {
	var raw struct {
		Type   string        `bson:"type"`
		Config bson.RawValue `bson:"config"`
	}
	if err := bson.Unmarshal(data, &raw); err != nil {
		return err
	}

	var node yaml.Node
	if raw.Config.Type != bson.TypeNull && len(raw.Config.Value) > 0 {
		var generic any
		if err := raw.Config.Unmarshal(&generic); err != nil {
			return err
		}
		yamlBytes, err := yaml.Marshal(bsonToPlain(generic))
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlBytes, &node); err != nil {
			return err
		}
	}

	s.Type, s.Config = raw.Type, node
	return nil
}

// bson.D/bson.A → map/slice, иначе yaml.Marshal пишет объекты как sequence пар Key/Value.
func bsonToPlain(v any) any {
	switch x := v.(type) {
	case bson.D:
		m := make(map[string]any, len(x))
		for _, e := range x {
			m[e.Key] = bsonToPlain(e.Value)
		}
		return m
	case bson.A:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = bsonToPlain(e)
		}
		return out
	default:
		return v
	}
}
