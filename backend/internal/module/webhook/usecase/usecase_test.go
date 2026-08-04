package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	engagementrepo "github.com/cia-da-vacina/crm/backend/internal/module/engagement/repository"
	engagementusecase "github.com/cia-da-vacina/crm/backend/internal/module/engagement/usecase"
	pricingrepo "github.com/cia-da-vacina/crm/backend/internal/module/pricing/repository"
	pricingusecase "github.com/cia-da-vacina/crm/backend/internal/module/pricing/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
)

// noopTriage stands in for the real triage usecase — AI triage behavior
// (phone_gate/handoff rules, prompt building) is already covered by
// internal/module/triage/usecase/usecase_test.go; webhook tests only need
// to confirm triage gets INVOKED at the right point, not what it decides,
// so a no-op keeps these tests focused on ingestion (identity/routing/
// idempotency) instead of coupling them to AI response parsing.
type noopTriage struct{ calls int }

func (n *noopTriage) RunTriage(ctx context.Context, conversationID string) error {
	n.calls++
	return nil
}

func newUseCase(t *testing.T, db *database.DB, triage *noopTriage) *usecase.UseCase {
	repo := repository.New(db)
	registry := meta.NewRegistry()
	registry.Register(meta.NewMockClient(meta.ChannelWhatsApp))
	registry.Register(meta.NewMockClient(meta.ChannelInstagram))
	registry.Register(meta.NewMockClient(meta.ChannelFacebook))
	engagementUC := engagementusecase.New(engagementrepo.New(db), registry, nil)
	pricingUC := pricingusecase.New(pricingrepo.New(db))
	return usecase.New(repo, triage, engagementUC, nil, pricingUC)
}

func whatsappTextPayload(phoneNumberID, waID, profileName, messageID, body string, ts time.Time) []byte {
	return []byte(fmt.Sprintf(`{
		"entry": [{"changes": [{"value": {
			"metadata": {"phone_number_id": %q},
			"contacts": [{"profile": {"name": %q}, "wa_id": %q}],
			"messages": [{"from": %q, "id": %q, "timestamp": %q, "type": "text", "text": {"body": %q}}]
		}}]}]
	}`, phoneNumberID, profileName, waID, waID, messageID, fmt.Sprintf("%d", ts.Unix()), body))
}

func messagingTextPayload(senderID, mid, body string, ts time.Time) []byte {
	return []byte(fmt.Sprintf(`{
		"entry": [{"messaging": [{
			"sender": {"id": %q},
			"timestamp": %d,
			"message": {"mid": %q, "text": %q, "is_echo": false}
		}]}]
	}`, senderID, ts.UnixMilli(), mid, body))
}

func storyReplyPayload(senderID, mid, storyID, body string, ts time.Time) []byte {
	return []byte(fmt.Sprintf(`{
		"entry": [{"messaging": [{
			"sender": {"id": %q},
			"timestamp": %d,
			"message": {"mid": %q, "text": %q, "is_echo": false, "reply_to": {"story": {"id": %q, "url": "https://example.test/story"}}}
		}]}]
	}`, senderID, ts.UnixMilli(), mid, body, storyID))
}

func commentPayload(fromID, commentID, mediaID, body string, ts time.Time) []byte {
	return []byte(fmt.Sprintf(`{
		"entry": [{"time": %d, "changes": [{"field": "comments", "value": {
			"from": {"id": %q}, "media": {"id": %q}, "id": %q, "text": %q
		}}]}]
	}`, ts.Unix(), fromID, mediaID, commentID, body))
}

// --- WhatsApp: identity resolution + routing + idempotency -----------------

