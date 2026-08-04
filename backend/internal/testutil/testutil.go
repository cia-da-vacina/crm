// Package testutil provides fixtures for integration tests that run against
// a real Postgres instance — the database wiring, table constraints (unique
// indexes, CHECKs, FKs) and idempotency behavior described throughout
// backend/ARCHITECTURE.md are only genuinely verified by hitting the real
// database, not a mock. Tests using this package are meant to run inside the
// api container (`docker compose exec api go test ./...`), where
// DATABASE_URL already points at the real dev Postgres and every secret env
// var (JWT_SECRET, APP_ENCRYPTION_KEY, PASSWORD_HASH_PEPPER) is already
// loaded — see backend/Makefile's `test` target.
//
// Every fixture helper inserts a row and registers a t.Cleanup that deletes
// it by id. Callers must create fixtures in parent-before-child order
// (Unit, then Customer, then Conversation, then Message/FollowUp/...) —
// t.Cleanup runs LIFO, so children are deleted before their parents
// automatically as long as creation order is respected. Deleting a row a
// second time (already gone via ON DELETE CASCADE from an explicit parent
// delete) is a no-op, not an error.
package testutil

import (
	"testing"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/domain/vo"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/cia-da-vacina/crm/backend/pkg/env"
	"github.com/google/uuid"
)

// TestPassword is the plaintext password behind every user fixture's hash —
// use it directly in login tests instead of re-hashing per test.
const TestPassword = "Test1234!"

