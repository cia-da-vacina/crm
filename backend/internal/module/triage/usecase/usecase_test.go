package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/cia-da-vacina/crm/backend/pkg/openai"
)

// fakeRepo é um double em memória do Repository do usecase — permite testar
// as regras determinísticas de phone_gate/handoff sem banco real.
type fakeRepo struct {
	conv           entity.Conversation
	identification entity.CustomerIdentification
	settings       entity.AppSettings
	updated        *entity.Conversation
	createdMsgs    []entity.Message
}

func (f *fakeRepo) GetConversation(ctx context.Context, id string) (entity.Conversation, error) {
	return f.conv, nil
}
func (f *fakeRepo) UpdateConversation(ctx context.Context, c entity.Conversation) error {
	cp := c
	f.updated = &cp
	return nil
}
func (f *fakeRepo) GetCustomerIdentification(ctx context.Context, customerID string) (entity.CustomerIdentification, error) {
	return f.identification, nil
}
func (f *fakeRepo) GetCustomerIdentityExternalID(ctx context.Context, customerID string, channel entity.Channel) (string, error) {
	return "external-id-123", nil
}
func (f *fakeRepo) GetRecentMessages(ctx context.Context, conversationID string, limit int) ([]entity.Message, error) {
	return nil, nil
}
func (f *fakeRepo) CreateMessage(ctx context.Context, msg entity.Message) error {
	f.createdMsgs = append(f.createdMsgs, msg)
	return nil
}
func (f *fakeRepo) GetAppSettings(ctx context.Context) (entity.AppSettings, error) {
	return f.settings, nil
}
func (f *fakeRepo) GetActiveCampaigns(ctx context.Context) ([]entity.AICampaign, error) {
	return nil, nil
}
func (f *fakeRepo) GetSuggestedPopIDs(ctx context.Context, intent string) ([]string, error) {
	return nil, nil
}
func (f *fakeRepo) GetActivePendingVerification(ctx context.Context, conversationID string) (entity.PhoneVerification, error) {
	return entity.PhoneVerification{}, sql.ErrNoRows
}

func newTestUseCase(repo *fakeRepo, aiResponse string) *UseCase {
	ai := openai.NewMockClient()
	ai.SetResponse(aiResponse)

	registry := meta.NewRegistry()
	registry.Register(meta.NewMockClient(meta.ChannelWhatsApp))
	registry.Register(meta.NewMockClient(meta.ChannelInstagram))
	registry.Register(meta.NewMockClient(meta.ChannelFacebook))

	return New(repo, ai, registry, nil, "gpt-4o-mini")
}

const respPhoneGateRequired = `{"intent":"agendar","confidence":0.9,"summary":"quer agendar",` +
	`"phone_gate_required":true,"ready_for_handoff":true,"reply":"Pode me passar seu telefone?","collected_fields":{}}`

func TestRunTriage_WhatsApp_NeverRequiresPhone(t *testing.T) {
	repo := &fakeRepo{
		conv: entity.Conversation{
			ID: "conv-1", CustomerID: "cust-1", Channel: entity.ChannelWhatsApp,
			UnitID: "unit-1", Mode: entity.ModeAITriage, PhoneGate: entity.PhoneGateNotNeeded,
		},
		identification: entity.IdentificationIdentified,
		settings:       entity.AppSettings{AIEnabled: true},
	}
	uc := newTestUseCase(repo, respPhoneGateRequired)

	if err := uc.RunTriage(context.Background(), "conv-1"); err != nil {
		t.Fatalf("RunTriage failed: %v", err)
	}
	if repo.updated == nil {
		t.Fatal("expected conversation to be updated")
	}
	if repo.updated.PhoneGate != entity.PhoneGateNotNeeded {
		t.Fatalf("expected phone_gate to stay not_needed on WhatsApp, got %q", repo.updated.PhoneGate)
	}
}