func TestIngestPayload_WhatsApp_NewSender_CreatesIdentifiedCustomerAndConversation(t *testing.T) {
	db := testutil.DB(t)
	testutil.SnapshotAppSettings(t, db)
	unit := testutil.NewUnit(t, db)
	cfg := testutil.NewMetaChannelConfig(t, db, entity.ChannelWhatsApp, &unit.ID, strPtr("phone-"+unit.ID))

	triage := &noopTriage{}
	uc := newUseCase(t, db, triage)

	waID := "55519" + unit.ID[:8]
	payload := whatsappTextPayload(*cfg.PhoneNumberID, waID, "Fulano de Tal", "wamid.TEST1", "Oi, quero agendar", time.Now())
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelWhatsApp, waID) })

	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, payload); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	var customerID, identification, primaryPhone string
	if err := db.QueryRow(`
		SELECT cu.id, cu.identification, cu.primary_phone
		FROM customer_identities ci JOIN customers cu ON cu.id = ci.customer_id
		WHERE ci.channel = 'whatsapp' AND ci.external_id = $1
	`, waID).Scan(&customerID, &identification, &primaryPhone); err != nil {
		t.Fatalf("expected a customer to be created for the new whatsapp sender: %v", err)
	}
	if identification != string(entity.IdentificationIdentified) {
		t.Fatalf("expected whatsapp customer to be born identified, got %q", identification)
	}
	if primaryPhone != "+"+waID {
		t.Fatalf("expected primary_phone=+%s, got %s", waID, primaryPhone)
	}

	var convUnitID, mode string
	var windowExpiresAt time.Time
	if err := db.QueryRow(`SELECT unit_id, mode, window_expires_at FROM conversations WHERE customer_id = $1`, customerID).
		Scan(&convUnitID, &mode, &windowExpiresAt); err != nil {
		t.Fatalf("expected a conversation to be created: %v", err)
	}
	if convUnitID != unit.ID {
		t.Fatalf("expected conversation routed to unit %s (via phone_number_id), got %s", unit.ID, convUnitID)
	}
	if windowExpiresAt.IsZero() {
		t.Fatal("expected window_expires_at to be set (24h service window)")
	}

	if triage.calls != 1 {
		t.Fatalf("expected RunTriage to be invoked once after the inbound message, got %d calls", triage.calls)
	}
}

func TestIngestPayload_WhatsApp_DuplicateMetaMessageID_Idempotent(t *testing.T) {
	db := testutil.DB(t)
	testutil.SnapshotAppSettings(t, db)
	unit := testutil.NewUnit(t, db)
	cfg := testutil.NewMetaChannelConfig(t, db, entity.ChannelWhatsApp, &unit.ID, strPtr("phone-"+unit.ID))
	uc := newUseCase(t, db, &noopTriage{})

	waID := "55519" + unit.ID[:8]
	payload := whatsappTextPayload(*cfg.PhoneNumberID, waID, "Fulano", "wamid.DUPLICATE", "mensagem repetida", time.Now())
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelWhatsApp, waID) })

	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, payload); err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, payload); err != nil {
		t.Fatalf("second (duplicate) ingest: %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM messages WHERE meta_message_id = 'wamid.DUPLICATE'`); err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 message despite the webhook being replayed, got %d", count)
	}
}

func TestIngestPayload_WhatsApp_SecondMessageSameSender_ReusesConversation(t *testing.T) {
	db := testutil.DB(t)
	testutil.SnapshotAppSettings(t, db)
	unit := testutil.NewUnit(t, db)
	cfg := testutil.NewMetaChannelConfig(t, db, entity.ChannelWhatsApp, &unit.ID, strPtr("phone-"+unit.ID))
	uc := newUseCase(t, db, &noopTriage{})

	waID := "55519" + unit.ID[:8]
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelWhatsApp, waID) })

	p1 := whatsappTextPayload(*cfg.PhoneNumberID, waID, "Fulano", "wamid.MSG1", "primeira mensagem", time.Now())
	p2 := whatsappTextPayload(*cfg.PhoneNumberID, waID, "Fulano", "wamid.MSG2", "segunda mensagem", time.Now().Add(time.Minute))

	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, p1); err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, p2); err != nil {
		t.Fatalf("second ingest: %v", err)
	}

	var convCount int
	if err := db.Get(&convCount, `
		SELECT COUNT(*) FROM conversations c JOIN customer_identities ci ON ci.customer_id = c.customer_id
		WHERE ci.channel = 'whatsapp' AND ci.external_id = $1
	`, waID); err != nil {
		t.Fatalf("count conversations: %v", err)
	}
	if convCount != 1 {
		t.Fatalf("expected both messages to land in the SAME conversation, found %d conversations", convCount)
	}

	var msgCount int
	if err := db.Get(&msgCount, `
		SELECT COUNT(*) FROM messages m JOIN conversations c ON c.id = m.conversation_id
		JOIN customer_identities ci ON ci.customer_id = c.customer_id
		WHERE ci.channel = 'whatsapp' AND ci.external_id = $1
	`, waID); err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if msgCount != 2 {
		t.Fatalf("expected 2 messages total, got %d", msgCount)
	}
}

