package usecase_test

import (
	"context"
	"testing"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/template/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/template/usecase"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
)

// fakeRepo is an in-memory stand-in for Repository — variable-count
// validation is pure business logic (no SQL involved), so it doesn't need
// a real Postgres to verify, unlike most other usecase tests in this repo.
type fakeRepo struct {
	created *entity.MessageTemplate
	updated *entity.MessageTemplate
}

func (f *fakeRepo) List(ctx context.Context, category string) ([]entity.MessageTemplate, error) {
	return nil, nil
}
func (f *fakeRepo) GetByID(ctx context.Context, id string) (entity.MessageTemplate, error) {
	if f.updated != nil && f.updated.ID == id {
		return *f.updated, nil
	}
	return entity.MessageTemplate{}, apperrors.NewNotFoundError("template")
}
func (f *fakeRepo) Create(ctx context.Context, t entity.MessageTemplate) error {
	f.created = &t
	f.updated = &t
	return nil
}
func (f *fakeRepo) Update(ctx context.Context, t entity.MessageTemplate) error {
	f.updated = &t
	return nil
}
func (f *fakeRepo) Delete(ctx context.Context, id string) error { return nil }

func respStatus(t *testing.T, err error) int {
	t.Helper()
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok {
		t.Fatalf("expected *apperrors.ResponseError, got %T (%v)", err, err)
	}
	return respErr.StatusCode
}

// TestCreate_VariableCountMismatch_Rejected covers the guide's own rejection
// checklist (WhatsApp_API_Optimization_Guide.md §4): "Deixar variáveis sem
// contexto" / declared count not matching {{n}} placeholders is exactly the
// kind of mistake that gets a template rejected by the Meta — catching it
// at creation time is the whole point of Frente B's catalog existing.
func TestCreate_VariableCountMismatch_Rejected(t *testing.T) {
	uc := usecase.New(&fakeRepo{})

	_, err := uc.Create(context.Background(), model.CreateTemplateRequest{
		Name:          "confirmacao_agendamento",
		Category:      "utility",
		Body:          "Confirmado para {{1}} às {{2}}!",
		VariableCount: 1, // body actually has two placeholders
	})
	if err == nil {
		t.Fatal("expected an error for a variable_count/placeholder mismatch")
	}
	if status := respStatus(t, err); status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}
}

func TestCreate_VariableCountMatches_Accepted(t *testing.T) {
	repo := &fakeRepo{}
	uc := usecase.New(repo)

	tpl, err := uc.Create(context.Background(), model.CreateTemplateRequest{
		Name:          "confirmacao_agendamento",
		Category:      "utility",
		Body:          "Confirmado para {{1}} às {{2}}, unidade {{3}}!",
		VariableCount: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.ApprovalStatus != string(entity.TemplateApprovalPending) {
		t.Fatalf("expected a freshly created template to start as pending, got %q", tpl.ApprovalStatus)
	}
	if repo.created == nil {
		t.Fatal("expected repo.Create to have been called")
	}
}

// TestUpdate_BodyChangedWithoutVariableCount_RevalidatesAgainstStoredCount
// makes sure a body-only edit can't silently desync from the declared
// variable_count — the mismatch is caught using the ALREADY-stored count,
// not just on creation.
func TestUpdate_BodyChangedWithoutVariableCount_RevalidatesAgainstStoredCount(t *testing.T) {
	existing := entity.MessageTemplate{
		ID: "tpl-1", Name: "faq_documentos", Category: entity.TemplateCategoryUtility,
		LanguageCode: "pt_BR", Body: "Aceitamos RG ou CNH.", VariableCount: 0,
		ApprovalStatus: entity.TemplateApprovalApproved, Active: true,
	}
	repo := &fakeRepo{updated: &existing}
	uc := usecase.New(repo)

	newBody := "Aceitamos {{1}} ou {{2}}."
	_, err := uc.Update(context.Background(), "tpl-1", model.UpdateTemplateRequest{Body: &newBody})
	if err == nil {
		t.Fatal("expected an error: body now has 2 placeholders but stored variable_count is still 0")
	}
	if status := respStatus(t, err); status != 400 {
		t.Fatalf("expected 400, got %d", status)
	}
}