func TestRunTriage_AnonymousIG_RequiresPhoneWhenAIAsks(t *testing.T) {
	repo := &fakeRepo{
		conv: entity.Conversation{
			ID: "conv-2", CustomerID: "cust-2", Channel: entity.ChannelInstagram,
			UnitID: "unit-1", Mode: entity.ModeAITriage, PhoneGate: entity.PhoneGateNotNeeded,
		},
		identification: entity.IdentificationAnonymous,
		settings:       entity.AppSettings{AIEnabled: true},
	}
	uc := newTestUseCase(repo, respPhoneGateRequired)

	if err := uc.RunTriage(context.Background(), "conv-2"); err != nil {
		t.Fatalf("RunTriage failed: %v", err)
	}
	if repo.updated.PhoneGate != entity.PhoneGateRequired {
		t.Fatalf("expected phone_gate=required for anonymous IG customer, got %q", repo.updated.PhoneGate)
	}
}

func TestRunTriage_AlreadyIdentifiedIG_NeverRequiresPhoneAgain(t *testing.T) {
	repo := &fakeRepo{
		conv: entity.Conversation{
			ID: "conv-3", CustomerID: "cust-3", Channel: entity.ChannelInstagram,
			UnitID: "unit-1", Mode: entity.ModeAITriage, PhoneGate: entity.PhoneGateNotNeeded,
		},
		identification: entity.IdentificationIdentified, // já confirmou telefone antes (merge)
		settings:       entity.AppSettings{AIEnabled: true},
	}
	uc := newTestUseCase(repo, respPhoneGateRequired)

	if err := uc.RunTriage(context.Background(), "conv-3"); err != nil {
		t.Fatalf("RunTriage failed: %v", err)
	}
	if repo.updated.PhoneGate != entity.PhoneGateNotNeeded {
		t.Fatalf("expected phone_gate to stay not_needed for already-identified customer, got %q", repo.updated.PhoneGate)
	}
}

func TestRunTriage_HandoffIntentsOverrideForcesReadyForHandoff(t *testing.T) {
	const respNotReady = `{"intent":"reclamacao","confidence":0.8,"summary":"reclamação",` +
		`"phone_gate_required":false,"ready_for_handoff":false,"reply":"Vou verificar.","collected_fields":{}}`

	repo := &fakeRepo{
		conv: entity.Conversation{
			ID: "conv-4", CustomerID: "cust-4", Channel: entity.ChannelWhatsApp,
			UnitID: "unit-1", Mode: entity.ModeAITriage, PhoneGate: entity.PhoneGateNotNeeded,
		},
		identification: entity.IdentificationIdentified,
		settings: entity.AppSettings{
			AIEnabled:            true,
			TriageHandoffIntents: entity.IntentTags{"reclamacao"},
		},
	}
	uc := newTestUseCase(repo, respNotReady)

	if err := uc.RunTriage(context.Background(), "conv-4"); err != nil {
		t.Fatalf("RunTriage failed: %v", err)
	}
	if !repo.updated.TriageReadyForHandoff {
		t.Fatal("expected ready_for_handoff to be forced true by triage_handoff_intents override")
	}
}

func TestRunTriage_AIDisabled_IsNoOp(t *testing.T) {
	repo := &fakeRepo{
		conv: entity.Conversation{
			ID: "conv-5", CustomerID: "cust-5", Channel: entity.ChannelWhatsApp,
			UnitID: "unit-1", Mode: entity.ModeAITriage,
		},
		settings: entity.AppSettings{AIEnabled: false},
	}
	uc := newTestUseCase(repo, respPhoneGateRequired)

	if err := uc.RunTriage(context.Background(), "conv-5"); err != nil {
		t.Fatalf("RunTriage failed: %v", err)
	}
	if repo.updated != nil {
		t.Fatal("expected no update when ai_enabled=false (kill-switch)")
	}
	if len(repo.createdMsgs) != 0 {
		t.Fatal("expected no reply message when ai_enabled=false")
	}
}

func TestRunTriage_AlreadyHuman_IsNoOp(t *testing.T) {
	repo := &fakeRepo{
		conv: entity.Conversation{
			ID: "conv-6", CustomerID: "cust-6", Channel: entity.ChannelWhatsApp,
			UnitID: "unit-1", Mode: entity.ModeHuman,
		},
		settings: entity.AppSettings{AIEnabled: true},
	}
	uc := newTestUseCase(repo, respPhoneGateRequired)

	if err := uc.RunTriage(context.Background(), "conv-6"); err != nil {
		t.Fatalf("RunTriage failed: %v", err)
	}
	if repo.updated != nil {
		t.Fatal("expected no update once a conversation has been claimed by a human")
	}
}