func TestIngestPayload_WhatsApp_UnroutedPhoneNumberID_DropsSilently(t *testing.T) {
	db := testutil.DB(t)
	testutil.SnapshotAppSettings(t, db)
	uc := newUseCase(t, db, &noopTriage{})

	waID := "5551900000000"
	payload := whatsappTextPayload("no-such-phone-number-id", waID, "Ninguem", "wamid.UNROUTED", "oi", time.Now())
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelWhatsApp, waID) })

	// IngestPayload must not return an error to the caller (the handler
	// always responds 200 to the Meta — see handler.Receive comment) even
	// though the individual message fails to route; the failure is only
	// logged.
	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, payload); err != nil {
		t.Fatalf("expected no error surfaced for an unrouted message, got: %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM customer_identities WHERE channel = 'whatsapp' AND external_id = $1`, waID); err != nil {
		t.Fatalf("count identities: %v", err)
	}
	if count != 0 {
		t.Fatal("expected no customer to be created for a message that couldn't be routed to a unit")
	}
}

// triage_enabled=false must make new WhatsApp conversations start directly
// in human mode, bypassing AI (docs/PRODUCT-V2.md §6 kill-switch).
func TestIngestPayload_TriageDisabled_ConversationStartsInHumanMode(t *testing.T) {
	db := testutil.DB(t)
	settings := testutil.SnapshotAppSettings(t, db)
	settings.TriageEnabled = false
	testutil.SetAppSettings(t, db, settings)

	unit := testutil.NewUnit(t, db)
	cfg := testutil.NewMetaChannelConfig(t, db, entity.ChannelWhatsApp, &unit.ID, strPtr("phone-"+unit.ID))
	uc := newUseCase(t, db, &noopTriage{})

	waID := "55519" + unit.ID[:8]
	payload := whatsappTextPayload(*cfg.PhoneNumberID, waID, "Fulano", "wamid.NOTRIAGE", "oi", time.Now())
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelWhatsApp, waID) })

	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, payload); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	var mode string
	if err := db.Get(&mode, `
		SELECT c.mode FROM conversations c JOIN customer_identities ci ON ci.customer_id = c.customer_id
		WHERE ci.channel = 'whatsapp' AND ci.external_id = $1
	`, waID); err != nil {
		t.Fatalf("read conversation mode: %v", err)
	}
	if mode != string(entity.ModeHuman) {
		t.Fatalf("expected mode=human with triage_enabled=false, got %q", mode)
	}
}

// --- Instagram/Facebook: centralized routing --------------------------------

func TestIngestPayload_Instagram_RoutesToDefaultUnit_AnonymousCustomer(t *testing.T) {
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	settings := testutil.SnapshotAppSettings(t, db)
	settings.DefaultUnitID = &unit.ID
	testutil.SetAppSettings(t, db, settings)

	uc := newUseCase(t, db, &noopTriage{})

	igsid := "igsid-" + unit.ID[:8]
	payload := messagingTextPayload(igsid, "mid.IG1", "quero saber sobre vacina", time.Now())
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelInstagram, igsid) })

	if err := uc.IngestPayload(context.Background(), entity.ChannelInstagram, payload); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	var identification, convUnitID string
	if err := db.QueryRow(`
		SELECT cu.identification, c.unit_id
		FROM customer_identities ci
		JOIN customers cu ON cu.id = ci.customer_id
		JOIN conversations c ON c.customer_id = cu.id
		WHERE ci.channel = 'instagram' AND ci.external_id = $1
	`, igsid).Scan(&identification, &convUnitID); err != nil {
		t.Fatalf("expected customer+conversation to be created: %v", err)
	}
	if identification != string(entity.IdentificationAnonymous) {
		t.Fatalf("expected instagram customer to be born anonymous, got %q", identification)
	}
	if convUnitID != unit.ID {
		t.Fatalf("expected conversation routed to default_unit_id=%s, got %s", unit.ID, convUnitID)
	}
}

