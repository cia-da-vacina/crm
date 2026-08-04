package seeder

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// seedDemoConversations cria uma conversa (com 1 mensagem inbound) pra cada
// cliente de demo — dá pra exercitar claim/pipeline/mensagens/OTP via API
// antes de existir qualquer webhook Meta real. Idempotência: se o customer
// já tem alguma conversa, pula (não há AI/webhook criando conversas extras
// em dev, então "já tem uma" é sinal suficiente de que o seed já rodou).
func seedDemoConversations(db *sqlx.DB) error {
	ctx := context.Background()

	maria, err := getDemoCustomer(ctx, db, "whatsapp", "5551999990000")
	if err != nil {
		return err
	}
	if err := seedConversation(ctx, db, maria, "whatsapp", "not_needed", nil,
		"Oi, gostaria de agendar a vacina da gripe"); err != nil {
		return err
	}

	joao, err := getDemoCustomer(ctx, db, "instagram", "ig_scoped_id_dev_123")
	if err != nil {
		return err
	}
	intent := "agendar"
	return seedConversation(ctx, db, joao, "instagram", "required", &intent,
		"Oi! Quero marcar um horário pra vacina, vocês atendem por aqui?")
}

type demoCustomer struct {
	ID     string `db:"id"`
	UnitID string `db:"unit_id"`
}

func getDemoCustomer(ctx context.Context, db *sqlx.DB, channel, externalID string) (demoCustomer, error) {
	var c demoCustomer
	err := db.GetContext(ctx, &c, `
		SELECT cu.id, cu.unit_id
		FROM customers cu
		JOIN customer_identities ci ON ci.customer_id = cu.id
		WHERE ci.channel = $1 AND ci.external_id = $2
	`, channel, externalID)
	return c, err
}

func seedConversation(ctx context.Context, db *sqlx.DB, customer demoCustomer, channel, phoneGate string, intent *string, firstMessage string) error {
	var exists bool
	if err := db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM conversations WHERE customer_id = $1)`, customer.ID); err != nil {
		return err
	}
	if exists {
		return nil
	}

	conversationID := uuid.Must(uuid.NewV7()).String()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO conversations (id, customer_id, channel, unit_id, phone_gate, intent, last_message_preview, last_message_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	`, conversationID, customer.ID, channel, customer.UnitID, phoneGate, intent, firstMessage); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO messages (id, conversation_id, direction, sender_type, kind, channel, body, status)
		VALUES ($1, $2, 'in', 'contact', 'text', $3, $4, 'delivered')
	`, uuid.Must(uuid.NewV7()).String(), conversationID, channel, firstMessage)
	return err
}
