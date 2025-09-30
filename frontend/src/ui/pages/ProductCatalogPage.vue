<template>
  <section class="space-y-6">
    <div class="card space-y-6">
      <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div class="flex items-center gap-3">
          <div class="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
            <Squares2X2Icon class="h-6 w-6" />
          </div>
          <div>
            <h2 class="text-xl font-semibold">Produk</h2>
            <p class="text-sm text-slate-500">Kelola katalog dan pantau stok setiap SKU.</p>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2 md:justify-end">
          <span v-if="products.length" class="text-sm text-slate-500">Total {{ products.length }} produk</span>
          <span
            v-if="outOfStockCount"
            class="inline-flex items-center gap-1 rounded-full bg-red-50 px-3 py-1 text-xs font-semibold text-red-600"
          >
            <ExclamationTriangleIcon class="h-4 w-4" />
            {{ outOfStockCount }} Out of Stock
          </span>
          <span
            v-if="warningStockCount"
            class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-3 py-1 text-xs font-semibold text-amber-600"
          >
            <ExclamationTriangleIcon class="h-4 w-4" />
            {{ warningStockCount }} stok menipis
          </span>
          <button v-if="products.length" type="button" class="btn-ghost" @click="productModalOpen = true">
            Lihat Semua
          </button>
        </div>
      </div>

      <div
        v-if="lowStockProducts.length"
        class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-amber-100 bg-amber-50/70 px-4 py-3 text-sm text-amber-700"
      >
        <div class="flex items-center gap-2">
          <ExclamationTriangleIcon class="h-5 w-5" />
          <span>
            {{ outOfStockCount }} produk habis stok · {{ warningStockCount }} stok menipis. Segera lakukan restock atau stock opname.
          </span>
        </div>
        <button type="button" class="text-xs font-semibold underline" @click="productModalOpen = true">
          Lihat detail
        </button>
      </div>

      <form class="grid grid-cols-1 gap-4 md:grid-cols-2" @submit.prevent="handleSubmit">
        <div>
          <label class="text-sm font-medium text-slate-600">Nama Produk</label>
          <input v-model="form.name" type="text" class="input mt-1" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">SKU</label>
          <input v-model="form.sku" type="text" class="input mt-1" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Harga Modal</label>
          <input
            v-model="costPriceDisplay"
            type="text"
            inputmode="numeric"
            class="input mt-1 font-mono text-left"
            required
            @input="onCostPriceInput"
            @blur="syncCostPriceDisplay"
          />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Harga Jual</label>
          <input
            v-model="salePriceDisplay"
            type="text"
            inputmode="numeric"
            class="input mt-1 font-mono text-left"
            required
            @input="onSalePriceInput"
            @blur="syncSalePriceDisplay"
          />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Stok Awal</label>
          <input v-model.number="form.stock" type="number" min="0" class="input mt-1" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Kategori</label>
          <input
            v-model="form.category"
            type="text"
            class="input mt-1"
            placeholder="Contoh: Fashion, Elektronik"
          />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Ambang Stok Minim</label>
          <input v-model.number="form.lowStockThreshold" type="number" min="1" class="input mt-1" />
          <p class="mt-1 text-xs text-slate-400">Notifikasi stok muncul jika jumlah ≤ nilai ini.</p>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Gambar Produk</label>
          <div class="mt-1 flex flex-col gap-3 md:flex-row md:items-center">
            <div class="flex h-20 w-20 items-center justify-center overflow-hidden rounded-lg bg-slate-100">
              <img v-if="productImagePreview" :src="productImagePreview" alt="Preview produk" class="h-full w-full object-cover" />
              <PhotoIcon v-else class="h-8 w-8 text-slate-400" />
            </div>
            <div class="flex flex-col gap-2">
              <input type="file" accept="image/*" class="input" @change="handleProductImageChange" />
              <button v-if="form.imageData || form.imageUrl" type="button" class="text-xs font-medium text-red-500" @click="clearProductImage">
                Hapus gambar
              </button>
            </div>
          </div>
        </div>
        <div class="md:col-span-2">
          <label class="text-sm font-medium text-slate-600">Deskripsi</label>
          <textarea v-model="form.description" rows="3" class="input mt-1"></textarea>
        </div>
        <div class="md:col-span-2 flex items-center gap-2">
          <button type="submit" class="btn-primary">
            <CheckBadgeIcon class="h-5 w-5" />
            {{ editing ? 'Update' : 'Tambah' }} Produk
          </button>
          <button v-if="editing" type="button" class="btn-secondary" @click="resetForm">
            <XMarkIcon class="h-5 w-5" />
            Batalkan
          </button>
        </div>
      </form>
    </div>

    <div class="card overflow-hidden">
      <div v-if="productsLoading" class="space-y-4 p-6">
        <div class="space-y-3">
          <div
            v-for="skeleton in 3"
            :key="skeleton"
            class="animate-pulse space-y-3 rounded-lg border border-slate-100 bg-slate-50/80 p-4"
          >
            <div class="flex items-center gap-3">
              <div class="h-12 w-12 rounded-lg bg-slate-200/70"></div>
              <div class="flex-1 space-y-2">
                <div class="h-3 w-1/3 rounded bg-slate-200/80"></div>
                <div class="h-3 w-1/2 rounded bg-slate-200/60"></div>
              </div>
            </div>
            <div class="grid grid-cols-3 gap-4 text-xs">
              <div class="h-3 rounded bg-slate-200/70"></div>
              <div class="h-3 rounded bg-slate-200/60"></div>
              <div class="h-3 rounded bg-slate-200/50"></div>
            </div>
          </div>
        </div>
      </div>
      <div v-else-if="loadError" class="space-y-4 p-6 text-sm">
        <p class="font-semibold text-red-500">{{ loadError }}</p>
        <p class="text-slate-500">Periksa koneksi lalu coba muat ulang.</p>
        <button type="button" class="btn-secondary w-fit" @click="retryLoadProducts">
          Muat ulang
        </button>
      </div>
      <template v-else>
        <div class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead class="text-left text-slate-500 uppercase tracking-wider">
              <tr>
                <th class="py-2">Produk</th>
                <th class="py-2">Harga</th>
                <th class="py-2">Stok</th>
                <th class="py-2">Status</th>
                <th class="py-2">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!products.length">
                <td colspan="5" class="py-6 text-center text-slate-500">Belum ada produk.</td>
              </tr>
              <tr v-for="product in paginatedProducts" :key="product.id" class="border-t border-slate-100 hover:bg-slate-50/80">
                <td class="py-3">
                  <div class="flex items-center gap-3">
                    <div class="h-12 w-12 overflow-hidden rounded-lg bg-slate-100">
                      <img
                        v-if="productThumbSource(product)"
                        :src="productThumbSource(product)"
                        :alt="product.name"
                        class="h-full w-full object-cover"
                      />
                      <PhotoIcon v-else class="mx-auto my-3 h-6 w-6 text-slate-300" />
                    </div>
                    <div>
                      <p class="font-semibold">{{ product.name }}</p>
                      <p class="text-xs text-slate-500">SKU {{ product.sku }}</p>
                      <p v-if="product.category" class="text-xs text-primary">Kategori: {{ product.category }}</p>
                    </div>
                  </div>
                </td>
                <td class="py-3">
                  <div class="space-y-1">
                    <p>Modal: Rp {{ formatCurrency(product.costPrice) }}</p>
                    <p class="text-slate-500">Jual: Rp {{ formatCurrency(product.salePrice) }}</p>
                  </div>
                </td>
                <td class="py-3">
                  <span class="block font-semibold">{{ product.stock }}</span>
                  <span class="text-xs text-slate-500">{{ thresholdLabel(product) }}</span>
                </td>
                <td class="py-3">
                  <span
                    :class="[
                      'inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs font-semibold',
                      productStatus(product).classes
                    ]"
                  >
                    <ExclamationTriangleIcon
                      v-if="productStatus(product).icon !== 'ready'"
                      class="h-4 w-4"
                    />
                    <CheckBadgeIcon v-else class="h-4 w-4" />
                    {{ productStatus(product).label }}
                  </span>
                </td>
                <td class="py-3 flex flex-wrap gap-2">
                  <button class="btn-secondary text-xs" @click="editProduct(product)">
                    <PencilSquareIcon class="h-4 w-4" />
                    Edit
                  </button>
                  <button class="btn-secondary text-xs" @click="openProductDetail(product)">
                    <InformationCircleIcon class="h-4 w-4" />
                    Detail
                  </button>
                  <button class="btn-secondary text-xs" @click="startOpnameWithProduct(product)">
                    <ArrowsUpDownIcon class="h-4 w-4" />
                    Stok
                  </button>
                  <button
                    class="inline-flex items-center gap-1 rounded-lg border border-red-200 px-3 py-1 text-xs font-semibold text-red-600 transition hover:bg-red-50"
                    @click="archiveProductAction(product)"
                  >
                    <ArchiveBoxXMarkIcon class="h-4 w-4" />
                    Arsipkan
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <footer
          v-if="products.length"
          class="flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between"
        >
          <span>{{ productRangeLabel }}</span>
          <div class="flex items-center gap-3">
            <span>Halaman {{ page }} / {{ totalPages }}</span>
            <div class="flex items-center gap-2">
              <button type="button" :class="paginationButtonClasses" :disabled="page === 1" @click="previousPage">
                <ChevronLeftIcon class="h-4 w-4" />
              </button>
              <button type="button" :class="paginationButtonClasses" :disabled="page === totalPages" @click="nextPage">
                <ChevronRightIcon class="h-4 w-4" />
              </button>
            </div>
          </div>
        </footer>
      </template>
    </div>

    <BaseModal v-model="productModalOpen" title="Daftar Produk">
      <div class="space-y-4">
        <div class="relative">
          <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
          <input
            v-model="productModalSearch"
            type="search"
            class="input pl-10"
            placeholder="Cari nama atau SKU produk"
          />
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead class="text-left text-slate-500 uppercase tracking-wider">
              <tr>
                <th class="py-2">Produk</th>
                <th class="py-2">Harga</th>
                <th class="py-2">Stok</th>
                <th class="py-2">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!productModalFiltered.length">
                <td colspan="4" class="py-6 text-center text-slate-500">Tidak ada produk yang cocok dengan pencarian.</td>
              </tr>
              <tr v-for="product in productModalItems" :key="product.id" class="border-t border-slate-100">
                <td class="py-3">
                  <div class="flex items-center gap-3">
                    <div class="h-10 w-10 overflow-hidden rounded-md bg-slate-100">
                      <img
                        v-if="productThumbSource(product)"
                        :src="productThumbSource(product)"
                        :alt="product.name"
                        class="h-full w-full object-cover"
                      />
                      <PhotoIcon v-else class="mx-auto my-2 h-5 w-5 text-slate-300" />
                    </div>
                    <div>
                      <p class="font-semibold">{{ product.name }}</p>
                      <p class="text-xs text-slate-500">SKU {{ product.sku }}</p>
                      <p v-if="product.category" class="text-xs text-primary">Kategori: {{ product.category }}</p>
                    </div>
                  </div>
                </td>
                <td class="py-3">
                  <div class="space-y-1">
                    <p>Modal: Rp {{ formatCurrency(product.costPrice) }}</p>
                    <p class="text-slate-500">Jual: Rp {{ formatCurrency(product.salePrice) }}</p>
                  </div>
                </td>
                <td class="py-3">
                  <span class="block font-semibold">{{ product.stock }}</span>
                  <span class="text-xs text-slate-500">{{ thresholdLabel(product) }}</span>
                </td>
                <td class="py-3">
                  <span
                    :class="[
                      'inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs font-semibold',
                      productStatus(product).classes
                    ]"
                  >
                    <ExclamationTriangleIcon
                      v-if="productStatus(product).icon !== 'ready'"
                      class="h-4 w-4"
                    />
                    <CheckBadgeIcon v-else class="h-4 w-4" />
                    {{ productStatus(product).label }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div
          v-if="productModalFiltered.length"
          class="flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between"
        >
          <span>{{ productModalRangeLabel }}</span>
          <div class="flex items-center gap-3">
            <span>Halaman {{ productModalPage }} / {{ productModalTotalPages }}</span>
            <div class="flex items-center gap-2">
              <button type="button" :class="paginationButtonClasses" :disabled="productModalPage === 1" @click="previousModalPage">
                <ChevronLeftIcon class="h-4 w-4" />
              </button>
              <button
                type="button"
                :class="paginationButtonClasses"
                :disabled="productModalPage === productModalTotalPages"
                @click="nextModalPage"
              >
                <ChevronRightIcon class="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="productModalOpen = false">Tutup</button>
      </template>
    </BaseModal>

    <BaseModal v-model="productDetailOpen" title="Detail Produk" :subtitle="productDetailSubtitle">
      <div v-if="productDetail" class="space-y-4 text-sm text-slate-600">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
          <div class="h-28 w-28 overflow-hidden rounded-xl bg-slate-100">
            <img
              v-if="productDetailImage"
              :src="productDetailImage"
              :alt="productDetail.name"
              class="h-full w-full object-cover"
            />
            <PhotoIcon v-else class="mx-auto my-6 h-8 w-8 text-slate-300" />
          </div>
          <div class="space-y-2">
            <h4 class="text-lg font-semibold text-slate-800">{{ productDetail.name }}</h4>
            <p class="text-xs uppercase tracking-wide text-slate-400">SKU {{ productDetail.sku }}</p>
            <p v-if="productDetail.category" class="text-xs text-primary">Kategori: {{ productDetail.category }}</p>
            <div class="grid gap-2 sm:grid-cols-2">
              <div>
                <p class="text-xs font-semibold uppercase text-slate-400">Harga Modal</p>
                <p>Rp {{ formatCurrency(productDetail.costPrice) }}</p>
              </div>
              <div>
                <p class="text-xs font-semibold uppercase text-slate-400">Harga Jual</p>
                <p>Rp {{ formatCurrency(productDetail.salePrice) }}</p>
              </div>
              <div>
                <p class="text-xs font-semibold uppercase text-slate-400">Stok</p>
                <p>{{ productDetail.stock }} unit</p>
              </div>
              <div>
                <p class="text-xs font-semibold uppercase text-slate-400">Ambang Minim</p>
                <p>{{ productDetail.lowStockThreshold || '-' }}</p>
              </div>
            </div>
          </div>
        </div>
        <div v-if="productDetail.description" class="rounded-lg bg-slate-50 p-3 text-xs text-slate-500">
          {{ productDetail.description }}
        </div>
        <div class="rounded-lg border border-slate-200 bg-white p-3 text-xs text-slate-500">
          Dibuat: {{ formatDate(productDetail.createdAt) || '-' }} · Diperbarui: {{ formatDate(productDetail.updatedAt) || '-' }}
        </div>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="productDetailOpen = false">Tutup</button>
        <button v-if="productDetail" type="button" class="btn-primary" @click="editProduct(productDetail)">
          <PencilSquareIcon class="h-4 w-4" />
          Edit Produk
        </button>
      </template>
    </BaseModal>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import BaseModal from '../components/BaseModal.vue';
import type { Product } from '../../modules/product';
import { archiveProduct as archiveProductApi, listProducts, saveProduct } from '../../modules/product';
import { useToastStore } from '../stores/toast';
import {
  ArrowsUpDownIcon,
  ArchiveBoxXMarkIcon,
  CheckBadgeIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  PhotoIcon,
  Squares2X2Icon,
  XMarkIcon
} from '@heroicons/vue/24/outline';

type ProductForm = {
  id?: string;
  name: string;
  sku: string;
  costPrice: number;
  salePrice: number;
  stock: number;
  category: string;
  lowStockThreshold: number;
  description: string;
  imageData?: string;
  imageMime?: string;
  imageUrl?: string;
  thumbUrl?: string;
  imagePath?: string;
  thumbPath?: string;
  imageHash?: string;
  imageWidth?: number;
  imageHeight?: number;
  imageSizeBytes?: number;
  thumbWidth?: number;
  thumbHeight?: number;
  thumbSizeBytes?: number;
};

const emit = defineEmits<{ (e: 'request-stock-opname', productId: string): void }>();
const props = defineProps<{ refreshToken: number }>();

const products = ref<Product[]>([]);
const productsLoading = ref(true);
const loadError = ref('');
const editing = ref(false);
const page = ref(1);
const pageSize = 6;
const paginationButtonClasses =
  'inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-600 transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-40';

const productModalOpen = ref(false);
const productModalSearch = ref('');
const productModalPage = ref(1);
const productModalPageSize = 12;

const productDetailOpen = ref(false);
const productDetail = ref<Product | null>(null);

const toast = useToastStore();

let loadRequestId = 0;

const form = reactive<ProductForm>({
  id: undefined,
  name: '',
  sku: '',
  costPrice: 0,
  salePrice: 0,
  stock: 0,
  category: '',
  lowStockThreshold: 5,
  description: '',
  imageData: '',
  imageMime: '',
  imageUrl: '',
  thumbUrl: '',
  imagePath: '',
  thumbPath: '',
  imageHash: '',
  imageWidth: 0,
  imageHeight: 0,
  imageSizeBytes: 0,
  thumbWidth: 0,
  thumbHeight: 0,
  thumbSizeBytes: 0
});

const costPriceDisplay = ref('');
const salePriceDisplay = ref('');

function makeImageSrc(data?: string, mime?: string): string {
  if (!data) return '';
  const type = (mime && mime.length ? mime : 'image/png').trim() || 'image/png';
  return `data:${type};base64,${data}`;
}

function resolveMediaPath(path?: string | null): string {
  if (!path) {
    return '';
  }
  const normalised = path
    .replace(/\\/g, '/')
    .replace(/^\.\//, '')
    .replace(/^media\//, '')
    .replace(/^\//, '');
  return normalised ? `/media/${normalised}` : '';
}

const productImagePreview = computed(() => {
  const dataUri = makeImageSrc(form.imageData, form.imageMime);
  if (dataUri) {
    return dataUri;
  }
  if (form.thumbUrl) {
    return form.thumbUrl;
  }
  if (form.imageUrl) {
    return form.imageUrl;
  }
  const fromThumbPath = resolveMediaPath(form.thumbPath);
  if (fromThumbPath) {
    return fromThumbPath;
  }
  const fromImagePath = resolveMediaPath(form.imagePath);
  if (fromImagePath) {
    return fromImagePath;
  }
  return '';
});

const productDetailImage = computed(() => {
  const detail = productDetail.value;
  if (!detail) {
    return '';
  }
  if (detail.thumbUrl) {
    return detail.thumbUrl;
  }
  if (detail.imageUrl) {
    return detail.imageUrl;
  }
  const fromThumbPath = resolveMediaPath(detail.thumbPath);
  if (fromThumbPath) {
    return fromThumbPath;
  }
  const fromImagePath = resolveMediaPath(detail.imagePath);
  if (fromImagePath) {
    return fromImagePath;
  }
  return '';
});

const productDetailSubtitle = computed(() => (productDetail.value ? `SKU ${productDetail.value.sku}` : ''));
watch(
  () => form.costPrice,
  (value) => {
    const formatted = formatPriceInput(value);
    if (costPriceDisplay.value !== formatted) {
      costPriceDisplay.value = formatted;
    }
  },
  { immediate: true }
);

watch(
  () => form.salePrice,
  (value) => {
    const formatted = formatPriceInput(value);
    if (salePriceDisplay.value !== formatted) {
      salePriceDisplay.value = formatted;
    }
  },
  { immediate: true }
);

const defaultThreshold = 5;

function resolveThreshold(product: Product): number {
  const candidate = Number(product.lowStockThreshold ?? defaultThreshold);
  if (!Number.isFinite(candidate) || candidate <= 0) {
    return defaultThreshold;
  }
  return Math.floor(candidate);
}

function stockSeverity(product: Product): number {
  if (product.stock <= 0) {
    return 2;
  }
  if (product.stock <= resolveThreshold(product)) {
    return 1;
  }
  return 0;
}

const lowStockProducts = computed(() => products.value.filter((product) => stockSeverity(product) >= 1));
const outOfStockProducts = computed(() => products.value.filter((product) => product.stock <= 0));
const warningStockProducts = computed(() => lowStockProducts.value.filter((product) => product.stock > 0));
const outOfStockCount = computed(() => outOfStockProducts.value.length);
const warningStockCount = computed(() => warningStockProducts.value.length);

const totalPages = computed(() => (products.value.length ? Math.ceil(products.value.length / pageSize) : 1));
const paginatedProducts = computed(() => {
  if (!products.value.length) {
    return [];
  }
  const start = (page.value - 1) * pageSize;
  return products.value.slice(start, start + pageSize);
});

const productRangeLabel = computed(() => {
  if (!products.value.length) {
    return 'Menampilkan 0 dari 0 produk';
  }
  const start = (page.value - 1) * pageSize;
  const from = start + 1;
  const to = Math.min(start + pageSize, products.value.length);
  return `Menampilkan ${from}-${to} dari ${products.value.length} produk`;
});

const productModalFiltered = computed(() => {
  const query = productModalSearch.value.trim().toLowerCase();
  if (!query) {
    return products.value;
  }
  return products.value.filter((product) => {
    return [product.name, product.sku]
      .filter(Boolean)
      .some((field) => (field as string).toLowerCase().includes(query));
  });
});

const productModalTotalPages = computed(() =>
  productModalFiltered.value.length ? Math.ceil(productModalFiltered.value.length / productModalPageSize) : 1
);

const productModalItems = computed(() => {
  if (!productModalFiltered.value.length) {
    return [] as Product[];
  }
  const start = (productModalPage.value - 1) * productModalPageSize;
  return productModalFiltered.value.slice(start, start + productModalPageSize);
});

const productModalRangeLabel = computed(() => {
  const total = productModalFiltered.value.length;
  if (!total) {
    return 'Menampilkan 0 dari 0 produk';
  }
  const start = (productModalPage.value - 1) * productModalPageSize;
  const from = start + 1;
  const to = Math.min(start + productModalPageSize, total);
  return `Menampilkan ${from}-${to} dari ${total} produk`;
});

watch(products, () => {
  if (!products.value.length) {
    page.value = 1;
    productModalPage.value = 1;
    return;
  }
  if (page.value > totalPages.value) {
    page.value = totalPages.value;
  }
  if (productModalPage.value > productModalTotalPages.value) {
    productModalPage.value = productModalTotalPages.value;
  }
});

watch(productModalFiltered, () => {
  if (productModalPage.value > productModalTotalPages.value) {
    productModalPage.value = productModalTotalPages.value;
  }
});

watch(
  () => form.lowStockThreshold,
  (value) => {
    if (!Number.isFinite(value) || value <= 0) {
      form.lowStockThreshold = 1;
    }
  }
);

watch(productModalOpen, (open) => {
  if (open) {
    productModalSearch.value = '';
    productModalPage.value = 1;
  }
});

watch(productDetailOpen, (open) => {
  if (!open) {
    productDetail.value = null;
  }
});

watch(
  () => props.refreshToken,
  () => {
    retryLoadProducts();
  }
);

async function loadProducts() {
  const requestId = ++loadRequestId;
  productsLoading.value = true;
  loadError.value = '';
  try {
    const fetched = await listProducts();
    const normalised = fetched.map((item) => ({
      ...item,
      lowStockThreshold: resolveThreshold(item)
    }));
    normalised.sort((a, b) => {
      const severityDiff = stockSeverity(b) - stockSeverity(a);
      if (severityDiff !== 0) return severityDiff;
      return a.name.localeCompare(b.name, 'id');
    });
    if (requestId === loadRequestId) {
      products.value = normalised;
    }
  } catch (error) {
    console.error(error);
    if (requestId === loadRequestId) {
      products.value = [];
      loadError.value = 'Gagal memuat data produk. Silakan coba lagi.';
      toast.push(loadError.value, 'error');
    }
  } finally {
    if (requestId === loadRequestId) {
      productsLoading.value = false;
    }
  }
}

function retryLoadProducts() {
  void loadProducts();
}

function resetForm() {
  Object.assign(form, {
    id: undefined,
    name: '',
    sku: '',
    costPrice: 0,
    salePrice: 0,
    stock: 0,
    category: '',
    lowStockThreshold: 5,
    description: '',
    imageData: '',
    imageMime: '',
    imageUrl: '',
    thumbUrl: '',
    imagePath: '',
    thumbPath: '',
    imageHash: '',
    imageWidth: 0,
    imageHeight: 0,
    imageSizeBytes: 0,
    thumbWidth: 0,
    thumbHeight: 0,
    thumbSizeBytes: 0
  });
  editing.value = false;
  costPriceDisplay.value = '';
  salePriceDisplay.value = '';
}

async function handleSubmit() {
  try {
    const payload = {
      ...form,
      category: form.category.trim(),
      costPrice: Math.max(0, Number.isFinite(form.costPrice) ? Number(form.costPrice) : 0),
      salePrice: Math.max(0, Number.isFinite(form.salePrice) ? Number(form.salePrice) : 0),
      lowStockThreshold: Math.max(1, Number.isFinite(form.lowStockThreshold) ? Number(form.lowStockThreshold) : 5),
      stock: Math.max(0, Number.isFinite(form.stock) ? Number(form.stock) : 0)
    };
    await saveProduct(payload);
    await loadProducts();
    toast.push(editing.value ? 'Produk diperbarui.' : 'Produk ditambahkan.', 'success');
    resetForm();
  } catch (error) {
    console.error(error);
    toast.push('Gagal menyimpan produk.', 'error');
  }
}

function editProduct(product: Product) {
  Object.assign(form, {
    ...product,
    category: product.category || '',
    lowStockThreshold: product.lowStockThreshold && product.lowStockThreshold > 0 ? product.lowStockThreshold : 5,
    stock: product.stock ?? 0
  });
  editing.value = true;
  productDetailOpen.value = false;
}

function clearProductImage() {
  form.imageData = '';
  form.imageMime = '';
  form.imageUrl = '';
  form.thumbUrl = '';
  form.imagePath = '';
  form.thumbPath = '';
  form.imageHash = '';
  form.imageWidth = 0;
  form.imageHeight = 0;
  form.imageSizeBytes = 0;
  form.thumbWidth = 0;
  form.thumbHeight = 0;
  form.thumbSizeBytes = 0;
}

function handleProductImageChange(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) {
    return;
  }
  const reader = new FileReader();
  reader.onload = () => {
    const result = reader.result as string;
    const base64 = result.includes(',') ? result.split(',')[1] : result;
    form.imageData = base64;
    form.imageMime = file.type || 'image/png';
    form.imageUrl = '';
    form.thumbUrl = '';
    form.imagePath = '';
    form.thumbPath = '';
  };
  reader.readAsDataURL(file);
  target.value = '';
}

async function archiveProductAction(product: Product) {
  if (!product.id) return;
  const confirmed = window.confirm(
    `Arsipkan ${product.name}? Produk yang diarsipkan tidak ditampilkan lagi tetapi histori order tetap aman.`
  );
  if (!confirmed) return;
  try {
    await archiveProductApi(product.id);
    if (form.id === product.id) {
      resetForm();
    }
    await loadProducts();
    toast.push(`${product.name} diarsipkan.`, 'success');
  } catch (error) {
    console.error(error);
    toast.push('Gagal mengarsipkan produk.', 'error');
  }
}

function startOpnameWithProduct(product: Product) {
  if (!product.id) {
    return;
  }
  emit('request-stock-opname', product.id);
}

function openProductDetail(product: Product) {
  productDetail.value = product;
  productDetailOpen.value = true;
}

function productThumbSource(product: Product): string {
  if (product.thumbUrl) {
    return product.thumbUrl;
  }
  if (product.imageUrl) {
    return product.imageUrl;
  }
  const fromThumbPath = resolveMediaPath(product.thumbPath);
  if (fromThumbPath) {
    return fromThumbPath;
  }
  const fromImagePath = resolveMediaPath(product.imagePath);
  if (fromImagePath) {
    return fromImagePath;
  }
  return '';
}

function previousPage() {
  if (page.value > 1) {
    page.value -= 1;
  }
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value += 1;
  }
}

function previousModalPage() {
  if (productModalPage.value > 1) {
    productModalPage.value -= 1;
  }
}

function nextModalPage() {
  if (productModalPage.value < productModalTotalPages.value) {
    productModalPage.value += 1;
  }
}

function productStatus(product: Product) {
  const severity = stockSeverity(product);
  if (severity === 2) {
    return {
      label: 'Out of Stock',
      classes: 'border-red-200 bg-red-50 text-red-600',
      icon: 'critical' as const
    };
  }
  if (severity === 1) {
    return {
      label: 'Stok Menipis',
      classes: 'border-amber-200 bg-amber-50 text-amber-600',
      icon: 'warning' as const
    };
  }
  return {
    label: 'Ready',
    classes: 'border-emerald-200 bg-emerald-50 text-emerald-600',
    icon: 'ready' as const
  };
}

function thresholdLabel(product: Product) {
  return `Min ${resolveThreshold(product)}`;
}

function formatCurrency(value: number) {
  return value.toLocaleString('id-ID');
}

function formatDate(value?: string | null) {
  if (!value) {
    return '';
  }
  try {
    return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
  } catch (error) {
    return value;
  }
}

function formatPriceInput(value: number) {
  const numeric = Number.isFinite(value) ? Math.max(0, Math.round(value)) : 0;
  if (numeric === 0) {
    return '';
  }
  return new Intl.NumberFormat('id-ID').format(numeric);
}

function parsePriceInput(raw: string) {
  const cleaned = raw.replace(/[^0-9]/g, '');
  if (!cleaned.length) {
    return { value: 0, display: '' };
  }
  const numeric = Number(cleaned);
  return { value: numeric, display: formatPriceInput(numeric) };
}

function onCostPriceInput(event: Event) {
  const target = event.target as HTMLInputElement;
  const { value, display } = parsePriceInput(target.value);
  if (form.costPrice !== value) {
    form.costPrice = value;
  }
  costPriceDisplay.value = display;
}

function onSalePriceInput(event: Event) {
  const target = event.target as HTMLInputElement;
  const { value, display } = parsePriceInput(target.value);
  if (form.salePrice !== value) {
    form.salePrice = value;
  }
  salePriceDisplay.value = display;
}

function syncCostPriceDisplay() {
  const { display } = parsePriceInput(costPriceDisplay.value);
  costPriceDisplay.value = display;
}

function syncSalePriceDisplay() {
  const { display } = parsePriceInput(salePriceDisplay.value);
  salePriceDisplay.value = display;
}

onMounted(async () => {
  await loadProducts();
});
</script>
