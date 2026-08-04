package usecase_test

import (
	"context"
	"testing"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	custrepo "github.com/cia-da-vacina/crm/backend/internal/module/customer/repository"
	custusecase "github.com/cia-da-vacina/crm/backend/internal/module/customer/usecase"

	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/usecase"
	pricingrepo "github.com/cia-da-vacina/crm/backend/internal/module/pricing/repository"
	pricingusecase "github.com/cia-da-vacina/crm/backend/internal/module/pricing/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/audit"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
)

func newUseCase(t *testing.T, db *database.DB) *usecase.UseCase {
	repo := repository.New(db)
	customerReader := custusecase.New(custrepo.New(db))
	registry := meta.NewRegistry()
	registry.Register(meta.NewMockClient(meta.ChannelWhatsApp))
	registry.Register(meta.NewMockClient(meta.ChannelInstagram))
	registry.Register(meta.NewMockClient(meta.ChannelFacebook))
	pricingReader := pricingusecase.New(pricingrepo.New(db))
	return usecase.New(repo, customerReader, nil, registry, audit.New(db), pricingReader)
}

func respStatus(t *testing.T, err error) int {
	t.Helper()
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok {
		t.Fatalf("expected *apperrors.ResponseError, got %T (%v)", err, err)
	}
	return respErr.StatusCode
}

// --- Claim -------------------------------------------------------------

func TestClaim_FreeConversation_AgentClaimsSuccessfully(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	detail, err := uc.Claim(context.Background(), conv.ID, access)
	if err != nil {
		t.Fatalf("expected claim to succeed on a free conversation, got: %v", err)
	}
	if detail.OwnerID == nil || *detail.OwnerID != agent.ID {
		t.Fatalf("expected owner_id=%s, got %v", agent.ID, detail.OwnerID)
	}
	if detail.Mode != string(entity.ModeHuman) {
		t.Fatalf("expected claim to flip mode to human, got %q", detail.Mode)
	}
}

// A second regular agent trying to claim an already-claimed conversation
// must get 409 — this is the "atendente ausente" guard rail; only
// admin/manager/supervisor may override it (see the next test).
func TestClaim_AlreadyClaimedByAnotherAgent_Conflict(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agentA := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	agentB := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	ownerID := agentA.ID
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{OwnerID: &ownerID, Mode: entity.ModeHuman})

	access := usecase.Access{UserID: agentB.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	_, err := uc.Claim(context.Background(), conv.ID, access)
	if err == nil {
		t.Fatal("expected 409 when a second agent claims an already-owned conversation")
	}
	if status := respStatus(t, err); status != 409 {
		t.Fatalf("expected 409 conflict, got %d", status)
	}
}

// The agent who already owns it re-claiming is a harmless no-op, not a
// conflict (e.g. UI double-submit).
func TestClaim_SameAgentReClaims_NoConflict(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	ownerID := agent.ID
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{OwnerID: &ownerID, Mode: entity.ModeHuman})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	if _, err := uc.Claim(context.Background(), conv.ID, access); err != nil {
		t.Fatalf("expected the owning agent to re-claim without conflict, got: %v", err)
	}
}

// Manager/supervisor/admin can reassign an already-claimed conversation —
// confirmed product decision (backend/ARCHITECTURE.md §5, "gerente realoca
// atendente ausente"). Table-driven over the three privileged roles.
func TestClaim_PrivilegedRoles_CanReassignAlreadyClaimedConversation(t *testing.T) {
	for _, role := range []entity.UserRole{entity.RoleManager, entity.RoleSupervisor, entity.RoleAdmin} {
		t.Run(string(role), func(t *testing.T) {
			db := testutil.DB(t)
			uc := newUseCase(t, db)
			unit := testutil.NewUnit(t, db)
			originalOwner := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
			reassigner := testutil.NewUser(t, db, role, unit.ID)
			customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
			ownerID := originalOwner.ID
			conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{OwnerID: &ownerID, Mode: entity.ModeHuman})

			access := usecase.Access{UserID: reassigner.ID, Role: string(role), UnitIDs: []string{unit.ID}}
			detail, err := uc.Claim(context.Background(), conv.ID, access)
			if err != nil {
				t.Fatalf("expected %s to reassign successfully, got: %v", role, err)
			}
			if detail.OwnerID == nil || *detail.OwnerID != reassigner.ID {
				t.Fatalf("expected owner_id=%s after reassignment, got %v", reassigner.ID, detail.OwnerID)
			}
		})
	}
}

