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
        <div
          v-if="customersTotal"
          class="flex flex-wrap items-center gap-3 text-sm text-slate-500 md:justify-end"
        >
          <span>{{ customerRangeLabel }}</span>
          <span>Halaman {{ page }} / {{ totalPages }}</span>
          <button type="button" class="btn-ghost text-xs" @click="customerModalOpen = true">
            Lihat Semua
          </button>
        </div>
      </div>
      <div v-if="customersLoading" class="py-6 text-center text-sm text-slate-500">Memuat kontak...</div>
      <div
        v-else-if="customersError"
        class="space-y-3 rounded-lg border border-red-100 bg-red-50 p-4 text-sm text-red-600"
      >
        <p>{{ customersError }}</p>
        <button type="button" class="btn-secondary text-xs" @click="loadCustomers">Coba lagi</button>
      </div>
      <template v-else>
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
                <button
                  type="button"
                  class="inline-flex items-center gap-1 rounded-lg border border-red-300 px-3 py-1 text-xs font-semibold text-red-600 transition hover:bg-red-50"
                  @click="deleteCustomerAction(customer)"
                >
                  <TrashIcon class="h-4 w-4" />
                  Hapus
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
        <p v-if="!paginatedCustomers.length" class="text-center text-sm text-slate-500">Belum ada customer.</p>
        <footer
          v-else
          class="mt-4 flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between"
        >
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
      </template>
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
          <div v-if="customerModalLoading" class="py-6 text-center text-sm text-slate-500">Memuat kontak...</div>
          <div
            v-else-if="customerModalError"
            class="space-y-3 rounded-lg border border-red-100 bg-red-50 p-4 text-sm text-red-600"
          >
            <p>{{ customerModalError }}</p>
            <button type="button" class="btn-secondary text-xs" @click="loadCustomerModal">Coba lagi</button>
          </div>
          <template v-else>
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
            <p v-if="!customerModalItems.length" class="text-center text-sm text-slate-500">
              Tidak ada kontak yang cocok dengan pencarian.
            </p>
          </template>
        </div>
        <div
          v-if="customerModalTotal && !customerModalLoading && !customerModalError"
          class="flex flex-col gap-3 border-t border-slate-100 pt-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between"
        >
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
import { deleteCustomer as deleteCustomerApi, listCustomers, saveCustomer } from '../../modules/customer';
import { useToastStore } from '../stores/toast';
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  EnvelopeIcon,
  MagnifyingGlassIcon,
  MapPinIcon,
  PencilSquareIcon,
  PhoneIcon,
  TrashIcon,
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
const customersTotal = ref(0);
const customerCounts = ref<Record<CustomerType, number>>({ customer: 0, marketer: 0, reseller: 0 });
const customersLoading = ref(false);
const customersError = ref('');

const editing = ref(false);
const page = ref(1);
const pageSize = 6;
const paginationButtonClasses =
  'inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-600 transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-40';

const customerModalOpen = ref(false);
const customerModalSearch = ref('');
const customerModalPage = ref(1);
const customerModalPageSize = 10;
const customerModalItems = ref<Customer[]>([]);
const customerModalTotal = ref(0);
const customerModalLoading = ref(false);
const customerModalError = ref('');
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

const counts = computed(() => customerCounts.value);

const totalPages = computed(() => (customersTotal.value ? Math.ceil(customersTotal.value / pageSize) : 1));
const paginatedCustomers = computed(() => customers.value);

const customerRangeLabel = computed(() => {
  if (!customers.value.length || !customersTotal.value) {
    return 'Menampilkan 0 kontak';
  }
  const start = (page.value - 1) * pageSize;
  const from = start + 1;
  const to = Math.min(start + customers.value.length, customersTotal.value);
  return `Menampilkan ${from}-${to} dari ${customersTotal.value} kontak`;
});

const customerModalTotalPages = computed(() =>
  customerModalTotal.value ? Math.ceil(customerModalTotal.value / customerModalPageSize) : 1
);

