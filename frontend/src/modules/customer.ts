import { getJson, postJson, putJson } from './http';

export type CustomerType = 'customer' | 'marketer' | 'reseller';

export interface Customer {
  id?: string;
  type: CustomerType;
  name: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postal?: string;
  notes?: string;
  createdAt?: string;
  updatedAt?: string;
}

type ApiCustomer = Customer & {
  createdAt?: string;
  updatedAt?: string;
};

function adaptCustomer(customer: ApiCustomer): Customer {
  return {
    id: customer.id,
    type: customer.type as CustomerType,
    name: customer.name,
    phone: customer.phone,
    email: customer.email,
    address: customer.address,
    city: customer.city,
    province: customer.province,
    postal: customer.postal,
    notes: customer.notes,
    createdAt: customer.createdAt,
    updatedAt: customer.updatedAt
  };
}

export async function listCustomers(): Promise<Customer[]> {
  const customers = await getJson<ApiCustomer[]>('/customers');
  return customers.map(adaptCustomer);
}

export async function saveCustomer(customer: Customer): Promise<Customer> {
  const payload: Record<string, unknown> = { ...customer };
  if (!payload.id) {
    delete payload.id;
    const created = await postJson<ApiCustomer>('/customers', payload);
    return adaptCustomer(created);
  }
  const id = String(payload.id);
  const updated = await putJson<ApiCustomer>(`/customers/${id}`, payload);
  return adaptCustomer(updated);
}
