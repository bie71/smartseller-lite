<template>
  <section class="space-y-6">
    <div class="card space-y-5">
      <header class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div class="flex items-center gap-3">
          <div class="h-12 w-12 rounded-full bg-primary/10 text-primary flex items-center justify-center">
            <TruckIcon class="h-6 w-6" />
          </div>
          <div>
            <h2 class="text-xl font-semibold">Daftar Ekspedisi</h2>
            <p class="text-sm text-slate-500">Kelola daftar ekspedisi favorit agen dan detail layanan mereka.</p>
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <button class="btn-secondary text-sm" @click="loadCouriers">Refresh</button>
          <button v-if="couriers.length" type="button" class="btn-ghost" @click="courierModalOpen = true">Lihat Semua</button>
        </div>
      </header>

      <form class="grid gap-4 md:grid-cols-2" @submit.prevent="handleSubmit">
        <div>
          <label class="text-sm font-medium text-slate-600">Nama Ekspedisi</label>
          <input v-model="form.name" type="text" class="input mt-1" placeholder="JNE" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Kode</label>
          <input v-model="form.code" type="text" class="input mt-1 uppercase" placeholder="JNE" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Layanan Populer</label>
          <input v-model="form.services" type="text" class="input mt-1" placeholder="REG · YES" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Kontak</label>
          <input v-model="form.contact" type="text" class="input mt-1" placeholder="(021) 2927 8888" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Tracking URL</label>
          <input v-model="form.trackingUrl" type="text" class="input mt-1" placeholder="https://" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Catatan</label>
          <input v-model="form.notes" type="text" class="input mt-1" placeholder="Jam pick-up, dsb" />
        </div>
        <div class="md:col-span-2">
          <label class="text-sm font-medium text-slate-600">Logo / Gambar</label>
          <div class="mt-1 flex flex-col gap-3 md:flex-row md:items-center">
            <div class="flex h-20 w-20 items-center justify-center overflow-hidden rounded-lg bg-slate-100">
              <img v-if="logoPreview" :src="logoPreview" alt="Preview ekspedisi" class="h-full w-full object-cover" />
              <PhotoIcon v-else class="h-8 w-8 text-slate-400" />
            </div>
            <div class="flex flex-col gap-2">
              <input type="file" accept="image/*" class="input" @change="handleLogoChange" />
              <button v-if="logoPreview" type="button" class="text-xs font-medium text-red-500" @click="clearLogo">
                Hapus logo
              </button>
            </div>
          </div>
        </div>
        <div class="md:col-span-2 flex items-center gap-2">
          <button type="submit" class="btn-primary">
            <CheckBadgeIcon class="h-5 w-5" />
            {{ editing ? 'Update Ekspedisi' : 'Tambah Ekspedisi' }}
          </button>
          <button v-if="editing" type="button" class="btn-secondary" @click="resetForm">Batal</button>
        </div>
      </form>
    </div>

    <div class="card">
      <div class="mb-4 flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 class="text-lg font-semibold">Ekspedisi Tersimpan</h3>
          <p class="text-sm text-slate-500" v-if="couriers.length">{{ courierRangeLabel }} · Halaman {{ page }} / {{ totalPages }}</p>
        </div>
        <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
          <div class="relative w-full sm:w-64">
            <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
            <input
              v-model="courierSearch"
              type="search"
              class="input pl-10"
              placeholder="Cari nama, kode, atau layanan"
            />
          </div>
          <button class="btn-secondary text-sm" @click="loadCouriers">Refresh</button>
          <button v-if="couriers.length" type="button" class="btn-ghost" @click="courierModalOpen = true">Lihat Semua</button>
        </div>
      </div>
      <div v-if="couriers.length" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <template v-if="paginatedCouriers.length">
          <article
            v-for="courier in paginatedCouriers"
            :key="courier.id"
            class="card border border-slate-200 hover:border-primary/40 transition-colors shadow-sm"
          >
            <header class="flex items-start justify-between gap-3">
              <div class="flex items-center gap-3">
                <div class="h-12 w-12 overflow-hidden rounded-xl bg-primary/5">
                  <img
                    v-if="courierLogoSrc(courier)"
                    :src="courierLogoSrc(courier)"
                    :alt="courier.name"
                    class="h-full w-full object-cover"
                  />
                  <PhotoIcon v-else class="mx-auto my-3 h-6 w-6 text-primary" />
                </div>
                <div>
                  <p class="text-xs font-semibold tracking-wide text-primary">{{ courier.code }}</p>
                  <h3 class="text-lg font-semibold">{{ courier.name }}</h3>
                </div>
              </div>
              <div class="flex gap-2">
                <button class="icon-btn" @click="editCourier(courier)">
                  <PencilSquareIcon class="h-5 w-5" />
                </button>
                <button class="icon-btn text-red-500 hover:text-red-600" @click="removeCourier(courier)">
                  <TrashIcon class="h-5 w-5" />
                </button>
              </div>
            </header>
            <button type="button" class="btn-secondary mt-3 w-full text-sm" @click="viewCourier(courier)">
              <InformationCircleIcon class="h-5 w-5" />
              Lihat Detail
            </button>
            <dl class="mt-3 space-y-2 text-sm text-slate-600">
              <div class="flex items-center gap-2">
                <SparklesIcon class="h-4 w-4 text-primary" />
                <span>{{ courier.services || 'Tidak ada informasi' }}</span>
              </div>
              <div v-if="courier.contact" class="flex items-center gap-2">
                <PhoneIcon class="h-4 w-4 text-primary" />
                <span>{{ courier.contact }}</span>
              </div>
              <div v-if="courier.trackingUrl" class="flex items-center gap-2">
                <LinkIcon class="h-4 w-4 text-primary" />
                <a :href="courier.trackingUrl" target="_blank" rel="noreferrer" class="underline">Lacak paket</a>
              </div>
              <p v-if="courier.notes" class="text-xs text-slate-500">Catatan: {{ courier.notes }}</p>
            </dl>
          </article>
        </template>
        <p v-else class="col-span-full text-center text-sm text-slate-500">Tidak ada ekspedisi yang cocok dengan pencarian.</p>
      </div>
      <p v-else class="text-center text-sm text-slate-500">Belum ada ekspedisi terdaftar.</p>
      <footer
        v-if="couriers.length && filteredCouriers.length"
        class="mt-4 flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between"
      >
        <span>{{ courierRangeLabel }}</span>
        <div class="flex items-center gap-2">
          <button type="button" :class="paginationButtonClasses" :disabled="page === 1" @click="previousPage">
            <ChevronLeftIcon class="h-4 w-4" />
          </button>
          <button type="button" :class="paginationButtonClasses" :disabled="page === totalPages" @click="nextPage">
            <ChevronRightIcon class="h-4 w-4" />
          </button>
        </div>
      </footer>
    </div>

    <BaseModal v-model="courierModalOpen" title="Semua Ekspedisi">
      <div class="space-y-4">
        <div class="relative">
          <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
          <input
            v-model="courierModalSearch"
            type="search"
            class="input pl-10"
            placeholder="Cari nama, kode, atau layanan"
          />
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead class="text-left text-slate-500 uppercase tracking-wider">
              <tr>
                <th class="py-2">Logo</th>
                <th class="py-2">Nama</th>
                <th class="py-2">Kode</th>
                <th class="py-2">Layanan</th>
                <th class="py-2">Kontak</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!courierModalFiltered.length">
                <td colspan="5" class="py-6 text-center text-slate-500">Tidak ada ekspedisi yang cocok dengan pencarian.</td>
              </tr>
              <tr v-for="courier in courierModalItems" :key="courier.id" class="border-t border-slate-100">
                <td class="py-3">
                  <div class="h-10 w-10 overflow-hidden rounded-md bg-slate-100">
                    <img
                      v-if="courierLogoSrc(courier)"
                      :src="courierLogoSrc(courier)"
                      :alt="courier.name"
                      class="h-full w-full object-cover"
                    />
                    <PhotoIcon v-else class="mx-auto my-2 h-5 w-5 text-slate-300" />
                  </div>
                </td>
                <td class="py-3 font-medium">{{ courier.name }}</td>
                <td class="py-3 text-slate-500">{{ courier.code }}</td>
                <td class="py-3">{{ courier.services || '—' }}</td>
                <td class="py-3">{{ courier.contact || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="courierModalFiltered.length" class="flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between">
          <span>{{ courierModalRangeLabel }}</span>
          <div class="flex items-center gap-2">
            <button type="button" :class="paginationButtonClasses" :disabled="courierModalPage === 1" @click="previousModalPage">
              <ChevronLeftIcon class="h-4 w-4" />
            </button>
            <button
              type="button"
              :class="paginationButtonClasses"
              :disabled="courierModalPage === courierModalTotalPages"
              @click="nextModalPage"
            >
              <ChevronRightIcon class="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="courierModalOpen = false">Tutup</button>
      </template>
    </BaseModal>

    <BaseModal v-model="detailModalOpen" :title="selectedCourier?.name || 'Detail Ekspedisi'">
      <div v-if="selectedCourier" class="space-y-5">
        <div class="flex items-center gap-4">
          <div class="h-16 w-16 overflow-hidden rounded-xl bg-primary/5 flex items-center justify-center">
            <img
              v-if="courierLogoSrc(selectedCourier)"
              :src="courierLogoSrc(selectedCourier)"
              :alt="selectedCourier?.name"
              class="h-full w-full object-cover"
            />
            <PhotoIcon v-else class="h-8 w-8 text-primary" />
          </div>
          <div>
            <p class="text-xs font-semibold tracking-wide text-primary">{{ selectedCourier.code }}</p>
            <p class="text-lg font-semibold">{{ selectedCourier.name }}</p>
            <p class="text-sm text-slate-500" v-if="selectedCourier.services">{{ selectedCourier.services }}</p>
          </div>
        </div>

        <dl class="space-y-3 text-sm">
          <div class="flex flex-col gap-1">
            <dt class="font-medium text-slate-500">Kontak</dt>
            <dd>{{ selectedCourier.contact || 'Tidak tersedia' }}</dd>
          </div>
          <div class="flex flex-col gap-1">
            <dt class="font-medium text-slate-500">Tracking</dt>
            <dd>
              <template v-if="selectedCourier.trackingUrl">
                <a :href="selectedCourier.trackingUrl" target="_blank" rel="noreferrer" class="text-primary underline"
                  >{{ selectedCourier.trackingUrl }}</a
                >
              </template>
              <template v-else>Belum diatur</template>
            </dd>
          </div>
          <div class="flex flex-col gap-1" v-if="selectedCourier.notes">
            <dt class="font-medium text-slate-500">Catatan</dt>
            <dd>{{ selectedCourier.notes }}</dd>
          </div>
          <div class="flex flex-col gap-1" v-if="selectedCourier.createdAt || selectedCourier.updatedAt">
            <dt class="font-medium text-slate-500">Riwayat</dt>
            <dd class="text-xs text-slate-500">
              <span v-if="selectedCourier.createdAt">Dibuat: {{ formatDate(selectedCourier.createdAt) }}</span>
              <span v-if="selectedCourier.updatedAt" class="block">Terakhir diubah: {{ formatDate(selectedCourier.updatedAt) }}</span>
            </dd>
          </div>
        </dl>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="detailModalOpen = false">Tutup</button>
      </template>
    </BaseModal>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import BaseModal from '../components/BaseModal.vue';
import { deleteCourier, listCouriers, saveCourier, type Courier } from '../../modules/settings';
import { useToastStore } from '../stores/toast';
import {
  CheckBadgeIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  InformationCircleIcon,
  LinkIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  PhoneIcon,
  PhotoIcon,
  SparklesIcon,
  TrashIcon,
  TruckIcon
} from '@heroicons/vue/24/outline';

const couriers = ref<Courier[]>([]);
const editing = ref(false);
const page = ref(1);
const pageSize = 6;
const paginationButtonClasses =
  'inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-600 transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-40';

const courierModalOpen = ref(false);
const courierModalSearch = ref('');
const courierModalPage = ref(1);
const courierModalPageSize = 12;
const courierSearch = ref('');
const detailModalOpen = ref(false);
const selectedCourier = ref<Courier | null>(null);
const toast = useToastStore();

const form = reactive<Courier>({
  id: undefined,
  code: '',
  name: '',
  services: '',
  trackingUrl: '',
  contact: '',
  notes: '',
  logoData: '',
  logoMime: '',
  logoUrl: '',
  logoPath: '',
  logoHash: '',
  logoWidth: 0,
  logoHeight: 0,
  logoSizeBytes: 0
});

function makeImageSrc(data?: string, mime?: string): string {
  if (!data) return '';
  const type = (mime && mime.length ? mime : 'image/png').trim() || 'image/png';
  return `data:${type};base64,${data}`;
}

function courierLogoSrc(candidate?: Partial<Courier> | null): string {
  if (!candidate) {
    return '';
  }
  if (candidate.logoData) {
    return makeImageSrc(candidate.logoData, candidate.logoMime);
  }
  if (candidate.logoUrl) {
    return candidate.logoUrl;
  }
  if (candidate.logoPath) {
    const normalised = candidate.logoPath
      .replace(/\\/g, '/')
      .replace(/^\.\//, '')
      .replace(/^media\//, '')
      .replace(/^\//, '');
    if (normalised) {
      return `/media/${normalised}`;
    }
  }
  return '';
}

const logoPreview = computed(() => courierLogoSrc(form));

const filteredCouriers = computed(() => {
  const query = courierSearch.value.trim().toLowerCase();
  if (!query) {
    return couriers.value;
  }
  return couriers.value.filter((courier) =>
    [courier.name, courier.code, courier.services, courier.notes, courier.contact]
      .filter(Boolean)
      .some((field) => (field as string).toLowerCase().includes(query))
  );
});

const totalPages = computed(() => (filteredCouriers.value.length ? Math.ceil(filteredCouriers.value.length / pageSize) : 1));
const paginatedCouriers = computed(() => {
  if (!filteredCouriers.value.length) {
    return [] as Courier[];
  }
  const start = (page.value - 1) * pageSize;
  return filteredCouriers.value.slice(start, start + pageSize);
});

const courierRangeLabel = computed(() => {
  const total = filteredCouriers.value.length;
  if (!total) {
    return couriers.value.length ? 'Tidak ada ekspedisi yang cocok dengan pencarian' : 'Menampilkan 0 dari 0 ekspedisi';
  }
  const start = (page.value - 1) * pageSize;
  const from = start + 1;
  const to = Math.min(start + pageSize, total);
  return `Menampilkan ${from}-${to} dari ${total} ekspedisi`;
});

const courierModalFiltered = computed(() => {
  const query = courierModalSearch.value.trim().toLowerCase();
  if (!query) {
    return couriers.value;
  }
  return couriers.value.filter((courier) => {
    return [courier.name, courier.code, courier.services, courier.contact]
      .filter(Boolean)
      .some((field) => (field as string).toLowerCase().includes(query));
  });
});

const courierModalTotalPages = computed(() =>
  courierModalFiltered.value.length ? Math.ceil(courierModalFiltered.value.length / courierModalPageSize) : 1
);

const courierModalItems = computed(() => {
  if (!courierModalFiltered.value.length) {
    return [] as Courier[];
  }
  const start = (courierModalPage.value - 1) * courierModalPageSize;
  return courierModalFiltered.value.slice(start, start + courierModalPageSize);
});

const courierModalRangeLabel = computed(() => {
  const total = courierModalFiltered.value.length;
  if (!total) {
    return 'Menampilkan 0 ekspedisi';
  }
  const start = (courierModalPage.value - 1) * courierModalPageSize;
  const from = start + 1;
  const to = Math.min(start + courierModalPageSize, total);
  return `Menampilkan ${from}-${to} dari ${total} ekspedisi`;
});

watch(couriers, () => {
  if (!couriers.value.length) {
    page.value = 1;
    courierModalPage.value = 1;
    return;
  }
  if (page.value > totalPages.value) {
    page.value = totalPages.value;
  }
  if (courierModalPage.value > courierModalTotalPages.value) {
    courierModalPage.value = courierModalTotalPages.value;
  }
});

watch(filteredCouriers, () => {
  if (!filteredCouriers.value.length) {
    page.value = 1;
  } else if (page.value > totalPages.value) {
    page.value = totalPages.value;
  }
});

watch(courierSearch, () => {
  page.value = 1;
});

watch(
  () => courierModalFiltered.value.length,
  () => {
    if (courierModalPage.value > courierModalTotalPages.value) {
      courierModalPage.value = courierModalTotalPages.value;
    }
  }
);

watch(courierModalOpen, (open) => {
  if (open) {
    courierModalSearch.value = '';
    courierModalPage.value = 1;
  }
});

async function loadCouriers() {
  try {
    couriers.value = await listCouriers();
  } catch (error) {
    console.error(error);
    toast.push('Gagal memuat daftar ekspedisi.', 'error');
  }
}

function resetForm() {
  Object.assign(form, {
    id: undefined,
    code: '',
    name: '',
    services: '',
    trackingUrl: '',
    contact: '',
    notes: '',
    logoData: '',
    logoMime: '',
    logoUrl: '',
    logoPath: '',
    logoHash: '',
    logoWidth: 0,
    logoHeight: 0,
    logoSizeBytes: 0,
    createdAt: undefined,
    updatedAt: undefined
  });
  editing.value = false;
}

async function handleSubmit() {
  try {
    const saved = await saveCourier({ ...form });
    toast.push(editing.value ? 'Data ekspedisi diperbarui.' : 'Ekspedisi ditambahkan.', 'success');
    await loadCouriers();
    if (editing.value) {
      const current = couriers.value.find((item) => item.id === saved.id);
      if (current) {
        Object.assign(form, current);
      }
    } else {
      resetForm();
    }
  } catch (error) {
    console.error(error);
    toast.push('Gagal menyimpan ekspedisi.', 'error');
  }
}

function editCourier(courier: Courier) {
  Object.assign(form, courier);
  editing.value = true;
}

function viewCourier(courier: Courier) {
  selectedCourier.value = courier;
  detailModalOpen.value = true;
}

async function removeCourier(courier: Courier) {
  if (!courier.id) return;
  if (!window.confirm(`Hapus ${courier.name}?`)) return;
  try {
    await deleteCourier(courier.id);
    await loadCouriers();
    if (editing.value && form.id === courier.id) {
      resetForm();
    }
    toast.push(`${courier.name} dihapus.`, 'success');
  } catch (error) {
    console.error(error);
    toast.push('Gagal menghapus ekspedisi.', 'error');
  }
}

function handleLogoChange(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    const result = reader.result as string;
    const base64 = result.includes(',') ? result.split(',')[1] : result;
    form.logoData = base64;
    form.logoMime = file.type || 'image/png';
    form.logoUrl = '';
    form.logoPath = '';
    form.logoHash = '';
    form.logoWidth = 0;
    form.logoHeight = 0;
    form.logoSizeBytes = 0;
  };
  reader.readAsDataURL(file);
  target.value = '';
}

function clearLogo() {
  form.logoData = '';
  form.logoMime = '';
  form.logoUrl = '';
  form.logoPath = '';
  form.logoHash = '';
  form.logoWidth = 0;
  form.logoHeight = 0;
  form.logoSizeBytes = 0;
}

function formatDate(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short'
  });
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
  if (courierModalPage.value > 1) {
    courierModalPage.value -= 1;
  }
}

function nextModalPage() {
  if (courierModalPage.value < courierModalTotalPages.value) {
    courierModalPage.value += 1;
  }
}

watch(detailModalOpen, (open) => {
  if (!open) {
    selectedCourier.value = null;
  }
});

onMounted(loadCouriers);
</script>
