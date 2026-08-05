/**
 * In-process router that answers backend-shaped paths while USE_MOCKS=true.
 * Paths are relative to `/api/v1` (same as `backendFetch` / proxy).
 */

import type {
  ConversationDetail,
  Intent,
  MetaSettings,
  PipelineStage,
  UpdateMetaSettingsPayload,
  User,
  UserRole,
} from "@/domain";
import { MOCK_PASSWORDS } from "./data";
import {
  buildDashboard,
  db,
  lossReasons,
  mockAdmin,
  mockUnits,
  newId,
  unitsForUser,
  userFromToken,
} from "./store";

export interface MockRequest {
  method: string;
  /** Path relative to API root, e.g. `/inbox` or `/conversations/abc/messages`. */
  path: string;
  search?: string;
  body?: string;
  token?: string | null;
}

export interface MockResponse {
  status: number;
  body?: unknown;
}

function json(status: number, body: unknown): MockResponse {
  return { status, body };
}

function noContent(): MockResponse {
  return { status: 204 };
}

function error(status: number, code: string, message: string): MockResponse {
  return { status, body: { code, message } };
}

function parseBody<T>(raw?: string): T {
  if (!raw) return {} as T;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return {} as T;
  }
}

function qs(search?: string): URLSearchParams {
  if (!search) return new URLSearchParams();
  return new URLSearchParams(search.startsWith("?") ? search.slice(1) : search);
}

function match(
  path: string,
  pattern: string,
): Record<string, string> | null {
  const pathParts = path.replace(/\/+$/, "").split("/").filter(Boolean);
  const patternParts = pattern.replace(/\/+$/, "").split("/").filter(Boolean);
  if (pathParts.length !== patternParts.length) return null;
  const params: Record<string, string> = {};
  for (let i = 0; i < patternParts.length; i++) {
    const p = patternParts[i];
    const v = pathParts[i];
    if (p.startsWith(":")) {
      params[p.slice(1)] = decodeURIComponent(v);
    } else if (p !== v) {
      return null;
    }
  }
  return params;
}

function toInboxItem(c: ConversationDetail) {
  const { customer: _customer, created_at: _ca, updated_at: _ua, ...summary } = c;
  return summary;
}

