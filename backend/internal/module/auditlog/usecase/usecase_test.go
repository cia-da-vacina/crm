package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/auditlog/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
	"github.com/cia-da-vacina/crm/backend/pkg/audit"
)

// pkg/audit.Logger.Log and this module's List are two different code paths
// (write side lives in pkg/audit, read side here) that both depend on the
// audit_logs schema staying in sync — the only way to prove the write
// actually produces what the read side can filter/find is to exercise both
// together against the real table, which is exactly what these tests do.

func TestList_WrittenEntry_IsRetrievableAndFilterable(t *testing.T) {
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	actor := testutil.NewUser(t, db, entity.RoleAdmin, unit.ID)
	logger := audit.New(db)
	uc := usecase.New(repository.New(db))
	t.Cleanup(func() { db.Exec(`DELETE FROM audit_logs WHERE actor_user_id = $1`, actor.ID) })

	actorID := actor.ID
	unitID := unit.ID
	logger.Log(context.Background(), audit.Entry{
		ActorUserID: &actorID, Action: "conversation.claim", ResourceType: "conversation",
		ResourceID: "some-conversation-id", UnitID: &unitID, Metadata: map[string]any{"owner_id": actorID},
	})

	page, err := uc.List(context.Background(), model.ListFilter{ActorUserID: actor.ID})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected exactly 1 audit entry for this actor, got %d", len(page.Items))
	}
	item := page.Items[0]
	if item.Action != "conversation.claim" || item.ResourceType != "conversation" {
		t.Fatalf("unexpected audit entry contents: %+v", item)
	}
	if item.ActorName == nil {
		t.Fatal("expected actor_name to be resolved via join with users")
	}
	if item.Metadata["owner_id"] != actorID {
		t.Fatalf("expected metadata.owner_id=%s to round-trip through JSONB, got %v", actorID, item.Metadata["owner_id"])
	}

	// Filtering by a DIFFERENT action must exclude it.
	page, err = uc.List(context.Background(), model.ListFilter{ActorUserID: actor.ID, Action: "user.create"})
	if err != nil {
		t.Fatalf("list with action filter: %v", err)
	}
	if len(page.Items) != 0 {
		t.Fatalf("expected 0 entries when filtering by an unrelated action, got %d", len(page.Items))
	}
}

// Audit logging must never block or fail the caller's action — a bad
// actor_user_id (FK violation) is swallowed internally and just logged to
// stdout (pkg/audit/audit.go), so List simply won't find anything, but the
// original business action (whatever called Log) is not affected. This
// pins down that "never returns an error" contract explicitly.
func TestLog_NeverPanicsOrErrors_EvenWithInvalidForeignKeys(t *testing.T) {
	db := testutil.DB(t)
	logger := audit.New(db)

	nonExistentActor := "00000000-0000-7000-8000-000000000000"
	// This must not panic and Log has no return value to check — the test
	// passes simply by completing without a panic/crash.
	logger.Log(context.Background(), audit.Entry{
		ActorUserID: &nonExistentActor, Action: "user.create", ResourceType: "user", ResourceID: "x",
	})
}

func TestList_MultipleActions_OrderedNewestFirst(t *testing.T) {
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	actor := testutil.NewUser(t, db, entity.RoleAdmin, unit.ID)
	logger := audit.New(db)
	uc := usecase.New(repository.New(db))
	t.Cleanup(func() { db.Exec(`DELETE FROM audit_logs WHERE actor_user_id = $1`, actor.ID) })

	actorID := actor.ID
	logger.Log(context.Background(), audit.Entry{ActorUserID: &actorID, Action: "user.create", ResourceType: "user", ResourceID: "1"})
	time.Sleep(5 * time.Millisecond)
	logger.Log(context.Background(), audit.Entry{ActorUserID: &actorID, Action: "user.delete", ResourceType: "user", ResourceID: "1"})

	page, err := uc.List(context.Background(), model.ListFilter{ActorUserID: actor.ID})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(page.Items))
	}
	if page.Items[0].Action != "user.delete" {
		t.Fatalf("expected the most recent action (user.delete) first, got %q", page.Items[0].Action)
	}
}
