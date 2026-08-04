package seeder

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/vo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// DemoSeeders são dados só de desenvolvimento: um admin fixo pra login local,
// duas unidades fictícias e dois clientes (um identified via WhatsApp, um
// anonymous via Instagram) — pra exercitar o fluxo de ponta a ponta com o
// frontend antes de existir dado real de produção. Idempotente — rodar
// `make db-seed` várias vezes não duplica nem falha. seedDemoCustomers roda
// depois de seedDemoUnits de propósito: depende da unidade já existir.
func DemoSeeders() []Seeder {
	return []Seeder{seedDemoAdmin, seedDemoUnits, seedDemoCustomers, seedDemoConversations}
}

func seedDemoAdmin(db *sqlx.DB) error {
	hash, err := vo.HashPassword("admin123", vo.DefaultPasswordConfig())
	if err != nil {
		return err
	}

	_, err = db.ExecContext(context.Background(), `
		INSERT INTO users (id, email, password_hash, name, role, active)
		VALUES ($1, 'admin@ciadavacina.dev', $2, 'Admin Dev', 'admin', true)
		ON CONFLICT (email) DO NOTHING
	`, uuid.Must(uuid.NewV7()).String(), hash)
	return err
}

func seedDemoUnits(db *sqlx.DB) error {
	units := []struct {
		Name, Code, City, Address string
	}{
		{"Unidade Centro (dev)", "DEV-CENTRO", "Porto Alegre", "Rua Fictícia, 100"},
		{"Unidade Norte (dev)", "DEV-NORTE", "Porto Alegre", "Av. Exemplo, 200"},
	}

	for _, u := range units {
		if _, err := db.ExecContext(context.Background(), `
			INSERT INTO units (id, name, code, city, address)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (code) DO NOTHING
		`, uuid.Must(uuid.NewV7()).String(), u.Name, u.Code, u.City, u.Address); err != nil {
			return err
		}
	}
	return nil
}

// seedDemoCustomers usa "checa antes de inserir" em vez de ON CONFLICT DO
// NOTHING puro: customers não tem uma chave natural sempre presente (cliente
// anônimo não tem primary_phone), então o ponto de idempotência real é a
// identidade (channel, external_id) — só ela garante não duplicar em reruns.
func seedDemoCustomers(db *sqlx.DB) error {
	ctx := context.Background()

	var unitID string
	if err := db.GetContext(ctx, &unitID, `SELECT id FROM units WHERE code = 'DEV-CENTRO'`); err != nil {
		return err
	}

	if err := seedIdentifiedCustomer(ctx, db, unitID); err != nil {
		return err
	}
	return seedAnonymousCustomer(ctx, db, unitID)
}

func seedIdentifiedCustomer(ctx context.Context, db *sqlx.DB, unitID string) error {
	var exists bool
	if err := db.GetContext(ctx, &exists, `
		SELECT EXISTS(SELECT 1 FROM customer_identities WHERE channel = 'whatsapp' AND external_id = '5551999990000')
	`); err != nil {
		return err
	}
	if exists {
		return nil
	}

	customerID := uuid.Must(uuid.NewV7()).String()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO customers (id, display_name, identification, primary_phone, unit_id)
		VALUES ($1, 'Maria Cliente (dev)', 'identified', '+5551999990000', $2)
	`, customerID, unitID); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO customer_identities (id, customer_id, channel, external_id, phone_e164, verified_at)
		VALUES ($1, $2, 'whatsapp', '5551999990000', '+5551999990000', now())
	`, uuid.Must(uuid.NewV7()).String(), customerID)
	return err
}

func seedAnonymousCustomer(ctx context.Context, db *sqlx.DB, unitID string) error {
	var exists bool
	if err := db.GetContext(ctx, &exists, `
		SELECT EXISTS(SELECT 1 FROM customer_identities WHERE channel = 'instagram' AND external_id = 'ig_scoped_id_dev_123')
	`); err != nil {
		return err
	}
	if exists {
		return nil
	}

	customerID := uuid.Must(uuid.NewV7()).String()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO customers (id, display_name, identification, unit_id)
		VALUES ($1, 'joaosilva_ig (dev)', 'anonymous', $2)
	`, customerID, unitID); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO customer_identities (id, customer_id, channel, external_id, display_handle)
		VALUES ($1, $2, 'instagram', 'ig_scoped_id_dev_123', '@joaosilva_ig')
	`, uuid.Must(uuid.NewV7()).String(), customerID)
	return err
}
