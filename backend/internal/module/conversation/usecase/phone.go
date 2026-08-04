package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/conversation/model"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/meta"
	"github.com/google/uuid"
)

const (
	// otpTTL, otpMaxAttempts e otpMaxResends não são especificados
	// numericamente pelo contrato ("5-10 min", "limite de reenvios") — são
	// defaults operacionais, ajustáveis sem migração de schema.
	otpTTL          = 10 * time.Minute
	otpMaxAttempts  = 5
	otpMaxResends   = 3
	otpCodeDigits   = 6
	otpTemplateName = "phone_verification_otp"
)

// InitiatePhoneVerification valida o E.164, cria/substitui a pendência e
// dispara o template OTP via WhatsApp — nunca seta primary_phone nem
// promove o customer aqui (docs/BACKEND-CONTRACT.md §3).
func (uc *UseCase) InitiatePhoneVerification(ctx context.Context, conversationID string, req model.StartPhoneVerificationRequest, access Access) (model.ConversationDetail, error) {
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
		}
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(conv.UnitID) {
		return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
	}

	code, err := genOTPCode()
	if err != nil {
		return model.ConversationDetail{}, apperrors.NewInternalError(err)
	}

	now := time.Now()
	pv := entity.PhoneVerification{
		ID:             uuid.Must(uuid.NewV7()).String(),
		ConversationID: conversationID,
		PhoneE164:      req.PhoneE164,
		CodeHash:       hashOTPCode(code),
		Attempts:       0,
		ResendCount:    0,
		ExpiresAt:      now.Add(otpTTL),
	}
	if err := uc.repo.UpsertPendingVerification(ctx, pv); err != nil {
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	if err := uc.sendOTP(ctx, req.PhoneE164, code); err != nil {
		return model.ConversationDetail{}, apperrors.NewBadGatewayError("failed to send verification template: " + err.Error())
	}

	conv.PhoneGate = entity.PhoneGatePendingVerification
	conv.UpdatedAt = now
	if err := uc.repo.UpdateConversation(ctx, conv); err != nil {
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	uc.publish("conversation.phone_gate_changed", conv.UnitID, map[string]string{"conversation_id": conv.ID, "phone_gate": string(conv.PhoneGate)})

	return uc.Get(ctx, conversationID, access)
}

// ResendPhoneVerification reusa a pendência atual (mesmo telefone), gera um
// novo código e reseta tentativas — 409 sem pendência ativa, 429 acima do
// limite de reenvios (docs/BACKEND-CONTRACT.md §3).
func (uc *UseCase) ResendPhoneVerification(ctx context.Context, conversationID string, access Access) (model.ConversationDetail, error) {
	if err := uc.checkAccess(ctx, conversationID, access); err != nil {
		return model.ConversationDetail{}, err
	}

	pv, err := uc.repo.GetActivePendingVerification(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewConflictErrorMessage("no pending phone verification")
		}
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}
	if pv.ResendCount >= otpMaxResends {
		return model.ConversationDetail{}, apperrors.NewTooManyRequestsError("resend limit exceeded")
	}

	code, err := genOTPCode()
	if err != nil {
		return model.ConversationDetail{}, apperrors.NewInternalError(err)
	}

	pv.CodeHash = hashOTPCode(code)
	pv.Attempts = 0
	pv.ResendCount++
	pv.ExpiresAt = time.Now().Add(otpTTL)
	if err := uc.repo.UpdatePendingVerification(ctx, pv); err != nil {
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	if err := uc.sendOTP(ctx, pv.PhoneE164, code); err != nil {
		return model.ConversationDetail{}, apperrors.NewBadGatewayError("failed to send verification template: " + err.Error())
	}

	return uc.Get(ctx, conversationID, access)
}

// ConfirmPhoneVerification é o único ponto do sistema onde promoção +
// merge acontecem — só depois do código confirmado (docs/BACKEND-CONTRACT.md
// §3: "só aqui ocorre promoção + merge por primary_phone").
func (uc *UseCase) ConfirmPhoneVerification(ctx context.Context, conversationID string, req model.ConfirmPhoneVerificationRequest, access Access) (model.ConversationDetail, error) {
	conv, err := uc.repo.GetConversation(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
		}
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}
	if !access.canAccessUnit(conv.UnitID) {
		return model.ConversationDetail{}, apperrors.NewNotFoundError("conversation")
	}

	pv, err := uc.repo.GetActivePendingVerification(ctx, conversationID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return model.ConversationDetail{}, apperrors.NewConflictErrorMessage("no pending phone verification")
		}
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	now := time.Now()
	if now.After(pv.ExpiresAt) {
		if err := uc.revertToRequired(ctx, conv, pv.ID); err != nil {
			return model.ConversationDetail{}, err
		}
		return model.ConversationDetail{}, apperrors.NewBadRequestError("verification code expired")
	}

	if hashOTPCode(req.Code) != pv.CodeHash {
		pv.Attempts++
		if pv.Attempts >= otpMaxAttempts {
			if err := uc.revertToRequired(ctx, conv, pv.ID); err != nil {
				return model.ConversationDetail{}, err
			}
			return model.ConversationDetail{}, apperrors.NewBadRequestError("too many invalid attempts, verification restarted")
		}
		if err := uc.repo.UpdatePendingVerification(ctx, pv); err != nil {
			return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
		}
		return model.ConversationDetail{}, apperrors.NewBadRequestError("invalid verification code")
	}

	targetCustomerID, err := uc.mergeOrPromote(ctx, conv, pv.PhoneE164)
	if err != nil {
		return model.ConversationDetail{}, err
	}

	pv.ConfirmedAt = &now
	if err := uc.repo.UpdatePendingVerification(ctx, pv); err != nil {
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	conv.CustomerID = targetCustomerID
	conv.PhoneGate = entity.PhoneGateCollected
	conv.UpdatedAt = now
	if err := uc.repo.UpdateConversation(ctx, conv); err != nil {
		return model.ConversationDetail{}, apperrors.NewDatabaseError(err)
	}

	uc.publish("conversation.phone_gate_changed", conv.UnitID, map[string]string{"conversation_id": conv.ID, "phone_gate": string(conv.PhoneGate)})

	return uc.Get(ctx, conversationID, access)
}

// mergeOrPromote implementa docs/BACKEND-CONTRACT.md §3: se já existe um
// Customer identified com esse primary_phone, funde (reparenta identities e
// conversations pro alvo, apaga a origem); senão promove o customer atual
// no lugar. Retorna o customer_id que a conversa deve passar a referenciar.
func (uc *UseCase) mergeOrPromote(ctx context.Context, conv entity.Conversation, phoneE164 string) (string, error) {
	existing, err := uc.repo.GetCustomerByPhone(ctx, phoneE164)
	switch {
	case err == nil && existing.ID != conv.CustomerID:
		if err := uc.repo.MergeCustomerInto(ctx, conv.CustomerID, existing.ID); err != nil {
			return "", apperrors.NewDatabaseError(err)
		}
		if err := uc.repo.SetCustomerIdentityVerified(ctx, existing.ID, conv.Channel, phoneE164); err != nil {
			return "", apperrors.NewDatabaseError(err)
		}
		return existing.ID, nil

	case err == nil:
		// existing.ID == conv.CustomerID: cliente já é o próprio (ex.: re-
		// verificação) — só garante que a identidade está marcada.
		if err := uc.repo.SetCustomerIdentityVerified(ctx, conv.CustomerID, conv.Channel, phoneE164); err != nil {
			return "", apperrors.NewDatabaseError(err)
		}
		return conv.CustomerID, nil

	case apperrors.IsNotFound(err):
		if err := uc.repo.PromoteCustomerToIdentified(ctx, conv.CustomerID, phoneE164); err != nil {
			return "", apperrors.NewDatabaseError(err)
		}
		if err := uc.repo.SetCustomerIdentityVerified(ctx, conv.CustomerID, conv.Channel, phoneE164); err != nil {
			return "", apperrors.NewDatabaseError(err)
		}
		return conv.CustomerID, nil

	default:
		return "", apperrors.NewDatabaseError(err)
	}
}

// revertToRequired é o "self-heal" de TTL/tentativas estouradas: apaga a
// pendência e volta phone_gate pra required, nunca deixando a conversa presa
// em pending_verification indefinidamente (docs/BACKEND-CONTRACT.md §3).
func (uc *UseCase) revertToRequired(ctx context.Context, conv entity.Conversation, pendingID string) error {
	if err := uc.repo.DeletePendingVerification(ctx, pendingID); err != nil {
		return apperrors.NewDatabaseError(err)
	}
	conv.PhoneGate = entity.PhoneGateRequired
	conv.UpdatedAt = time.Now()
	if err := uc.repo.UpdateConversation(ctx, conv); err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// sendOTP sempre envia pelo WhatsApp, mesmo quando a conversa é de outro
// canal — é o WhatsApp que prova posse do número, não o canal de origem
// (docs/BACKEND-CONTRACT.md §3, "Por que OTP no WhatsApp").
func (uc *UseCase) sendOTP(ctx context.Context, phoneE164, code string) error {
	sender, err := uc.meta.Sender(meta.ChannelWhatsApp)
	if err != nil {
		return err
	}
	_, err = sender.SendTemplate(ctx, meta.SendTemplateInput{
		Recipient:    meta.Recipient{Channel: meta.ChannelWhatsApp, ExternalID: strings.TrimPrefix(phoneE164, "+")},
		TemplateName: otpTemplateName,
		LanguageCode: "pt_BR",
		Params:       []string{code},
	})
	return err
}

func genOTPCode() (string, error) {
	max := big.NewInt(1)
	for range otpCodeDigits {
		max.Mul(max, big.NewInt(10))
	}
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate otp code: %w", err)
	}
	return fmt.Sprintf("%0*d", otpCodeDigits, n.Int64()), nil
}

func hashOTPCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}
