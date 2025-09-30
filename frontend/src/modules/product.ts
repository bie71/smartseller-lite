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

export async function listProducts(): Promise<Product[]> {
  const products = await getJson<ApiProduct[]>('/products');
  return products.map(adaptProduct);
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