// DB opens a real connection to the database pointed at by DATABASE_URL
// (or TEST_DATABASE_URL, if set to something different). Skips the test
// (not fails) when no database is reachable, so `go test ./...` degrades
// gracefully outside the docker network instead of failing hard.
func DB(t *testing.T) *database.DB {
	t.Helper()
	dsn := env.GetOrDefault("TEST_DATABASE_URL", env.GetOrDefault("DATABASE_URL", ""))
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}
	db, err := database.NewDB(dsn)
	if err != nil {
		t.Skipf("cannot connect to test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("cannot ping test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newID() string { return uuid.Must(uuid.NewV7()).String() }

// suffix returns a short unique token for building collision-free emails,
// unit codes, external_ids, etc. across concurrently-registered fixtures.
func suffix() string { return newID()[24:] }

// cleanupRow deletes are best-effort but NOT silent: under concurrent load
// (the full `go test ./...` suite hitting the same dev Postgres from many
// packages back-to-back) a transient connection/lock hiccup can occasionally
// make a single DELETE fail. That's rare and harmless (an orphaned empty
// fixture row, never a real leak of dependent data — every child row this
// package creates has ITS OWN cleanup registered too), but staying silent
// about it would hide a real signal, so failures are logged instead of
// swallowed. Run `DELETE FROM units WHERE name LIKE 'Test Unit%'` (and the
// equivalent for other fixture-prefixed tables) if debris ever accumulates.
func cleanupRow(t *testing.T, db *database.DB, table, id string) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM `+table+` WHERE id = $1`, id); err != nil {
			t.Logf("testutil: cleanup of %s/%s failed (best-effort, non-fatal): %v", table, id, err)
		}
	})
}

// NewUnit inserts a fresh unit fixture.
func NewUnit(t *testing.T, db *database.DB) entity.Unit {
	t.Helper()
	now := time.Now()
	u := entity.Unit{
		ID: newID(), Name: "Test Unit " + suffix(), Code: "TEST-" + suffix(),
		Timezone: "America/Sao_Paulo", Active: true,
		Address: "Rua de Teste, 123", City: "Porto Alegre",
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NamedExec(`
		INSERT INTO units (id, name, code, timezone, active, address, city, created_at, updated_at)
		VALUES (:id, :name, :code, :timezone, :active, :address, :city, :created_at, :updated_at)
	`, u); err != nil {
		t.Fatalf("testutil: create unit: %v", err)
	}
	cleanupRow(t, db, "units", u.ID)
	return u
}

// NewUser inserts a user fixture (password hash of TestPassword) and links
// it to the given units via user_unit_relation.
func NewUser(t *testing.T, db *database.DB, role entity.UserRole, unitIDs ...string) entity.User {
	t.Helper()
	hash, err := vo.HashPassword(TestPassword, vo.DefaultPasswordConfig())
	if err != nil {
		t.Fatalf("testutil: hash test password: %v", err)
	}
	now := time.Now()
	u := entity.User{
		ID: newID(), Email: "test-" + suffix() + "@example.test", PasswordHash: hash,
		Name: "Test User", Role: role, Active: true, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NamedExec(`
		INSERT INTO users (id, email, password_hash, name, role, active, created_at, updated_at)
		VALUES (:id, :email, :password_hash, :name, :role, :active, :created_at, :updated_at)
	`, u); err != nil {
		t.Fatalf("testutil: create user: %v", err)
	}
	cleanupRow(t, db, "users", u.ID)

	for _, unitID := range unitIDs {
		if _, err := db.Exec(`INSERT INTO user_unit_relation (id, user_id, unit_id) VALUES ($1, $2, $3)`,
			newID(), u.ID, unitID); err != nil {
			t.Fatalf("testutil: link user to unit: %v", err)
		}
	}
	return u
}

// SetUserActive flips users.active — used to exercise the "inactive user
// can't log in" rule.
func SetUserActive(t *testing.T, db *database.DB, userID string, active bool) {
	t.Helper()
	if _, err := db.Exec(`UPDATE users SET active = $1 WHERE id = $2`, active, userID); err != nil {
		t.Fatalf("testutil: set user active: %v", err)
	}
}

// NewCustomer inserts a customer fixture.
func NewCustomer(t *testing.T, db *database.DB, unitID *string, identification entity.CustomerIdentification, phone *string) entity.Customer {
	t.Helper()
	now := time.Now()
	c := entity.Customer{
		ID: newID(), DisplayName: "Test Customer " + suffix(), Identification: identification,
		PrimaryPhone: phone, UnitID: unitID, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NamedExec(`
		INSERT INTO customers (id, display_name, identification, primary_phone, unit_id, created_at, updated_at)
		VALUES (:id, :display_name, :identification, :primary_phone, :unit_id, :created_at, :updated_at)
	`, c); err != nil {
		t.Fatalf("testutil: create customer: %v", err)
	}
	cleanupRow(t, db, "customers", c.ID)
	return c
}

// NewCustomerIdentity inserts a customer_identities fixture. externalID, if
// empty, gets a random unique value — pass one explicitly when the test
// needs to address a specific wa_id/IGSID/PSID (e.g. to send a webhook
// payload that must resolve to this identity).
func NewCustomerIdentity(t *testing.T, db *database.DB, customerID string, channel entity.Channel, externalID string, phone *string, verified bool) entity.CustomerIdentity {
	t.Helper()
	if externalID == "" {
		externalID = "ext-" + suffix()
	}
	now := time.Now()
	ci := entity.CustomerIdentity{
		ID: newID(), CustomerID: customerID, Channel: channel, ExternalID: externalID,
		PhoneE164: phone, CreatedAt: now,
	}
	if verified {
		ci.VerifiedAt = &now
	}
	if _, err := db.NamedExec(`
		INSERT INTO customer_identities (id, customer_id, channel, external_id, display_handle, phone_e164, verified_at, created_at)
		VALUES (:id, :customer_id, :channel, :external_id, :display_handle, :phone_e164, :verified_at, :created_at)
	`, ci); err != nil {
		t.Fatalf("testutil: create customer identity: %v", err)
	}
	cleanupRow(t, db, "customer_identities", ci.ID)
	return ci
}

// ConversationOpts configures NewConversation — zero values fall back to
// sane defaults (em_atendimento / ai_triage / not_needed).
type ConversationOpts struct {
	Channel       entity.Channel
	PipelineStage entity.PipelineStage
	Mode          entity.ConversationMode
	PhoneGate     entity.PhoneGate
	OwnerID       *string
}

// NewConversation inserts a conversation fixture directly (no module owns a
// generic "create conversation for any state" repository method — usecases
// only create conversations in the specific states their own flow produces
// — so tests need their own fixture writer to set up arbitrary starting
// states like "already claimed" or "pending_verification").
func NewConversation(t *testing.T, db *database.DB, customerID, unitID string, opts ConversationOpts) entity.Conversation {
	t.Helper()
	if opts.Channel == "" {
		opts.Channel = entity.ChannelWhatsApp
	}
	if opts.PipelineStage == "" {
		opts.PipelineStage = entity.StageEmAtendimento
	}
	if opts.Mode == "" {
		opts.Mode = entity.ModeAITriage
	}
	if opts.PhoneGate == "" {
		opts.PhoneGate = entity.PhoneGateNotNeeded
	}
	now := time.Now()
	c := entity.Conversation{
		ID: newID(), CustomerID: customerID, Channel: opts.Channel, UnitID: unitID,
		PipelineStage: opts.PipelineStage, Mode: opts.Mode, PhoneGate: opts.PhoneGate,
		OwnerID: opts.OwnerID, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NamedExec(`
		INSERT INTO conversations (id, customer_id, channel, unit_id, pipeline_stage, mode, phone_gate, owner_id,
		                            last_message_preview, created_at, updated_at)
		VALUES (:id, :customer_id, :channel, :unit_id, :pipeline_stage, :mode, :phone_gate, :owner_id,
		        '', :created_at, :updated_at)
	`, c); err != nil {
		t.Fatalf("testutil: create conversation: %v", err)
	}
	cleanupRow(t, db, "conversations", c.ID)
	// conversation/usecase.Claim and .UpdatePipeline both audit-log with
	// ResourceID = conversation id (pkg/audit is append-only in production —
	// no test should leave synthetic entries mixed into real audit history).
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM audit_logs WHERE resource_id = $1`, c.ID) })
	return c
}

