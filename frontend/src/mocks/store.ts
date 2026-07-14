import {
  DEFAULT_AI_CONTEXT,
  DEFAULT_AI_SYSTEM_PROMPT,
} from "@/lib/ai-defaults";
import type {
  DashboardSummary,
  FollowUp,
  InboxItem,
  Message,
  PipelineStage,
  Pop,
  User,
  WhatsAppSettings,
} from "@/lib/types";
import { mockAdmin, mockAgent, mockInbox, mockUnits } from "./data";

export const db = {
  inbox: structuredClone(mockInbox) as InboxItem[],
  messages: {} as Record<string, Message[]>,
  followups: [] as FollowUp[],
  pops: [] as Pop[],
  users: structuredClone([mockAdmin, mockAgent]) as User[],
  settings: {
    waba_id: "WABA-DEMO-001",
    phone_number_id: "PHONE-CENTRO-001",
    display_phone: "+55 11 4000-1000",
    token_masked: "EAA•••••••••••••••xyz",
    ai_enabled: true,
    webhook_verified: true,
    ai_system_prompt: DEFAULT_AI_SYSTEM_PROMPT,
    ai_context: DEFAULT_AI_CONTEXT,
    ai_campaigns: [
      {
        id: "camp-gripe-2026",
        title: "Campanha Gripe 2026",
        description:
          "Vacina da gripe com 15% off em dose individual até o fim do período. Pacote família a partir de R$ 480.",
        starts_on: "2026-07-01",
        ends_on: "2026-08-31",
        active: true,
      },
    ],
  } as WhatsAppSettings,
};

const c1 = "33333333-3333-3333-3333-333333333301";
const c2 = "33333333-3333-3333-3333-333333333302";
const c3 = "33333333-3333-3333-3333-333333333303";

db.messages[c1] = [
  {
    id: "m1",
    conversation_id: c1,
    direction: "in",
    sender_type: "contact",
    body: "Olá, boa tarde!",
    status: "read",
    created_at: new Date(Date.now() - 120000).toISOString(),
  },
  {
    id: "m2",
    conversation_id: c1,
    direction: "out",
    sender_type: "ai",
    body: "Olá! Sou a assistente da Cia da Vacina. Posso ajudar com agendamento, preços ou dúvidas. O que você precisa?",
    status: "read",
    created_at: new Date(Date.now() - 110000).toISOString(),
  },
  {
    id: "m3",
    conversation_id: c1,
    direction: "in",
    sender_type: "contact",
    body: "Quero agendar a vacina da gripe para a minha filha.",
    status: "read",
    created_at: new Date(Date.now() - 60000).toISOString(),
  },
  {
    id: "m4",
    conversation_id: c1,
    direction: "out",
    sender_type: "ai",
    body: "Perfeito! Identifiquei intenção de agendamento. Vou direcionar você para um atendente da Unidade Centro.",
    status: "delivered",
    created_at: new Date(Date.now() - 30000).toISOString(),
  },
];

db.messages[c2] = [
  {
    id: "m5",
    conversation_id: c2,
    direction: "in",
    sender_type: "contact",
    body: "Qual o valor do pacote família?",
    status: "read",
    created_at: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: "m6",
    conversation_id: c2,
    direction: "out",
    sender_type: "agent",
    body: "O pacote família (até 4 pessoas) está a partir de R$ 480 nesta campanha. Prefere Unidade Centro?",
    status: "read",
    created_at: new Date(Date.now() - 3500000).toISOString(),
  },
];

db.messages[c3] = [
  {
    id: "m7",
    conversation_id: c3,
    direction: "out",
    sender_type: "agent",
    body: "Combinamos o horário de amanhã às 10h. Posso confirmar?",
    status: "read",
    created_at: new Date(Date.now() - 90000000).toISOString(),
  },
  {
    id: "m8",
    conversation_id: c3,
    direction: "in",
    sender_type: "contact",
    body: "Vou falar com meu marido e retorno",
    status: "read",
    created_at: new Date(Date.now() - 86400000).toISOString(),
  },
];

