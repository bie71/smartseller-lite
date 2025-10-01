<template>
  <section class="space-y-6">
    <div class="space-y-6 xl:grid xl:grid-cols-5 xl:gap-6 xl:space-y-0">
      <aside class="space-y-6 xl:col-span-2 xl:order-1">
        <div class="card space-y-5 xl:sticky xl:top-28">
      <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div class="flex items-center gap-3">
          <div class="h-12 w-12 rounded-full bg-primary/10 text-primary flex items-center justify-center">
            <UsersIcon class="h-6 w-6" />
          </div>
          <div>
            <h2 class="text-xl font-semibold">Customer & Relasi</h2>
            <p class="text-sm text-slate-500">Simpan detail pelanggan, marketer, dan reseller dalam satu tempat.</p>
          </div>
        </div>
        <div class="flex w-full flex-wrap items-center gap-3 md:ml-auto md:justify-end">
          <div class="flex gap-2 text-sm text-slate-500">
            <span>{{ counts.customer }} Customer</span>
            <span>{{ counts.marketer }} Marketer</span>
            <span>{{ counts.reseller }} Reseller</span>
          </div>
        </div>
      </div>

      <form class="grid grid-cols-1 gap-4 md:grid-cols-2" @submit.prevent="handleSubmit">
        <div>
          <label class="text-sm font-medium text-slate-600">Nama</label>
          <input v-model="form.name" type="text" class="input mt-1" required />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Kategori</label>
          <select v-model="form.type" class="input mt-1">
            <option v-for="option in customerTypes" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Telepon</label>
          <input v-model="form.phone" type="tel" class="input mt-1" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Email</label>
          <input v-model="form.email" type="email" class="input mt-1" />
        </div>
        <div class="md:col-span-2">
          <label class="text-sm font-medium text-slate-600">Alamat</label>
          <textarea v-model="form.address" rows="2" class="input mt-1"></textarea>
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Kota</label>
          <input v-model="form.city" type="text" class="input mt-1" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Provinsi</label>
          <input v-model="form.province" type="text" class="input mt-1" />
        </div>
        <div>
          <label class="text-sm font-medium text-slate-600">Kode Pos</label>
          <input v-model="form.postal" type="text" class="input mt-1" />
        </div>
        <div class="md:col-span-2">
          <label class="text-sm font-medium text-slate-600">Catatan</label>
          <textarea v-model="form.notes" rows="2" class="input mt-1"></textarea>
        </div>
        <div class="md:col-span-2 flex flex-col gap-2">
          <div class="flex items-center gap-2">
            <button type="submit" class="btn-primary" :disabled="savingCustomer">
            <UserPlusIcon class="h-5 w-5" />
            {{ savingCustomer ? 'Menyimpan...' : editing ? 'Update' : 'Simpan' }} Kontak
          </button>
          <button type="button" v-if="editing" class="btn-secondary" @click="resetForm()">
            <XMarkIcon class="h-5 w-5" />
            Batal
          </button>
        </div>
        </div>
      </form>
        </div>
      </aside>
      <div class="space-y-6 xl:col-span-3 xl:order-2">
        <div class="card">
      <div class="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <h3 class="text-lg font-semibold">Daftar Kontak</h3>
        <div v-if="customers.length" class="flex flex-wrap items-center gap-3 text-sm text-slate-500 md:justify-end">
          <span>{{ customerRangeLabel }}</span>
          <span>Halaman {{ page }} / {{ totalPages }}</span>
          <button type="button" class="btn-ghost text-xs" @click="customerModalOpen = true">
            Lihat Semua
          </button>
        </div>
      </div>
      <div class="grid gap-4 md:grid-cols-2">
        <article
          v-for="customer in paginatedCustomers"
          :key="customer.id"
          class="border border-slate-200 rounded-lg p-4 flex flex-col gap-2"
        >
          <header class="flex items-start justify-between gap-3">
            <div>
              <h4 class="font-semibold">{{ customer.name }}</h4>
              <span class="text-xs uppercase tracking-wide text-primary">{{ labelType(customer.type) }}</span>
            </div>
            <div class="flex gap-2">
              <button type="button" class="btn-ghost text-xs" @click="openCustomerDetail(customer)">
                Detail
              </button>
              <button class="btn-secondary text-xs" @click="editCustomer(customer)">
                <PencilSquareIcon class="h-4 w-4" />
                Edit
              </button>
            </div>
          </header>
          <p v-if="customer.address" class="text-sm text-slate-600 flex items-start gap-2">
            <MapPinIcon class="h-4 w-4 text-primary mt-1" />
            <span>{{ customer.address }}</span>
          </p>
          <p class="text-sm text-slate-500 flex items-center gap-2">
            <MapPinIcon class="h-4 w-4 text-primary" />
            <span>{{ customer.city }} {{ customer.province }} {{ customer.postal }}</span>
          </p>
          <p v-if="customer.phone" class="text-sm flex items-center gap-2">
            <PhoneIcon class="h-4 w-4 text-primary" />
            <span>{{ customer.phone }}</span>
          </p>
          <p v-if="customer.email" class="text-sm flex items-center gap-2">
            <EnvelopeIcon class="h-4 w-4 text-primary" />
            <span>{{ customer.email }}</span>
          </p>
          <p v-if="customer.notes" class="text-xs text-slate-500">Catatan: {{ customer.notes }}</p>
        </article>
      </div>
      <p v-if="!customers.length" class="text-center text-sm text-slate-500">Belum ada customer.</p>
      <footer v-else class="mt-4 flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between">
        <span>{{ customerRangeLabel }}</span>
        <div class="flex items-center gap-3">
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
        </div>
      </div>
    </div>

    <BaseModal v-model="customerModalOpen" title="Seluruh Kontak">
      <div class="space-y-4">
        <div class="relative">
          <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
          <input
            v-model="customerModalSearch"
            type="search"
            class="input pl-10"
            placeholder="Cari nama, email, atau telepon"
          />
        </div>
        <div class="max-h-[60vh] space-y-3">
          <article
            v-for="customer in customerModalItems"
            :key="customer.id"
            class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
          >
            <header class="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h4 class="text-base font-semibold">{{ customer.name }}</h4>
                <span class="text-xs uppercase tracking-wide text-primary">{{ labelType(customer.type) }}</span>
              </div>
              <div class="text-sm text-slate-500 flex flex-wrap gap-3">
                <span v-if="customer.phone" class="flex items-center gap-1">
                  <PhoneIcon class="h-4 w-4 text-primary" />
                  {{ customer.phone }}
                </span>
                <span v-if="customer.email" class="flex items-center gap-1">
                  <EnvelopeIcon class="h-4 w-4 text-primary" />
                  {{ customer.email }}
                </span>
              </div>
            </header>
            <p v-if="customer.address" class="mt-2 text-sm text-slate-600 flex items-start gap-2">
              <MapPinIcon class="h-4 w-4 text-primary mt-1" />
              <span>{{ customer.address }}</span>
            </p>
            <p class="text-xs text-slate-500">{{ customer.city }} {{ customer.province }} {{ customer.postal }}</p>
            <p v-if="customer.notes" class="text-xs text-slate-500">Catatan: {{ customer.notes }}</p>
          </article>
          <p v-if="!customerModalFiltered.length" class="text-center text-sm text-slate-500">
            Tidak ada kontak yang cocok dengan pencarian.
          </p>
        </div>
        <div v-if="customerModalFiltered.length" class="flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between">
          <span>{{ customerModalRangeLabel }}</span>
          <div class="flex items-center gap-2">
            <button type="button" :class="paginationButtonClasses" :disabled="customerModalPage === 1" @click="previousModalPage">
              <ChevronLeftIcon class="h-4 w-4" />
            </button>
            <button
              type="button"
              :class="paginationButtonClasses"
              :disabled="customerModalPage === customerModalTotalPages"
              @click="nextModalPage"
            >
              <ChevronRightIcon class="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="customerModalOpen = false">Tutup</button>
      </template>
    </BaseModal>
    <BaseModal v-model="customerDetailOpen" title="Detail Kontak" :subtitle="customerDetailSubtitle">
      <div v-if="activeCustomer" class="space-y-4 text-sm text-slate-600">
        <div>
          <p class="text-xs font-semibold uppercase text-slate-400">Nama</p>
          <p class="text-base font-semibold text-slate-800">{{ activeCustomer.name }}</p>
          <p class="text-xs uppercase tracking-wide text-primary">{{ labelType(activeCustomer.type) }}</p>
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <div v-if="activeCustomer.phone">
            <p class="text-xs font-semibold uppercase text-slate-400">Telepon</p>
            <p class="flex items-center gap-2">
              <PhoneIcon class="h-4 w-4 text-primary" />
              <span>{{ activeCustomer.phone }}</span>
            </p>
          </div>
          <div v-if="activeCustomer.email">
            <p class="text-xs font-semibold uppercase text-slate-400">Email</p>
            <p class="flex items-center gap-2">
              <EnvelopeIcon class="h-4 w-4 text-primary" />
              <span>{{ activeCustomer.email }}</span>
            </p>
          </div>
        </div>
        <div v-if="activeCustomer.address">
          <p class="text-xs font-semibold uppercase text-slate-400">Alamat</p>
          <p class="flex items-start gap-2">
            <MapPinIcon class="h-4 w-4 text-primary mt-1" />
            <span>{{ activeCustomer.address }}</span>
          </p>
        </div>
        <div class="grid gap-2 text-xs text-slate-500 sm:grid-cols-3">
          <div>
            <p class="font-semibold uppercase text-slate-400">Kota</p>
            <p>{{ activeCustomer.city || '—' }}</p>
          </div>
          <div>
            <p class="font-semibold uppercase text-slate-400">Provinsi</p>
            <p>{{ activeCustomer.province || '—' }}</p>
          </div>
          <div>
            <p class="font-semibold uppercase text-slate-400">Kode Pos</p>
            <p>{{ activeCustomer.postal || '—' }}</p>
          </div>
        </div>
        <p v-if="activeCustomer.notes" class="rounded-lg bg-slate-50 p-3 text-xs text-slate-500">
          {{ activeCustomer.notes }}
        </p>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="customerDetailOpen = false">Tutup</button>
        <button
          v-if="activeCustomer"
          type="button"
          class="btn-primary"
          @click="editCustomerAndClose(activeCustomer)"
        >
          <PencilSquareIcon class="h-4 w-4" />
          Edit Kontak
        </button>
      </template>
    </BaseModal>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import BaseModal from '../components/BaseModal.vue';
