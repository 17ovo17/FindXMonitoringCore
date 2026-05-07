package store

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/spf13/viper"
)

func resetUsersFallbackForTest(t *testing.T) {
	t.Helper()
	oldDB, oldMySQLOK := db, mysqlOK
	oldFallbackFile := viper.GetString("storage.fallback_file")
	viper.Set("storage.fallback_file", filepath.Join(t.TempDir(), "memory-store.json"))

	mu.Lock()
	users = map[string]*model.User{}
	mysqlOK = false
	db = nil
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()
		users = map[string]*model.User{}
		mysqlOK = oldMySQLOK
		db = oldDB
		mu.Unlock()
		if oldFallbackFile == "" {
			viper.Set("storage.fallback_file", "")
			return
		}
		viper.Set("storage.fallback_file", oldFallbackFile)
	})
}

func TestUserFallbackCreateGetCountAndUpdatePassword(t *testing.T) {
	resetUsersFallbackForTest(t)

	created := &model.User{
		ID:            "user-1",
		Username:      "admin",
		PasswordHash:  "hash-before",
		Role:          "admin",
		MustChangePwd: true,
	}
	if err := CreateUser(created); err != nil {
		t.Fatalf("create fallback user failed: %v", err)
	}
	if UserCount() != 1 {
		t.Fatalf("expected one fallback user, got %d", UserCount())
	}

	got := GetUserByUsername("admin")
	if got == nil {
		t.Fatalf("expected fallback user")
	}
	if got.Username != created.Username || got.Role != created.Role || got.PasswordHash != created.PasswordHash {
		t.Fatalf("fallback user mismatch: got %#v", got)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set: created=%v updated=%v", got.CreatedAt, got.UpdatedAt)
	}

	beforeUpdate := got.UpdatedAt
	time.Sleep(time.Millisecond)
	if err := UpdateUserPassword("user-1", "hash-after"); err != nil {
		t.Fatalf("update fallback password failed: %v", err)
	}

	updated := GetUserByUsername("admin")
	if updated == nil {
		t.Fatalf("expected fallback user after password update")
	}
	if updated.PasswordHash != "hash-after" {
		t.Fatalf("expected updated hash, got %q", updated.PasswordHash)
	}
	if updated.MustChangePwd {
		t.Fatalf("expected must_change_pwd to be false")
	}
	if !updated.UpdatedAt.After(beforeUpdate) {
		t.Fatalf("expected updated_at to advance: before=%v after=%v", beforeUpdate, updated.UpdatedAt)
	}
}

func TestUserFallbackSnapshotPersistsUsers(t *testing.T) {
	resetUsersFallbackForTest(t)

	if err := CreateUser(&model.User{
		ID:            "user-2",
		Username:      "snapshot-admin",
		PasswordHash:  "snapshot-hash-before",
		Role:          "admin",
		MustChangePwd: true,
	}); err != nil {
		t.Fatalf("create fallback user failed: %v", err)
	}
	if err := UpdateUserPassword("user-2", "snapshot-hash-after"); err != nil {
		t.Fatalf("update fallback password failed: %v", err)
	}

	mu.Lock()
	users = map[string]*model.User{}
	mu.Unlock()
	loadFallbackSnapshot()

	got := GetUserByUsername("snapshot-admin")
	if got == nil {
		t.Fatalf("expected user restored from fallback snapshot")
	}
	if got.PasswordHash != "snapshot-hash-after" {
		t.Fatalf("expected persisted updated hash, got %q", got.PasswordHash)
	}
	if got.MustChangePwd {
		t.Fatalf("expected persisted must_change_pwd=false")
	}
	if UserCount() != 1 {
		t.Fatalf("expected restored fallback user count, got %d", UserCount())
	}
}

func TestUserFallbackCreateRollsBackWhenSnapshotCannotPersist(t *testing.T) {
	resetUsersFallbackForTest(t)
	blockingDir := filepath.Join(t.TempDir(), "snapshot-blocker")
	if err := os.Mkdir(blockingDir, 0700); err != nil {
		t.Fatalf("create blocking directory failed: %v", err)
	}
	viper.Set("storage.fallback_file", blockingDir)

	err := CreateUser(&model.User{
		ID:            "user-blocked",
		Username:      "blocked-admin",
		PasswordHash:  "blocked-hash",
		Role:          "admin",
		MustChangePwd: true,
	})
	if err == nil {
		t.Fatalf("expected create user to fail when fallback snapshot cannot persist")
	}
	if got := GetUserByUsername("blocked-admin"); got != nil {
		t.Fatalf("expected fallback user rollback, got %#v", got)
	}
	if UserCount() != 0 {
		t.Fatalf("expected no fallback users after rollback, got %d", UserCount())
	}
}

func TestUserFallbackUpdatePasswordRollsBackWhenSnapshotCannotPersist(t *testing.T) {
	resetUsersFallbackForTest(t)

	if err := CreateUser(&model.User{
		ID:            "user-rollback",
		Username:      "rollback-admin",
		PasswordHash:  "hash-before",
		Role:          "admin",
		MustChangePwd: true,
	}); err != nil {
		t.Fatalf("create fallback user failed: %v", err)
	}

	blockingDir := filepath.Join(t.TempDir(), "snapshot-blocker")
	if err := os.Mkdir(blockingDir, 0700); err != nil {
		t.Fatalf("create blocking directory failed: %v", err)
	}
	viper.Set("storage.fallback_file", blockingDir)

	err := UpdateUserPassword("user-rollback", "hash-after")
	if err == nil {
		t.Fatalf("expected password update to fail when fallback snapshot cannot persist")
	}
	got := GetUserByUsername("rollback-admin")
	if got == nil {
		t.Fatalf("expected fallback user after failed update")
	}
	if got.PasswordHash != "hash-before" {
		t.Fatalf("expected password hash rollback, got %q", got.PasswordHash)
	}
	if !got.MustChangePwd {
		t.Fatalf("expected must_change_pwd rollback")
	}
}

func TestUserFallbackSnapshotFilePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows does not preserve POSIX file permission bits")
	}
	resetUsersFallbackForTest(t)

	if err := CreateUser(&model.User{
		ID:            "user-3",
		Username:      "permission-admin",
		PasswordHash:  "permission-hash",
		Role:          "admin",
		MustChangePwd: true,
	}); err != nil {
		t.Fatalf("create fallback user failed: %v", err)
	}

	info, err := os.Stat(fallbackSnapshotPath())
	if err != nil {
		t.Fatalf("stat fallback snapshot failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("expected fallback snapshot permission 0600, got %o", perm)
	}
}
