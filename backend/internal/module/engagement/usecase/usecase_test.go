package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/engagement/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/google/uuid"
)

func newUseCase(t *testing.T, db *database.DB) (*usecase.UseCase, *meta.MockClient) {
	registry := meta.NewRegistry()
	waMock := meta.NewMockClient(meta.ChannelWhatsApp)
	igMock := meta.NewMockClient(meta.ChannelInstagram)
	fbMock := meta.NewMockClient(meta.ChannelFacebook)
	registry.Register(waMock)
	registry.Register(igMock)
	registry.Register(fbMock)
	return usecase.New(repository.New(db), registry, nil), igMock
}

func respStatus(t *testing.T, err error) int {
	t.Helper()
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok {
		t.Fatalf("expected *apperrors.ResponseError, got %T (%v)", err, err)
	}
	return respErr.StatusCode
}

func newID() string { return uuid.Must(uuid.NewV7()).String() }

func newEngagementFixture(t *testing.T, db *database.DB, unitID string, status entity.EngagementStatus) entity.SocialEngagement {
	t.Helper()
	e := entity.SocialEngagement{
		ID: newID(), Channel: entity.ChannelInstagram, Type: entity.EngagementPostComment,
		Status: status, UnitID: unitID, Body: "quanto custa a vacina da gripe?",
		ExternalID: "ext-" + newID(), AuthorExternalID: "author-" + newID(), CreatedAt: time.Now(),
	}
	if _, err := db.NamedExec(`
		INSERT INTO social_engagements (id, channel, type, status, unit_id, body, external_id, author_external_id, created_at)
		VALUES (:id, :channel, :type, :status, :unit_id, :body, :external_id, :author_external_id, :created_at)
	`, e); err != nil {
		t.Fatalf("create engagement fixture: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM social_engagements WHERE id = $1`, e.ID) })
	return e
}

// --- IngestFromWebhook idempotency -----------------------------------------

func TestIngestFromWebhook_DuplicateExternalID_Idempotent(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)

	e := entity.SocialEngagement{
		ID: newID(), Channel: entity.ChannelInstagram, Type: entity.EngagementPostComment,
		UnitID: unit.ID, Body: "oi", ExternalID: "dup-ext-id", AuthorExternalID: "author-x", CreatedAt: time.Now(),
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM social_engagements WHERE external_id = 'dup-ext-id'`) })

	if err := uc.IngestFromWebhook(context.Background(), e); err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	e.ID = newID() // simulate a resend where the Meta assigns nothing meaningful to reuse — only external_id matters for idempotency.
	if err := uc.IngestFromWebhook(context.Background(), e); err != nil {
		t.Fatalf("second (duplicate) ingest: %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM social_engagements WHERE external_id = 'dup-ext-id'`); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 engagement row despite the duplicate ingest, got %d", count)
	}
}

