package usecase_test

import (
	"context"
	"testing"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/metasettings/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/audit"
	"github.com/cia-da-vacina/crm/backend/pkg/crypto"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

// testCipher uses a fixed, test-only AES-256 key (64 hex chars) — never
// touches APP_ENCRYPTION_KEY from the real environment, so these tests are
// self-contained regardless of what secret the container happens to have.
func testCipher(t *testing.T) *crypto.Cipher {
	t.Setenv("APP_ENCRYPTION_KEY", "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1")
	c, err := crypto.NewCipherFromEnv()
	if err != nil {
		t.Fatalf("new test cipher: %v", err)
	}
	return c
}

func newUseCase(t *testing.T, db *database.DB) *usecase.UseCase {
	return usecase.New(repository.New(db), testCipher(t), audit.New(db))
}

func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }

// --- Channel CHECK constraint mirroring ------------------------------------

// WhatsApp without unit_id must be rejected by the usecase itself — not
// just rely on the DB CHECK constraint to 500 it (backend/ARCHITECTURE.md
// §5: whatsapp is always per-unit).
func TestUpdate_WhatsAppWithoutUnitID_Rejected(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)

	_, err := uc.Update(context.Background(), model.UpdateSettingsRequest{
		Channels: []model.ChannelUpdateItem{{Channel: "whatsapp", Token: strPtr("some-token")}},
	}, "00000000-0000-7000-8000-000000000001")
	if err == nil {
		t.Fatal("expected whatsapp channel update without unit_id to be rejected")
	}
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok || respErr.StatusCode != 400 {
		t.Fatalf("expected 400 ResponseError, got %T: %v", err, err)
	}
}

// Instagram/Facebook are centralized — setting unit_id on them must be
// rejected, not silently accepted as if it were per-unit.
func TestUpdate_CentralizedChannelWithUnitID_Rejected(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)

	_, err := uc.Update(context.Background(), model.UpdateSettingsRequest{
		Channels: []model.ChannelUpdateItem{{Channel: "instagram", UnitID: &unit.ID, Token: strPtr("tok")}},
	}, "00000000-0000-7000-8000-000000000001")
	if err == nil {
		t.Fatal("expected instagram channel update WITH unit_id to be rejected")
	}
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok || respErr.StatusCode != 400 {
		t.Fatalf("expected 400 ResponseError, got %T: %v", err, err)
	}
}

func TestUpdate_WhatsAppWithUnitID_CreatesPerUnitConfig(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM meta_channel_configs WHERE channel = 'whatsapp' AND unit_id = $1`, unit.ID)
	})

	settings, err := uc.Update(context.Background(), model.UpdateSettingsRequest{
		Channels: []model.ChannelUpdateItem{{
			Channel: "whatsapp", UnitID: &unit.ID, Enabled: boolPtr(true),
			DisplayName: strPtr("Unidade Teste WhatsApp"), Token: strPtr("EAAGtoken1234567890abcdef"),
		}},
	}, "00000000-0000-7000-8000-000000000001")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	found := false
	for _, ch := range settings.Channels {
		if ch.Channel == "whatsapp" && ch.UnitID != nil && *ch.UnitID == unit.ID {
			found = true
			if ch.TokenMasked == nil {
				t.Fatal("expected token_masked to be populated after setting a token")
			}
			if *ch.TokenMasked == "EAAGtoken1234567890abcdef" {
				t.Fatal("token_masked must never equal the raw token")
			}
		}
	}
	if !found {
		t.Fatal("expected the new per-unit whatsapp channel config to appear in the settings response")
	}
}

// --- Token encryption round-trip -------------------------------------------

// The plaintext token must never be recoverable from what's returned by the
// API (only token_masked) — but the ciphertext stored in the DB must
// decrypt back to the exact original plaintext, proving the encrypt path is
// correct and not just "some bytes got stored".
func TestUpdate_TokenEncryption_RoundTripsThroughCipher(t *testing.T) {
	db := testutil.DB(t)
	cipher := testCipher(t)
	uc := usecase.New(repository.New(db), cipher, audit.New(db))
	unit := testutil.NewUnit(t, db)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM meta_channel_configs WHERE channel = 'whatsapp' AND unit_id = $1`, unit.ID)
	})

	const plaintext = "EAAG_super_secret_meta_token_value"
	if _, err := uc.Update(context.Background(), model.UpdateSettingsRequest{
		Channels: []model.ChannelUpdateItem{{Channel: "whatsapp", UnitID: &unit.ID, Token: strPtr(plaintext)}},
	}, "00000000-0000-7000-8000-000000000001"); err != nil {
		t.Fatalf("update: %v", err)
	}

	var ciphertext []byte
	if err := db.Get(&ciphertext, `SELECT token_ciphertext FROM meta_channel_configs WHERE channel = 'whatsapp' AND unit_id = $1`, unit.ID); err != nil {
		t.Fatalf("read ciphertext: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("expected a non-empty ciphertext to be stored")
	}

	decrypted, err := cipher.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt stored ciphertext: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("expected decrypted ciphertext to equal the original token, got %q", decrypted)
	}
}