db.followups = [
  {
    id: "f1",
    conversation_id: c3,
    contact_name: "Ana Costa",
    contact_phone: "+5511999990003",
    unit_id: mockUnits[1].id,
    pipeline_stage: "aguardando_fechamento",
    due_at: new Date(Date.now() - 3600000).toISOString(),
    status: "open",
    note: "Retomar confirmação de horário",
  },
  {
    id: "f2",
    conversation_id: c2,
    contact_name: "João Pereira",
    contact_phone: "+5511999990002",
    unit_id: mockUnits[0].id,
    pipeline_stage: "em_negociacao",
    due_at: new Date(Date.now() + 86400000).toISOString(),
    status: "open",
    note: "Enviar proposta pacote família",
  },
];

db.pops = [
  {
    id: "p1",
    title: "Saudação padrão",
    body: "Olá! Sou da Cia da Vacina. Em que posso ajudar hoje?",
    intent_tags: ["outro", "duvidas"],
    active: true,
  },
  {
    id: "p2",
    title: "Agendamento — gripe",
    body: "Para agendar a vacina da gripe, preciso do nome completo, data de nascimento e unidade de preferência. Tem preferência de dia/período?",
    intent_tags: ["agendar"],
    active: true,
  },
  {
    id: "p3",
    title: "Tabela de preços — campanha",
    body: "Nesta campanha: dose individual a partir de R$ 150; pacote família (até 4) a partir de R$ 480. Valores podem variar por unidade.",
    intent_tags: ["precos"],
    active: true,
  },
  {
    id: "p4",
    title: "Follow-up aguardando fechamento",
    body: "Oi! Passando para saber se conseguiu confirmar o horário da vacinação. Posso reservar um encaixe para você ainda esta semana?",
    intent_tags: ["agendar", "outro"],
    active: true,
  },
];

export const lossReasons = [
  { code: "preco", label: "Preço elevado" },
  { code: "concorrente", label: "Foi para concorrente" },
  { code: "sem_retorno", label: "Cliente sem retorno" },
  { code: "prazo", label: "Sem disponibilidade de agenda" },
  { code: "nao_interesse", label: "Perdeu o interesse" },
  { code: "outro", label: "Outro" },
];

export function buildDashboard(unitId?: string | null): DashboardSummary {
  const items = unitId ? db.inbox.filter((i) => i.unit_id === unitId) : db.inbox;
  const by_stage = {
    em_atendimento: 0,
    em_negociacao: 0,
    aguardando_fechamento: 0,
    fechado: 0,
    nao_fechado: 0,
  } as Record<PipelineStage, number>;
  for (const i of items) by_stage[i.pipeline_stage] += 1;
  const closed = by_stage.fechado;
  const not_closed = by_stage.nao_fechado;
  const decided = closed + not_closed;
  const open = items.filter((i) => i.pipeline_stage !== "fechado" && i.pipeline_stage !== "nao_fechado").length;

  const units = mockUnits.map((u) => {
    const uItems = db.inbox.filter((i) => i.unit_id === u.id);
    const uClosed = uItems.filter((i) => i.pipeline_stage === "fechado").length;
    const uLost = uItems.filter((i) => i.pipeline_stage === "nao_fechado").length;
    const d = uClosed + uLost;
    return {
      unit_id: u.id,
      unit_name: u.name,
      open: uItems.filter((i) => i.pipeline_stage !== "fechado" && i.pipeline_stage !== "nao_fechado").length,
      closed: uClosed,
      conversion_rate: d === 0 ? 0 : Math.round((uClosed / d) * 100),
    };
  });

  return {
    open_conversations: open,
    by_stage,
    closed,
    not_closed,
    conversion_rate: decided === 0 ? 0 : Math.round((closed / decided) * 100),
    ai_triage: items.filter((i) => i.mode === "ai_triage").length,
    human: items.filter((i) => i.mode === "human").length,
    awaiting_followup: db.followups.filter((f) => f.status === "open").length,
    units,
  };
}

export { mockAdmin, mockAgent, mockUnits };
