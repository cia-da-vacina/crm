import type { Paginated, User, UserRole } from "@/domain";
import { bffDelete, bffGet, bffPatch, bffPost, bffPut } from "./http";

export interface CreateUserPayload {
  email: string;
  password: string;
  name: string;
  role: UserRole;
  unit_ids: string[];
}

export interface UpdateUserPayload {
  name?: string;
  role?: UserRole;
  active?: boolean;
  password?: string;
}

/** GET /api/proxy/users */
export async function list(): Promise<Paginated<User>> {
  return bffGet<Paginated<User>>("/api/proxy/users");
}

/** GET /api/proxy/users/:id */
export async function get(id: string): Promise<User> {
  return bffGet<User>(`/api/proxy/users/${id}`);
}

/** POST /api/proxy/users */
export async function create(payload: CreateUserPayload): Promise<User> {
  return bffPost<User>("/api/proxy/users", payload);
}

/** PATCH /api/proxy/users/:id */
export async function update(id: string, payload: UpdateUserPayload): Promise<User> {
  return bffPatch<User>(`/api/proxy/users/${id}`, payload);
}

/** DELETE /api/proxy/users/:id */
export async function remove(id: string): Promise<void> {
  await bffDelete<void>(`/api/proxy/users/${id}`);
}

/** PUT /api/proxy/users/:id/units */
export async function setUnits(id: string, unitIds: string[]): Promise<void> {
  await bffPut<void>(`/api/proxy/users/${id}/units`, { unit_ids: unitIds });
}

export const usersService = {
  list,
  get,
  create,
  update,
  remove,
  setUnits,
};