import type { Customer, CustomerType } from '../../modules/customer';
import { listCustomers, saveCustomer } from '../../modules/customer';
import { useToastStore } from '../stores/toast';
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  EnvelopeIcon,
  MagnifyingGlassIcon,
  MapPinIcon,
  PencilSquareIcon,
  PhoneIcon,
  UserPlusIcon,
  UsersIcon,
  XMarkIcon
} from '@heroicons/vue/24/outline';

const customerTypes: Array<{ value: CustomerType; label: string }> = [
  { value: 'customer', label: 'Customer' },
  { value: 'marketer', label: 'Marketer' },
  { value: 'reseller', label: 'Reseller' }
];

const customers = ref<Customer[]>([]);
const editing = ref(false);
const page = ref(1);
const pageSize = 6;
const paginationButtonClasses =
  'inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-600 transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-40';

const customerModalOpen = ref(false);
const customerModalSearch = ref('');
const customerModalPage = ref(1);
const customerModalPageSize = 10;
const customerDetailOpen = ref(false);
const activeCustomer = ref<Customer | null>(null);

const savingCustomer = ref(false);
const toast = useToastStore();

const form = reactive<Customer>({
  id: undefined,
  name: '',
  type: 'customer',
  phone: '',
  email: '',
  address: '',
  city: '',
  province: '',
  postal: '',
  notes: ''
});

