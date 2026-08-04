package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

// Este arquivo acessa customers/customer_identities diretamente em vez de
// depender do módulo customer — mesma convenção de módulo autocontido usada
// em internal/module/auth/repository (que também tem sua própria query de
// users, independente do módulo user). Só o fluxo de confirmação de OTP
// precisa mutar Customer, então a duplicação é pequena e localizada.

// GetCustomerIdentityExternalID busca o external_id (wa_id/IGSID/PSID) da
// identidade do cliente NO CANAL da conversa — é esse valor, não
// channel_thread_id, que endereça o destinatário na API da Meta.
func (r *Repository) GetCustomerIdentityExternalID(ctx context.Context, customerID string, channel entity.Channel) (string, error) {
	var externalID string
	err := r.db.GetContext(ctx, &externalID, `
		SELECT external_id FROM customer_identities WHERE customer_id = $1 AND channel = $2
	`, customerID, channel)
	return externalID, err
}

func (r *Repository) GetCustomerByPhone(ctx context.Context, phone string) (entity.Customer, error) {
	var c entity.Customer
	err := r.db.GetContext(ctx, &c, `
		SELECT id, display_name, identification, primary_phone, unit_id, created_at, updated_at
		FROM customers WHERE primary_phone = $1
	`, phone)
	return c, err
}

// MergeCustomerInto reparenta identities e conversations de source pra
// target e apaga source — usado quando o telefone confirmado via OTP já
// pertence a um Customer identified existente (docs/BACKEND-CONTRACT.md §3:
// "funde se já existir Customer canônico com esse telefone").
func (r *Repository) MergeCustomerInto(ctx context.Context, sourceID, targetID string) error {
	return r.db.Transaction(ctx, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			UPDATE customer_identities SET customer_id = $1 WHERE customer_id = $2
		`, targetID, sourceID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE conversations SET customer_id = $1 WHERE customer_id = $2
		`, targetID, sourceID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM customers WHERE id = $1`, sourceID)
		return err
	})
}

func (r *Repository) PromoteCustomerToIdentified(ctx context.Context, customerID, phone string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE customers SET identification = 'identified', primary_phone = $1, updated_at = NOW() WHERE id = $2
	`, phone, customerID)
	return err
}

// SetCustomerIdentityVerified marca a identidade do canal atual da conversa
// (a que efetivamente foi provada por posse via OTP) com o telefone e o
// timestamp de verificação.
func (r *Repository) SetCustomerIdentityVerified(ctx context.Context, customerID string, channel entity.Channel, phone string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE customer_identities SET phone_e164 = $1, verified_at = NOW()
		WHERE customer_id = $2 AND channel = $3
	`, phone, customerID, channel)
	return err
}
