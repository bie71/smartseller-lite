import { deleteJson, getJson, postJson } from './http';
import type { Customer } from './customer';
import type { Product } from './product';

export interface OrderItemInput {
  productId: string;
  quantity: number;
  unitPrice: number;
  discount: number;
}

export interface CreateOrderPayload {
  buyerId: string;
  recipientId: string;
  items: OrderItemInput[];
  discount: number;
  notes: string;
  courier: string;
  serviceLevel: string;
  trackingCode: string;
  shippingCost: number;
}

export interface OrderItem {
  id: string;
  orderId: string;
  productId: string;
  sku: string;
  quantity: number;
  unitPrice: number;
  discount: number;
  costPrice: number;
  profit: number;
}

export interface Order {
  id: string;
  code: string;
  buyerId: string;
  recipientId: string;
  shipment: {
    courier: string;
    serviceLevel: string;
    trackingCode: string;
    shippingCost: number;
  };
  items: OrderItem[];
  discount: number;
  total: number;
  profit: number;
  notes: string;
  createdAt: string;
  updatedAt: string;
}

type ApiOrderItem = OrderItem;

type ApiOrder = {
  id: string;
  code: string;
  buyerId: string;
  recipientId: string;
  shipment: {
    courier: string;
    serviceLevel: string;
    trackingCode: string;
    shippingCost: number;
  };
  items: ApiOrderItem[];
  discount: number;
  total: number;
  profit: number;
  notes: string;
  createdAt: string;
  updatedAt: string;
};

function adaptOrderItem(item: ApiOrderItem): OrderItem {
  return {
    id: item.id,
    orderId: item.orderId,
    productId: item.productId,
    sku: item.sku,
    quantity: item.quantity,
    unitPrice: item.unitPrice,
    discount: item.discount,
    costPrice: item.costPrice,
    profit: item.profit
  };
}

function adaptOrder(order: ApiOrder): Order {
  return {
    id: order.id,
    code: order.code,
    buyerId: order.buyerId,
    recipientId: order.recipientId,
    shipment: {
      courier: order.shipment.courier,
      serviceLevel: order.shipment.serviceLevel,
      trackingCode: order.shipment.trackingCode,
      shippingCost: order.shipment.shippingCost
    },
    items: order.items.map(adaptOrderItem),
    discount: order.discount,
    total: order.total,
    profit: order.profit,
    notes: order.notes,
    createdAt: order.createdAt,
    updatedAt: order.updatedAt
  };
}

export async function listOrders(limit?: number): Promise<Order[]> {
  const searchParams = new URLSearchParams();
  if (limit) {
    searchParams.set('limit', String(limit));
  }
  const query = searchParams.toString();
  const path = query ? `/orders?${query}` : '/orders';
  const orders = await getJson<ApiOrder[]>(path);
  return orders.map(adaptOrder);
}

export async function createOrder(payload: CreateOrderPayload): Promise<Order> {
  const order = await postJson<ApiOrder>('/orders', payload);
  return adaptOrder(order);
}

export async function deleteOrder(orderId: string): Promise<void> {
  await deleteJson(`/orders/${orderId}`);
}

export async function generateLabel(orderId: string): Promise<string> {
  const response = await postJson<{ base64: string }>(`/orders/${orderId}/label`);
  return response.base64;
}

function pdfBlobFromBase64(base64: string): Blob {
  const payload = (base64 || '').replace(/\s+/g, '');
  if (!payload) {
    throw new Error('label payload kosong');
  }
  let byteString: string;
  try {
    byteString = atob(payload);
  } catch (error) {
    throw new Error('label tidak valid');
  }
  const buffer = new Uint8Array(byteString.length);
  for (let i = 0; i < byteString.length; i += 1) {
    buffer[i] = byteString.charCodeAt(i);
  }
  return new Blob([buffer], { type: 'application/pdf' });
}

export function downloadLabel(order: Order, base64: string) {
  const blob = pdfBlobFromBase64(base64);
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = `${order.code}-label.pdf`;
  anchor.style.display = 'none';
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(url);
}

export function ensureLabelBlob(base64: string): Blob {
  return pdfBlobFromBase64(base64);
}

export interface UiOrderItem extends OrderItemInput {
  product?: Product;
}

export interface UiFormState {
  buyer?: Customer;
  recipient?: Customer;
  courier: string;
  serviceLevel: string;
  trackingCode: string;
  discount: number;
  shippingCost: number;
  notes: string;
  items: UiOrderItem[];
}