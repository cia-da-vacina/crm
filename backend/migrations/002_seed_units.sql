-- Seed: 5 unidades Cia da Vacina (IDs fixos para demos)
INSERT INTO units (id, name, code, timezone, active) VALUES
  ('11111111-1111-1111-1111-111111111101', 'Unidade Centro', 'centro', 'America/Sao_Paulo', TRUE),
  ('11111111-1111-1111-1111-111111111102', 'Unidade Norte', 'norte', 'America/Sao_Paulo', TRUE),
  ('11111111-1111-1111-1111-111111111103', 'Unidade Sul', 'sul', 'America/Sao_Paulo', TRUE),
  ('11111111-1111-1111-1111-111111111104', 'Unidade Leste', 'leste', 'America/Sao_Paulo', TRUE),
  ('11111111-1111-1111-1111-111111111105', 'Unidade Oeste', 'oeste', 'America/Sao_Paulo', TRUE)
ON CONFLICT (code) DO NOTHING;
