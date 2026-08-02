import type { Customer, CustomerIdentity, Paginated } from "@/domain";
import { bffGet, toQueryString } from "./http";

export interface ListCustomersParams {
  q?: string;
  unit_id?: string;
}

/** GET /api/proxy/customers/:id */
export async function get(id: string): Promise<Customer> {
  return bffGet<Customer>(`/api/proxy/customers/${id}`);
}

/** GET /api/proxy/customers */
export async function list(
  params: ListCustomersParams = {},
): Promise<Paginated<Customer>> {
  return bffGet<Paginated<Customer>>(`/api/proxy/customers${toQueryString(params)}`);
}

/** GET /api/proxy/customers/:id/identities */
export async function listIdentities(customerId: string): Promise<CustomerIdentity[]> {
  return bffGet<CustomerIdentity[]>(`/api/proxy/customers/${customerId}/identities`);
}

export const customersService = {
  get,
  list,
  listIdentities,
};