const counts = computed(() => {
  return customers.value.reduce(
    (acc, customer) => {
      const key = (customer.type || 'customer') as CustomerType;
      acc[key] += 1;
      return acc;
    },
    { customer: 0, marketer: 0, reseller: 0 } as Record<CustomerType, number>
  );
});

const totalPages = computed(() => (customers.value.length ? Math.ceil(customers.value.length / pageSize) : 1));
const paginatedCustomers = computed(() => {
  if (!customers.value.length) {
    return [] as Customer[];
  }
  const start = (page.value - 1) * pageSize;
  return customers.value.slice(start, start + pageSize);
});

const customerRangeLabel = computed(() => {
  if (!customers.value.length) {
    return 'Menampilkan 0 kontak';
  }
  const start = (page.value - 1) * pageSize;
  const from = start + 1;
  const to = Math.min(start + pageSize, customers.value.length);
  return `Menampilkan ${from}-${to} dari ${customers.value.length} kontak`;
});

const customerModalFiltered = computed(() => {
  const query = customerModalSearch.value.trim().toLowerCase();
  if (!query) {
    return customers.value;
  }
  return customers.value.filter((customer) => {
    return [customer.name, customer.email, customer.phone, customer.city, customer.province]
      .filter(Boolean)
      .some((field) => (field as string).toLowerCase().includes(query));
  });
});

