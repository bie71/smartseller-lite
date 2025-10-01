import { deleteJson, getJson, postJson, putJson } from './http';

export interface AppSettings {
  brandName: string;
  logoPath?: string;
  logoUrl?: string;
  logoHash?: string;
  logoWidth?: number;
  logoHeight?: number;
  logoSizeBytes?: number;
  logoMime?: string;
  logoData?: string;
}

export interface BackupOptions {
  includeSchema: boolean;
  includeData: boolean;
  includeMedia?: boolean;
}

export interface RestoreOptions {
  disableForeignKeyChecks: boolean;
  useTransaction: boolean;
}

export interface RestoreResult {
  statements: number;
  durationMs: number;
  executionDriver: string;
}

export interface Courier {
  id?: string;
  code: string;
  name: string;
  services: string;
  trackingUrl: string;
  contact: string;
  notes: string;
  logoPath?: string;
  logoUrl?: string;
  logoHash?: string;
  logoWidth?: number;
  logoHeight?: number;
  logoSizeBytes?: number;
  logoMime?: string;
  logoData?: string;
  createdAt?: string;
  updatedAt?: string;
}

function normaliseDate(value: unknown): string | undefined {
  if (!value) {
    return undefined;
  }

  if (typeof value === 'string') {
    const date = new Date(value);
    if (!Number.isNaN(date.getTime())) {
      return date.toISOString();
    }
    return value;
  }

  if (typeof value === 'object') {
    const candidate = (value as Record<string, unknown>).Time ?? (value as Record<string, unknown>).time;
    if (typeof candidate === 'string') {
      const date = new Date(candidate);
      if (!Number.isNaN(date.getTime())) {
        return date.toISOString();
      }
      return candidate;
    }
  }

  return undefined;
}

type ApiSettings = AppSettings;

type ApiCourier = Courier & {
  createdAt?: string;
  updatedAt?: string;
};

function adaptSettings(settings: ApiSettings): AppSettings {
  return {
    brandName: settings.brandName,
    logoPath: settings.logoPath,
    logoUrl: settings.logoUrl,
    logoHash: settings.logoHash,
    logoWidth: settings.logoWidth,
    logoHeight: settings.logoHeight,
    logoSizeBytes: settings.logoSizeBytes,
    logoMime: settings.logoMime,
    logoData: settings.logoData
  };
}

function adaptCourier(courier: ApiCourier): Courier {
  return {
    id: courier.id,
    code: courier.code,
    name: courier.name,
    services: courier.services,
    trackingUrl: courier.trackingUrl,
    contact: courier.contact,
    notes: courier.notes,
    logoPath: courier.logoPath,
    logoUrl: courier.logoUrl,
    logoHash: courier.logoHash,
    logoWidth: courier.logoWidth,
    logoHeight: courier.logoHeight,
    logoSizeBytes: courier.logoSizeBytes,
    logoMime: courier.logoMime,
    logoData: courier.logoData,
    createdAt: normaliseDate(courier.createdAt),
    updatedAt: normaliseDate(courier.updatedAt)
  };
}

export async function getSettings(): Promise<AppSettings> {
  const settings = await getJson<ApiSettings>('/settings');
  return adaptSettings(settings);
}

export async function updateSettings(payload: AppSettings): Promise<AppSettings> {
  const data: Record<string, unknown> = { ...payload };
  delete data.logoUrl;
  const updated = await putJson<ApiSettings>('/settings', data);
  return adaptSettings(updated);
}

export interface CourierListParams {
  page?: number;
  pageSize?: number;
  query?: string;
}

export interface CourierListResponse {
  items: Courier[];
  total: number;
  page: number;
  pageSize: number;
}

type ApiCourierListResponse = {
  items: ApiCourier[];
  total: number;
  page: number;
  pageSize: number;
};

function buildCourierQuery(params?: CourierListParams): string {
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

export async function listCouriers(params?: CourierListParams): Promise<CourierListResponse> {
  const response = await getJson<ApiCourierListResponse>(`/couriers${buildCourierQuery(params)}`);
  return {
    items: response.items.map(adaptCourier),
    total: response.total,
    page: response.page,
    pageSize: response.pageSize
  };
}

export async function saveCourier(payload: Courier): Promise<Courier> {
  const data: Record<string, unknown> = { ...payload };
  delete data.logoUrl;
  if (!data.id) {
    delete data.id;
    const courier = await postJson<ApiCourier>('/couriers', data);
    return adaptCourier(courier);
  }
  const id = String(data.id);
  const courier = await putJson<ApiCourier>(`/couriers/${id}`, data);
  return adaptCourier(courier);
}

export async function deleteCourier(id: string): Promise<void> {
  await deleteJson(`/couriers/${id}`);
}

export async function createBackup(options: BackupOptions): Promise<string> {
  const response = await postJson<{ payload: string }>('/backups', {
    includeSchema: options.includeSchema,
    includeData: options.includeData,
    includeMedia: options.includeMedia ?? true
  });
  return response.payload;
}

export async function restoreBackup(payload: string, options: RestoreOptions): Promise<RestoreResult> {
  const response = await postJson<RestoreResult>('/backups/restore', {
    payload,
    disableForeignKeyChecks: options.disableForeignKeyChecks,
    useTransaction: options.useTransaction
  });
  return {
    statements: Number(response.statements ?? 0),
    durationMs: Number(response.durationMs ?? 0),
    executionDriver: response.executionDriver || 'driver'
  };
}
