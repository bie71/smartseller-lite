<template>
  <div class="min-h-screen bg-slate-900/5 flex flex-col">
    <header class="sticky top-0 z-30 shadow-lg background-color-secondary">
      <div class="relative overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-r from-primary-dark via-primary to-primary-light opacity-95"></div>
        <div class="relative mx-auto flex max-w-screen-2xl flex-col gap-6 px-8 py-6 text-white lg:flex-row lg:items-center lg:justify-between">
          <div class="flex items-center gap-4">
            <div class="h-16 w-16 rounded-2xl bg-white/20 backdrop-blur flex items-center justify-center">
              <img :src="logoPreview" alt="Brand logo" class="h-12 w-12 object-contain" />
          </div>
          <div>
            <h1 class="text-3xl font-semibold tracking-tight">
              {{ brandName }}
            </h1>
            <p class="text-sm text-white/80 flex items-center gap-1">
              <SparklesIcon class="h-4 w-4" />
              Dashboard fulfilment serba cepat untuk agen pengiriman.
            </p>
          </div>
        </div>
        <nav class="flex flex-wrap gap-2">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            :class="['nav-pill', activeTab === tab.id ? 'nav-pill-active' : 'nav-pill-idle']"
            @click="activeTab = tab.id"
          >
            <component :is="tab.icon" class="h-5 w-5" />
            <span>{{ tab.label }}</span>
          </button>
        </nav>
        </div>
      </div>
    </header>

    <main class="mx-auto w-full max-w-screen-2xl flex-1 px-8 pb-16 pt-8 space-y-6">
      <AnalyticsPage v-if="activeTab === 'analytics'" />
      <OrderPage v-else-if="activeTab === 'orders'" />
      <ProductCatalogPage
        v-else-if="activeTab === 'products'"
        key="products"
        :refresh-token="productRefreshToken"
        @request-stock-opname="openStockOpname"
      />
      <StockOpnamePage
        v-else-if="activeTab === 'stock'"
        key="stock"
        :incoming-opname-product-id="pendingOpnameProductId"
        @opname-product-consumed="handleOpnameProductConsumed"
        @stock-adjusted="handleStockAdjusted"
      />
      <CustomerPage v-else-if="activeTab === 'customers'" />
      <ExpeditionPage v-else-if="activeTab === 'expeditions'" />
      <SettingsPage v-else-if="activeTab === 'settings'" :initial-settings="settings" @updated="handleSettingsUpdated" />
    </main>
    <ToastStack />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import OrderPage from './pages/OrderPage.vue';
import ProductCatalogPage from './pages/ProductCatalogPage.vue';
import StockOpnamePage from './pages/StockOpnamePage.vue';
import CustomerPage from './pages/CustomerPage.vue';
import ExpeditionPage from './pages/ExpeditionPage.vue';
import SettingsPage from './pages/SettingsPage.vue';
import AnalyticsPage from './pages/AnalyticsPage.vue';
import { getSettings, type AppSettings } from '../modules/settings';
import defaultLogo from '../assets/logo-default.svg';
import ToastStack from './components/ToastStack.vue';
import {
  Cog6ToothIcon,
  SparklesIcon,
  ClipboardDocumentListIcon,
  ClipboardDocumentCheckIcon,
  Squares2X2Icon,
  UsersIcon,
  TruckIcon,
  PresentationChartLineIcon
} from '@heroicons/vue/24/outline';

type TabId = 'analytics' | 'orders' | 'products' | 'stock' | 'customers' | 'expeditions' | 'settings';

const tabs: Array<{ id: TabId; label: string; icon: any }> = [
  { id: 'analytics', label: 'Analitik', icon: PresentationChartLineIcon },
  { id: 'orders', label: 'Orders', icon: ClipboardDocumentListIcon },
  { id: 'products', label: 'Produk', icon: Squares2X2Icon },
  { id: 'stock', label: 'Stock Opname', icon: ClipboardDocumentCheckIcon },
  { id: 'customers', label: 'Customer', icon: UsersIcon },
  { id: 'expeditions', label: 'Ekspedisi', icon: TruckIcon },
  { id: 'settings', label: 'Pengaturan', icon: Cog6ToothIcon }
];

const activeTab = ref<TabId>('orders');
const settings = ref<AppSettings | null>(null);
const productRefreshToken = ref(0);
const pendingOpnameProductId = ref<string | null>(null);

function resolveMediaPath(path?: string | null) {
  if (!path) return '';
  const normalised = path
    .replace(/\\/g, '/')
    .replace(/^\.\//, '')
    .replace(/^media\//, '')
    .replace(/^\//, '');
  return normalised ? `/media/${normalised}` : '';
}

const logoPreview = computed(() => {
  if (settings.value?.logoData) {
    const mime = settings.value.logoMime || 'image/png';
    return `data:${mime};base64,${settings.value.logoData}`;
  }
  if (settings.value?.logoUrl) {
    return settings.value.logoUrl;
  }
  if (settings.value?.logoPath) {
    const path = resolveMediaPath(settings.value.logoPath);
    if (path) {
      return path;
    }
  }
  return defaultLogo;
});

const brandName = computed(() => settings.value?.brandName || 'SmartSeller Lite');

async function loadSettings() {
  try {
    settings.value = await getSettings();
  } catch (error) {
    console.error(error);
  }
}

function handleSettingsUpdated(updated: AppSettings) {
  settings.value = { ...updated };
}

function openStockOpname(productId?: string) {
  if (productId) {
    pendingOpnameProductId.value = productId;
  }
  activeTab.value = 'stock';
}

function handleOpnameProductConsumed(productId: string) {
  if (pendingOpnameProductId.value === productId) {
    pendingOpnameProductId.value = null;
  }
}

function handleStockAdjusted() {
  productRefreshToken.value += 1;
}

watch(activeTab, (tab) => {
  if (tab !== 'stock') {
    pendingOpnameProductId.value = null;
  }
});

onMounted(async () => {
  await loadSettings();
});
</script>
