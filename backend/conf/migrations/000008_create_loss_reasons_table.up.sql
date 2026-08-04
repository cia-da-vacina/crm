CREATE TABLE loss_reasons (
    code       TEXT PRIMARY KEY,
    label      TEXT NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Catálogo seed fixo (docs/decisions.md D-06) — dado de produção real, não
-- fixture de dev, por isso entra na migration e não no seeder de demo.
INSERT INTO loss_reasons (code, label) VALUES
    ('preco',        'Preço elevado'),
    ('concorrente',  'Foi para concorrente'),
    ('sem_retorno',  'Cliente sem retorno'),
    ('prazo',        'Sem disponibilidade de agenda'),
    ('nao_interesse','Perdeu o interesse'),
    ('outro',        'Outro (detalhar em texto)');