func TestIngestPayload_Instagram_NoDefaultUnitConfigured_DropsSilently(t *testing.T) {
	db := testutil.DB(t)
	settings := testutil.SnapshotAppSettings(t, db)
	settings.DefaultUnitID = nil
	testutil.SetAppSettings(t, db, settings)

	uc := newUseCase(t, db, &noopTriage{})

	igsid := "igsid-nodefault"
	payload := messagingTextPayload(igsid, "mid.IGNODEFAULT", "oi", time.Now())
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelInstagram, igsid) })

	if err := uc.IngestPayload(context.Background(), entity.ChannelInstagram, payload); err != nil {
		t.Fatalf("expected no error surfaced, got: %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM customer_identities WHERE channel = 'instagram' AND external_id = $1`, igsid); err != nil {
		t.Fatalf("count identities: %v", err)
	}
	if count != 0 {
		t.Fatal("expected the message to be dropped (not create a customer) when default_unit_id is unset")
	}
}

// --- Engagements: story reply / comment idempotency -------------------------

// A DM that is actually a reply to a story must NOT become a 1:1 Message —
// it becomes a SocialEngagement instead (docs/BACKEND-CONTRACT.md §5).
func TestIngestPayload_Instagram_StoryReply_CreatesEngagementNotMessage(t *testing.T) {
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	settings := testutil.SnapshotAppSettings(t, db)
	settings.DefaultUnitID = &unit.ID
	testutil.SetAppSettings(t, db, settings)

	uc := newUseCase(t, db, &noopTriage{})

	senderID := "igsid-story-" + unit.ID[:8]
	payload := storyReplyPayload(senderID, "mid.STORY1", "story-abc", "adorei o story!", time.Now())
	t.Cleanup(func() {
		db.Exec(`DELETE FROM social_engagements WHERE channel = 'instagram' AND external_id = 'mid.STORY1'`)
		cleanupByExternalID(db, entity.ChannelInstagram, senderID)
	})

	if err := uc.IngestPayload(context.Background(), entity.ChannelInstagram, payload); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	var engType string
	if err := db.Get(&engType, `SELECT type FROM social_engagements WHERE channel = 'instagram' AND external_id = 'mid.STORY1'`); err != nil {
		t.Fatalf("expected a story_reply engagement to be created: %v", err)
	}
	if engType != string(entity.EngagementStoryReply) {
		t.Fatalf("expected type=story_reply, got %q", engType)
	}

	var msgCount int
	if err := db.Get(&msgCount, `SELECT COUNT(*) FROM messages WHERE meta_message_id = 'mid.STORY1'`); err != nil {
		t.Fatalf("count messages: %v", err)
	}
	if msgCount != 0 {
		t.Fatal("expected the story reply to NOT also be created as a regular 1:1 message")
	}
}