const customerModalRangeLabel = computed(() => {
  if (!customerModalItems.value.length || !customerModalTotal.value) {
    return 'Menampilkan 0 kontak';
  }
  const start = (customerModalPage.value - 1) * customerModalPageSize;
  const from = start + 1;
  const to = Math.min(start + customerModalItems.value.length, customerModalTotal.value);
  return `Menampilkan ${from}-${to} dari ${customerModalTotal.value} kontak`;
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

watch(customerModalOpen, (open) => {
  if (open) {
    customerModalSearch.value = '';
    customerModalPage.value = 1;
    void loadCustomerModal();
  } else {
    customerModalItems.value = [];
    customerModalTotal.value = 0;
    customerModalError.value = '';
  }
});

watch(customerModalSearch, () => {
  customerModalPage.value = 1;
  scheduleLoadCustomerModal();
});

watch(customerModalPage, (value, oldValue) => {
  if (value !== oldValue && customerModalOpen.value) {
    void loadCustomerModal();
  }
});

watch(page, (value, oldValue) => {
  if (value !== oldValue) {
    void loadCustomers();
  }
});

watch(customerDetailOpen, (open) => {
  if (!open) {
    activeCustomer.value = null;
  }
});

let customerFetchId = 0;
let customerModalFetchId = 0;
let customerModalTimer: ReturnType<typeof setTimeout> | null = null;

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
  const requestId = ++customerFetchId;
  customersLoading.value = true;
  customersError.value = '';
  try {
    const response = await listCustomers({
      page: page.value,
      pageSize
    });

    if (requestId !== customerFetchId) {
      return;
    }

    const totalPagesCount = Math.max(1, Math.ceil(response.total / pageSize));
    if (page.value > totalPagesCount) {
      page.value = totalPagesCount;
      return;
    }

    customers.value = response.items;
    customersTotal.value = response.total;
    customerCounts.value = response.counts;
  } catch (error) {
    console.error(error);
    if (requestId === customerFetchId) {
      customers.value = [];
      customersTotal.value = 0;
      customerCounts.value = { customer: 0, marketer: 0, reseller: 0 };
      customersError.value = 'Gagal memuat daftar kontak.';
      toast.push(customersError.value, 'error');
    }
  } finally {
    if (requestId === customerFetchId) {
      customersLoading.value = false;
    }
  }
}

async function loadCustomerModal() {
  const requestId = ++customerModalFetchId;
  customerModalLoading.value = true;
  customerModalError.value = '';
  try {
    const response = await listCustomers({
      page: customerModalPage.value,
      pageSize: customerModalPageSize,
      query: customerModalSearch.value.trim() || undefined
    });

    if (requestId !== customerModalFetchId) {
      return;
    }

    const totalPagesCount = Math.max(1, Math.ceil(response.total / customerModalPageSize));
    if (customerModalPage.value > totalPagesCount) {
      customerModalPage.value = totalPagesCount;
      return;
    }

    customerModalItems.value = response.items;
    customerModalTotal.value = response.total;
  } catch (error) {
    console.error(error);
    if (requestId === customerModalFetchId) {
      customerModalItems.value = [];
      customerModalTotal.value = 0;
      customerModalError.value = 'Gagal memuat daftar kontak.';
      toast.push(customerModalError.value, 'error');
    }
  } finally {
    if (requestId === customerModalFetchId) {
      customerModalLoading.value = false;
    }
  }
}

function scheduleLoadCustomerModal(delay = 250) {
  if (customerModalTimer) {
    clearTimeout(customerModalTimer);
  }
  customerModalTimer = setTimeout(() => {
    customerModalTimer = null;
    if (customerModalOpen.value) {
      void loadCustomerModal();
    }
  }, delay);
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

async function deleteCustomerAction(customer: Customer) {
  if (!customer.id) {
    return;
  }
  const confirmed = window.confirm(
    `Hapus ${customer.name}? Tindakan ini akan menghapus data kontak dan tidak dapat dibatalkan.`
  );
  if (!confirmed) {
    return;
  }
  try {
    await deleteCustomerApi(customer.id);
    if (form.id === customer.id) {
      resetForm();
    }
    if (activeCustomer.value?.id === customer.id) {
      customerDetailOpen.value = false;
      activeCustomer.value = null;
    }
    await loadCustomers();
    toast.push(`${customer.name} dihapus.`, 'success');
  } catch (error) {
    console.error(error);
    const message = error instanceof Error && error.message ? error.message : 'Gagal menghapus kontak.';
    toast.push(message, 'error');
  }
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
