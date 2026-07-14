import { http, HttpResponse } from "msw";
import {
  DEFAULT_AI_CONTEXT,
  DEFAULT_AI_SYSTEM_PROMPT,
} from "@/lib/ai-defaults";
import {
  buildDashboard,
  db,
  lossReasons,
  mockAdmin,
  mockAgent,
  mockUnits,
} from "./store";
import type { PipelineStage, User } from "@/lib/types";

/** Match any host so prod (Vercel) and local share the same handlers. */
const API = "*/api/v1";

function currentUser(auth: string | null) {
  return auth?.includes("admin") ? mockAdmin : mockAgent;
}

export const handlers = [
  http.post(`${API}/auth/login`, async ({ request }) => {
    const body = (await request.json()) as { email?: string; password?: string };
    const email = body.email?.toLowerCase();
    if (email === mockAdmin.email && body.password === "admin123") {
      return HttpResponse.json({
        access_token: "mock-access-admin",
        refresh_token: "mock-refresh-admin",
        expires_in: 900,
        user: mockAdmin,
      });
    }
    if (email === mockAgent.email && body.password === "agent123") {
      return HttpResponse.json({
        access_token: "mock-access-agent",
        refresh_token: "mock-refresh-agent",
        expires_in: 900,
        user: mockAgent,
      });
    }
    return HttpResponse.json(
      { code: "unauthorized", message: "Credenciais inválidas" },
      { status: 401 },
    );
  }),

  http.post(`${API}/auth/logout`, () => new HttpResponse(null, { status: 204 })),

  http.get(`${API}/me`, ({ request }) => {
    const user = currentUser(request.headers.get("Authorization"));
    const units =
      user.role === "admin"
        ? mockUnits
        : mockUnits.filter((u) => user.unit_ids?.includes(u.id));
    return HttpResponse.json({ ...user, units });
  }),

  http.get(`${API}/units`, () => HttpResponse.json({ items: mockUnits })),

  http.get(`${API}/users`, () =>
    HttpResponse.json({ items: db.users, total: db.users.length }),
  ),

  http.post(`${API}/users`, async ({ request }) => {
    const body = (await request.json()) as {
      email?: string;
      password?: string;
      name?: string;
      role?: User["role"];
      unit_ids?: string[];
    };
    const email = body.email?.trim().toLowerCase() ?? "";
    const name = body.name?.trim() ?? "";
    const password = body.password ?? "";
    const role = body.role;
    if (!email || !name || !role || password.length < 8) {
      return HttpResponse.json(
        { code: "bad_request", message: "Campos obrigatórios inválidos" },
        { status: 400 },
      );
    }
    if (db.users.some((u) => u.email === email)) {
      return HttpResponse.json(
        { code: "conflict", message: "Email já cadastrado" },
        { status: 409 },
      );
    }
    const user: User = {
      id: crypto.randomUUID(),
      email,
      name,
      role,
      active: true,
      unit_ids: (body.unit_ids ?? []).filter((id) =>
        mockUnits.some((u) => u.id === id),
      ),
    };
    db.users = [user, ...db.users];
    return HttpResponse.json(user, { status: 201 });
  }),

  http.patch(`${API}/users/:id`, async ({ params, request }) => {
    const id = String(params.id);
    const idx = db.users.findIndex((u) => u.id === id);
    if (idx < 0) {
      return HttpResponse.json(
        { code: "not_found", message: "Usuário não encontrado" },
        { status: 404 },
      );
    }
    const body = (await request.json()) as {
      name?: string;
      role?: User["role"];
      active?: boolean;
      password?: string;
    };
    if (body.password !== undefined && body.password.length < 8) {
      return HttpResponse.json(
        { code: "bad_request", message: "Senha deve ter no mínimo 8 caracteres" },
        { status: 400 },
      );
    }
    const next = { ...db.users[idx] };
    if (body.name !== undefined) next.name = body.name.trim();
    if (body.role !== undefined) next.role = body.role;
    if (body.active !== undefined) next.active = body.active;
    db.users[idx] = next;
    return HttpResponse.json(next);
  }),

  http.delete(`${API}/users/:id`, ({ params, request }) => {
    const id = String(params.id);
    const auth = request.headers.get("Authorization") ?? "";
    const self = currentUser(auth);
    if (self.id === id) {
      return HttpResponse.json(
        { code: "bad_request", message: "Você não pode remover a própria conta" },
        { status: 400 },
      );
    }
    const before = db.users.length;
    db.users = db.users.filter((u) => u.id !== id);
    if (db.users.length === before) {
      return HttpResponse.json(
        { code: "not_found", message: "Usuário não encontrado" },
        { status: 404 },
      );
    }
    return new HttpResponse(null, { status: 204 });
  }),

  http.put(`${API}/users/:id/units`, async ({ params, request }) => {
    const id = String(params.id);
    const idx = db.users.findIndex((u) => u.id === id);
    if (idx < 0) {
      return HttpResponse.json(
        { code: "not_found", message: "Usuário não encontrado" },
        { status: 404 },
      );
    }
    const body = (await request.json()) as { unit_ids?: string[] };
    const unit_ids = (body.unit_ids ?? []).filter((uid) =>
      mockUnits.some((u) => u.id === uid),
    );
    if ((body.unit_ids ?? []).length !== unit_ids.length) {
      return HttpResponse.json(
        { code: "bad_request", message: "Unidade inválida" },
        { status: 400 },
      );
    }
    db.users[idx] = { ...db.users[idx], unit_ids };
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${API}/inbox`, ({ request }) => {
    const url = new URL(request.url);
    const unitId = url.searchParams.get("unit_id");
    const stage = url.searchParams.get("stage");
    let items = [...db.inbox];
    if (unitId) items = items.filter((i) => i.unit_id === unitId);
    if (stage) items = items.filter((i) => i.pipeline_stage === stage);
    items.sort(
      (a, b) =>
        new Date(b.last_message_at).getTime() -
        new Date(a.last_message_at).getTime(),
    );
    return HttpResponse.json({ items, next_cursor: null });
  }),

  http.get(`${API}/conversations/:id`, ({ params }) => {
    const item = db.inbox.find((i) => i.id === params.id);
    if (!item) {
      return HttpResponse.json(
        { code: "not_found", message: "Conversa não encontrada" },
        { status: 404 },
      );
    }
    return HttpResponse.json(item);
  }),

  http.get(`${API}/conversations/:id/messages`, ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json({ items: db.messages[id] ?? [] });
  }),

  http.post(`${API}/conversations/:id/messages`, async ({ params, request }) => {
    const id = String(params.id);
    const body = (await request.json()) as { body?: string };
    const item = db.inbox.find((i) => i.id === id);
    if (!item) {
      return HttpResponse.json(
        { code: "not_found", message: "Conversa não encontrada" },
        { status: 404 },
      );
    }
    if (item.mode === "ai_triage") {
      return HttpResponse.json(
        { code: "forbidden", message: "Assuma a conversa antes de responder" },
        { status: 403 },
      );
    }
    if (!body.body?.trim()) {
      return HttpResponse.json(
        { code: "bad_request", message: "Mensagem vazia" },
        { status: 400 },
      );
    }
    const msg = {
      id: crypto.randomUUID(),
      conversation_id: id,
      direction: "out" as const,
      sender_type: "agent" as const,
      body: body.body.trim(),
      status: "sent" as const,
      created_at: new Date().toISOString(),
    };
    db.messages[id] = [...(db.messages[id] ?? []), msg];
    item.last_message_preview = msg.body;
    item.last_message_at = msg.created_at;
    return HttpResponse.json(msg, { status: 201 });
  }),

  http.post(`${API}/conversations/:id/claim`, ({ params, request }) => {
    const user = currentUser(request.headers.get("Authorization"));
    const item = db.inbox.find((i) => i.id === params.id);
    if (!item) {
      return HttpResponse.json(
        { code: "not_found", message: "Conversa não encontrada" },
        { status: 404 },
      );
    }
    if (item.owner_id && item.owner_id !== user.id) {
      return HttpResponse.json(
        { code: "conflict", message: "Conversa já atribuída" },
        { status: 409 },
      );
    }
    item.owner_id = user.id;
    item.mode = "human";
    const sys = {
      id: crypto.randomUUID(),
      conversation_id: item.id,
      direction: "out" as const,
      sender_type: "system" as const,
      body: `${user.name} assumiu o atendimento.`,
      status: "sent" as const,
      created_at: new Date().toISOString(),
    };
    db.messages[item.id] = [...(db.messages[item.id] ?? []), sys];
    return HttpResponse.json(item);
  }),

  http.patch(`${API}/conversations/:id/pipeline`, async ({ params, request }) => {
    const body = (await request.json()) as {
      stage: PipelineStage;
      reason_code?: string;
      reason_text?: string;
    };
    const item = db.inbox.find((i) => i.id === params.id);
    if (!item) {
      return HttpResponse.json(
        { code: "not_found", message: "Conversa não encontrada" },
        { status: 404 },
      );
    }
    if (body.stage === "nao_fechado" && !body.reason_code) {
      return HttpResponse.json(
        { code: "unprocessable", message: "Motivo obrigatório para Não fechado" },
        { status: 422 },
      );
    }
    item.pipeline_stage = body.stage;
    if (body.stage === "aguardando_fechamento" || body.stage === "nao_fechado") {
      const exists = db.followups.some(
        (f) => f.conversation_id === item.id && f.status === "open",
      );
      if (!exists) {
        db.followups.unshift({
          id: crypto.randomUUID(),
          conversation_id: item.id,
          contact_name: item.contact_name,
          contact_phone: item.contact_phone,
          unit_id: item.unit_id,
          pipeline_stage: item.pipeline_stage,
          due_at: new Date(Date.now() + 86400000).toISOString(),
          status: "open",
          note:
            body.stage === "nao_fechado"
              ? `Motivo: ${body.reason_code}${body.reason_text ? ` — ${body.reason_text}` : ""}`
              : "Aguardando fechamento — retomar contato",
        });
      }
    }
    return HttpResponse.json(item);
  }),

  http.get(`${API}/followups`, ({ request }) => {
    const url = new URL(request.url);
    const unitId = url.searchParams.get("unit_id");
    const status = url.searchParams.get("status") ?? "open";
    let items = [...db.followups];
    if (unitId) items = items.filter((f) => f.unit_id === unitId);
    if (status) items = items.filter((f) => f.status === status);
    return HttpResponse.json({ items });
  }),

  http.post(`${API}/followups/:id/complete`, ({ params }) => {
    const item = db.followups.find((f) => f.id === params.id);
    if (!item) {
      return HttpResponse.json(
        { code: "not_found", message: "Follow-up não encontrado" },
        { status: 404 },
      );
    }
    item.status = "done";
    return HttpResponse.json(item);
  }),

  http.get(`${API}/pops`, ({ request }) => {
    const intent = new URL(request.url).searchParams.get("intent");
    let items = db.pops.filter((p) => p.active);
    if (intent) {
      items = items.filter((p) => p.intent_tags.includes(intent));
    }
    return HttpResponse.json({ items });
  }),

  http.post(`${API}/pops`, async ({ request }) => {
    const body = (await request.json()) as {
      title?: string;
      body?: string;
      intent_tags?: string[];
    };
    const title = body.title?.trim() ?? "";
    const text = body.body?.trim() ?? "";
    if (!title || !text) {
      return HttpResponse.json(
        { code: "bad_request", message: "Título e texto são obrigatórios" },
        { status: 400 },
      );
    }
    const seen = new Set<string>();
    const intent_tags = (body.intent_tags ?? [])
      .map((t) => t.trim().toLowerCase())
      .filter((t) => {
        if (!t || seen.has(t)) return false;
        seen.add(t);
        return true;
      });
    const pop = {
      id: `p${Date.now()}`,
      title,
      body: text,
      intent_tags: intent_tags.length ? intent_tags : ["outro"],
      active: true,
    };
    db.pops = [pop, ...db.pops];
    return HttpResponse.json(pop, { status: 201 });
  }),

  http.patch(`${API}/pops/:id`, async ({ params, request }) => {
    const id = String(params.id);
    const idx = db.pops.findIndex((p) => p.id === id && p.active);
    if (idx < 0) {
      return HttpResponse.json(
        { code: "not_found", message: "POP não encontrado" },
        { status: 404 },
      );
    }
    const body = (await request.json()) as {
      title?: string;
      body?: string;
      intent_tags?: string[];
    };
    const title = body.title?.trim() ?? "";
    const text = body.body?.trim() ?? "";
    if (!title || !text) {
      return HttpResponse.json(
        { code: "bad_request", message: "Título e texto são obrigatórios" },
        { status: 400 },
      );
    }
    const seen = new Set<string>();
    const intent_tags = (body.intent_tags ?? [])
      .map((t) => t.trim().toLowerCase())
      .filter((t) => {
        if (!t || seen.has(t)) return false;
        seen.add(t);
        return true;
      });
    db.pops[idx] = {
      ...db.pops[idx],
      title,
      body: text,
      intent_tags: intent_tags.length ? intent_tags : ["outro"],
    };
    return HttpResponse.json(db.pops[idx]);
  }),

  http.delete(`${API}/pops/:id`, ({ params }) => {
    const id = String(params.id);
    const idx = db.pops.findIndex((p) => p.id === id && p.active);
    if (idx < 0) {
      return HttpResponse.json(
        { code: "not_found", message: "POP não encontrado" },
        { status: 404 },
      );
    }
    db.pops[idx] = { ...db.pops[idx], active: false };
    return new HttpResponse(null, { status: 204 });
  }),

  http.get(`${API}/loss-reasons`, () => HttpResponse.json({ items: lossReasons })),

  http.get(`${API}/dashboard/summary`, ({ request }) => {
    const unitId = new URL(request.url).searchParams.get("unit_id");
    return HttpResponse.json(buildDashboard(unitId));
  }),

  http.get(`${API}/settings/whatsapp`, () => {
    if (!db.settings.ai_system_prompt?.trim()) {
      db.settings.ai_system_prompt = DEFAULT_AI_SYSTEM_PROMPT;
    }
    if (!db.settings.ai_context?.trim()) {
      db.settings.ai_context = DEFAULT_AI_CONTEXT;
    }
    if (!db.settings.ai_campaigns) db.settings.ai_campaigns = [];
    return HttpResponse.json(db.settings);
  }),

  http.put(`${API}/settings/whatsapp`, async ({ request }) => {
    const body = (await request.json()) as Partial<typeof db.settings> & {
      token?: string;
    };
    if (typeof body.ai_enabled === "boolean") db.settings.ai_enabled = body.ai_enabled;
    if (body.display_phone !== undefined) db.settings.display_phone = body.display_phone;
    if (body.waba_id !== undefined) db.settings.waba_id = body.waba_id;
    if (body.phone_number_id !== undefined) db.settings.phone_number_id = body.phone_number_id;
    if (body.ai_system_prompt !== undefined)
      db.settings.ai_system_prompt = body.ai_system_prompt;
    if (body.ai_context !== undefined) db.settings.ai_context = body.ai_context;
    if (body.ai_campaigns !== undefined) {
      db.settings.ai_campaigns = body.ai_campaigns
        .filter((c) => c.title?.trim())
        .map((c) => ({
          ...c,
          id: c.id || `camp-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
          title: c.title.trim(),
          description: (c.description ?? "").trim(),
        }));
    }
    if (body.token) db.settings.token_masked = `${body.token.slice(0, 3)}••••••••`;
    return HttpResponse.json(db.settings);
  }),
];
