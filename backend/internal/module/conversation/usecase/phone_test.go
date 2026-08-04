package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
)

func hashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

func agentAccess(userID, unitID string) usecase.Access {
	return usecase.Access{UserID: userID, Role: string(entity.RoleAgent), UnitIDs: []string{unitID}}
}

// --- Initiate --------------------------------------------------------------

func TestInitiatePhoneVerification_SetsGateAndSendsOTP(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "ig-"+customer.ID, nil, false)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{
		Channel: entity.ChannelInstagram, PhoneGate: entity.PhoneGateRequired,
	})
	t.Cleanup(func() { db.Exec(`DELETE FROM phone_verifications WHERE conversation_id = $1`, conv.ID) })

	access := agentAccess(agent.ID, unit.ID)
	detail, err := uc.InitiatePhoneVerification(context.Background(), conv.ID, model.StartPhoneVerificationRequest{PhoneE164: "+5551999990000"}, access)
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	if detail.PhoneGate != string(entity.PhoneGatePendingVerification) {
		t.Fatalf("expected phone_gate=pending_verification, got %q", detail.PhoneGate)
	}
	if detail.PendingPhoneMasked == nil {
		t.Fatal("expected pending_phone_masked to be set")
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM phone_verifications WHERE conversation_id = $1 AND confirmed_at IS NULL`, conv.ID); err != nil {
		t.Fatalf("count pending verifications: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 pending verification row, got %d", count)
	}
}

// --- Confirm: wrong code / lockout / expiry --------------------------------

func TestConfirmPhoneVerification_WrongCode_IncrementsAttempts(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "ig-"+customer.ID, nil, false)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{
		Channel: entity.ChannelInstagram, PhoneGate: entity.PhoneGatePendingVerification,
	})
	pv := testutil.NewPendingVerification(t, db, conv.ID, "+5551999990000", hashCode("111111"), time.Now().Add(10*time.Minute))

	access := agentAccess(agent.ID, unit.ID)
	_, err := uc.ConfirmPhoneVerification(context.Background(), conv.ID, model.ConfirmPhoneVerificationRequest{Code: "000000"}, access)
	if err == nil {
		t.Fatal("expected wrong code to be rejected")
	}
	if status := respStatus(t, err); status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}

	var attempts int
	if err := db.Get(&attempts, `SELECT attempts FROM phone_verifications WHERE id = $1`, pv.ID); err != nil {
		t.Fatalf("read attempts: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected attempts=1 after one wrong guess, got %d", attempts)
	}
}

// After otpMaxAttempts (5) wrong guesses in a row, the pending verification
// self-heals: it's deleted and phone_gate reverts to required — never left
// stuck in pending_verification forever (docs/BACKEND-CONTRACT.md §3).
func TestConfirmPhoneVerification_MaxAttemptsExceeded_RevertsToRequired(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "ig-"+customer.ID, nil, false)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{
		Channel: entity.ChannelInstagram, PhoneGate: entity.PhoneGatePendingVerification,
	})
	pv := testutil.NewPendingVerification(t, db, conv.ID, "+5551999990000", hashCode("111111"), time.Now().Add(10*time.Minute))
	t.Cleanup(func() { db.Exec(`DELETE FROM phone_verifications WHERE id = $1`, pv.ID) })

	access := agentAccess(agent.ID, unit.ID)
	req := model.ConfirmPhoneVerificationRequest{Code: "000000"}
	var lastErr error
	for i := 0; i < 5; i++ {
		_, lastErr = uc.ConfirmPhoneVerification(context.Background(), conv.ID, req, access)
	}
	if lastErr == nil {
		t.Fatal("expected the 5th wrong attempt to return an error")
	}
	if status := respStatus(t, lastErr); status != 400 {
		t.Fatalf("expected 400 on lockout, got %d", status)
	}

	var gate string
	if err := db.Get(&gate, `SELECT phone_gate FROM conversations WHERE id = $1`, conv.ID); err != nil {
		t.Fatalf("read phone_gate: %v", err)
	}
	if gate != string(entity.PhoneGateRequired) {
		t.Fatalf("expected phone_gate to revert to required after lockout, got %q", gate)
	}

	var remaining int
	if err := db.Get(&remaining, `SELECT COUNT(*) FROM phone_verifications WHERE id = $1`, pv.ID); err != nil {
		t.Fatalf("count remaining pending verification: %v", err)
	}
	if remaining != 0 {
		t.Fatal("expected the pending verification row to be deleted after lockout")
	}
}

func TestConfirmPhoneVerification_Expired_RevertsToRequired(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "ig-"+customer.ID, nil, false)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{
		Channel: entity.ChannelInstagram, PhoneGate: entity.PhoneGatePendingVerification,
	})
	pv := testutil.NewPendingVerification(t, db, conv.ID, "+5551999990000", hashCode("111111"), time.Now().Add(-1*time.Minute))
	t.Cleanup(func() { db.Exec(`DELETE FROM phone_verifications WHERE id = $1`, pv.ID) })

	access := agentAccess(agent.ID, unit.ID)
	_, err := uc.ConfirmPhoneVerification(context.Background(), conv.ID, model.ConfirmPhoneVerificationRequest{Code: "111111"}, access)
	if err == nil {
		t.Fatal("expected expired code to be rejected even if correct")
	}
	if status := respStatus(t, err); status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}

	var gate string
	if err := db.Get(&gate, `SELECT phone_gate FROM conversations WHERE id = $1`, conv.ID); err != nil {
		t.Fatalf("read phone_gate: %v", err)
	}
	if gate != string(entity.PhoneGateRequired) {
		t.Fatalf("expected phone_gate to revert to required after expiry, got %q", gate)
	}
}

// --- Confirm: promote (no existing customer with that phone) ---------------

func TestConfirmPhoneVerification_NoExistingCustomer_PromotesInPlace(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "ig-"+customer.ID, nil, false)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{
		Channel: entity.ChannelInstagram, PhoneGate: entity.PhoneGatePendingVerification,
	})
	phone := "+5551988887777"
	pv := testutil.NewPendingVerification(t, db, conv.ID, phone, hashCode("123456"), time.Now().Add(10*time.Minute))
	t.Cleanup(func() { db.Exec(`DELETE FROM phone_verifications WHERE id = $1`, pv.ID) })

	access := agentAccess(agent.ID, unit.ID)
	detail, err := uc.ConfirmPhoneVerification(context.Background(), conv.ID, model.ConfirmPhoneVerificationRequest{Code: "123456"}, access)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if detail.PhoneGate != string(entity.PhoneGateCollected) {
		t.Fatalf("expected phone_gate=collected, got %q", detail.PhoneGate)
	}
	if detail.CustomerID != customer.ID {
		t.Fatalf("expected the same customer to be promoted in place (id=%s), got %s", customer.ID, detail.CustomerID)
	}

	var identification, primaryPhone string
	if err := db.QueryRow(`SELECT identification, primary_phone FROM customers WHERE id = $1`, customer.ID).Scan(&identification, &primaryPhone); err != nil {
		t.Fatalf("read customer: %v", err)
	}
	if identification != string(entity.IdentificationIdentified) {
		t.Fatalf("expected customer to become identified, got %q", identification)
	}
	if primaryPhone != phone {
		t.Fatalf("expected primary_phone=%s, got %s", phone, primaryPhone)
	}
}

// --- Confirm: merge (phone already belongs to another identified customer) -

// This is the cross-channel unification case: a customer already exists as
// "identified" via WhatsApp with primary_phone X. The SAME phone gets
// confirmed via OTP from an Instagram conversation belonging to a DIFFERENT
// (anonymous) customer record. mergeOrPromote must fold the Instagram
// customer into the WhatsApp one: identities and conversations reparented,
// the source customer row deleted, and the conversation now points at the
// canonical (WhatsApp) customer.
func TestConfirmPhoneVerification_ExistingCustomerWithPhone_MergesRecords(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)

	phone := "+5551977776666"
	canonical := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationIdentified, &phone)
	testutil.NewCustomerIdentity(t, db, canonical.ID, entity.ChannelWhatsApp, "wa-"+canonical.ID, &phone, true)
	canonicalConv := testutil.NewConversation(t, db, canonical.ID, unit.ID, testutil.ConversationOpts{Channel: entity.ChannelWhatsApp})

	anon := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, anon.ID, entity.ChannelInstagram, "ig-"+anon.ID, nil, false)
	igConv := testutil.NewConversation(t, db, anon.ID, unit.ID, testutil.ConversationOpts{
		Channel: entity.ChannelInstagram, PhoneGate: entity.PhoneGatePendingVerification,
	})
	pv := testutil.NewPendingVerification(t, db, igConv.ID, phone, hashCode("654321"), time.Now().Add(10*time.Minute))
	t.Cleanup(func() { db.Exec(`DELETE FROM phone_verifications WHERE id = $1`, pv.ID) })

	access := agentAccess(agent.ID, unit.ID)
	detail, err := uc.ConfirmPhoneVerification(context.Background(), igConv.ID, model.ConfirmPhoneVerificationRequest{Code: "654321"}, access)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if detail.CustomerID != canonical.ID {
		t.Fatalf("expected conversation to be reparented onto the canonical customer %s, got %s", canonical.ID, detail.CustomerID)
	}

	// The anonymous (source) customer must be gone — merged, not duplicated.
	var sourceCount int
	if err := db.Get(&sourceCount, `SELECT COUNT(*) FROM customers WHERE id = $1`, anon.ID); err != nil {
		t.Fatalf("count source customer: %v", err)
	}
	if sourceCount != 0 {
		t.Fatal("expected the source (anonymous) customer row to be deleted after merge")
	}

	// Both conversations (the pre-existing WhatsApp one and the just-
	// confirmed Instagram one) must now belong to the canonical customer.
	var waCustomerID string
	if err := db.Get(&waCustomerID, `SELECT customer_id FROM conversations WHERE id = $1`, canonicalConv.ID); err != nil {
		t.Fatalf("read whatsapp conversation: %v", err)
	}
	if waCustomerID != canonical.ID {
		t.Fatalf("expected pre-existing whatsapp conversation to remain on canonical customer, got %s", waCustomerID)
	}

	// The Instagram identity must have been reparented onto the canonical
	// customer (not left dangling on the deleted source row).
	var igIdentityCustomerID string
	if err := db.Get(&igIdentityCustomerID, `SELECT customer_id FROM customer_identities WHERE customer_id = $1 AND channel = 'instagram'`, canonical.ID); err != nil {
		t.Fatalf("expected instagram identity to be reparented onto canonical customer: %v", err)
	}

	_ = igIdentityCustomerID
	t.Cleanup(func() { db.Exec(`DELETE FROM conversations WHERE id = $1`, canonicalConv.ID) })
}

// --- Resend ------------------------------------------------------------

func TestResendPhoneVerification_NoActivePending_Conflict(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "ig-"+customer.ID, nil, false)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{Channel: entity.ChannelInstagram})

	access := agentAccess(agent.ID, unit.ID)
	_, err := uc.ResendPhoneVerification(context.Background(), conv.ID, access)
	if err == nil {
		t.Fatal("expected resend without an active pending verification to fail")
	}
	if status := respStatus(t, err); status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

// After otpMaxResends (3), further resends must be rate-limited (429), not
// silently keep extending the OTP indefinitely.
func TestResendPhoneVerification_ExceedsLimit_TooManyRequests(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelInstagram, "ig-"+customer.ID, nil, false)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{
		Channel: entity.ChannelInstagram, PhoneGate: entity.PhoneGatePendingVerification,
	})
	pv := testutil.NewPendingVerification(t, db, conv.ID, "+5551999990000", hashCode("111111"), time.Now().Add(10*time.Minute))
	t.Cleanup(func() { db.Exec(`DELETE FROM phone_verifications WHERE id = $1`, pv.ID) })

	access := agentAccess(agent.ID, unit.ID)
	var lastErr error
	for i := 0; i < 4; i++ {
		_, lastErr = uc.ResendPhoneVerification(context.Background(), conv.ID, access)
	}
	if lastErr == nil {
		t.Fatal("expected the 4th resend (limit is 3) to be rejected")
	}
	if status := respStatus(t, lastErr); status != 429 {
		t.Fatalf("expected 429 too many requests, got %d", status)
	}
}