// --- Campaign upsert semantics ---------------------------------------------

func TestUpdate_Campaign_NoID_Creates(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)

	settings, err := uc.Update(context.Background(), model.UpdateSettingsRequest{
		AICampaigns: []model.CampaignUpdateItem{{
			Title: "Campanha de Teste", StartsOn: "2026-01-01", EndsOn: "2026-12-31",
		}},
	}, "00000000-0000-7000-8000-000000000001")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	var found *model.Campaign
	for i := range settings.AICampaigns {
		if settings.AICampaigns[i].Title == "Campanha de Teste" {
			found = &settings.AICampaigns[i]
		}
	}
	if found == nil {
		t.Fatal("expected the new campaign to appear in the settings response")
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM ai_campaigns WHERE id = $1`, found.ID) })
}

// Updating with an unknown campaign ID must 404, not silently create one
// with a client-supplied ID (model.go comment: "evita criar silenciosamente").
func TestUpdate_Campaign_UnknownID_NotFound(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)

	fakeID := "00000000-0000-7000-8000-000000000000"
	_, err := uc.Update(context.Background(), model.UpdateSettingsRequest{
		AICampaigns: []model.CampaignUpdateItem{{
			ID: &fakeID, Title: "Nao Deveria Existir", StartsOn: "2026-01-01", EndsOn: "2026-12-31",
		}},
	}, "00000000-0000-7000-8000-000000000001")
	if err == nil {
		t.Fatal("expected update with an unknown campaign id to fail")
	}
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok || respErr.StatusCode != 404 {
		t.Fatalf("expected 404, got %T: %v", err, err)
	}
}

// --- Audit trail -------------------------------------------------------

func TestUpdate_WritesAuditLogEntry(t *testing.T) {
	db := testutil.DB(t)
	testutil.SnapshotAppSettings(t, db) // Update mutates the ai_enabled singleton below — must be restored.
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	actor := testutil.NewUser(t, db, entity.RoleAdmin, unit.ID)

	before := auditCountForActor(t, db, actor.ID)

	if _, err := uc.Update(context.Background(), model.UpdateSettingsRequest{AIEnabled: boolPtr(true)}, actor.ID); err != nil {
		t.Fatalf("update: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM audit_logs WHERE actor_user_id = $1`, actor.ID) })

	after := auditCountForActor(t, db, actor.ID)
	if after != before+1 {
		t.Fatalf("expected exactly 1 new audit log entry for settings.meta.update, got %d -> %d", before, after)
	}
}

func auditCountForActor(t *testing.T, db *database.DB, actorID string) int {
	t.Helper()
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM audit_logs WHERE action = 'settings.meta.update' AND actor_user_id = $1`, actorID); err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	return count
}
