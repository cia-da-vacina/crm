import type { UserRole } from "./enums";
import type { Unit } from "./unit";

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  active: boolean;
  /** Units this user is allowed to operate in. Absent for global roles. */
  unit_ids?: string[];
}

/** Response of `GET /me`: the authenticated user plus the units they can access. */
export interface MeResponse extends User {
  units: Unit[];
}
