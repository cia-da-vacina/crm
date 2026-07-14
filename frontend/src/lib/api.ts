import { getAccessToken } from "./auth-storage";
import { API_BASE } from "./api-base";
import type {
  DashboardSummary,
  FollowUp,
  InboxItem,
  LoginResponse,
  LossReason,
  MeResponse,
  Message,
  PipelineStage,
  Pop,
  Unit,
  User,
  WhatsAppSettings,
} from "./types";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  const token = getAccessToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new ApiError(res.status, data.code ?? "error", data.message ?? "Erro na API");
  }
  return data as T;
}

export const api = {
  login: (email: string, password: string) =>
    request<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  logout: () => request<void>("/auth/logout", { method: "POST" }),
  me: () => request<MeResponse>("/me"),
  listUnits: () => request<{ items: Unit[] }>("/units"),
  listUsers: () => request<{ items: User[]; total: number }>("/users"),

  createUser: (payload: {
    email: string;
    password: string;
    name: string;
    role: User["role"];
    unit_ids: string[];
  }) =>
    request<User>("/users", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  updateUser: (
    id: string,
    payload: {
      name?: string;
      role?: User["role"];
      active?: boolean;
      password?: string;
    },
  ) =>
    request<User>(`/users/${id}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),

  deleteUser: (id: string) =>
    request<void>(`/users/${id}`, { method: "DELETE" }),

  setUserUnits: (id: string, unit_ids: string[]) =>
    request<void>(`/users/${id}/units`, {
      method: "PUT",
      body: JSON.stringify({ unit_ids }),
    }),

  listInbox: (params?: { unit_id?: string; stage?: string }) => {
    const q = new URLSearchParams();
    if (params?.unit_id) q.set("unit_id", params.unit_id);
    if (params?.stage) q.set("stage", params.stage);
    const qs = q.toString();
    return request<{ items: InboxItem[]; next_cursor: string | null }>(
      `/inbox${qs ? `?${qs}` : ""}`,
    );
  },

  getConversation: (id: string) => request<InboxItem>(`/conversations/${id}`),

  listMessages: (conversationId: string) =>
    request<{ items: Message[] }>(`/conversations/${conversationId}/messages`),

  sendMessage: (conversationId: string, body: string) =>
    request<Message>(`/conversations/${conversationId}/messages`, {
      method: "POST",
      body: JSON.stringify({ body }),
    }),

  claim: (conversationId: string) =>
    request<InboxItem>(`/conversations/${conversationId}/claim`, { method: "POST" }),

  updatePipeline: (
    conversationId: string,
    payload: { stage: PipelineStage; reason_code?: string; reason_text?: string },
  ) =>
    request<InboxItem>(`/conversations/${conversationId}/pipeline`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),

  listFollowUps: (params?: { unit_id?: string; status?: string }) => {
    const q = new URLSearchParams();
    if (params?.unit_id) q.set("unit_id", params.unit_id);
    if (params?.status) q.set("status", params.status);
    const qs = q.toString();
    return request<{ items: FollowUp[] }>(`/followups${qs ? `?${qs}` : ""}`);
  },

  completeFollowUp: (id: string) =>
    request<FollowUp>(`/followups/${id}/complete`, { method: "POST" }),

  listPops: (intent?: string) => {
    const q = intent ? `?intent=${encodeURIComponent(intent)}` : "";
    return request<{ items: Pop[] }>(`/pops${q}`);
  },

  createPop: (payload: { title: string; body: string; intent_tags: string[] }) =>
    request<Pop>("/pops", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  updatePop: (
    id: string,
    payload: { title: string; body: string; intent_tags: string[] },
  ) =>
    request<Pop>(`/pops/${id}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),

  deletePop: (id: string) =>
    request<void>(`/pops/${id}`, { method: "DELETE" }),

  listLossReasons: () => request<{ items: LossReason[] }>("/loss-reasons"),

  dashboard: (unitId?: string) => {
    const q = unitId ? `?unit_id=${unitId}` : "";
    return request<DashboardSummary>(`/dashboard/summary${q}`);
  },

  getWhatsAppSettings: () => request<WhatsAppSettings>("/settings/whatsapp"),

  updateWhatsAppSettings: (payload: Partial<WhatsAppSettings> & { token?: string }) =>
    request<WhatsAppSettings>("/settings/whatsapp", {
      method: "PUT",
      body: JSON.stringify(payload),
    }),
};
