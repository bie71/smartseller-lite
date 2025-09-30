<template>
  <section class="space-y-6">
    <div class="card space-y-5">
      <header class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div class="flex items-center gap-3">
          <div class="flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-600">
            <ClipboardDocumentCheckIcon class="h-6 w-6" />
          </div>
          <div>
            <h3 class="text-lg font-semibold">Stock Opname</h3>
            <p class="text-sm text-slate-500">Catat hasil hitung fisik dan sinkronkan stok produk secara otomatis.</p>
          </div>
        </div>
        <div class="text-sm text-slate-500">
          {{ opnameSummary.totalProducts }} produk dipilih · {{ opnameSummary.changed }} butuh penyesuaian · Petugas:
          {{ opnameUser.trim() || '—' }}
        </div>
      </header>

      <div class="grid gap-6 lg:grid-cols-3">
        <div class="lg:col-span-2 space-y-4">
          <div>
            <label class="text-sm font-medium text-slate-600">Cari Produk</label>
            <div class="relative mt-2">
              <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
              <input
                ref="opnameSearchInput"
                v-model="opnameSearch"
                type="search"
                class="input pl-10"
                placeholder="Cari nama atau SKU produk"
                @keydown.enter.prevent="addFirstCandidate"
              />
            </div>
            <ul
              v-if="opnameCandidates.length"
              class="mt-2 space-y-1 rounded-lg border border-slate-200 bg-white p-3 shadow-sm"
            >
              <li v-for="candidate in opnameCandidates" :key="candidate.id">
                <button
                  type="button"
                  class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm hover:bg-primary/10"
                  @click="addOpnameProduct(candidate)"
                >
                  <span>
                    <span class="block font-medium">{{ candidate.name }}</span>
                    <span class="text-xs text-slate-500">SKU {{ candidate.sku }}</span>
                  </span>
                  <span class="text-xs text-slate-500">Stok {{ candidate.stock }}</span>
                </button>
              </li>
            </ul>
          </div>

          <div
            v-if="!opnameDrafts.length"
            class="rounded-lg border border-dashed border-slate-300 bg-slate-50/70 p-6 text-sm text-slate-500"
          >
            Pilih produk yang ingin dihitung dengan mengetik nama atau SKU di atas.
          </div>
          <div v-else class="overflow-x-auto rounded-lg border border-slate-200">
            <table class="min-w-full text-sm">
              <thead class="bg-slate-50 text-left text-xs uppercase tracking-wider text-slate-500">
                <tr>
                  <th class="px-4 py-3">Produk</th>
                  <th class="px-4 py-3 text-right">Stok Sistem</th>
                  <th class="px-4 py-3 text-right">Hasil Hitung</th>
                  <th class="px-4 py-3 text-right">Selisih</th>
                  <th class="px-4 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="draft in opnameDrafts"
                  :key="draft.productId"
                  :class="[
                    'border-t border-slate-100 transition-colors hover:bg-slate-50/80',
                    highlightedDraftId === draft.productId ? 'bg-primary/5 ring-1 ring-primary/30' : ''
                  ]"
                >
                  <td class="px-4 py-3">
                    <div class="font-medium">{{ productMap.get(draft.productId)?.name || 'Produk' }}</div>
                    <div class="text-xs text-slate-500">{{ productMap.get(draft.productId)?.sku }}</div>
                  </td>
                  <td class="px-4 py-3 text-right font-mono">{{ productMap.get(draft.productId)?.stock ?? 0 }}</td>
                  <td class="px-4 py-3 text-right">
                    <input v-model.number="draft.counted" type="number" min="0" class="input h-9 w-28 text-right" />
                  </td>
                  <td class="px-4 py-3 text-right">
                    <span
                      :class="[
                        'font-semibold',
                        opnameDifference(draft.productId, draft.counted) > 0
                          ? 'text-emerald-600'
                          : opnameDifference(draft.productId, draft.counted) < 0
                          ? 'text-red-500'
                          : 'text-slate-500'
                      ]"
                    >
                      {{ opnameDifference(draft.productId, draft.counted) }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-right">
                    <button
                      type="button"
                      class="text-xs font-medium text-red-500 hover:text-red-600"
                      @click="removeOpnameProduct(draft.productId)"
                    >
                      Hapus
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <aside class="space-y-3 rounded-xl border border-slate-200 bg-slate-50/60 p-4">
          <div>
            <label class="text-sm font-medium text-slate-600">Petugas</label>
            <input
              v-model="opnameUser"
              type="text"
              class="input mt-1"
              placeholder="Nama petugas opname"
            />
            <p v-if="opnameDrafts.length && !opnameUser.trim()" class="mt-1 text-xs font-medium text-amber-600">
              Isi nama petugas sebelum menyimpan penyesuaian.
            </p>
          </div>
          <div>
            <label class="text-sm font-medium text-slate-600">Catatan</label>
            <textarea
              v-model="opnameNote"
              rows="3"
              class="input mt-1"
              placeholder="Contoh: opname mingguan gudang"
            ></textarea>
          </div>
          <div class="rounded-lg bg-white p-3 text-sm shadow-inner">
            <p class="flex items-center justify-between">
              <span>Total produk</span>
              <span class="font-semibold">{{ opnameSummary.totalProducts }}</span>
            </p>
            <p class="flex items-center justify-between text-emerald-600">
              <span>Tambah stok</span>
              <span class="font-semibold">+{{ opnameSummary.increase }}</span>
            </p>
            <p class="flex items-center justify-between text-red-500">
              <span>Kurangi stok</span>
              <span class="font-semibold">-{{ opnameSummary.decrease }}</span>
            </p>
            <p class="mt-1 text-xs text-slate-500">Perubahan dicatat sebagai mutasi stok otomatis.</p>
          </div>
          <button
            type="button"
            class="btn-primary w-full justify-center"
            :disabled="!canSubmitOpname"
            @click="submitStockOpname"
          >
            <ClipboardDocumentCheckIcon class="h-5 w-5" />
            Simpan Penyesuaian
          </button>
          <button
            type="button"
            class="btn-secondary w-full justify-center"
            :disabled="!opnameDrafts.length || opnameSaving"
            @click="resetOpnameSession"
          >
            <XMarkIcon class="h-5 w-5" />
            Reset
          </button>
        </aside>
      </div>

      <div class="border-t border-slate-100 pt-4">
        <h4 class="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-600">
          <ClockIcon class="h-4 w-4 text-primary" />
          Riwayat Stock Opname Terakhir
        </h4>
        <div v-if="opnamesLoading" class="space-y-3">
          <div
            v-for="placeholder in 3"
            :key="placeholder"
            class="h-20 animate-pulse rounded-lg border border-slate-100 bg-slate-50/80"
          ></div>
        </div>
        <div v-else-if="!recentOpnames.length" class="text-sm text-slate-500">Belum ada stock opname tersimpan.</div>
        <ul v-else class="space-y-3">
          <li
            v-for="opname in recentOpnames"
            :key="opname.id"
            class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm"
          >
            <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <div>
                <p class="font-semibold">{{ formatOpnameDate(opname.performedAt) }}</p>
                <p class="text-xs text-slate-500">{{ opname.note || 'Tanpa catatan tambahan' }}</p>
                <p class="text-xs text-slate-400">Petugas: {{ opname.performedBy || 'Tidak diketahui' }}</p>
              </div>
              <div class="flex flex-col items-start gap-2 text-xs text-slate-500 sm:flex-row sm:items-center sm:gap-3">
                <span>
                  {{ opname.items.length }} produk · +{{ opnameDiffLabel(opname).increase }} / -{{ opnameDiffLabel(opname).decrease }}
                </span>
                <button type="button" class="btn-secondary text-xs" @click="openOpnameDetail(opname)">
                  Detail
                </button>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </div>
    <BaseModal
      v-model="opnameDetailOpen"
      title="Detail Stock Opname"
      :subtitle="opnameDetailSubtitle"
    >
      <div v-if="activeOpname" class="space-y-4 text-sm text-slate-600">
        <div class="grid gap-3 sm:grid-cols-2">
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Tanggal</p>
            <p class="font-medium text-slate-800">{{ formatOpnameDate(activeOpname.performedAt) }}</p>
          </div>
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Petugas</p>
            <p>{{ activeOpname.performedBy || 'Tidak diketahui' }}</p>
          </div>
          <div class="sm:col-span-2" v-if="activeOpname.note">
            <p class="text-xs font-semibold uppercase text-slate-400">Catatan</p>
            <p>{{ activeOpname.note }}</p>
          </div>
        </div>
        <div class="max-h-[50vh] overflow-y-auto rounded-lg border border-slate-200">
          <table class="min-w-full text-sm">
            <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
              <tr>
                <th class="px-4 py-2 text-left">Produk</th>
                <th class="px-4 py-2 text-right">Sebelum</th>
                <th class="px-4 py-2 text-right">Hasil</th>
                <th class="px-4 py-2 text-right">Selisih</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in activeOpname.items"
                :key="item.id"
                class="border-t border-slate-100"
              >
                <td class="px-4 py-2">
                  <div class="font-medium text-slate-700">{{ item.productName || 'Produk' }}</div>
                  <div class="text-xs text-slate-500">SKU {{ item.productSku || '—' }}</div>
                </td>
                <td class="px-4 py-2 text-right font-mono">{{ item.previousStock }}</td>
                <td class="px-4 py-2 text-right font-mono">{{ item.counted }}</td>
                <td
                  class="px-4 py-2 text-right font-semibold"
                  :class="[
                    item.difference > 0 ? 'text-emerald-600' : item.difference < 0 ? 'text-red-500' : 'text-slate-500'
                  ]"
                >
                  {{ item.difference > 0 ? `+${item.difference}` : item.difference }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="opnameDetailOpen = false">Tutup</button>
      </template>
    </BaseModal>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import type { Product } from '../../modules/product';
import { listProducts } from '../../modules/product';
import { listStockOpnames, performStockOpname, type StockOpname } from '../../modules/stock';
import { ClockIcon, ClipboardDocumentCheckIcon, MagnifyingGlassIcon, XMarkIcon } from '@heroicons/vue/24/outline';
import BaseModal from '../components/BaseModal.vue';
import { useToastStore } from '../stores/toast';

interface OpnameDraft {
  productId: string;
  counted: number;
}

const props = defineProps<{ incomingOpnameProductId?: string | null }>();
const emit = defineEmits<{
  (e: 'opname-product-consumed', productId: string): void;
  (e: 'stock-adjusted'): void;
}>();

const products = ref<Product[]>([]);
const opnamesLoading = ref(true);
const opnameDrafts = ref<OpnameDraft[]>([]);
const opnameSearch = ref('');
const opnameNote = ref('');
const opnameUser = ref('');
const opnameSaving = ref(false);
const recentOpnames = ref<StockOpname[]>([]);
const highlightedDraftId = ref<string | null>(null);
const opnameSearchInput = ref<HTMLInputElement | null>(null);
let highlightTimer: ReturnType<typeof setTimeout> | undefined;
const pendingIncomingProductId = ref<string | null>(null);
const opnameDetailOpen = ref(false);
const activeOpname = ref<StockOpname | null>(null);
const toast = useToastStore();

const productMap = computed(() => {
  const map = new Map<string, Product>();
  products.value.forEach((item) => {
    if (item.id) {
      map.set(item.id, item);
    }
  });
  return map;
});

const opnameCandidates = computed(() => {
  const query = opnameSearch.value.trim().toLowerCase();
  if (!query) {
    return [] as Product[];
  }
  return products.value
    .filter((product) =>
      [product.name, product.sku]
        .filter(Boolean)
        .some((field) => (field as string).toLowerCase().includes(query))
    )
    .filter((product) => !opnameDrafts.value.some((draft) => draft.productId === product.id))
    .slice(0, 6);
});

const opnameRows = computed(() =>
  opnameDrafts.value
    .map((draft) => {
      const product = draft.productId ? productMap.value.get(draft.productId) : undefined;
      if (!product) {
        return null;
      }
      const counted = Number.isFinite(draft.counted) ? draft.counted : 0;
      const difference = counted - product.stock;
      return {
        product,
        counted,
        difference
      };
    })
    .filter(Boolean) as Array<{ product: Product; counted: number; difference: number }>
);

const opnameSummary = computed(() => {
  const rows = opnameRows.value;
  const increase = rows.filter((row) => row.difference > 0).reduce((sum, row) => sum + row.difference, 0);
  const decrease = rows.filter((row) => row.difference < 0).reduce((sum, row) => sum + Math.abs(row.difference), 0);
  const changed = rows.filter((row) => row.difference !== 0).length;
  return {
    totalProducts: rows.length,
    increase,
    decrease,
    changed
  };
});

const canSubmitOpname = computed(
  () => opnameDrafts.value.length > 0 && !opnameSaving.value && opnameUser.value.trim().length > 0
);

const opnameDetailSubtitle = computed(() => {
  if (!activeOpname.value) {
    return '';
  }
  const diff = opnameDiffLabel(activeOpname.value);
  return `${activeOpname.value.items.length} produk · +${diff.increase} / -${diff.decrease}`;
});

watch(
  () => opnameDrafts.value,
  (drafts) => {
    drafts.forEach((draft) => {
      if (!Number.isFinite(draft.counted) || draft.counted < 0) {
        draft.counted = 0;
      }
    });
    if (highlightedDraftId.value && !drafts.some((draft) => draft.productId === highlightedDraftId.value)) {
      highlightedDraftId.value = null;
    }
  },
  { deep: true }
);

watch(
  () => props.incomingOpnameProductId,
  (value) => {
    if (!value) {
      pendingIncomingProductId.value = null;
      return;
    }
    pendingIncomingProductId.value = value;
    tryConsumeIncoming();
  },
  { immediate: true }
);

watch(products, () => {
  tryConsumeIncoming();
});

function tryConsumeIncoming() {
  const productId = pendingIncomingProductId.value;
  if (!productId) {
    return;
  }
  if (!productMap.value.has(productId)) {
    return;
  }
  if (handleIncomingOpnameProduct(productId)) {
    emit('opname-product-consumed', productId);
    pendingIncomingProductId.value = null;
  }
}

function highlightDraft(productId: string) {
  highlightedDraftId.value = productId;
  if (highlightTimer) {
    clearTimeout(highlightTimer);
  }
  highlightTimer = setTimeout(() => {
    highlightedDraftId.value = null;
  }, 2200);
}

async function focusOpnameTools() {
  await nextTick();
  opnameSearchInput.value?.focus();
}

function enqueueOpnameProduct(productId: string, counted?: number) {
  const product = productMap.value.get(productId);
  if (!product) {
    return false;
  }
  const draft = opnameDrafts.value.find((item) => item.productId === productId);
  const nextCounted = Number.isFinite(counted) ? Number(counted) : product.stock;
  if (draft) {
    draft.counted = nextCounted;
  } else {
    opnameDrafts.value.push({ productId, counted: nextCounted });
  }
  highlightDraft(productId);
  return true;
}

function handleIncomingOpnameProduct(productId: string) {
  const added = enqueueOpnameProduct(productId);
  if (!added) {
    return false;
  }
  opnameSearch.value = '';
  void focusOpnameTools();
  return true;
}

function addOpnameProduct(product: Product) {
  if (!product.id) {
    return;
  }
  enqueueOpnameProduct(product.id, product.stock);
}

function addFirstCandidate() {
  if (opnameCandidates.value.length) {
    addOpnameProduct(opnameCandidates.value[0]);
  }
}

function removeOpnameProduct(productId: string) {
  opnameDrafts.value = opnameDrafts.value.filter((draft) => draft.productId !== productId);
  if (highlightedDraftId.value === productId) {
    highlightedDraftId.value = null;
  }
}

function resetOpnameSession() {
  opnameDrafts.value = [];
  opnameSearch.value = '';
  opnameNote.value = '';
  highlightedDraftId.value = null;
}

function opnameDifference(productId: string, counted: number) {
  const base = productId ? productMap.value.get(productId)?.stock ?? 0 : 0;
  return counted - base;
}

async function submitStockOpname() {
  if (!opnameDrafts.value.length || opnameSaving.value) {
    return;
  }
  const actor = opnameUser.value.trim();
  if (!actor) {
    toast.push('Nama petugas opname wajib diisi.', 'error');
    return;
  }
  opnameSaving.value = true;
  try {
    const payload = {
      note: opnameNote.value,
      user: actor,
      items: opnameDrafts.value.map((draft) => ({
        productId: draft.productId,
        counted: draft.counted
      }))
    };
    await performStockOpname(payload);
    toast.push(`Stock opname tersimpan oleh ${actor}.`, 'success');
    resetOpnameSession();
    await loadProducts();
    await loadRecentOpnames();
    emit('stock-adjusted');
  } catch (error) {
    console.error(error);
    toast.push('Gagal menyimpan stock opname.', 'error');
  } finally {
    opnameSaving.value = false;
  }
}

function formatOpnameDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short'
  }).format(date);
}

function opnameDiffLabel(opname: StockOpname) {
  const increase = opname.items.filter((item) => item.difference > 0).reduce((sum, item) => sum + item.difference, 0);
  const decrease = opname.items.filter((item) => item.difference < 0).reduce((sum, item) => sum + Math.abs(item.difference), 0);
  return { increase, decrease };
}

function openOpnameDetail(opname: StockOpname) {
  activeOpname.value = opname;
  opnameDetailOpen.value = true;
}

async function loadProducts() {
  try {
    const fetched = await listProducts();
    products.value = fetched;
  } catch (error) {
    console.error(error);
  }
}

async function loadRecentOpnames() {
  try {
    opnamesLoading.value = true;
    recentOpnames.value = await listStockOpnames(5);
  } catch (error) {
    console.error(error);
  } finally {
    opnamesLoading.value = false;
  }
}

onMounted(async () => {
  await loadProducts();
  await loadRecentOpnames();
  tryConsumeIncoming();
});

onBeforeUnmount(() => {
  if (highlightTimer) {
    clearTimeout(highlightTimer);
  }
});

watch(opnameDetailOpen, (open) => {
  if (!open) {
    activeOpname.value = null;
  }
});
</script>
