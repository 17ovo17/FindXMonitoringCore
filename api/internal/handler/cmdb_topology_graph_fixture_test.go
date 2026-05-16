package handler

import (
	"testing"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"
)

func createCmdbRelationGraphFixture(t *testing.T, prefix string) (string, string) {
	t.Helper()
	rootObjectID := prefix + "-app"
	targetObjectID := prefix + "-db"
	for _, obj := range []model.CmdbObject{
		{ID: rootObjectID, Name: "app", CategoryID: prefix + "-cat", ObjectType: 101},
		{ID: targetObjectID, Name: "database", CategoryID: prefix + "-cat", ObjectType: 101},
	} {
		obj := obj
		if err := store.CreateCmdbObject(&obj); err != nil {
			t.Fatalf("create object %s: %v", obj.ID, err)
		}
	}
	root := &model.CmdbInstance{
		ObjectID: rootObjectID,
		Data:     `{"name":"api-service","password":"root-password-marker","dsn":"secret-dsn-marker"}`,
		Creator:  "test",
		Updater:  "test",
	}
	if err := store.CreateCmdbInstance(root); err != nil {
		t.Fatalf("create root instance: %v", err)
	}
	target := &model.CmdbInstance{
		ObjectID: targetObjectID,
		Data:     `{"name":"mysql-primary","token":"target-token-marker"}`,
		Creator:  "test",
		Updater:  "test",
	}
	if err := store.CreateCmdbInstance(target); err != nil {
		t.Fatalf("create target instance: %v", err)
	}
	if err := store.CreateCmdbRelationType(&model.CmdbRelationType{ID: "depends_on", Name: "depends_on", Label: "depends on"}); err != nil {
		t.Fatalf("create relation type: %v", err)
	}
	if err := store.CreateCmdbInstanceRelation(&model.CmdbInstanceRelation{
		ID:               prefix + "-rel",
		SourceInstanceID: root.ID,
		TargetInstanceID: target.ID,
		RelationTypeID:   "depends_on",
	}); err != nil {
		t.Fatalf("create instance relation: %v", err)
	}
	return root.ID, target.ID
}

func createCmdbRecursiveRelationFixture(t *testing.T, prefix string) (string, string, string) {
	t.Helper()
	rootID, middleID := createCmdbRelationGraphFixture(t, prefix)
	leafObjectID := prefix + "-cache"
	obj := model.CmdbObject{ID: leafObjectID, Name: "cache", CategoryID: prefix + "-cat", ObjectType: 101}
	if err := store.CreateCmdbObject(&obj); err != nil {
		t.Fatalf("create object %s: %v", obj.ID, err)
	}
	leaf := &model.CmdbInstance{
		ObjectID: leafObjectID,
		Data:     `{"name":"redis-leaf","token":"leaf-token-marker"}`,
		Creator:  "test",
		Updater:  "test",
	}
	if err := store.CreateCmdbInstance(leaf); err != nil {
		t.Fatalf("create leaf instance: %v", err)
	}
	if err := store.CreateCmdbInstanceRelation(&model.CmdbInstanceRelation{
		ID:               prefix + "-rel-middle-leaf",
		SourceInstanceID: middleID,
		TargetInstanceID: leaf.ID,
		RelationTypeID:   "depends_on",
	}); err != nil {
		t.Fatalf("create recursive instance relation: %v", err)
	}
	return rootID, middleID, leaf.ID
}

func createCmdbDanglingRelationFixture(t *testing.T, prefix string) string {
	t.Helper()
	rootObjectID := prefix + "-app"
	obj := model.CmdbObject{ID: rootObjectID, Name: "app", CategoryID: prefix + "-cat", ObjectType: 101}
	if err := store.CreateCmdbObject(&obj); err != nil {
		t.Fatalf("create object %s: %v", obj.ID, err)
	}
	root := &model.CmdbInstance{
		ObjectID: rootObjectID,
		Data:     `{"name":"api-service"}`,
		Creator:  "test",
		Updater:  "test",
	}
	if err := store.CreateCmdbInstance(root); err != nil {
		t.Fatalf("create root instance: %v", err)
	}
	if err := store.CreateCmdbInstanceRelation(&model.CmdbInstanceRelation{
		ID:               prefix + "-rel",
		SourceInstanceID: root.ID,
		TargetInstanceID: prefix + "-missing-target",
		RelationTypeID:   prefix + "-missing-type",
	}); err != nil {
		t.Fatalf("create dangling relation: %v", err)
	}
	return root.ID
}