func TestIngestPayload_Comment_DuplicateExternalID_Idempotent(t *testing.T) {
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	settings := testutil.SnapshotAppSettings(t, db)
	settings.DefaultUnitID = &unit.ID
	testutil.SetAppSettings(t, db, settings)

	uc := newUseCase(t, db, &noopTriage{})

	payload := commentPayload("author-"+unit.ID[:8], "comment-DUP", "media-1", "quanto custa a vacina?", time.Now())
	t.Cleanup(func() {
		db.Exec(`DELETE FROM social_engagements WHERE channel = 'instagram' AND external_id = 'comment-DUP'`)
	})

	if err := uc.IngestPayload(context.Background(), entity.ChannelInstagram, payload); err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	if err := uc.IngestPayload(context.Background(), entity.ChannelInstagram, payload); err != nil {
		t.Fatalf("second (duplicate) ingest: %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM social_engagements WHERE channel = 'instagram' AND external_id = 'comment-DUP'`); err != nil {
		t.Fatalf("count engagements: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 engagement despite the webhook being replayed, got %d", count)
	}
}

func whatsappStatusPayload(metaMessageID, status, category string, ts time.Time) []byte {
	return []byte(fmt.Sprintf(`{
		"entry": [{"changes": [{"value": {
			"statuses": [{
				"id": %q, "status": %q, "timestamp": %q,
				"pricing": {"billable": true, "pricing_model": "CBP", "category": %q}
			}]
		}}]}]
	}`, metaMessageID, status, fmt.Sprintf("%d", ts.Unix()), category))
}

// TestIngestPayload_WhatsApp_StatusWithPricing_ReconcilesMessageCost covers
// Frente A of the WhatsApp 2026 adaptation plan: once a real Meta status
// webhook exists, it should overwrite the LOCAL estimate (see
// conversation/usecase.applyEstimatedPricing) with what the Meta actually
// confirmed, and flip pricing_confirmed to true.
func TestIngestPayload_WhatsApp_StatusWithPricing_ReconcilesMessageCost(t *testing.T) {
	db := testutil.DB(t)
	testutil.SnapshotAppSettings(t, db)
	unit := testutil.NewUnit(t, db)
	cfg := testutil.NewMetaChannelConfig(t, db, entity.ChannelWhatsApp, &unit.ID, strPtr("phone-"+unit.ID))
	uc := newUseCase(t, db, &noopTriage{})

	waID := "55519" + unit.ID[:8]
	t.Cleanup(func() { cleanupByExternalID(db, entity.ChannelWhatsApp, waID) })

	// Uma mensagem inbound primeiro, só pra existir uma linha em messages —
	// o evento de status abaixo referencia um meta_message_id de uma
	// mensagem OUTBOUND fictícia (o backend não teria como ter mandado essa
	// mensagem via mock e recebido um status real dela no mesmo teste, já
	// que não existe client real — ver ARCHITECTURE.md §5), então o teste
	// insere a mensagem outbound direto, simulando o que SendMessage teria
	// criado.
	inbound := whatsappTextPayload(*cfg.PhoneNumberID, waID, "Fulano", "wamid.INBOUND1", "oi", time.Now())
	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, inbound); err != nil {
		t.Fatalf("inbound ingest: %v", err)
	}
	var conversationID string
	if err := db.Get(&conversationID, `SELECT conversation_id FROM messages WHERE meta_message_id = 'wamid.INBOUND1'`); err != nil {
		t.Fatalf("find conversation: %v", err)
	}
	outboundID := "wamid.OUTBOUND1"
	if _, err := db.Exec(`
		INSERT INTO messages (id, conversation_id, direction, sender_type, kind, channel, body, status, meta_message_id, created_at,
		                       pricing_category, pricing_confirmed, cost_brl)
		VALUES (gen_random_uuid(), $1, 'out', 'agent', 'text', 'whatsapp', 'confirmado!', 'sent', $2, now(),
		        'service', false, 0.0350)
	`, conversationID, outboundID); err != nil {
		t.Fatalf("seed outbound message: %v", err)
	}

	statusPayload := whatsappStatusPayload(outboundID, "delivered", "utility", time.Now())
	if err := uc.IngestPayload(context.Background(), entity.ChannelWhatsApp, statusPayload); err != nil {
		t.Fatalf("status ingest: %v", err)
	}

	var expectedRate float64
	if err := db.Get(&expectedRate, `SELECT rate_brl FROM message_pricing_rates WHERE category = 'utility'`); err != nil {
		t.Fatalf("read seeded utility rate: %v", err)
	}

	var status, category string
	var confirmed bool
	var costBRL float64
	if err := db.QueryRow(`
		SELECT status, pricing_category, pricing_confirmed, cost_brl FROM messages WHERE meta_message_id = $1
	`, outboundID).Scan(&status, &category, &confirmed, &costBRL); err != nil {
		t.Fatalf("read reconciled message: %v", err)
	}
	if status != "delivered" {
		t.Fatalf("expected status=delivered, got %q", status)
	}
	if category != "utility" {
		t.Fatalf("expected pricing_category reconciled to utility (the Meta-reported category), got %q", category)
	}
	if !confirmed {
		t.Fatal("expected pricing_confirmed=true after reconciling with a status webhook that carries a pricing object")
	}
	if costBRL != expectedRate {
		t.Fatalf("expected cost_brl recomputed to the utility rate (%v), got %v", expectedRate, costBRL)
	}
}

func strPtr(s string) *string { return &s }

func cleanupByExternalID(db *database.DB, channel entity.Channel, externalID string) {
	var customerID string
	if err := db.Get(&customerID, `SELECT customer_id FROM customer_identities WHERE channel = $1 AND external_id = $2`, channel, externalID); err == nil {
		db.Exec(`DELETE FROM customers WHERE id = $1`, customerID)
	}
}
