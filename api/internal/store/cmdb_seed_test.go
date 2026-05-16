package store

import "testing"

func TestSeedCmdbContractProbeTopologyCreatesRealRelationGraph(t *testing.T) {
	SeedCmdbContractProbeTopology()

	root, ok := GetCmdbInstance("contract-probe")
	if !ok {
		t.Fatal("contract-probe instance was not seeded")
	}
	if root.ObjectID != "OperatingSystem1" {
		t.Fatalf("contract-probe object_id = %q, want OperatingSystem1", root.ObjectID)
	}

	rels := ListCmdbInstanceRelations(root.ID)
	if len(rels) != 1 {
		t.Fatalf("contract-probe relation count = %d, want 1: %#v", len(rels), rels)
	}
	rel := rels[0]
	if rel.ID != "contract-probe-relation-default-user" {
		t.Fatalf("relation id = %q, want contract-probe-relation-default-user", rel.ID)
	}
	if rel.SourceInstanceID != "contract-probe" || rel.TargetInstanceID != "contract-probe-user" {
		t.Fatalf("relation source/target mismatch: %#v", rel)
	}

	typ, ok := GetCmdbRelationType(rel.RelationTypeID)
	if !ok {
		t.Fatalf("relation type %q was not seeded", rel.RelationTypeID)
	}
	if typ.ID != "OperatingSystem1_default_j6p8Wb2xkV1666171515" {
		t.Fatalf("relation type id = %q, want mature capture relation id", typ.ID)
	}
	if typ.Name != "default" || typ.Mapping != "n:1" || typ.RuleExpression == "" || typ.RulesJSON == "" {
		t.Fatalf("relation type missing mature mapping/rules fields: %#v", typ)
	}
	if typ.Visible == nil || !*typ.Visible {
		t.Fatalf("relation type visible must be true: %#v", typ)
	}
}

func TestSeedCmdbContractProbeTopologyIsIdempotent(t *testing.T) {
	SeedCmdbContractProbeTopology()
	SeedCmdbContractProbeTopology()

	rels := ListCmdbInstanceRelations("contract-probe")
	if len(rels) != 1 {
		t.Fatalf("contract-probe relation count after repeated seed = %d, want 1: %#v", len(rels), rels)
	}
}
