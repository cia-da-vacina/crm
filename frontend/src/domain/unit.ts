/** A physical/operational vaccination unit (clinic). */
export interface Unit {
  id: string;
  name: string;
  code: string;
  timezone: string;
  active: boolean;
  /** Street and number. */
  address: string;
  /** Neighborhood / bairro. */
  district?: string | null;
  /** Suite, floor, room, etc. */
  complement?: string | null;
  /** Landmark / how to find. */
  reference?: string | null;
  /** City name (often matches `name` for Cia da Vacina clinics). */
  city: string;
}

export type CreateUnitPayload = {
  name: string;
  code: string;
  timezone?: string;
  active?: boolean;
  address: string;
  district?: string | null;
  complement?: string | null;
  reference?: string | null;
  city: string;
};

export type UpdateUnitPayload = Partial<CreateUnitPayload>;