// A conversation outside the agent's units must 404, not 403 — leaking
// "this exists but you can't touch it" would confirm existence to someone
// who shouldn't even know about it (see usecase.go Get comment).
func TestClaim_OutsideUnitScope_NotFound(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unitA := testutil.NewUnit(t, db)
	unitB := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unitA.ID)
	customer := testutil.NewCustomer(t, db, &unitB.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unitB.ID, testutil.ConversationOpts{})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unitA.ID}}
	_, err := uc.Claim(context.Background(), conv.ID, access)
	if err == nil {
		t.Fatal("expected out-of-scope claim to fail")
	}
	if status := respStatus(t, err); status != 404 {
		t.Fatalf("expected 404 (not 403) for out-of-scope conversation, got %d", status)
	}
}

// Admin has no unit restriction at all.
func TestClaim_Admin_CanClaimAnyUnit(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	admin := testutil.NewUser(t, db, entity.RoleAdmin)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})

	access := usecase.Access{UserID: admin.ID, Role: string(entity.RoleAdmin)}
	if _, err := uc.Claim(context.Background(), conv.ID, access); err != nil {
		t.Fatalf("expected admin to claim across units, got: %v", err)
	}
}

// --- UpdatePipeline ------------------------------------------------------

func TestUpdatePipeline_NaoFechado_RequiresReasonCode(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	_, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{Stage: string(entity.StageNaoFechado)}, access)
	if err == nil {
		t.Fatal("expected missing reason_code to be rejected for nao_fechado")
	}
	if status := respStatus(t, err); status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}
}

func TestUpdatePipeline_NaoFechado_UnknownReasonCode_Rejected(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	_, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{
		Stage: string(entity.StageNaoFechado), ReasonCode: "not-a-real-code",
	}, access)
	if err == nil {
		t.Fatal("expected unknown reason_code to be rejected")
	}
	if status := respStatus(t, err); status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}
}

func TestUpdatePipeline_InvalidStage_Rejected(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	_, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{Stage: "not_a_real_stage"}, access)
	if err == nil {
		t.Fatal("expected invalid pipeline stage to be rejected")
	}
	if status := respStatus(t, err); status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}
}

// Any-to-any transition is allowed (no imposed order) — confirmed product
// decision, backend/ARCHITECTURE.md §5.
func TestUpdatePipeline_AnyToAnyTransitionAllowed(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{PipelineStage: entity.StageFechado})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	// fechado -> em_atendimento: "backwards" transition, no sequence enforced.
	detail, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{Stage: string(entity.StageEmAtendimento)}, access)
	if err != nil {
		t.Fatalf("expected free transition between valid stages, got: %v", err)
	}
	if detail.PipelineStage != string(entity.StageEmAtendimento) {
		t.Fatalf("expected stage em_atendimento, got %q", detail.PipelineStage)
	}
}

// Moving into aguardando_fechamento must create an open follow-up.
func TestUpdatePipeline_AguardandoFechamento_CreatesFollowUp(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	if _, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{Stage: string(entity.StageAguardandoFechamento)}, access); err != nil {
		t.Fatalf("update pipeline: %v", err)
	}

	count := countOpenFollowUps(t, db, conv.ID)
	if count != 1 {
		t.Fatalf("expected exactly 1 open follow-up after moving to aguardando_fechamento, got %d", count)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM follow_ups WHERE conversation_id = $1`, conv.ID) })
}

// Oscillating between the two follow-up-triggering stages must NOT stack
// duplicate open follow-ups — the partial unique index +
// CreateFollowUpIfNotOpen contract (backend/ARCHITECTURE.md §5).
func TestUpdatePipeline_OscillatingStages_DoesNotDuplicateFollowUp(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})
	t.Cleanup(func() { db.Exec(`DELETE FROM follow_ups WHERE conversation_id = $1`, conv.ID) })

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}

	if _, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{Stage: string(entity.StageAguardandoFechamento)}, access); err != nil {
		t.Fatalf("1st transition: %v", err)
	}
	if _, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{
		Stage: string(entity.StageNaoFechado), ReasonCode: "preco",
	}, access); err != nil {
		t.Fatalf("2nd transition: %v", err)
	}
	if _, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{Stage: string(entity.StageAguardandoFechamento)}, access); err != nil {
		t.Fatalf("3rd transition: %v", err)
	}

	count := countOpenFollowUps(t, db, conv.ID)
	if count != 1 {
		t.Fatalf("expected oscillating between stages to leave exactly 1 open follow-up, got %d", count)
	}
}

// Moving OUT of nao_fechado must clear the loss reason fields — otherwise a
// stale reason_code would linger on a conversation that's no longer lost.
func TestUpdatePipeline_LeavingNaoFechado_ClearsLossReason(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{})
	t.Cleanup(func() { db.Exec(`DELETE FROM follow_ups WHERE conversation_id = $1`, conv.ID) })

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	if _, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{
		Stage: string(entity.StageNaoFechado), ReasonCode: "preco",
	}, access); err != nil {
		t.Fatalf("move to nao_fechado: %v", err)
	}

	var code *string
	if err := db.Get(&code, `SELECT loss_reason_code FROM conversations WHERE id = $1`, conv.ID); err != nil {
		t.Fatalf("read loss_reason_code: %v", err)
	}
	if code == nil || *code != "preco" {
		t.Fatalf("expected loss_reason_code=preco to be persisted, got %v", code)
	}

	if _, err := uc.UpdatePipeline(context.Background(), conv.ID, model.UpdatePipelineRequest{Stage: string(entity.StageEmNegociacao)}, access); err != nil {
		t.Fatalf("move back to em_negociacao: %v", err)
	}
	if err := db.Get(&code, `SELECT loss_reason_code FROM conversations WHERE id = $1`, conv.ID); err != nil {
		t.Fatalf("read loss_reason_code after leaving nao_fechado: %v", err)
	}
	if code != nil {
		t.Fatalf("expected loss_reason_code to be cleared after leaving nao_fechado, got %v", *code)
	}
}

func countOpenFollowUps(t *testing.T, db *database.DB, conversationID string) int {
	t.Helper()
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM follow_ups WHERE conversation_id = $1 AND status = 'open'`, conversationID); err != nil {
		t.Fatalf("count open follow-ups: %v", err)
	}
	return count
}

