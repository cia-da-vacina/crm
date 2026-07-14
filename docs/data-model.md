# Modelo de Dados — CRM Cia da Vacina

Banco: **PostgreSQL**. Sem migrations neste documento — apenas modelagem lógica.

## Enums

| Nome | Valores |
|------|---------|
| `user_role` | `admin`, `manager`, `supervisor`, `agent` |
| `conversation_status` | `open`, `pending`, `resolved` |
| `conversation_mode` | `ai_triage`, `human` |
| `pipeline_stage` | `em_atendimento`, `em_negociacao`, `aguardando_fechamento`, `fechado`, `nao_fechado` |
| `message_direction` | `in`, `out` |
| `sender_type` | `contact`, `agent`, `ai`, `system` |
| `message_status` | `accepted`, `sent`, `delivered`, `read`, `failed` |
| `followup_status` | `open`, `done`, `canceled` |
| `intent` (catálogo MVP) | `agendar`, `precos`, `duvidas`, `reclamacao`, `outro` |

## Entidades

### users
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| email | TEXT | UNIQUE, NOT NULL |
| password_hash | TEXT | NOT NULL |
| name | TEXT | NOT NULL |
| role | user_role | NOT NULL |
| active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### units
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| name | TEXT | NOT NULL |
| code | TEXT | UNIQUE, NOT NULL |
| timezone | TEXT | NOT NULL DEFAULT 'America/Sao_Paulo' |
| active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### user_units
| Campo | Tipo | Restrições |
|-------|------|------------|
| user_id | UUID | PK, FK → users |
| unit_id | UUID | PK, FK → units |

### whatsapp_numbers
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| unit_id | UUID | FK → units, NOT NULL |
| phone_number_id | TEXT | UNIQUE, NOT NULL (Meta) |
| display_phone | TEXT | NOT NULL |
| waba_id | TEXT | NOT NULL |
| active | BOOLEAN | NOT NULL DEFAULT true |

**Decisão:** 1 número WhatsApp por unidade (N números sob 1 WABA). Ver [decisions.md](./decisions.md).

### contacts
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| wa_id | TEXT | UNIQUE, NOT NULL |
| phone_e164 | TEXT | NOT NULL |
| name | TEXT | |
| unit_id | UUID | FK → units, nullable |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

**Índices:** `contacts(phone_e164)`

### conversations
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| contact_id | UUID | FK → contacts, NOT NULL |
| unit_id | UUID | FK → units, NOT NULL |
| status | conversation_status | NOT NULL DEFAULT 'open' |
| mode | conversation_mode | NOT NULL DEFAULT 'ai_triage' |
| owner_id | UUID | FK → users, nullable |
| pipeline_stage | pipeline_stage | NOT NULL DEFAULT 'em_atendimento' |
| intent | TEXT | nullable |
| ai_summary | TEXT | nullable |
| last_message_at | TIMESTAMPTZ | |
| window_expires_at | TIMESTAMPTZ | janela 24h WhatsApp |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

**Índices:** `(unit_id, status, last_message_at DESC)`

**Nota de modelagem:** no MVP, a oportunidade comercial **é** a conversa (`pipeline_stage`). Sem entidade Deal separada.

### messages
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| conversation_id | UUID | FK → conversations, NOT NULL |
| direction | message_direction | NOT NULL |
| sender_type | sender_type | NOT NULL |
| sender_user_id | UUID | FK → users, nullable |
| body | TEXT | NOT NULL |
| wa_message_id | TEXT | UNIQUE, nullable |
| status | message_status | NOT NULL DEFAULT 'accepted' |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

**Índices:** `(conversation_id, created_at)`

### pipeline_transitions
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| conversation_id | UUID | FK → conversations |
| from_stage | pipeline_stage | NOT NULL |
| to_stage | pipeline_stage | NOT NULL |
| reason_code | TEXT | FK lógico → loss_reasons.code; obrigatório se to_stage = nao_fechado |
| reason_text | TEXT | |
| by_user_id | UUID | FK → users |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### loss_reasons
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| code | TEXT | UNIQUE, NOT NULL |
| label | TEXT | NOT NULL |
| active | BOOLEAN | NOT NULL DEFAULT true |

### pops
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| title | TEXT | NOT NULL |
| body | TEXT | NOT NULL |
| intent_tags | TEXT[] | DEFAULT '{}' |
| unit_id | UUID | FK → units, nullable (global se null) |
| active | BOOLEAN | NOT NULL DEFAULT true |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### followups
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| conversation_id | UUID | FK → conversations |
| due_at | TIMESTAMPTZ | NOT NULL |
| status | followup_status | NOT NULL DEFAULT 'open' |
| note | TEXT | |
| created_by | UUID | FK → users |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

**Índices:** `(status, due_at)`

### ai_runs
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| conversation_id | UUID | FK → conversations |
| model | TEXT | NOT NULL |
| prompt_version | TEXT | NOT NULL |
| input_ref | TEXT | |
| output_json | JSONB | |
| latency_ms | INT | |
| error | TEXT | |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### audit_logs
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| actor_user_id | UUID | FK → users, nullable |
| action | TEXT | NOT NULL |
| entity_type | TEXT | NOT NULL |
| entity_id | TEXT | NOT NULL |
| payload_json | JSONB | |
| ip | TEXT | |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

**Índices:** `(created_at)`, `(entity_type, entity_id)`  
**Restrição:** append-only (sem UPDATE/DELETE na aplicação).

### message_templates
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| name | TEXT | NOT NULL |
| language | TEXT | NOT NULL DEFAULT 'pt_BR' |
| status | TEXT | NOT NULL |
| body_preview | TEXT | |
| meta_template_name | TEXT | NOT NULL |

### configs
| Campo | Tipo | Restrições |
|-------|------|------------|
| key | TEXT | PK |
| value_encrypted | TEXT | |
| updated_by | UUID | FK → users |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### jobs (fila leve em DB — MVP)
| Campo | Tipo | Restrições |
|-------|------|------------|
| id | UUID | PK |
| kind | TEXT | NOT NULL |
| payload_json | JSONB | NOT NULL |
| run_at | TIMESTAMPTZ | NOT NULL |
| attempts | INT | NOT NULL DEFAULT 0 |
| last_error | TEXT | |
| done_at | TIMESTAMPTZ | |

## Relacionamentos (resumo)

```
units 1──* whatsapp_numbers
users *──* units (via user_units)
contacts 1──* conversations
units 1──* conversations
users 1──* conversations (owner)
conversations 1──* messages
conversations 1──* pipeline_transitions
conversations 1──* followups
conversations 1──* ai_runs
```

## Regras de integridade

1. Motivo (`reason_code`) obrigatório quando `to_stage = nao_fechado`.
2. `wa_message_id` único para idempotência de webhook.
3. Escopo de leitura/escrita: usuário só acessa conversas de unidades em `user_units` (admin vê todas).
4. Retenção LGPD: **24 meses** (configurável) — ver decisions.md.
