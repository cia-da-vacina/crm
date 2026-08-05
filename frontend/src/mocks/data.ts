/**
 * Demo fixtures for local UI development without a backend.
 *
 * Remove this entire `src/mocks/` folder (and the two USE_MOCKS hooks in
 * `server/backend.ts` + `app/api/proxy/[...path]/route.ts`) when the real
 * API is ready. See README.md in this folder.
 */

import type {
  ConversationDetail,
  Customer,
  FollowUp,
  Message,
  MetaSettings,
  Pop,
  SocialEngagement,
  Unit,
  User,
} from "@/domain";

export const mockUnits: Unit[] = [
  {
    id: "11111111-1111-1111-1111-111111111101",
    name: "Unidade Centro",
    code: "centro",
    timezone: "America/Sao_Paulo",
    active: true,
    address: "Rua das Flores, 100",
    city: "Querência",
    district: "Centro",
  },
  {
    id: "11111111-1111-1111-1111-111111111102",
    name: "Unidade Norte",
    code: "norte",
    timezone: "America/Sao_Paulo",
    active: true,
    address: "Av. Brasil, 250",
    city: "Querência",
    district: "Norte",
  },
  {
    id: "11111111-1111-1111-1111-111111111103",
    name: "Unidade Sul",
    code: "sul",
    timezone: "America/Sao_Paulo",
    active: true,
    address: "Rua do Comércio, 80",
    city: "Querência",
    district: "Sul",
  },
  {
    id: "11111111-1111-1111-1111-111111111104",
    name: "Unidade Leste",
    code: "leste",
    timezone: "America/Sao_Paulo",
    active: true,
    address: "Rua Leste, 45",
    city: "Querência",
    district: "Leste",
  },
  {
    id: "11111111-1111-1111-1111-111111111105",
    name: "Unidade Oeste",
    code: "oeste",
    timezone: "America/Sao_Paulo",
    active: true,
    address: "Av. Oeste, 300",
    city: "Querência",
    district: "Oeste",
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

const now = Date.now();
const iso = (msAgo: number) => new Date(now - msAgo).toISOString();
const windowExpires = () => new Date(now + 20 * 3600_000).toISOString();

const CUST = {
  maria: "44444444-4444-4444-4444-444444444401",
  joao: "44444444-4444-4444-4444-444444444402",
  ana: "44444444-4444-4444-4444-444444444403",
  lucas: "44444444-4444-4444-4444-444444444404",
} as const;

const CONV = {
  maria: "33333333-3333-3333-3333-333333333301",
  joao: "33333333-3333-3333-3333-333333333302",
  ana: "33333333-3333-3333-3333-333333333303",
  lucas: "33333333-3333-3333-3333-333333333304",
} as const;

export const mockCustomers: Customer[] = [
  {
    id: CUST.maria,
    display_name: "Maria Silva",
    identification: "identified",
    primary_phone: "+5511999990001",
    unit_id: mockUnits[0].id,
    identities: [
      {
        id: "55555555-5555-5555-5555-555555555501",
        customer_id: CUST.maria,
        channel: "whatsapp",
        external_id: "5511999990001",
        phone_e164: "+5511999990001",
        verified_at: iso(86400_000),
        created_at: iso(86400_000),
      },
    ],
    created_at: iso(7 * 86400_000),
    updated_at: iso(60_000),
  },
  {
    id: CUST.joao,
    display_name: "João Pereira",
    identification: "identified",
    primary_phone: "+5511999990002",
    unit_id: mockUnits[0].id,
    identities: [
      {
        id: "55555555-5555-5555-5555-555555555502",
        customer_id: CUST.joao,
        channel: "whatsapp",
        external_id: "5511999990002",
        phone_e164: "+5511999990002",
        verified_at: iso(2 * 86400_000),
        created_at: iso(2 * 86400_000),
      },
    ],
    created_at: iso(14 * 86400_000),
    updated_at: iso(3600_000),
  },
  {
    id: CUST.ana,
    display_name: "Ana Costa",
    identification: "identified",
    primary_phone: "+5511999990003",
    unit_id: mockUnits[1].id,
    identities: [
      {
        id: "55555555-5555-5555-5555-555555555503",
        customer_id: CUST.ana,
        channel: "whatsapp",
        external_id: "5511999990003",
        phone_e164: "+5511999990003",
        verified_at: iso(3 * 86400_000),
        created_at: iso(3 * 86400_000),
      },
      {
        id: "55555555-5555-5555-5555-555555555513",
        customer_id: CUST.ana,
        channel: "instagram",
        external_id: "igsid-ana-costa",
        display_handle: "@anacosta",
        phone_e164: "+5511999990003",
        verified_at: iso(2 * 86400_000),
        created_at: iso(3 * 86400_000),
      },
    ],
    created_at: iso(21 * 86400_000),
    updated_at: iso(86400_000),
  },
  {
    id: CUST.lucas,
    display_name: "Lucas Mendes",
    identification: "anonymous",
    primary_phone: null,
    unit_id: mockUnits[0].id,
    identities: [
      {
        id: "55555555-5555-5555-5555-555555555504",
        customer_id: CUST.lucas,
        channel: "instagram",
        external_id: "igsid-lucas-mendes",
        display_handle: "@lucasmendes",
        created_at: iso(120_000),
      },
    ],
    created_at: iso(120_000),
    updated_at: iso(30_000),
  },
];

function customerById(id: string): Customer {
  const c = mockCustomers.find((x) => x.id === id);
  if (!c) throw new Error(`mock customer missing: ${id}`);
  return structuredClone(c);
}

export const mockConversations: ConversationDetail[] = [
  {
    id: CONV.maria,
    customer_id: CUST.maria,
    customer_name: "Maria Silva",
    customer_phone: "+5511999990001",
    identification: "identified",
    phone_gate: "collected",
    channel: "whatsapp",
    channel_thread_id: "wamid-thread-maria",
    unit_id: mockUnits[0].id,
    pipeline_stage: "em_atendimento",
    mode: "ai_triage",
    status: "open",
    owner_id: null,
    intent: "agendar",
    ai_summary: "Cliente quer agendar vacina da gripe para a filha.",
    triage_notes: "Filha ~8 anos. Preferência Unidade Centro.",
    last_message_preview: "Quero agendar a vacina da gripe",
    last_message_at: iso(30_000),
    window_expires_at: windowExpires(),
    unread_count: 1,
    customer: customerById(CUST.maria),
    created_at: iso(600_000),
    updated_at: iso(30_000),
  },
  {
    id: CONV.joao,
    customer_id: CUST.joao,
    customer_name: "João Pereira",
    customer_phone: "+5511999990002",
    identification: "identified",
    phone_gate: "collected",
    channel: "whatsapp",
    channel_thread_id: "wamid-thread-joao",
    unit_id: mockUnits[0].id,
    pipeline_stage: "em_negociacao",
    mode: "human",
    status: "open",
    owner_id: mockAgent.id,
    intent: "precos",
    ai_summary: "Interessado no pacote família.",
    last_message_preview: "Qual o valor do pacote?",
    last_message_at: iso(3600_000),
    window_expires_at: windowExpires(),
    unread_count: 0,
    customer: customerById(CUST.joao),
    created_at: iso(2 * 3600_000),
    updated_at: iso(3600_000),
  },
  {
    id: CONV.ana,
    customer_id: CUST.ana,
    customer_name: "Ana Costa",
    customer_phone: "+5511999990003",
    identification: "identified",
    phone_gate: "collected",
    channel: "facebook",
    channel_thread_id: "fb-thread-ana",
    unit_id: mockUnits[1].id,
    pipeline_stage: "aguardando_fechamento",
    mode: "human",
    status: "pending",
    owner_id: mockAgent.id,
    intent: "agendar",
    ai_summary: "Combinou valores; falta confirmar horário.",
    last_message_preview: "Vou falar com meu marido e retorno",
    last_message_at: iso(86400_000),
    window_expires_at: windowExpires(),
    unread_count: 0,
    customer: customerById(CUST.ana),
    created_at: iso(2 * 86400_000),
    updated_at: iso(86400_000),
  },
  {
    id: CONV.lucas,
    customer_id: CUST.lucas,
    customer_name: "Lucas Mendes",
    customer_phone: null,
    identification: "anonymous",
    phone_gate: "required",
    channel: "instagram",
    channel_thread_id: "ig-thread-lucas",
    unit_id: mockUnits[0].id,
    pipeline_stage: "em_atendimento",
    mode: "ai_triage",
    status: "open",
    owner_id: null,
    intent: "agendar",
    ai_summary: "Quer agendar no Instagram; ainda sem telefone confirmado.",
    last_message_preview: "Quero marcar a gripe pra mim",
    last_message_at: iso(45_000),
    unread_count: 2,
    customer: customerById(CUST.lucas),
    created_at: iso(180_000),
    updated_at: iso(45_000),
  },
];

export const mockMessages: Record<string, Message[]> = {
  [CONV.maria]: [
    {
      id: "m1",
      conversation_id: CONV.maria,
      direction: "in",
      sender_type: "contact",
      kind: "text",
      channel: "whatsapp",
      body: "Olá, boa tarde!",
      status: "read",
      created_at: iso(120_000),
    },
    {
      id: "m2",
      conversation_id: CONV.maria,
      direction: "out",
      sender_type: "ai",
      kind: "text",
      channel: "whatsapp",
      body: "Olá! Sou a assistente da Cia da Vacina. Posso ajudar com agendamento, preços ou dúvidas?",
      status: "read",
      created_at: iso(110_000),
    },
    {
      id: "m3",
      conversation_id: CONV.maria,
      direction: "in",
      sender_type: "contact",
      kind: "text",
      channel: "whatsapp",
      body: "Quero agendar a vacina da gripe para a minha filha.",
      status: "read",
      created_at: iso(60_000),
    },
    {
      id: "m4",
      conversation_id: CONV.maria,
      direction: "out",
      sender_type: "ai",
      kind: "text",
      channel: "whatsapp",
      body: "Perfeito! Vou direcionar você para um atendente da Unidade Centro.",
      status: "delivered",
      created_at: iso(30_000),
    },
  ],
  [CONV.joao]: [
    {
      id: "m5",
      conversation_id: CONV.joao,
      direction: "in",
      sender_type: "contact",
      kind: "text",
      channel: "whatsapp",
      body: "Qual o valor do pacote família?",
      status: "read",
      created_at: iso(3600_000),
    },
    {
      id: "m6",
      conversation_id: CONV.joao,
      direction: "out",
      sender_type: "agent",
      kind: "text",
      channel: "whatsapp",
      body: "Pacote família a partir de R$ 480. Prefere Unidade Centro?",
      status: "read",
      created_at: iso(3500_000),
    },
  ],
  [CONV.ana]: [
    {
      id: "m7",
      conversation_id: CONV.ana,
      direction: "out",
      sender_type: "agent",
      kind: "text",
      channel: "facebook",
      body: "Combinamos amanhã às 10h. Posso confirmar?",
      status: "read",
      created_at: iso(90_000_000),
    },
    {
      id: "m8",
      conversation_id: CONV.ana,
      direction: "in",
      sender_type: "contact",
      kind: "text",
      channel: "facebook",
      body: "Vou falar com meu marido e retorno",
      status: "read",
      created_at: iso(86400_000),
    },
  ],
  [CONV.lucas]: [
    {
      id: "m9",
      conversation_id: CONV.lucas,
      direction: "in",
      sender_type: "contact",
      kind: "text",
      channel: "instagram",
      body: "Oi! Vi o story de vocês",
      status: "read",
      created_at: iso(150_000),
    },
    {
      id: "m10",
      conversation_id: CONV.lucas,
      direction: "out",
      sender_type: "ai",
      kind: "text",
      channel: "instagram",
      body: "Olá! Para agendar, preciso do seu WhatsApp para confirmar o cadastro.",
      status: "delivered",
      created_at: iso(90_000),
    },
    {
      id: "m11",
      conversation_id: CONV.lucas,
      direction: "in",
      sender_type: "contact",
      kind: "text",
      channel: "instagram",
      body: "Quero marcar a gripe pra mim",
      status: "read",
      created_at: iso(45_000),
    },
  ],
};

export const mockFollowUps: FollowUp[] = [
  {
    id: "66666666-6666-6666-6666-666666666601",
    conversation_id: CONV.ana,
    customer_id: CUST.ana,
    customer_name: "Ana Costa",
    customer_phone: "+5511999990003",
    unit_id: mockUnits[1].id,
    pipeline_stage: "aguardando_fechamento",
    due_at: iso(3600_000),
    status: "open",
    note: "Retomar confirmação de horário",
    created_at: iso(86400_000),
  },
  {
    id: "66666666-6666-6666-6666-666666666602",
    conversation_id: CONV.joao,
    customer_id: CUST.joao,
    customer_name: "João Pereira",
    customer_phone: "+5511999990002",
    unit_id: mockUnits[0].id,
    pipeline_stage: "em_negociacao",
    due_at: new Date(now + 86400_000).toISOString(),
    status: "open",
    note: "Enviar proposta pacote família",
    created_at: iso(7200_000),
  },
];

export const mockPops: Pop[] = [
  {
    id: "77777777-7777-7777-7777-777777777701",
    title: "Saudação padrão",
    body: "Olá! Sou da Cia da Vacina. Em que posso ajudar hoje?",
    intent_tags: ["outro", "duvidas"],
    active: true,
  },
  {
    id: "77777777-7777-7777-7777-777777777702",
    title: "Agendamento — gripe",
    body: "Para agendar a vacina da gripe, preciso do nome completo, data de nascimento e unidade de preferência.",
    intent_tags: ["agendar"],
    active: true,
  },
  {
    id: "77777777-7777-7777-7777-777777777703",
    title: "Tabela de preços",
    body: "Dose individual a partir de R$ 150; pacote família a partir de R$ 480.",
    intent_tags: ["precos"],
    active: true,
  },
  {
    id: "77777777-7777-7777-7777-777777777704",
    title: "Follow-up fechamento",
    body: "Oi! Conseguiu confirmar o horário da vacinação? Posso reservar um encaixe esta semana?",
    intent_tags: ["agendar", "outro"],
    active: true,
  },
];

export const mockEngagements: SocialEngagement[] = [
  {
    id: "88888888-8888-8888-8888-888888888801",
    customer_id: CUST.lucas,
    customer_name: "Lucas Mendes",
    channel: "instagram",
    type: "story_reply",
    status: "open",
    unit_id: mockUnits[0].id,
    media_id: "ig-media-story-01",
    media_caption: "Campanha gripe 2026 — 15% off",
    body: "Quanto custa a dose?",
    external_id: "ig-story-reply-01",
    author_external_id: "igsid-lucas-mendes",
    created_at: iso(300_000),
  },
  {
    id: "88888888-8888-8888-8888-888888888802",
    customer_id: null,
    customer_name: null,
    channel: "instagram",
    type: "post_comment",
    status: "open",
    unit_id: mockUnits[0].id,
    media_id: "ig-media-post-02",
    media_caption: "Pacote família",
    body: "Atende Unidade Norte?",
    external_id: "ig-comment-02",
    author_external_id: "igsid-desconhecido",
    created_at: iso(900_000),
  },
  {
    id: "88888888-8888-8888-8888-888888888803",
    customer_id: CUST.ana,
    customer_name: "Ana Costa",
    channel: "facebook",
    type: "live_comment",
    status: "replied",
    unit_id: mockUnits[1].id,
    media_id: "fb-live-01",
    body: "Vocês aplicam em crianças?",
    external_id: "fb-live-comment-01",
    author_external_id: "psid-ana",
    replied_at: iso(400_000),
    created_at: iso(500_000),
  },
];

export const mockMetaSettings: MetaSettings = {
  channels: [
    {
      channel: "whatsapp",
      enabled: true,
      account_id: "WABA-DEMO-001",
      display_name: "Cia da Vacina WhatsApp",
      phone_number_id: "PHONE-CENTRO-001",
      webhook_verified: true,
      token_masked: "EAA••••••••xyz",
    },
    {
      channel: "instagram",
      enabled: true,
      account_id: "IG-DEMO-001",
      display_name: "@ciadavacina",
      webhook_verified: true,
      token_masked: "IGQ••••••••abc",
    },
    {
      channel: "facebook",
      enabled: true,
      account_id: "PAGE-DEMO-001",
      display_name: "Cia da Vacina",
      webhook_verified: false,
      token_masked: "EAA••••••••page",
    },
  ],
  ai_enabled: true,
  ai_system_prompt: `Você é a assistente de triagem da Cia da Vacina.

Como agir:
- Seja cordial, objetiva e em português do Brasil.
- Faça perguntas curtas para entender a necessidade.
- Ao identificar intenção clara, resuma e prepare handoff para humano.

O que NÃO fazer:
- Não invente preços fora do contexto/campanhas.
- Não confirme agendamento definitivo.
- Não colete dados sensíveis desnecessários.`,
  ai_context:
    "Unidades: Centro, Norte, Sul, Leste e Oeste. Horário: seg–sáb 8h–18h. Foco: gripe e pacotes família.",
  ai_campaigns: [
    {
      id: "camp-gripe-2026",
      title: "Campanha Gripe 2026",
      description:
        "Vacina da gripe com 15% off em dose individual. Pacote família a partir de R$ 480.",
      starts_on: "2026-07-01",
      ends_on: "2026-08-31",
      active: true,
    },
  ],
  triage_enabled: true,
  triage_handoff_intents: ["agendar", "reclamacao"],
};

export const lossReasons = [
  { code: "preco", label: "Preço elevado" },
  { code: "concorrente", label: "Foi para concorrente" },
  { code: "sem_retorno", label: "Cliente sem retorno" },
  { code: "prazo", label: "Sem disponibilidade de agenda" },
  { code: "nao_interesse", label: "Perdeu o interesse" },
  { code: "outro", label: "Outro" },
];

export const MOCK_PASSWORDS: Record<string, string> = {
  [mockAdmin.email]: "admin123",
  [mockAgent.email]: "agent123",
};

export { CONV, CUST };
