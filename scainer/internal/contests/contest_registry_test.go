package contests

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v3"

	ejimporter "scainer/internal/importer/ejudge"
)

func TestSourceSpecBSONRoundtrip_EjudgeConfig(t *testing.T) {
	var cfgNode yaml.Node
	if err := cfgNode.Encode(ejimporter.Config{ContestID: 50051}); err != nil {
		t.Fatal(err)
	}
	orig := SourceSpec{Type: "ejudge", Config: cfgNode}

	raw, err := bson.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	var got SourceSpec
	if err := bson.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "ejudge" {
		t.Fatalf("type = %q", got.Type)
	}

	var cfg ejimporter.Config
	if err := got.Config.Decode(&cfg); err != nil {
		t.Fatalf("Decode ejimporter.Config: %v", err)
	}
	if cfg.ContestID != 50051 {
		t.Fatalf("ContestID = %d", cfg.ContestID)
	}
}