const customerModalTotalPages = computed(() =>
  customerModalFiltered.value.length ? Math.ceil(customerModalFiltered.value.length / customerModalPageSize) : 1
);

const customerModalItems = computed(() => {
  if (!customerModalFiltered.value.length) {
    return [] as Customer[];
  }
  const start = (customerModalPage.value - 1) * customerModalPageSize;
  return customerModalFiltered.value.slice(start, start + customerModalPageSize);
});

const customerModalRangeLabel = computed(() => {
  const total = customerModalFiltered.value.length;
  if (!total) {
    return 'Menampilkan 0 kontak';
  }
  const start = (customerModalPage.value - 1) * customerModalPageSize;
  const from = start + 1;
  const to = Math.min(start + customerModalPageSize, total);
  return `Menampilkan ${from}-${to} dari ${total} kontak`;
});

const customerDetailSubtitle = computed(() => {
  if (!activeCustomer.value) {
    return '';
  }
  const parts: string[] = [];
  if (activeCustomer.value.city) {
    parts.push(String(activeCustomer.value.city));
  }
  if (activeCustomer.value.province) {
    parts.push(String(activeCustomer.value.province));
  }
  return parts.length ? parts.join(', ') : 'Kontak tersimpan dalam database';
});

watch(customers, () => {
  if (!customers.value.length) {
    page.value = 1;
    customerModalPage.value = 1;
    return;
  }
  if (page.value > totalPages.value) {
    page.value = totalPages.value;
  }
  if (customerModalPage.value > customerModalTotalPages.value) {
    customerModalPage.value = customerModalTotalPages.value;
  }
});

watch(
  () => customerModalFiltered.value.length,
  () => {
    if (customerModalPage.value > customerModalTotalPages.value) {
      customerModalPage.value = customerModalTotalPages.value;
    }
  }
);

watch(customerModalOpen, (open) => {
  if (open) {
    customerModalSearch.value = '';
    customerModalPage.value = 1;
  }
});

watch(customerDetailOpen, (open) => {
  if (!open) {
    activeCustomer.value = null;
  }
});

function resetForm() {
  Object.assign(form, {
    id: undefined,
    name: '',
    type: 'customer' as CustomerType,
    phone: '',
    email: '',
    address: '',
    city: '',
    province: '',
    postal: '',
    notes: ''
  });
  editing.value = false;
}

async function loadCustomers() {
  try {
    customers.value = await listCustomers();
  } catch (error) {
    console.error(error);
    toast.push('Gagal memuat daftar kontak.', 'error');
  }
}

async function handleSubmit() {
  if (savingCustomer.value) {
    return;
  }
  savingCustomer.value = true;
  try {
    await saveCustomer({ ...form });
    await loadCustomers();
    toast.push(editing.value ? 'Kontak diperbarui.' : 'Kontak tersimpan.', 'success');
    resetForm();
  } catch (error) {
    console.error(error);
    const message = error instanceof Error && error.message ? error.message : 'Gagal menyimpan kontak.';
    toast.push(message, 'error', { timeout: 5000 });
  } finally {
    savingCustomer.value = false;
  }
}

function editCustomer(customer: Customer) {
  customerDetailOpen.value = false;
  Object.assign(form, customer);
  editing.value = true;
}

function openCustomerDetail(customer: Customer) {
  activeCustomer.value = customer;
  customerDetailOpen.value = true;
}

function editCustomerAndClose(customer: Customer) {
  editCustomer(customer);
  customerDetailOpen.value = false;
}

function labelType(type: CustomerType) {
  const option = customerTypes.find((c) => c.value === type);
  return option ? option.label : 'Customer';
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
  if (customerModalPage.value > 1) {
    customerModalPage.value -= 1;
  }
}

function nextModalPage() {
  if (customerModalPage.value < customerModalTotalPages.value) {
    customerModalPage.value += 1;
  }
}

onMounted(loadCustomers);
</script>
