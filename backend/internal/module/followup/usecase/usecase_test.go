package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/followup/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/followup/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/followup/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/google/uuid"
)

func newID() string { return uuid.Must(uuid.NewV7()).String() }

// newFollowUpFixture inserts directly against follow_ups — the automatic
// creation path (pipeline -> aguardando_fechamento/nao_fechado) is already
// covered end-to-end in conversation/usecase/usecase_test.go; this module's
// own tests only need to exercise List/Complete/Cancel over a known row.
func newFollowUpFixture(t *testing.T, db *database.DB, conversationID, customerID, unitID string, status entity.FollowUpStatus) entity.FollowUp {
	t.Helper()
	fu := entity.FollowUp{
		ID: newID(), ConversationID: conversationID, CustomerID: customerID, UnitID: unitID,
		PipelineStage: entity.StageAguardandoFechamento, DueAt: time.Now().Add(72 * time.Hour),
		Status: status, Note: "test follow-up", CreatedAt: time.Now(),
	}
	if _, err := db.NamedExec(`
		INSERT INTO follow_ups (id, conversation_id, customer_id, unit_id, pipeline_stage, due_at, status, note, created_at)
		VALUES (:id, :conversation_id, :customer_id, :unit_id, :pipeline_stage, :due_at, :status, :note, :created_at)
	`, fu); err != nil {
		t.Fatalf("create follow-up fixture: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM follow_ups WHERE id = $1`, fu.ID) })
	return fu
}

func setup(t *testing.T) (*usecase.UseCase, *database.DB, entity.Unit, entity.FollowUp) {
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})
	fu := newFollowUpFixture(t, db, conv.ID, customer.ID, unit.ID, entity.FollowUpOpen)
	return usecase.New(repository.New(db)), db, unit, fu
}

func TestComplete_OpenFollowUp_SetsDoneAndCompletedAt(t *testing.T) {
	uc, db, unit, fu := setup(t)
	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}

	result, err := uc.Complete(context.Background(), fu.ID, access)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if result.Status != string(entity.FollowUpDone) {
		t.Fatalf("expected status=done, got %q", result.Status)
	}
	if result.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}

	var status string
	if err := db.Get(&status, `SELECT status FROM follow_ups WHERE id = $1`, fu.ID); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != string(entity.FollowUpDone) {
		t.Fatalf("expected persisted status=done, got %q", status)
	}
}

func TestCancel_OpenFollowUp_SetsCanceled(t *testing.T) {
	uc, db, unit, fu := setup(t)
	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}

	result, err := uc.Cancel(context.Background(), fu.ID, access)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if result.Status != string(entity.FollowUpCanceled) {
		t.Fatalf("expected status=canceled, got %q", result.Status)
	}

	var status string
	if err := db.Get(&status, `SELECT status FROM follow_ups WHERE id = $1`, fu.ID); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != string(entity.FollowUpCanceled) {
		t.Fatalf("expected persisted status=canceled, got %q", status)
	}
}

// Completing an already-canceled follow-up must not silently overwrite its
// status — this exercises the read-modify-write path, not a state machine
// gate (setStatus has no "already done" guard the way engagement.Reply
// does), but it must still respect unit scope on the way in.
func TestCompleteCancel_OutsideUnitScope_NotFound(t *testing.T) {
	db := testutil.DB(t)
	unitA := testutil.NewUnit(t, db)
	unitB := testutil.NewUnit(t, db)
	customer := testutil.NewCustomer(t, db, &unitB.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unitB.ID, testutil.ConversationOpts{})
	fu := newFollowUpFixture(t, db, conv.ID, customer.ID, unitB.ID, entity.FollowUpOpen)
	uc := usecase.New(repository.New(db))

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unitA.ID}}
	_, err := uc.Complete(context.Background(), fu.ID, access)
	if err == nil {
		t.Fatal("expected completing a follow-up outside the agent's units to fail")
	}
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok || respErr.StatusCode != 404 {
		t.Fatalf("expected 404, got %T: %v", err, err)
	}
}

func TestList_FiltersByStatus(t *testing.T) {
	uc, db, unit, openFU := setup(t)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})
	doneFU := newFollowUpFixture(t, db, conv.ID, customer.ID, unit.ID, entity.FollowUpDone)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	page, err := uc.List(context.Background(), model.ListFilter{Status: "open", UnitID: unit.ID, RequesterUnitIDs: access.UnitIDs})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	foundOpen, foundDone := false, false
	for _, item := range page.Items {
		if item.ID == openFU.ID {
			foundOpen = true
		}
		if item.ID == doneFU.ID {
			foundDone = true
		}
	}
	if !foundOpen {
		t.Fatal("expected the open follow-up to appear in a status=open filtered list")
	}
	if foundDone {
		t.Fatal("expected the done follow-up to be EXCLUDED from a status=open filtered list")
	}
}

func TestList_UnitScoped_ExcludesOtherUnits(t *testing.T) {
	db := testutil.DB(t)
	unitA := testutil.NewUnit(t, db)
	unitB := testutil.NewUnit(t, db)
	customerA := testutil.NewCustomer(t, db, &unitA.ID, entity.IdentificationAnonymous, nil)
	convA := testutil.NewConversation(t, db, customerA.ID, unitA.ID, testutil.ConversationOpts{})
	fuA := newFollowUpFixture(t, db, convA.ID, customerA.ID, unitA.ID, entity.FollowUpOpen)

	customerB := testutil.NewCustomer(t, db, &unitB.ID, entity.IdentificationAnonymous, nil)
	convB := testutil.NewConversation(t, db, customerB.ID, unitB.ID, testutil.ConversationOpts{})
	fuB := newFollowUpFixture(t, db, convB.ID, customerB.ID, unitB.ID, entity.FollowUpOpen)

	uc := usecase.New(repository.New(db))
	page, err := uc.List(context.Background(), model.ListFilter{RequesterUnitIDs: []string{unitA.ID}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	foundA, foundB := false, false
	for _, item := range page.Items {
		if item.ID == fuA.ID {
			foundA = true
		}
		if item.ID == fuB.ID {
			foundB = true
		}
	}
	if !foundA {
		t.Fatal("expected the agent's own unit follow-up to be visible")
	}
	if foundB {
		t.Fatal("expected a follow-up from a DIFFERENT unit to be excluded from a scoped list")
	}
}
