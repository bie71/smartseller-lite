<template>
  <div class="grid grid-cols-1 lg:grid-cols-5 gap-8">
    <!-- Form Column -->
    <div class="lg:col-span-2 space-y-4">
      <div class="inline-flex rounded-lg border border-slate-300 overflow-hidden">
        <button class="px-3 py-2 text-sm" :class="tab==='auto' ? 'bg-primary text-white' : 'bg-white text-slate-700 hover:bg-slate-50'"
                @click="tab='auto'">Auto (Order)</button>
        <button class="px-3 py-2 text-sm border-l border-slate-300" :class="tab==='manual' ? 'bg-primary text-white' : 'bg-white text-slate-700 hover:bg-slate-50'"
                @click="tab='manual'">Manual</button>
      </div>

      <form class="space-y-6 bg-white p-6 rounded-lg shadow" @submit.prevent="handleSubmit">
        <h2 class="text-xl font-semibold text-slate-800 border-b pb-3">{{ isEditing ? 'Edit Detail Pengiriman' : 'Detail Pengiriman' }}</h2>

        <!-- Sender & Recipient -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <h3 class="font-medium text-slate-600">Pengirim</h3>
            <div>
              <label class="block text-sm font-medium text-slate-700 mb-1">Nama Pengirim</label>
              <input v-model="form.senderName" required class="input w-full" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-700 mb-1">No. Telepon</label>
              <input v-model="form.senderPhone" class="input w-full" />
            </div>
          </div>
          <div class="space-y-4">
            <h3 class="font-medium text-slate-600">Penerima</h3>
            <div>
              <label class="block text-sm font-medium text-slate-700 mb-1">Nama Penerima</label>
              <input v-model="form.recipientName" required class="input w-full" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-700 mb-1">No. Telepon</label>
              <input v-model="form.recipientPhone" class="input w-full" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-700 mb-1">Alamat Lengkap</label>
              <textarea v-model="form.recipientAddress" required class="input w-full"></textarea>
            </div>
          </div>
        </div>

        <!-- Order Details -->
        <div class="space-y-4 border-t pt-6">
          <h3 class="font-medium text-slate-600">Detail Pesanan</h3>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">ID Pesanan</label>
            <input v-model="form.orderCode" class="input w-full" />
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label class="block text-sm font-medium text-slate-700 mb-1">No. Resi</label>
              <input v-model="form.trackingCode" class="input w-full" />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-700 mb-1">Kurir Pengiriman</label>
              <input v-model="form.courier" required class="input w-full" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Berat (kg)</label>
            <input v-model.number="form.weight" type="number" step="0.1" class="input w-full" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">Isi Paket</label>
            <textarea v-model="form.notes" class="input w-full"></textarea>
          </div>
        </div>

        <!-- COD -->
        <div class="space-y-4 border-t pt-6">
          <div class="flex items-start">
            <div class="flex items-center h-5">
              <input id="isCOD" type="checkbox" v-model="form.isCOD" class="focus:ring-blue-500 h-4 w-4 text-blue-600 border-slate-300 rounded" />
            </div>
            <div class="ml-3 text-sm">
              <label for="isCOD" class="font-medium text-slate-700">Cash on Delivery (COD)</label>
              <p class="text-slate-500">Aktifkan jika pesanan ini menggunakan metode COD.</p>
            </div>
          </div>
          <div v-if="form.isCOD">
            <label class="block text-sm font-medium text-slate-700 mb-1">Jumlah COD (IDR)</label>
            <input v-model="codAmountFormatted" type="text" inputmode="numeric" class="input w-full" />
          </div>
        </div>

        <div class="pt-4 border-t space-y-3">
          <button type="submit" class="w-full bg-slate-800 hover:bg-slate-700 text-white font-bold py-3 px-4 rounded-lg flex items-center justify-center transition-colors">
            {{ isEditing ? 'Update Label' : 'Tambahkan Label ke Antrian' }}
          </button>
          <button v-if="isEditing" type="button" @click="cancelEdit" class="w-full bg-slate-200 hover:bg-slate-300 text-slate-800 font-bold py-3 px-4 rounded-lg flex items-center justify-center transition-colors">
            Batal Edit
          </button>
        </div>
      </form>
    </div>

    <!-- Preview Column -->
    <div class="lg:col-span-3">
      <div class="bg-white p-6 rounded-lg shadow sticky top-8">
        <div class="flex justify-between items-center border-b pb-3 mb-4">
          <h2 class="text-xl font-semibold text-slate-800">Antrian Cetak ({{ queue.length }})</h2>
          <div class="flex items-center space-x-2">
            <button v-if="queue.length > 0" @click="printQueuePdf" :disabled="isPrinting" class="btn-primary">
              {{ isPrinting ? 'Mencetak...' : `Cetak Semua Label` }}
            </button>
          </div>
        </div>

        <div v-if="queue.length > 0" class="max-h-[80vh] overflow-y-auto pr-2 -mr-2 space-y-4 grid grid-cols-1 xl:grid-cols-1 gap-4">
          <LabelPreview
            v-for="label in queue"
            :key="label.id"
            :label="label"
            :app-settings="appSettings"
            @remove="removeLabel"
            @edit="startEdit"
            @print="printSingleLabel"
          />
        </div>
        <div v-else class="text-center py-16 px-6 border-2 border-dashed border-slate-300 rounded-lg">
          <h3 class="mt-2 text-sm font-medium text-slate-900">Antrian Cetak Kosong</h3>
          <p class="mt-1 text-sm text-slate-500">Isi formulir di samping untuk menambahkan label.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue';