// NewMessage inserts a message fixture (e.g. to seed an "out" message so
// unread_count computation has something to compare against).
func NewMessage(t *testing.T, db *database.DB, conversationID string, direction entity.MessageDirection, metaMessageID *string) entity.Message {
	t.Helper()
	now := time.Now()
	m := entity.Message{
		ID: newID(), ConversationID: conversationID, Direction: direction, SenderType: entity.SenderContact,
		Kind: entity.KindText, Channel: entity.ChannelWhatsApp, Body: "test message",
		Status: entity.MessageStatusDelivered, MetaMessageID: metaMessageID, CreatedAt: now,
	}
	if direction == entity.DirectionOut {
		m.SenderType = entity.SenderAgent
	}
	if _, err := db.NamedExec(`
		INSERT INTO messages (id, conversation_id, direction, sender_type, kind, channel, body, status, meta_message_id, created_at)
		VALUES (:id, :conversation_id, :direction, :sender_type, :kind, :channel, :body, :status, :meta_message_id, :created_at)
	`, m); err != nil {
		t.Fatalf("testutil: create message: %v", err)
	}
	cleanupRow(t, db, "messages", m.ID)
	return m
}

// NewPendingVerification inserts a phone_verifications fixture. code is the
// plaintext OTP code — the fixture stores its SHA-256 hash, same as
// production (see conversation/usecase/phone.go hashOTPCode), so tests can
// exercise ConfirmPhoneVerification with a known code.
func NewPendingVerification(t *testing.T, db *database.DB, conversationID, phoneE164, codeHash string, expiresAt time.Time) entity.PhoneVerification {
	t.Helper()
	pv := entity.PhoneVerification{
		ID: newID(), ConversationID: conversationID, PhoneE164: phoneE164, CodeHash: codeHash,
		ExpiresAt: expiresAt, CreatedAt: time.Now(),
	}
	if _, err := db.NamedExec(`
		INSERT INTO phone_verifications (id, conversation_id, phone_e164, code_hash, attempts, resend_count, expires_at)
		VALUES (:id, :conversation_id, :phone_e164, :code_hash, :attempts, :resend_count, :expires_at)
	`, pv); err != nil {
		t.Fatalf("testutil: create pending verification: %v", err)
	}
	cleanupRow(t, db, "phone_verifications", pv.ID)
	return pv
}

