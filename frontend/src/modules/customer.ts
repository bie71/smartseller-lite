import { deleteJson, getJson, postJson, putJson } from './http';

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

export interface CustomerListParams {
  page?: number;
  pageSize?: number;
  query?: string;
}

export interface CustomerListResponse {
  items: Customer[];
  total: number;
  page: number;
  pageSize: number;
  counts: {
    customer: number;
    marketer: number;
    reseller: number;
  };
}

type ApiCustomerListResponse = {
  items: ApiCustomer[];
  total: number;
  page: number;
  pageSize: number;
  counts: {
    customer: number;
    marketer: number;
    reseller: number;
  };
};

function buildCustomerQuery(params?: CustomerListParams): string {
  const searchParams = new URLSearchParams();
  if (params?.page) {
    searchParams.set('page', String(params.page));
  }
  if (params?.pageSize) {
    searchParams.set('pageSize', String(params.pageSize));
  }
  if (params?.query && params.query.trim().length > 0) {
    searchParams.set('q', params.query.trim());
  }
  const query = searchParams.toString();
  return query ? `?${query}` : '';
}

export async function listCustomers(params?: CustomerListParams): Promise<CustomerListResponse> {
  const response = await getJson<ApiCustomerListResponse>(`/customers${buildCustomerQuery(params)}`);
  return {
    items: response.items.map(adaptCustomer),
    total: response.total,
    page: response.page,
    pageSize: response.pageSize,
    counts: response.counts
  };
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

export async function deleteCustomer(customerID: string): Promise<void> {
  await deleteJson(`/customers/${customerID}`);
}
