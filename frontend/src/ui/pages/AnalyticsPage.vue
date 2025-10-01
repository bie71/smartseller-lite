<template>
  <section class="space-y-6">
    <div class="card space-y-6">
      <header class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div class="flex items-center gap-3">
          <div class="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
            <PresentationChartLineIcon class="h-6 w-6" />
          </div>
          <div>
            <h2 class="text-xl font-semibold">Analitik Bisnis</h2>
            <p class="text-sm text-slate-500">Pantau performa order, pendapatan, dan stok secara ringkas.</p>
          </div>
        </div>
        <div class="flex flex-wrap gap-2 md:justify-end">
          <button type="button" class="btn-secondary text-sm" :disabled="loading" @click="refreshAnalytics">
            <ArrowPathIcon class="h-5 w-5" />
            {{ loading ? 'Memuat...' : 'Refresh Data' }}
          </button>
        </div>
      </header>

      <p v-if="lastUpdated" class="text-xs text-slate-500">
        Pembaruan terakhir {{ formatUpdatedAt(lastUpdated) }}
      </p>
      <div v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
        {{ error }}
      </div>
      <div v-if="loading" class="rounded-lg border border-dashed border-slate-200 px-4 py-6 text-sm text-slate-500">
        Menyiapkan analitik terbaru...
      </div>
      <div v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
          <p class="text-xs uppercase text-slate-500">Total Order</p>
          <p class="mt-2 text-2xl font-semibold text-slate-800">{{ totalOrders }}</p>
          <p class="text-xs text-slate-500">Bulan ini {{ monthSummary.orders }} order ({{ formatGrowth(monthSummary.orders, lastMonthSummary.orders) }})</p>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
          <p class="text-xs uppercase text-slate-500">Pendapatan</p>
          <p class="mt-2 text-2xl font-semibold text-emerald-600">Rp {{ formatCurrency(totalRevenue) }}</p>
          <p class="text-xs text-slate-500">Rata-rata Rp {{ formatCurrency(averageOrderValue) }} per order</p>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
          <p class="text-xs uppercase text-slate-500">Profit</p>
          <p class="mt-2 text-2xl font-semibold text-indigo-600">Rp {{ formatCurrency(totalProfit) }}</p>
          <p class="text-xs text-slate-500">Margin {{ formatPercent(profitMargin) }}</p>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
          <p class="text-xs uppercase text-slate-500">Customer Aktif</p>
          <p class="mt-2 text-2xl font-semibold text-slate-800">{{ uniqueCustomers }}</p>
          <p class="text-xs text-slate-500">Database {{ totalCustomers }} customer · {{ activeProducts }} produk aktif</p>
        </div>
      </div>
    </div>

    <div class="grid gap-6 xl:grid-cols-2">
      <div class="card space-y-4">
        <header class="flex items-center justify-between">
          <h3 class="text-lg font-semibold flex items-center gap-2">
            <ChartBarIcon class="h-5 w-5 text-primary" /> Tren 7 Hari Terakhir
          </h3>
          <span class="text-xs text-slate-500">Ringkasan order harian</span>
        </header>
        <div v-if="!dailyPerformance.length" class="text-sm text-slate-500">Belum ada order untuk dihitung.</div>
        <ul v-else class="space-y-3">
          <li
            v-for="day in dailyPerformance"
            :key="day.date"
            class="flex flex-col gap-2 rounded-lg border border-slate-200 bg-slate-50/60 p-3 sm:flex-row sm:items-center sm:justify-between"
          >
            <div>
              <p class="font-semibold text-slate-700">{{ formatDate(day.date) }}</p>
              <p class="text-xs text-slate-500">{{ day.orders }} order · {{ day.items }} item</p>
            </div>
            <div class="text-sm text-slate-600 sm:text-right">
              <p>Pendapatan Rp {{ formatCurrency(day.revenue) }}</p>
              <p class="text-xs text-emerald-600">Profit Rp {{ formatCurrency(day.profit) }}</p>
            </div>
          </li>
        </ul>
      </div>

      <div class="card space-y-4">
        <header class="flex items-center justify-between">
          <h3 class="text-lg font-semibold flex items-center gap-2">
            <ArrowTrendingUpIcon class="h-5 w-5 text-primary" /> Produk Terlaris
          </h3>
          <span class="text-xs text-slate-500">Top 5 berdasarkan omzet</span>
        </header>
        <div v-if="!topProducts.length" class="text-sm text-slate-500">Belum ada data order untuk menghitung peringkat produk.</div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead class="text-left text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th class="py-2">Produk</th>
                <th class="py-2 text-right">Qty</th>
                <th class="py-2 text-right">Omzet</th>
                <th class="py-2 text-right">Profit</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="product in topProducts" :key="product.productId" class="border-t border-slate-100">
                <td class="py-3">
                  <p class="font-semibold">{{ product.name }}</p>
                  <p class="text-xs text-slate-500">SKU {{ product.sku || '—' }}</p>
                </td>
                <td class="py-3 text-right">{{ product.quantity }}</td>
                <td class="py-3 text-right">Rp {{ formatCurrency(product.revenue) }}</td>
                <td class="py-3 text-right">Rp {{ formatCurrency(product.profit) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="card space-y-4">
      <header class="flex items-center justify-between">
        <h3 class="text-lg font-semibold flex items-center gap-2">
          <TruckIcon class="h-5 w-5 text-primary" /> Kinerja Ekspedisi
        </h3>
        <span class="text-xs text-slate-500">Dari 30 order terbaru</span>
      </header>
      <div v-if="!courierPerformance.length" class="text-sm text-slate-500">Belum ada data ekspedisi yang tersedia.</div>
      <ul v-else class="space-y-3">
        <li
          v-for="courier in courierPerformance"
          :key="courier.courier"
          class="flex flex-col gap-2 rounded-lg border border-slate-200 bg-white p-3 shadow-sm sm:flex-row sm:items-center sm:justify-between"
        >
          <div>
            <p class="font-semibold text-slate-700">{{ courier.courier }}</p>
            <p class="text-xs text-slate-500">{{ courier.orders }} order · {{ courier.services }} layanan</p>
          </div>
          <div class="text-sm text-slate-600 sm:text-right">
            <p>Omzet Rp {{ formatCurrency(courier.revenue) }}</p>
            <p class="text-xs text-emerald-600">Profit Rp {{ formatCurrency(courier.profit) }}</p>
          </div>
        </li>
      </ul>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import {
  ArrowPathIcon,
  ArrowTrendingUpIcon,
  ChartBarIcon,
  PresentationChartLineIcon,
  TruckIcon
} from '@heroicons/vue/24/outline';
import { listOrders, type Order } from '../../modules/order';
import { listProducts, type Product } from '../../modules/product';
import { listCustomers, type Customer } from '../../modules/customer';

type DailyPerformanceRow = {
  date: string;
  orders: number;
  items: number;
  revenue: number;
  profit: number;
};

type ProductPerformance = {
  productId: string;
  name: string;
  sku: string;
  quantity: number;
  revenue: number;
  profit: number;
};

type CourierPerformance = {
  courier: string;
  orders: number;
  services: number;
  revenue: number;
  profit: number;
};

const loading = ref(true);
const error = ref('');
const orders = ref<Order[]>([]);
const products = ref<Product[]>([]);
const customers = ref<Customer[]>([]);
const ordersTotal = ref(0);
const productsTotal = ref(0);
const customersTotal = ref(0);
const lastUpdated = ref<Date | null>(null);

const productMap = computed(() => {
  const map = new Map<string, Product>();
  products.value.forEach((product) => {
    if (product.id) {
      map.set(product.id, product);
    }
  });
  return map;
});

const totalOrders = computed(() => (ordersTotal.value ? ordersTotal.value : orders.value.length));
const totalRevenue = computed(() => orders.value.reduce((sum, order) => sum + order.total, 0));
const totalProfit = computed(() => orders.value.reduce((sum, order) => sum + order.profit, 0));
const averageOrderValue = computed(() => (totalOrders.value ? totalRevenue.value / totalOrders.value : 0));
const profitMargin = computed(() => (totalRevenue.value ? totalProfit.value / totalRevenue.value : 0));
const totalCustomers = computed(() => (customersTotal.value ? customersTotal.value : customers.value.length));
const uniqueCustomers = computed(() => {
  const ids = new Set<string>();
  orders.value.forEach((order) => {
    if (order.buyerId) ids.add(order.buyerId);
    if (order.recipientId) ids.add(order.recipientId);
  });
  return ids.size || totalCustomers.value;
});
const activeProducts = computed(() => (productsTotal.value ? productsTotal.value : products.value.length));

function monthKey(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
}

const monthBuckets = computed(() => {
  const buckets = new Map<string, { orders: number; revenue: number; profit: number }>();
  orders.value.forEach((order) => {
    const parsed = new Date(order.createdAt);
    if (Number.isNaN(parsed.getTime())) {
      return;
    }
    const key = monthKey(parsed);
    if (!buckets.has(key)) {
      buckets.set(key, { orders: 0, revenue: 0, profit: 0 });
    }
    const current = buckets.get(key)!;
    current.orders += 1;
    current.revenue += order.total;
    current.profit += order.profit;
  });
  return buckets;
});

const now = new Date();
const currentMonthKey = monthKey(now);
const previousMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
const previousMonthKey = monthKey(previousMonth);

const monthSummary = computed(() => monthBuckets.value.get(currentMonthKey) || { orders: 0, revenue: 0, profit: 0 });
const lastMonthSummary = computed(
  () => monthBuckets.value.get(previousMonthKey) || { orders: 0, revenue: 0, profit: 0 }
);

const dailyPerformance = computed<DailyPerformanceRow[]>(() => {
  const bucket = new Map<string, DailyPerformanceRow>();
  orders.value.forEach((order) => {
    const parsed = new Date(order.createdAt);
    if (Number.isNaN(parsed.getTime())) {
      return;
    }
    const key = parsed.toISOString().slice(0, 10);
    if (!bucket.has(key)) {
      bucket.set(key, { date: key, orders: 0, items: 0, revenue: 0, profit: 0 });
    }
    const row = bucket.get(key)!;
    row.orders += 1;
    row.revenue += order.total;
    row.profit += order.profit;
    row.items += order.items.reduce((sum, item) => sum + item.quantity, 0);
  });
  return Array.from(bucket.values())
    .sort((a, b) => (a.date < b.date ? 1 : -1))
    .slice(0, 7)
    .sort((a, b) => (a.date > b.date ? 1 : -1));
});

const topProducts = computed<ProductPerformance[]>(() => {
  const bucket = new Map<string, ProductPerformance>();
  orders.value.forEach((order) => {
    order.items.forEach((item) => {
      if (!bucket.has(item.productId)) {
        const product = productMap.value.get(item.productId);
        bucket.set(item.productId, {
          productId: item.productId,
          name: product?.name || 'Produk',
          sku: product?.sku || '',
          quantity: 0,
          revenue: 0,
          profit: 0
        });
      }
      const entry = bucket.get(item.productId)!;
      entry.quantity += item.quantity;
      entry.revenue += item.quantity * item.unitPrice - item.discount;
      entry.profit += item.profit;
    });
  });
  return Array.from(bucket.values())
    .sort((a, b) => (b.revenue === a.revenue ? b.quantity - a.quantity : b.revenue - a.revenue))
    .slice(0, 5);
});

const courierPerformance = computed<CourierPerformance[]>(() => {
  type CourierBucket = {
    courier: string;
    orders: number;
    services: Set<string>;
    revenue: number;
    profit: number;
  };
  const bucket = new Map<string, CourierBucket>();
  const recentOrders = orders.value.slice(0, 30);
  recentOrders.forEach((order) => {
    const courierName = order.shipment?.courier || 'Tanpa ekspedisi';
    if (!bucket.has(courierName)) {
      bucket.set(courierName, {
        courier: courierName,
        orders: 0,
        services: new Set<string>(),
        revenue: 0,
        profit: 0
      });
    }
    const entry = bucket.get(courierName)!;
    entry.orders += 1;
    entry.revenue += order.total;
    entry.profit += order.profit;
    const serviceLevel = order.shipment?.serviceLevel?.trim();
    if (serviceLevel) {
      entry.services.add(serviceLevel);
    }
  });
  return Array.from(bucket.values())
    .map((entry) => ({
      courier: entry.courier,
      orders: entry.orders,
      services: entry.services.size,
      revenue: entry.revenue,
      profit: entry.profit
    }))
    .sort((a, b) => (b.orders === a.orders ? b.revenue - a.revenue : b.orders - a.orders));
});

function formatCurrency(value: number) {
  return Math.round(value).toLocaleString('id-ID');
}

function formatPercent(value: number) {
  if (!Number.isFinite(value)) {
    return '0%';
  }
  return `${(value * 100).toFixed(1)}%`;
}

function formatGrowth(current: number, previous: number) {
  if (!previous) {
    return current > 0 ? '+100%' : '0%';
  }
  const delta = ((current - previous) / previous) * 100;
  const prefix = delta >= 0 ? '+' : '';
  return `${prefix}${delta.toFixed(1)}%`; // limited to one decimal
}

function formatDate(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat('id-ID', { weekday: 'short', day: '2-digit', month: 'short' }).format(parsed);
}

function formatUpdatedAt(date: Date) {
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(date);
}

async function loadAnalytics() {
  loading.value = true;
  error.value = '';
  try {
    const [orderResults, productResults, customerResults] = await Promise.all([
      listOrders({ page: 1, pageSize: 200 }),
      listProducts({ page: 1, pageSize: 500 }),
      listCustomers({ page: 1, pageSize: 500 })
    ]);
    orders.value = orderResults.items;
    ordersTotal.value = orderResults.total;
    products.value = productResults.items;
    productsTotal.value = productResults.total;
    customers.value = customerResults.items;
    customersTotal.value = customerResults.total;
    lastUpdated.value = new Date();
  } catch (err) {
    console.error(err);
    error.value = 'Gagal memuat data analitik.';
  } finally {
    loading.value = false;
  }
}

function refreshAnalytics() {
  if (!loading.value) {
    loadAnalytics();
  }
}

onMounted(() => {
  loadAnalytics();
});
</script>
