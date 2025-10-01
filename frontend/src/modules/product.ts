import { getJson, postJson, putJson } from './http';

export interface Product {
  id?: string;
  name: string;
  sku: string;
  costPrice: number;
  salePrice: number;
  stock: number;
  category?: string;
  lowStockThreshold?: number;
  description?: string;
  imagePath?: string;
  thumbPath?: string;
  imageUrl?: string;
  thumbUrl?: string;
  imageHash?: string;
  imageWidth?: number;
  imageHeight?: number;
  imageSizeBytes?: number;
  thumbWidth?: number;
  thumbHeight?: number;
  thumbSizeBytes?: number;
  imageData?: string;
  createdAt?: string;
  updatedAt?: string;
  deletedAt?: string | null;
}

type ApiProduct = Product & {
  deletedAt?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

function adaptProduct(product: ApiProduct): Product {
  return {
    id: product.id,
    name: product.name,
    sku: product.sku,
    costPrice: product.costPrice,
    salePrice: product.salePrice,
    stock: product.stock,
    category: product.category,
    lowStockThreshold: product.lowStockThreshold,
    description: product.description,
    imagePath: product.imagePath,
    thumbPath: product.thumbPath,
    imageUrl: product.imageUrl,
    thumbUrl: product.thumbUrl,
    imageHash: product.imageHash,
    imageWidth: product.imageWidth,
    imageHeight: product.imageHeight,
    imageSizeBytes: product.imageSizeBytes,
    thumbWidth: product.thumbWidth,
    thumbHeight: product.thumbHeight,
    thumbSizeBytes: product.thumbSizeBytes,
    imageData: product.imageData,
    deletedAt: product.deletedAt ?? null,
    createdAt: product.createdAt,
    updatedAt: product.updatedAt
  };
}

export interface ProductListParams {
  page?: number;
  pageSize?: number;
  query?: string;
}

export interface ProductListResponse {
  items: Product[];
  total: number;
  page: number;
  pageSize: number;
  outOfStockCount: number;
  warningStockCount: number;
  lowStockHighlights: Product[];
}

type ApiProductListResponse = {
  items: ApiProduct[];
  total: number;
  page: number;
  pageSize: number;
  outOfStockCount: number;
  warningStockCount: number;
  lowStockHighlights: ApiProduct[];
};

function buildQuery(params?: ProductListParams): string {
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

export async function listProducts(params?: ProductListParams): Promise<ProductListResponse> {
  const response = await getJson<ApiProductListResponse>(`/products${buildQuery(params)}`);
  return {
    items: response.items.map(adaptProduct),
    total: response.total,
    page: response.page,
    pageSize: response.pageSize,
    outOfStockCount: response.outOfStockCount,
    warningStockCount: response.warningStockCount,
    lowStockHighlights: response.lowStockHighlights.map(adaptProduct)
  };
}

export async function saveProduct(product: Product): Promise<Product> {
  const payload: Record<string, unknown> = { ...product };
  delete payload.imageUrl;
  delete payload.thumbUrl;
  delete payload.imageMime;
  if (!payload.id) {
    delete payload.id;
    const created = await postJson<ApiProduct>('/products', payload);
    return adaptProduct(created);
  }
  const id = String(payload.id);
  const updated = await putJson<ApiProduct>(`/products/${id}`, payload);
  return adaptProduct(updated);
}

export async function adjustStock(productID: string, delta: number, reason: string): Promise<void> {
  await postJson(`/products/${productID}/adjust-stock`, { delta, reason });
}

export async function archiveProduct(productID: string): Promise<void> {
  await postJson(`/products/${productID}/archive`);
}
