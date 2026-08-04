-- pgcrypto habilita gen_random_uuid() no Postgres, disponível para uso em SQL
-- ad-hoc (seeds, scripts). O código Go gera UUIDs em aplicação (uuid.NewV7),
-- então isto não é usado por nenhuma migration de domínio ainda — é só
-- infraestrutura de banco disponível desde o início.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