// NewMetaChannelConfig inserts a meta_channel_configs fixture directly —
// bypasses the metasettings usecase's whatsapp/unit_id CHECK-mirroring
// validation on purpose, so callers must respect the same rule the DB CHECK
// enforces (unit_id required iff channel == whatsapp) or the insert itself
// will fail, which is exactly what the CHECK-constraint tests want to see.
func NewMetaChannelConfig(t *testing.T, db *database.DB, channel entity.Channel, unitID *string, phoneNumberID *string) entity.MetaChannelConfig {
	t.Helper()
	now := time.Now()
	cfg := entity.MetaChannelConfig{
		ID: newID(), Channel: channel, UnitID: unitID, Enabled: true,
		AccountID: "acct-" + suffix(), DisplayName: "Test Channel", PhoneNumberID: phoneNumberID,
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NamedExec(`
		INSERT INTO meta_channel_configs (id, channel, unit_id, enabled, account_id, display_name, phone_number_id, created_at, updated_at)
		VALUES (:id, :channel, :unit_id, :enabled, :account_id, :display_name, :phone_number_id, :created_at, :updated_at)
	`, cfg); err != nil {
		t.Fatalf("testutil: create meta channel config: %v", err)
	}
	cleanupRow(t, db, "meta_channel_configs", cfg.ID)
	return cfg
}

// SnapshotAppSettings reads the current (singleton, id=1) app_settings row
// and registers a t.Cleanup that restores it exactly — app_settings is
// shared mutable global state across the whole database, so any test that
// mutates it (directly, or indirectly via a usecase) MUST call this first,
// and tests using it must not run with t.Parallel() or alongside other
// packages that touch it (run the suite with `go test -p 1` — see
// backend/Makefile).
func SnapshotAppSettings(t *testing.T, db *database.DB) entity.AppSettings {
	t.Helper()
	var s entity.AppSettings
	if err := db.Get(&s, `
		SELECT id, ai_enabled, ai_system_prompt, ai_context, triage_enabled, triage_handoff_intents, default_unit_id, updated_at
		FROM app_settings WHERE id = 1
	`); err != nil {
		t.Fatalf("testutil: snapshot app_settings: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.NamedExec(`
			UPDATE app_settings SET ai_enabled = :ai_enabled, ai_system_prompt = :ai_system_prompt,
				ai_context = :ai_context, triage_enabled = :triage_enabled,
				triage_handoff_intents = :triage_handoff_intents, default_unit_id = :default_unit_id,
				updated_at = :updated_at
			WHERE id = 1
		`, s)
	})
	return s
}

// SetAppSettings overwrites the singleton row — call SnapshotAppSettings
// first so it gets restored after the test.
func SetAppSettings(t *testing.T, db *database.DB, s entity.AppSettings) {
	t.Helper()
	s.ID = 1
	s.UpdatedAt = time.Now()
	if _, err := db.NamedExec(`
		UPDATE app_settings SET ai_enabled = :ai_enabled, ai_system_prompt = :ai_system_prompt,
			ai_context = :ai_context, triage_enabled = :triage_enabled,
			triage_handoff_intents = :triage_handoff_intents, default_unit_id = :default_unit_id,
			updated_at = :updated_at
		WHERE id = 1
	`, s); err != nil {
		t.Fatalf("testutil: set app_settings: %v", err)
	}
}
