import type { Intent, Pop } from "@/domain";
import { bffDelete, bffGet, bffPatch, bffPost, toQueryString } from "./http";

export interface ListPopsParams {
  intent?: Intent;
}

export interface PopPayload {
  title: string;
  body: string;
  intent_tags: Intent[];
  active?: boolean;
}

/** GET /api/proxy/pops — backend returns `{ items }`; service unwraps for callers. */
export async function list(params: ListPopsParams = {}): Promise<Pop[]> {
  const data = await bffGet<{ items: Pop[] }>(
    `/api/proxy/pops${toQueryString(params)}`,
  );
  return data.items;
}

/** GET /api/proxy/pops/:id */
export async function get(id: string): Promise<Pop> {
  return bffGet<Pop>(`/api/proxy/pops/${id}`);
}

/** POST /api/proxy/pops */
export async function create(payload: PopPayload): Promise<Pop> {
  return bffPost<Pop>("/api/proxy/pops", payload);
}

/** PATCH /api/proxy/pops/:id */
export async function update(id: string, payload: PopPayload): Promise<Pop> {
  return bffPatch<Pop>(`/api/proxy/pops/${id}`, payload);
}

/** DELETE /api/proxy/pops/:id */
export async function remove(id: string): Promise<void> {
  await bffDelete<void>(`/api/proxy/pops/${id}`);
}

export const popsService = {
  list,
  get,
  create,
  update,
  remove,
};