func TestIngestFromWebhook_KnownAuthor_LinksCustomer(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	identity := testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "", nil, false)

	e := entity.SocialEngagement{
		ID: newID(), Channel: entity.ChannelInstagram, Type: entity.EngagementPostComment,
		UnitID: unit.ID, Body: "oi de novo", ExternalID: "known-author-ext", AuthorExternalID: identity.ExternalID, CreatedAt: time.Now(),
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM social_engagements WHERE external_id = 'known-author-ext'`) })

	if err := uc.IngestFromWebhook(context.Background(), e); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	var customerID *string
	if err := db.Get(&customerID, `SELECT customer_id FROM social_engagements WHERE external_id = 'known-author-ext'`); err != nil {
		t.Fatalf("read customer_id: %v", err)
	}
	if customerID == nil || *customerID != customer.ID {
		t.Fatalf("expected the engagement to be linked to the existing customer %s, got %v", customer.ID, customerID)
	}
}

// --- Reply: always private ---------------------------------------------

// post_comment/live_comment replies must always go through ReplyPrivate
// (DM), never ReplyPublic — a public reply under a comment could expose a
// health question to any visitor of the post (backend/ARCHITECTURE.md §5).
func TestReply_UsesReplyPrivate_NeverReplyPublic(t *testing.T) {
	db := testutil.DB(t)
	uc, igMock := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	e := newEngagementFixture(t, db, unit.ID, entity.EngagementOpen)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	if _, err := uc.Reply(context.Background(), e.ID, model.ReplyRequest{Body: "Custa R$150!"}, access); err != nil {
		t.Fatalf("reply: %v", err)
	}

	sent := igMock.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected exactly 1 send recorded on the mock client, got %d", len(sent))
	}
	if sent[0].Kind != "reply_private" {
		t.Fatalf("expected the reply to use reply_private, got %q", sent[0].Kind)
	}

	var status string
	if err := db.Get(&status, `SELECT status FROM social_engagements WHERE id = $1`, e.ID); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != string(entity.EngagementReplied) {
		t.Fatalf("expected status=replied, got %q", status)
	}
}

func TestReply_AlreadyTreated_Conflict(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	e := newEngagementFixture(t, db, unit.ID, entity.EngagementDismissed)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	_, err := uc.Reply(context.Background(), e.ID, model.ReplyRequest{Body: "oi"}, access)
	if err == nil {
		t.Fatal("expected replying to an already-dismissed engagement to fail")
	}
	if status := respStatus(t, err); status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

// --- Dismiss -------------------------------------------------------------

func TestDismiss_OpenEngagement_Succeeds(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	e := newEngagementFixture(t, db, unit.ID, entity.EngagementOpen)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	result, err := uc.Dismiss(context.Background(), e.ID, access)
	if err != nil {
		t.Fatalf("dismiss: %v", err)
	}
	if result.Status != string(entity.EngagementDismissed) {
		t.Fatalf("expected status=dismissed, got %q", result.Status)
	}
}

func TestDismiss_AlreadyTreated_Conflict(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	e := newEngagementFixture(t, db, unit.ID, entity.EngagementReplied)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	_, err := uc.Dismiss(context.Background(), e.ID, access)
	if err == nil {
		t.Fatal("expected dismissing an already-replied engagement to fail")
	}
	if status := respStatus(t, err); status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

// --- Convert ---------------------------------------------------------------

// Convert must create the conversation in mode:"human" (not ai_triage) — an
// agent explicitly pulled this engagement into the queue, so it shouldn't
// bounce back to the AI (backend/ARCHITECTURE.md §6).
func TestConvert_CreatesConversationInHumanMode(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	e := newEngagementFixture(t, db, unit.ID, entity.EngagementOpen)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	result, err := uc.Convert(context.Background(), e.ID, access)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if result.ConversationID == nil {
		t.Fatal("expected a conversation_id to be set after convert")
	}
	t.Cleanup(func() {
		var customerID string
		db.Get(&customerID, `SELECT customer_id FROM conversations WHERE id = $1`, *result.ConversationID)
		db.Exec(`DELETE FROM customers WHERE id = $1`, customerID)
	})

	var mode string
	if err := db.Get(&mode, `SELECT mode FROM conversations WHERE id = $1`, *result.ConversationID); err != nil {
		t.Fatalf("read mode: %v", err)
	}
	if mode != string(entity.ModeHuman) {
		t.Fatalf("expected mode=human, got %q", mode)
	}
}

func TestConvert_AlreadyConverted_Conflict(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	e := newEngagementFixture(t, db, unit.ID, entity.EngagementOpen)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	first, err := uc.Convert(context.Background(), e.ID, access)
	if err != nil {
		t.Fatalf("first convert: %v", err)
	}
	t.Cleanup(func() {
		var customerID string
		db.Get(&customerID, `SELECT customer_id FROM conversations WHERE id = $1`, *first.ConversationID)
		db.Exec(`DELETE FROM customers WHERE id = $1`, customerID)
	})

	_, err = uc.Convert(context.Background(), e.ID, access)
	if err == nil {
		t.Fatal("expected converting an already-converted engagement to fail")
	}
	if status := respStatus(t, err); status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

// --- Unit scoping ------------------------------------------------------

func TestGet_OutsideUnitScope_NotFound(t *testing.T) {
	db := testutil.DB(t)
	uc, _ := newUseCase(t, db)
	unitA := testutil.NewUnit(t, db)
	unitB := testutil.NewUnit(t, db)
	e := newEngagementFixture(t, db, unitB.ID, entity.EngagementOpen)

	access := usecase.Access{Role: string(entity.RoleAgent), UnitIDs: []string{unitA.ID}}
	_, err := uc.Get(context.Background(), e.ID, access)
	if err == nil {
		t.Fatal("expected out-of-scope engagement to 404")
	}
	if status := respStatus(t, err); status != 404 {
		t.Fatalf("expected 404, got %d", status)
	}
}