import type { LabelData } from '../../modules/label';
import { generateLabelsPdf, generateSingleLabelPdf } from '../../modules/label';
import { getSettings, type AppSettings } from '../../modules/settings';
import LabelPreview from './LabelPreview.vue';

const props = defineProps<{
  autoData?: Partial<LabelData> | null
}>();

const tab = ref<'auto'|'manual'>('manual');
const queue = ref<LabelData[]>([]);
const isPrinting = ref(false);
const editingLabelId = ref<string | null>(null);
const isEditing = ref(false);
const appSettings = ref<AppSettings | null>(null);

onMounted(async () => {
  try {
    appSettings.value = await getSettings();
  } catch (e) {
    console.error('Failed to load app settings', e);
  }
});

const createFreshForm = (): LabelData => ({
  id: ``, // ID will be set on add
  senderName: appSettings.value?.brandName || 'Toko SmartSeller',
  senderPhone: '081234567890',
  recipientName: '',
  recipientPhone: '',
  recipientAddress: '',
  orderCode: `INV-${Date.now()}`,
  trackingCode: `SS${Date.now()}`,
  courier: 'JNE REG',
  weight: 1,
  notes: '',
  isCOD: false,
  codAmount: 0,
});

const form = ref<LabelData>(createFreshForm());

const codAmountFormatted = computed({
  get() {
    if (form.value.codAmount === undefined || form.value.codAmount === 0) return '';
    return new Intl.NumberFormat('id-ID').format(form.value.codAmount);
  },
  set(value) {
    const num = parseInt(value.replace(/[^0-9]/g, ''), 10);
    form.value.codAmount = isNaN(num) ? 0 : num;
  }
});

watch(() => props.autoData, (v) => {
  if (v) {
    Object.assign(form.value, v);
    tab.value = 'auto';
  }
}, { immediate: true, deep: true });

// Update sender name if settings load after form is created
watch(appSettings, (settings) => {
  if (settings && !form.value.senderName) {
    form.value.senderName = settings.brandName;
  }
});

const handleSubmit = () => {
  if (!form.value.recipientName || !form.value.courier) {
    alert('Nama Penerima dan Kurir wajib diisi.');
    return;
  }

  if (isEditing.value && editingLabelId.value) {
    const index = queue.value.findIndex(l => l.id === editingLabelId.value);
    if (index !== -1) {
      queue.value[index] = { ...form.value };
    }
  } else {
    queue.value.push({
      ...form.value,
      id: `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
    });
  }
  resetForm();
};

const startEdit = (id: string) => {
  const labelToEdit = queue.value.find(label => label.id === id);
  if (labelToEdit) {
    form.value = { ...labelToEdit };
    editingLabelId.value = id;
    isEditing.value = true;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
};

const cancelEdit = () => {
  resetForm();
};

const removeLabel = (id: string) => {
  queue.value = queue.value.filter(label => label.id !== id);
};

function resetForm() {
  form.value = createFreshForm();
  editingLabelId.value = null;
  isEditing.value = false;
}

async function printSingleLabel(id: string) {
  const label = queue.value.find(l => l.id === id);
  if (!label) return;
  try {
    const blob = await generateSingleLabelPdf(label, appSettings.value);
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank');
  } catch (e: any) {
    alert(e?.message || 'Gagal membuat PDF');
  }
}

async function printQueuePdf() {
  if (!queue.value.length) return alert('Antrian kosong.');
  isPrinting.value = true;
  try {
    const blob = await generateLabelsPdf(queue.value, appSettings.value);
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank');
  } catch (e:any) {
    alert(e?.message || 'Gagal membuat PDF');
  } finally {
    isPrinting.value = false;
  }
}
</script>