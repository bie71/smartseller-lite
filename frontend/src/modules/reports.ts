import { API_BASE } from './http';

export interface OrderExportFilters {
  search?: string;
  start?: string;
  end?: string;
  courier?: string;
}

export async function fetchOrdersCsv(filters: OrderExportFilters = {}): Promise<Blob> {
  const params = new URLSearchParams();
  if (filters.search) {
    params.set('search', filters.search);
  }
  if (filters.start) {
    params.set('start', filters.start);
  }
  if (filters.end) {
    params.set('end', filters.end);
  }
  if (filters.courier && filters.courier !== 'all') {
    params.set('courier', filters.courier);
  }
  const query = params.toString();
  const url = `${API_BASE}/orders/export.csv${query ? `?${query}` : ''}`;

  const response = await fetch(url, {
    headers: {
      Accept: 'text/csv'
    }
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || 'Gagal mengekspor CSV.');
  }
  return await response.blob();
}
