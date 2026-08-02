import type {
  CreateUnitPayload,
  Paginated,
  Unit,
  UpdateUnitPayload,
} from "@/domain";
import { bffGet, bffPatch, bffPost } from "./http";

/** GET /api/proxy/units */
export async function list(): Promise<Paginated<Unit>> {
  return bffGet<Paginated<Unit>>("/api/proxy/units");
}

/** GET /api/proxy/units/:id */
export async function get(id: string): Promise<Unit> {
  return bffGet<Unit>(`/api/proxy/units/${id}`);
}

/** POST /api/proxy/units */
export async function create(payload: CreateUnitPayload): Promise<Unit> {
  return bffPost<Unit>("/api/proxy/units", payload);
}

/** PATCH /api/proxy/units/:id */
export async function update(
  id: string,
  payload: UpdateUnitPayload,
): Promise<Unit> {
  return bffPatch<Unit>(`/api/proxy/units/${id}`, payload);
}

export const unitsService = {
  list,
  get,
  create,
  update,
};
