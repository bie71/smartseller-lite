import { getJson, postJson } from './http';

export interface StockOpnameItem {
  id: string;
  stockOpnameId: string;
  productId: string;
  productName: string;
  productSku: string;
  counted: number;
  previousStock: number;
  difference: number;
}

export interface StockOpname {
  id: string;
  note: string;
  performedBy: string;
  performedAt: string;
  createdAt: string;
  items: StockOpnameItem[];
}

export interface PerformStockOpnameItemInput {
  productId: string;
  counted: number;
}

export interface PerformStockOpnamePayload {
  note: string;
  user: string;
  items: PerformStockOpnameItemInput[];
}

type ApiStockOpnameItem = StockOpnameItem;

type ApiStockOpname = {
  id: string;
  note: string;
  performedBy: string;
  performedAt: string;
  createdAt: string;
  items: ApiStockOpnameItem[];
};

function adaptItem(item: ApiStockOpnameItem): StockOpnameItem {
  return {
    id: item.id,
    stockOpnameId: item.stockOpnameId,
    productId: item.productId,
    productName: item.productName,
    productSku: item.productSku,
    counted: item.counted,
    previousStock: item.previousStock,
    difference: item.difference
  };
}

function adaptOpname(opname: ApiStockOpname): StockOpname {
  const performed = new Date(opname.performedAt as unknown as string);
  const created = new Date(opname.createdAt as unknown as string);
  return {
    id: opname.id,
    note: opname.note,
    performedBy: opname.performedBy,
    performedAt: Number.isNaN(performed.getTime()) ? (opname.performedAt as unknown as string) : performed.toISOString(),
    createdAt: Number.isNaN(created.getTime()) ? (opname.createdAt as unknown as string) : created.toISOString(),
    items: opname.items.map(adaptItem)
  };
}

export async function listStockOpnames(limit = 10): Promise<StockOpname[]> {
  const searchParams = new URLSearchParams();
  if (limit) {
    searchParams.set('limit', String(limit));
  }
  const query = searchParams.toString();
  const path = query ? `/stock-opnames?${query}` : '/stock-opnames';
  const results = await getJson<ApiStockOpname[] | null | undefined>(path);
  if (!results) {
    return [];
  }
  return results.map(adaptOpname);
}

export async function performStockOpname(payload: PerformStockOpnamePayload): Promise<StockOpname> {
  const response = await postJson<ApiStockOpname>('/stock-opnames', payload);
  return adaptOpname(response);
}