// --- SendMessage ---------------------------------------------------------

// Sending is only allowed once a human has claimed the conversation
// (mode=human) — sending into an ai_triage conversation must 409 regardless
// of what the UI already blocks (docs/BACKEND-CONTRACT.md §4).
func TestSendMessage_AITriageMode_Conflict(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{Mode: entity.ModeAITriage})
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelWhatsApp, "", nil, false)

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	_, err := uc.SendMessage(context.Background(), conv.ID, model.SendMessageRequest{Body: "oi"}, access)
	if err == nil {
		t.Fatal("expected sending into an ai_triage conversation to be rejected")
	}
	if status := respStatus(t, err); status != 409 {
		t.Fatalf("expected 409, got %d", status)
	}
}

func TestSendMessage_HumanMode_Succeeds(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{Mode: entity.ModeHuman})
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelWhatsApp, "", nil, false)
	t.Cleanup(func() { db.Exec(`DELETE FROM messages WHERE conversation_id = $1`, conv.ID) })

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	msg, err := uc.SendMessage(context.Background(), conv.ID, model.SendMessageRequest{Body: "oi, tudo bem?"}, access)
	if err != nil {
		t.Fatalf("expected send to succeed in human mode, got: %v", err)
	}
	if msg.Direction != string(entity.DirectionOut) || msg.SenderType != string(entity.SenderAgent) {
		t.Fatalf("unexpected message shape: %+v", msg)
	}
}

// TestSendMessage_EstimatesServiceCostFromRateCard covers Frente A of the
// WhatsApp 2026 adaptation plan: free-form text sent by an agent/AI is
// always the "service" pricing category (guide's Regra 1 — no more free
// in-window replies from Oct/2026), and its cost is looked up from
// message_pricing_rates (migration 000021 seeds it, not hardcoded here) —
// not confirmed yet (pricing_confirmed=false) since no real Meta status
// webhook exists to reconcile it against.
func TestSendMessage_EstimatesServiceCostFromRateCard(t *testing.T) {
	db := testutil.DB(t)
	uc := newUseCase(t, db)
	unit := testutil.NewUnit(t, db)
	agent := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	customer := testutil.NewCustomer(t, db, &unit.ID, entity.IdentificationAnonymous, nil)
	conv := testutil.NewConversation(t, db, customer.ID, unit.ID, testutil.ConversationOpts{Mode: entity.ModeHuman})
	testutil.NewCustomerIdentity(t, db, customer.ID, entity.ChannelWhatsApp, "", nil, false)
	t.Cleanup(func() { db.Exec(`DELETE FROM messages WHERE conversation_id = $1`, conv.ID) })

	var expectedRate float64
	if err := db.Get(&expectedRate, `SELECT rate_brl FROM message_pricing_rates WHERE category = 'service'`); err != nil {
		t.Fatalf("read seeded service rate: %v", err)
	}

	access := usecase.Access{UserID: agent.ID, Role: string(entity.RoleAgent), UnitIDs: []string{unit.ID}}
	msg, err := uc.SendMessage(context.Background(), conv.ID, model.SendMessageRequest{Body: "seu exame está pronto"}, access)
	if err != nil {
		t.Fatalf("expected send to succeed, got: %v", err)
	}

	if msg.PricingCategory == nil || *msg.PricingCategory != string(entity.PricingService) {
		t.Fatalf("expected pricing_category=service, got %v", msg.PricingCategory)
	}
	if msg.CostBRL == nil || *msg.CostBRL != expectedRate {
		t.Fatalf("expected cost_brl=%v (from rate card), got %v", expectedRate, msg.CostBRL)
	}
	if msg.PricingConfirmed {
		t.Fatal("expected pricing_confirmed=false for a locally-estimated cost (no real Meta status webhook yet)")
	}
}
