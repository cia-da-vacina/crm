import type { InboxItem, Unit, User } from "@/lib/types";

export const mockUnits: Unit[] = [
  {
    id: "11111111-1111-1111-1111-111111111101",
    name: "Unidade Centro",
    code: "centro",
    timezone: "America/Sao_Paulo",
    active: true,
  },
  {
    id: "11111111-1111-1111-1111-111111111102",
    name: "Unidade Norte",
    code: "norte",
    timezone: "America/Sao_Paulo",
    active: true,
  },
  {
    id: "11111111-1111-1111-1111-111111111103",
    name: "Unidade Sul",
    code: "sul",
    timezone: "America/Sao_Paulo",
    active: true,
  },
  {
    id: "11111111-1111-1111-1111-111111111104",
    name: "Unidade Leste",
    code: "leste",
    timezone: "America/Sao_Paulo",
    active: true,
  },
  {
    id: "11111111-1111-1111-1111-111111111105",
    name: "Unidade Oeste",
    code: "oeste",
    timezone: "America/Sao_Paulo",
    active: true,
  },
];

export const mockAdmin: User = {
  id: "22222222-2222-2222-2222-222222222201",
  email: "admin@ciadavacina.com.br",
  name: "Administrador",
  role: "admin",
  active: true,
  unit_ids: mockUnits.map((u) => u.id),
};

export const mockAgent: User = {
  id: "22222222-2222-2222-2222-222222222202",
  email: "atendente@ciadavacina.com.br",
  name: "Atendente Demo",
  role: "agent",
  active: true,
  unit_ids: [mockUnits[0].id],
};

export const mockInbox: InboxItem[] = [
  {
    id: "33333333-3333-3333-3333-333333333301",
    contact_name: "Maria Silva",
    contact_phone: "+5511999990001",
    unit_id: mockUnits[0].id,
    pipeline_stage: "em_atendimento",
    mode: "ai_triage",
    owner_id: null,
    intent: "agendar",
    ai_summary: "Cliente quer agendar vacina da gripe para a filha.",
    last_message_preview: "Quero agendar a vacina da gripe",
    last_message_at: new Date().toISOString(),
  },
  {
    id: "33333333-3333-3333-3333-333333333302",
    contact_name: "João Pereira",
    contact_phone: "+5511999990002",
    unit_id: mockUnits[0].id,
    pipeline_stage: "em_negociacao",
    mode: "human",
    owner_id: mockAgent.id,
    intent: "precos",
    ai_summary: "Interessado no pacote família.",
    last_message_preview: "Qual o valor do pacote?",
    last_message_at: new Date(Date.now() - 3600_000).toISOString(),
  },
  {
    id: "33333333-3333-3333-3333-333333333303",
    contact_name: "Ana Costa",
    contact_phone: "+5511999990003",
    unit_id: mockUnits[1].id,
    pipeline_stage: "aguardando_fechamento",
    mode: "human",
    owner_id: mockAgent.id,
    intent: "agendar",
    ai_summary: "Combinou valores; falta confirmar horário.",
    last_message_preview: "Vou falar com meu marido e retorno",
    last_message_at: new Date(Date.now() - 86400_000).toISOString(),
  },
];