export function routeMockRequest(req: MockRequest): MockResponse {
  const method = req.method.toUpperCase();
  const path = req.path.startsWith("/") ? req.path : `/${req.path}`;
  const params = qs(req.search);
  const authUser = userFromToken(req.token);

  // --- Auth ---
  if (method === "POST" && match(path, "/auth/login")) {
    const body = parseBody<{ email?: string; password?: string }>(req.body);
    const email = body.email?.trim().toLowerCase() ?? "";
    const expected = MOCK_PASSWORDS[email];
    if (!expected || body.password !== expected) {
      return error(401, "unauthorized", "Credenciais inválidas");
    }
    const user = email === mockAdmin.email ? mockAdmin : authUser;
    const resolved =
      email === mockAdmin.email
        ? mockAdmin
        : db.users.find((u) => u.email === email) ?? authUser;
    const roleTag = resolved.role === "admin" ? "admin" : "agent";
    return json(200, {
      access_token: `mock-access-${roleTag}`,
      refresh_token: `mock-refresh-${roleTag}`,
      expires_in: 3600,
      user: resolved,
    });
  }

  if (method === "POST" && match(path, "/auth/refresh")) {
    const body = parseBody<{ refresh_token?: string }>(req.body);
    const refresh = body.refresh_token ?? "";
    if (!refresh.startsWith("mock-refresh-")) {
      return error(401, "unauthorized", "Refresh inválido");
    }
    const roleTag = refresh.includes("admin") ? "admin" : "agent";
    const user = roleTag === "admin" ? mockAdmin : userFromToken(`mock-access-${roleTag}`);
    return json(200, {
      access_token: `mock-access-${roleTag}`,
      refresh_token: `mock-refresh-${roleTag}`,
      expires_in: 3600,
      user,
    });
  }

  if (method === "POST" && match(path, "/auth/logout")) {
    return noContent();
  }

  if (method === "GET" && match(path, "/me")) {
    if (!req.token) return error(401, "unauthorized", "Não autenticado");
    const user = userFromToken(req.token);
    return json(200, { ...user, units: unitsForUser(user) });
  }

  // --- Units / users ---
  if (method === "GET" && match(path, "/units")) {
    const items = unitsForUser(authUser);
    return json(200, {
      items,
      total: items.length,
      page: 1,
      page_size: items.length || 20,
    });
  }

  {
    const m = match(path, "/units/:id");
    if (method === "GET" && m) {
      const unit = mockUnits.find((u) => u.id === m.id);
      if (!unit) return error(404, "not_found", "Unidade não encontrada");
      return json(200, unit);
    }
  }

  if (method === "GET" && match(path, "/users")) {
    return json(200, {
      items: db.users,
      total: db.users.length,
      page: 1,
      page_size: db.users.length || 20,
    });
  }

  if (method === "POST" && match(path, "/users")) {
    const body = parseBody<{
      email?: string;
      password?: string;
      name?: string;
      role?: UserRole;
      unit_ids?: string[];
    }>(req.body);
    if (!body.email || !body.name || !body.role || (body.password?.length ?? 0) < 8) {
      return error(400, "bad_request", "Campos obrigatórios inválidos");
    }
    const user: User = {
      id: newId("user"),
      email: body.email.trim().toLowerCase(),
      name: body.name.trim(),
      role: body.role,
      active: true,
      unit_ids: body.unit_ids ?? [],
    };
    db.users.push(user);
    return json(201, user);
  }

  {
    const m = match(path, "/users/:id");
    if (m) {
      const idx = db.users.findIndex((u) => u.id === m.id);
      if (idx < 0) return error(404, "not_found", "Usuário não encontrado");
      if (method === "GET") return json(200, db.users[idx]);
      if (method === "PATCH") {
        const body = parseBody<Partial<User> & { password?: string }>(req.body);
        db.users[idx] = {
          ...db.users[idx],
          name: body.name ?? db.users[idx].name,
          role: body.role ?? db.users[idx].role,
          active: body.active ?? db.users[idx].active,
        };
        return json(200, db.users[idx]);
      }
      if (method === "DELETE") {
        db.users.splice(idx, 1);
        return noContent();
      }
    }
  }

  {
    const m = match(path, "/users/:id/units");
    if (method === "PUT" && m) {
      const idx = db.users.findIndex((u) => u.id === m.id);
      if (idx < 0) return error(404, "not_found", "Usuário não encontrado");
      const body = parseBody<{ unit_ids?: string[] }>(req.body);
      db.users[idx] = { ...db.users[idx], unit_ids: body.unit_ids ?? [] };
      return noContent();
    }
  }

  // --- Customers ---
  if (method === "GET" && match(path, "/customers")) {
    const q = (params.get("q") ?? "").toLowerCase();
    const unitId = params.get("unit_id");
    let items = db.customers;
    if (unitId) items = items.filter((c) => c.unit_id === unitId);
    if (q) {
      items = items.filter(
        (c) =>
          c.display_name.toLowerCase().includes(q) ||
          (c.primary_phone ?? "").includes(q),
      );
    }
    return json(200, {
      items,
      total: items.length,
      page: 1,
      page_size: items.length || 20,
    });
  }

  {
    const m = match(path, "/customers/:id");
    if (method === "GET" && m) {
      const c = db.customers.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Cliente não encontrado");
      return json(200, c);
    }
  }

  {
    const m = match(path, "/customers/:id/identities");
    if (method === "GET" && m) {
      const c = db.customers.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Cliente não encontrado");
      return json(200, c.identities);
    }
  }

  // --- Inbox / conversations ---
  if (method === "GET" && match(path, "/inbox")) {
    const unitId = params.get("unit_id");
    const stage = params.get("stage");
    const channel = params.get("channel");
    const mode = params.get("mode");
    let items = db.conversations.map(toInboxItem);
    if (unitId) items = items.filter((c) => c.unit_id === unitId);
    if (stage) items = items.filter((c) => c.pipeline_stage === stage);
    if (channel) items = items.filter((c) => c.channel === channel);
    if (mode) items = items.filter((c) => c.mode === mode);
    items = [...items].sort(
      (a, b) =>
        new Date(b.last_message_at).getTime() -
        new Date(a.last_message_at).getTime(),
    );
    return json(200, { items, next_cursor: null });
  }

  {
    const m = match(path, "/conversations/:id");
    if (method === "GET" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      return json(200, c);
    }
  }

  {
    const m = match(path, "/conversations/:id/messages");
    if (method === "GET" && m) {
      return json(200, {
        items: db.messages[m.id] ?? [],
        next_cursor: null,
      });
    }
    if (method === "POST" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      if (c.mode === "ai_triage") {
        return error(403, "forbidden", "Assuma a conversa antes de responder");
      }
      const body = parseBody<{ body?: string; kind?: string }>(req.body);
      const text = body.body?.trim() ?? "";
      if (!text) return error(400, "bad_request", "Mensagem vazia");
      const msg = {
        id: newId("msg"),
        conversation_id: m.id,
        direction: "out" as const,
        sender_type: "agent" as const,
        kind: (body.kind ?? "text") as "text",
        channel: c.channel,
        body: text,
        status: "sent" as const,
        created_at: new Date().toISOString(),
      };
      db.messages[m.id] = [...(db.messages[m.id] ?? []), msg];
      c.last_message_preview = text;
      c.last_message_at = msg.created_at;
      c.updated_at = msg.created_at;
      return json(201, msg);
    }
  }

  {
    const m = match(path, "/conversations/:id/claim");
    if (method === "POST" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      if (c.owner_id && c.owner_id !== authUser.id) {
        return error(409, "conflict", "Conversa já atribuída");
      }
      c.owner_id = authUser.id;
      c.mode = "human";
      c.updated_at = new Date().toISOString();
      const systemMsg = {
        id: newId("msg"),
        conversation_id: m.id,
        direction: "out" as const,
        sender_type: "system" as const,
        kind: "system" as const,
        channel: c.channel,
        body: `${authUser.name} assumiu o atendimento.`,
        status: "sent" as const,
        created_at: c.updated_at,
      };
      db.messages[m.id] = [...(db.messages[m.id] ?? []), systemMsg];
      return json(200, c);
    }
  }

  {
    const m = match(path, "/conversations/:id/pipeline");
    if (method === "PATCH" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      const body = parseBody<{
        stage?: PipelineStage;
        reason_code?: string;
        reason_text?: string;
      }>(req.body);
      if (!body.stage) return error(400, "bad_request", "stage obrigatório");
      if (body.stage === "nao_fechado" && !body.reason_code) {
        return error(422, "unprocessable", "Motivo obrigatório para Não fechado");
      }
      c.pipeline_stage = body.stage;
      c.updated_at = new Date().toISOString();
      if (
        body.stage === "aguardando_fechamento" ||
        body.stage === "nao_fechado"
      ) {
        db.followups.push({
          id: newId("fu"),
          conversation_id: c.id,
          customer_id: c.customer_id,
          customer_name: c.customer_name,
          customer_phone: c.customer_phone,
          unit_id: c.unit_id,
          pipeline_stage: body.stage,
          due_at: new Date(Date.now() + 86400_000).toISOString(),
          status: "open",
          note:
            body.stage === "nao_fechado"
              ? `Motivo: ${body.reason_code}${body.reason_text ? ` — ${body.reason_text}` : ""}`
              : "Aguardando fechamento — retomar contato",
          created_at: new Date().toISOString(),
        });
      }
      return json(200, c);
    }
  }

  {
    const m = match(path, "/conversations/:id/triage");
    if (method === "GET" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      if (c.mode !== "ai_triage") {
        return error(404, "not_found", "Triagem não aplicável");
      }
      return json(200, {
        conversation_id: c.id,
        intent: c.intent ?? "outro",
        confidence: 0.86,
        summary: c.ai_summary ?? "Coletando necessidade…",
        suggested_pops: db.pops
          .filter((p) => c.intent && p.intent_tags.includes(c.intent as Intent))
          .map((p) => p.id),
        ready_for_handoff: c.phone_gate === "collected" || c.phone_gate === "not_needed",
        phone_gate: c.phone_gate,
        pending_phone_masked: c.pending_phone_masked,
        collected_fields: c.intent
          ? { vacina: "gripe", unidade: "Centro" }
          : undefined,
      });
    }
  }

  {
    const m = match(path, "/conversations/:id/phone");
    if (method === "POST" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      const body = parseBody<{ phone_e164?: string }>(req.body);
      const phone = body.phone_e164?.trim() ?? "";
      if (!/^\+\d{10,15}$/.test(phone)) {
        return error(400, "bad_request", "Telefone E.164 inválido");
      }
      c.phone_gate = "pending_verification";
      c.pending_phone_masked = `${phone.slice(0, 3)}•••••${phone.slice(-4)}`;
      c.updated_at = new Date().toISOString();
      return json(202, c);
    }
  }

  {
    const m = match(path, "/conversations/:id/phone/confirm");
    if (method === "POST" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      const body = parseBody<{ code?: string }>(req.body);
      if ((body.code ?? "") !== "123456") {
        return error(400, "bad_request", "Código inválido (use 123456 no mock)");
      }
      const phone = "+5511988887777";
      c.identification = "identified";
      c.phone_gate = "collected";
      c.customer_phone = phone;
      c.pending_phone_masked = null;
      c.customer = {
        ...c.customer,
        identification: "identified",
        primary_phone: phone,
      };
      c.updated_at = new Date().toISOString();
      return json(200, c);
    }
  }

  {
    const m = match(path, "/conversations/:id/phone/resend");
    if (method === "POST" && m) {
      const c = db.conversations.find((x) => x.id === m.id);
      if (!c) return error(404, "not_found", "Conversa não encontrada");
      return json(200, c);
    }
  }

  // --- Engagements ---
  if (method === "GET" && match(path, "/engagements")) {
    const unitId = params.get("unit_id");
    const status = params.get("status");
    const channel = params.get("channel");
    const type = params.get("type");
    let items = [...db.engagements];
    if (unitId) items = items.filter((e) => e.unit_id === unitId);
    if (status) items = items.filter((e) => e.status === status);
    if (channel) items = items.filter((e) => e.channel === channel);
    if (type) items = items.filter((e) => e.type === type);
    items.sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
    return json(200, { items, next_cursor: null });
  }

  {
    const m = match(path, "/engagements/:id");
    if (method === "GET" && m) {
      const e = db.engagements.find((x) => x.id === m.id);
      if (!e) return error(404, "not_found", "Engagement não encontrado");
      return json(200, e);
    }
  }

  {
    const m = match(path, "/engagements/:id/reply");
    if (method === "POST" && m) {
      const e = db.engagements.find((x) => x.id === m.id);
      if (!e) return error(404, "not_found", "Engagement não encontrado");
      e.status = "replied";
      e.replied_at = new Date().toISOString();
      return json(200, e);
    }
  }

  {
    const m = match(path, "/engagements/:id/dismiss");
    if (method === "POST" && m) {
      const e = db.engagements.find((x) => x.id === m.id);
      if (!e) return error(404, "not_found", "Engagement não encontrado");
      e.status = "dismissed";
      return json(200, e);
    }
  }

  {
    const m = match(path, "/engagements/:id/convert");
    if (method === "POST" && m) {
      const e = db.engagements.find((x) => x.id === m.id);
      if (!e) return error(404, "not_found", "Engagement não encontrado");
      const now = new Date().toISOString();
      const customerId = e.customer_id ?? newId("cust");
      const customerName = e.customer_name ?? "Contato social";
      const conv: ConversationDetail = {
        id: newId("conv"),
        customer_id: customerId,
        customer_name: customerName,
        customer_phone: null,
        identification: "anonymous",
        phone_gate: e.channel === "whatsapp" ? "collected" : "required",
        channel: e.channel,
        unit_id: e.unit_id,
        pipeline_stage: "em_atendimento",
        mode: "human",
        status: "open",
        owner_id: authUser.id,
        intent: null,
        ai_summary: `Convertido de ${e.type}`,
        last_message_preview: e.body,
        last_message_at: now,
        customer: {
          id: customerId,
          display_name: customerName,
          identification: "anonymous",
          primary_phone: null,
          unit_id: e.unit_id,
          identities: [],
          created_at: now,
          updated_at: now,
        },
        created_at: now,
        updated_at: now,
      };
      db.conversations.unshift(conv);
      db.messages[conv.id] = [
        {
          id: newId("msg"),
          conversation_id: conv.id,
          direction: "in",
          sender_type: "contact",
          kind: "text",
          channel: e.channel,
          body: e.body,
          status: "read",
          reply_to_engagement_id: e.id,
          created_at: now,
        },
      ];
      e.status = "converted_to_conversation";
      e.conversation_id = conv.id;
      return json(201, conv);
    }
  }

  // --- Follow-ups ---
  if (method === "GET" && match(path, "/followups")) {
    const unitId = params.get("unit_id");
    const status = params.get("status") || "open";
    let items = [...db.followups];
    if (unitId) items = items.filter((f) => f.unit_id === unitId);
    if (status) items = items.filter((f) => f.status === status);
    return json(200, { items, next_cursor: null });
  }

  {
    const m = match(path, "/followups/:id/complete");
    if (method === "POST" && m) {
      const f = db.followups.find((x) => x.id === m.id);
      if (!f) return error(404, "not_found", "Follow-up não encontrado");
      f.status = "done";
      f.completed_at = new Date().toISOString();
      return json(200, f);
    }
  }

  {
    const m = match(path, "/followups/:id/cancel");
    if (method === "POST" && m) {
      const f = db.followups.find((x) => x.id === m.id);
      if (!f) return error(404, "not_found", "Follow-up não encontrado");
      f.status = "canceled";
      f.completed_at = new Date().toISOString();
      return json(200, f);
    }
  }

  // --- POPs ---
  if (method === "GET" && match(path, "/pops")) {
    const intent = params.get("intent");
    let items = db.pops.filter((p) => p.active);
    if (intent) {
      items = items.filter((p) => p.intent_tags.includes(intent as Intent));
    }
    return json(200, { items });
  }

  if (method === "POST" && match(path, "/pops")) {
    const body = parseBody<{
      title?: string;
      body?: string;
      intent_tags?: Intent[];
    }>(req.body);
    if (!body.title?.trim() || !body.body?.trim()) {
      return error(400, "bad_request", "Título e texto são obrigatórios");
    }
    const pop = {
      id: newId("pop"),
      title: body.title.trim(),
      body: body.body.trim(),
      intent_tags: (body.intent_tags?.length ? body.intent_tags : ["outro"]) as Intent[],
      active: true,
    };
    db.pops.unshift(pop);
    return json(201, pop);
  }

  {
    const m = match(path, "/pops/:id");
    if (m) {
      const idx = db.pops.findIndex((p) => p.id === m.id);
      if (idx < 0) return error(404, "not_found", "POP não encontrado");
      if (method === "GET") return json(200, db.pops[idx]);
      if (method === "PATCH") {
        const body = parseBody<{
          title?: string;
          body?: string;
          intent_tags?: Intent[];
          active?: boolean;
        }>(req.body);
        db.pops[idx] = {
          ...db.pops[idx],
          title: body.title?.trim() || db.pops[idx].title,
          body: body.body?.trim() || db.pops[idx].body,
          intent_tags: body.intent_tags ?? db.pops[idx].intent_tags,
          active: body.active ?? db.pops[idx].active,
        };
        return json(200, db.pops[idx]);
      }
      if (method === "DELETE") {
        db.pops[idx] = { ...db.pops[idx], active: false };
        return noContent();
      }
    }
  }

  if (method === "GET" && match(path, "/loss-reasons")) {
    return json(200, { items: lossReasons });
  }

  if (method === "GET" && match(path, "/dashboard/summary")) {
    return json(200, buildDashboard(params.get("unit_id")));
  }

  // --- Meta settings ---
  if (method === "GET" && match(path, "/settings/meta")) {
    return json(200, db.settings);
  }

  if (method === "PUT" && match(path, "/settings/meta")) {
    const body = parseBody<UpdateMetaSettingsPayload>(req.body);
    const next = { ...db.settings };
    if (body.ai_enabled !== undefined) next.ai_enabled = body.ai_enabled;
    if (body.ai_system_prompt !== undefined) next.ai_system_prompt = body.ai_system_prompt;
    if (body.ai_context !== undefined) next.ai_context = body.ai_context;
    if (body.ai_campaigns !== undefined) next.ai_campaigns = body.ai_campaigns;
    if (body.triage_enabled !== undefined) next.triage_enabled = body.triage_enabled;
    if (body.triage_handoff_intents !== undefined) {
      next.triage_handoff_intents = body.triage_handoff_intents;
    }
    if (body.channels) {
      next.channels = next.channels.map((ch) => {
        const patch = body.channels?.find((c) => c.channel === ch.channel);
        if (!patch) return ch;
        const token = body.channel_tokens?.[ch.channel];
        return {
          ...ch,
          ...patch,
          channel: ch.channel,
          token_masked: token
            ? `${token.slice(0, 3)}••••••••`
            : ch.token_masked,
        };
      });
    }
    db.settings = next as MetaSettings;
    return json(200, db.settings);
  }

  return error(404, "not_found", `Mock sem rota para ${method} ${path}`);
}
