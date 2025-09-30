<template>
  <section class="space-y-6">
    <div class="grid gap-6 lg:grid-cols-[1.3fr,1fr]">
      <div class="card space-y-5">
        <header class="flex items-center gap-3">
          <div class="h-12 w-12 rounded-full bg-primary/10 text-primary flex items-center justify-center">
            <Cog6ToothIcon class="h-6 w-6" />
          </div>
          <div>
            <h2 class="text-xl font-semibold">Branding Aplikasi</h2>
            <p class="text-sm text-slate-500">Sesuaikan logo dan nama brand yang tampil di dashboard dan label PDF.</p>
          </div>
        </header>

        <form class="space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label class="text-sm font-medium text-slate-600">Nama Brand</label>
            <input v-model="form.brandName" type="text" class="input mt-1" placeholder="SmartSeller Lite" />
          </div>

          <div class="space-y-3">
            <label class="text-sm font-medium text-slate-600">Logo</label>
            <div class="flex flex-col gap-3 md:flex-row md:items-center">
              <div class="h-24 w-24 rounded-xl border border-dashed border-slate-300 bg-slate-50 flex items-center justify-center overflow-hidden">
                <img v-if="logoPreview" :src="logoPreview" alt="Logo preview" class="max-h-20 max-w-20 object-contain" />
                <PhotoIcon v-else class="h-10 w-10 text-slate-400" />
              </div>
              <div class="flex-1 space-y-2">
                <label class="btn-secondary cursor-pointer w-full md:w-auto">
                  <ArrowUpTrayIcon class="h-5 w-5" />
                  <span>Pilih logo</span>
                  <input type="file" accept="image/*" class="hidden" @change="handleFile" />
                </label>
                <p class="text-xs text-slate-500">Direkomendasikan PNG dengan latar transparan, maksimal 1 MB.</p>
                <button v-if="form.logoData || form.logoPath" type="button" class="text-xs text-red-500" @click="clearLogo">Hapus logo</button>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <button type="submit" class="btn-primary">
              <CheckCircleIcon class="h-5 w-5" />
              Simpan Pengaturan
            </button>
          </div>
        </form>
      </div>

      <aside class="card bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white shadow-lg">
        <div class="space-y-4">
          <h3 class="text-lg font-semibold flex items-center gap-2">
            <SparklesIcon class="h-5 w-5" />
            Pratinjau Label
          </h3>
          <p class="text-sm text-white/70">
            Label pengiriman akan menampilkan logo dan nama brand Anda secara otomatis setiap kali mencetak.
          </p>
          <div class="rounded-xl bg-white/10 p-4 space-y-3">
            <div class="flex items-center gap-3">
              <div class="h-14 w-14 bg-white/10 rounded-lg flex items-center justify-center overflow-hidden">
                <img v-if="logoPreview" :src="logoPreview" alt="preview" class="max-h-12 max-w-12 object-contain" />
                <PhotoIcon v-else class="h-6 w-6 text-white/60" />
              </div>
              <div>
                <p class="text-base font-semibold">{{ form.brandName || 'SmartSeller Lite' }}</p>
                <p class="text-xs text-white/60">Contoh label: #INV-00123</p>
              </div>
            </div>
            <div class="text-xs space-y-1 text-white/80">
              <p>Courier: JNE · REG</p>
              <p>Penerima: Budi Santoso</p>
              <p>Alamat: Jl. Melati No. 12, Jakarta</p>
            </div>
          </div>
        </div>
      </aside>
    </div>

    <div class="card space-y-5">
      <header class="flex items-center gap-3">
        <div class="h-12 w-12 rounded-full bg-primary/10 text-primary flex items-center justify-center">
          <ArrowDownTrayIcon class="h-6 w-6" />
        </div>
        <div>
          <h2 class="text-xl font-semibold">Backup &amp; Restore</h2>
          <p class="text-sm text-slate-500">Simpan salinan SQL lengkap dan pulihkan saat berpindah perangkat atau server.</p>
        </div>
      </header>

      <div class="space-y-4">
        <div class="flex flex-col gap-3 md:flex-row md:items-center">
          <button type="button" class="btn-primary" :disabled="isBackingUp" @click="handleBackup">
            <ArrowDownTrayIcon class="h-5 w-5" />
            <span>{{ isBackingUp ? 'Menyiapkan...' : 'Unduh Backup SQL' }}</span>
          </button>
          <label :class="['btn-secondary cursor-pointer w-full md:w-auto', isRestoring ? 'opacity-60 pointer-events-none' : '']">
            <ArrowPathIcon class="h-5 w-5" />
            <span>{{ isRestoring ? 'Memulihkan...' : 'Pilih file backup' }}</span>
            <input
              ref="restoreInput"
              type="file"
              accept=".sql,application/sql,text/plain,.zip,application/zip"
              class="hidden"
              :disabled="isRestoring"
              @change="handleRestoreFile"
            />
          </label>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="rounded-lg border border-slate-200 bg-slate-50/60 p-3">
            <p class="text-xs font-semibold uppercase text-slate-500">Mode Backup SQL</p>
            <div class="mt-2 space-y-2 text-sm">
              <label class="flex items-center gap-2">
                <input
                  type="checkbox"
                  :checked="backupMode === 'all'"
                  @change="backupMode = 'all'"
                />
                <span>Schema + Data (default)</span>
              </label>
              <label class="flex items-center gap-2">
                <input
                  type="checkbox"
                  :checked="backupMode === 'data'"
                  @change="backupMode = 'data'"
                />
                <span>Data saja</span>
              </label>
              <label class="flex items-center gap-2">
                <input
                  type="checkbox"
                  :checked="backupMode === 'schema'"
                  @change="backupMode = 'schema'"
                />
                <span>Schema saja</span>
              </label>
            </div>
          </div>
          <div class="rounded-lg border border-slate-200 bg-slate-50/60 p-3">
            <p class="text-xs font-semibold uppercase text-slate-500">Opsi Restore</p>
            <div class="mt-2 space-y-2 text-sm">
              <label class="flex items-center gap-2">
                <input type="checkbox" v-model="restoreOptions.disableForeignKeyChecks" />
                <span>Nonaktifkan pengecekan foreign key saat restore</span>
              </label>
              <label class="flex items-center gap-2">
                <input type="checkbox" v-model="restoreOptions.useTransaction" />
                <span>Jalankan restore dalam transaksi</span>
              </label>
            </div>
          </div>
        </div>
        <p class="text-xs text-slate-500">
          Pastikan file backup tersimpan aman. Proses restore SQL akan menggantikan seluruh data yang ada pada database.
        </p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import type { AppSettings } from '../../modules/settings';
import { createBackup, getSettings, restoreBackup, updateSettings } from '../../modules/settings';
import type { RestoreResult } from '../../modules/settings';
import { ArrowDownTrayIcon, ArrowPathIcon, ArrowUpTrayIcon, CheckCircleIcon, Cog6ToothIcon, PhotoIcon, SparklesIcon } from '@heroicons/vue/24/outline';
import { useToastStore } from '../stores/toast';

interface Props {
  initialSettings: AppSettings | null;
}

const props = defineProps<Props>();
const emit = defineEmits<{ (e: 'updated', payload: AppSettings): void }>();

const form = reactive<AppSettings>({
  brandName: '',
  logoPath: '',
  logoUrl: '',
  logoHash: '',
  logoWidth: 0,
  logoHeight: 0,
  logoSizeBytes: 0,
  logoMime: '',
  logoData: ''
});

const isBackingUp = ref(false);
const isRestoring = ref(false);
const restoreInput = ref<HTMLInputElement | null>(null);
const backupMode = ref<'all' | 'data' | 'schema'>('all');
const restoreOptions = reactive({
  disableForeignKeyChecks: true,
  useTransaction: true
});

const toast = useToastStore();

watch(
  () => props.initialSettings,
  (value) => {
    if (value) {
      Object.assign(
        form,
        { brandName: '', logoPath: '', logoUrl: '', logoHash: '', logoWidth: 0, logoHeight: 0, logoSizeBytes: 0, logoMime: '', logoData: '' },
        value
      );
    }
  },
  { immediate: true }
);

function resolveMediaPath(path?: string | null) {
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

const logoPreview = computed(() => {
  if (form.logoData) {
    const mime = form.logoMime || 'image/png';
    return `data:${mime};base64,${form.logoData}`;
  }
  if (form.logoUrl) {
    return form.logoUrl;
  }
  if (form.logoPath) {
    const fromPath = resolveMediaPath(form.logoPath);
    if (fromPath) {
      return fromPath;
    }
  }
  return '';
});

function uint8ArrayToBase64(bytes: Uint8Array): string {
  let binary = '';
  const chunkSize = 0x8000;
  for (let i = 0; i < bytes.length; i += chunkSize) {
    const chunk = bytes.subarray(i, Math.min(i + chunkSize, bytes.length));
    let chunkString = '';
    for (let j = 0; j < chunk.length; j += 1) {
      chunkString += String.fromCharCode(chunk[j]);
    }
    binary += chunkString;
  }
  return btoa(binary);
}

function clearLogo() {
  form.logoData = '';
  form.logoPath = '';
  form.logoUrl = '';
  form.logoHash = '';
  form.logoWidth = 0;
  form.logoHeight = 0;
  form.logoSizeBytes = 0;
  form.logoMime = '';
}

function handleFile(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  if (file.size > 1024 * 1024) {
    toast.push('Ukuran file melebihi 1 MB.', 'error');
    return;
  }
  const reader = new FileReader();
  reader.onload = () => {
    const result = reader.result as string;
    const base64 = result.split(',')[1];
    form.logoData = base64;
    form.logoMime = file.type;
    form.logoUrl = '';
    form.logoPath = '';
  };
  reader.readAsDataURL(file);
}

async function handleSubmit() {
  try {
    const saved = await updateSettings({ ...form });
    emit('updated', saved);
    toast.push('Pengaturan tersimpan!', 'success');
  } catch (error) {
    console.error(error);
    toast.push('Gagal menyimpan pengaturan.', 'error');
  }
}

function currentBackupOptions(): { includeSchema: boolean; includeData: boolean } {
  switch (backupMode.value) {
    case 'data':
      return { includeSchema: false, includeData: true };
    case 'schema':
      return { includeSchema: true, includeData: false };
    default:
      return { includeSchema: true, includeData: true };
  }
}

function base64ToBlob(base64: string, mime: string): Blob {
  const clean = base64.replace(/\s+/g, '');
  const binary = atob(clean);
  const buffer = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    buffer[i] = binary.charCodeAt(i);
  }
  return new Blob([buffer], { type: mime });
}

function downloadSqlBackup(base64: string, filename: string) {
  const blob = base64ToBlob(base64, 'application/zip');
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  anchor.style.display = 'none';
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(url);
}

function formatRestoreSummary(summary: RestoreResult): string {
  const statements = summary.statements.toLocaleString('id-ID');
  const duration = `${summary.durationMs.toLocaleString('id-ID')} ms`;
  const driver = summary.executionDriver === 'mysql' ? 'mysql client' : 'engine internal';
  return `Restore selesai · ${statements} statement · ${duration} · via ${driver}`;
}

async function handleBackup() {
  try {
    isBackingUp.value = true;
    const options = currentBackupOptions();
    const base64 = await createBackup({ ...options, includeMedia: true });
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const filename = `smartseller-backup-${timestamp}.zip`;
    downloadSqlBackup(base64, filename);
    toast.push('Backup SQL berhasil diunduh.', 'success');
  } catch (error) {
    console.error(error);
    if (error instanceof Error && error.message) {
      toast.push(`Gagal membuat backup: ${error.message}`, 'error', { timeout: 6000 });
    } else {
      toast.push('Gagal membuat backup.', 'error');
    }
  } finally {
    isBackingUp.value = false;
  }
}

async function handleRestoreFile(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    return;
  }

  isRestoring.value = true;
  try {
    const buffer = await file.arrayBuffer();
    const base64 = uint8ArrayToBase64(new Uint8Array(buffer));
    const summary = await restoreBackup(base64, {
      disableForeignKeyChecks: restoreOptions.disableForeignKeyChecks,
      useTransaction: restoreOptions.useTransaction
    });
    const latest = await getSettings();
    Object.assign(
      form,
      { brandName: '', logoPath: '', logoUrl: '', logoHash: '', logoWidth: 0, logoHeight: 0, logoSizeBytes: 0, logoMime: '', logoData: '' },
      latest
    );
    emit('updated', latest);
    toast.push(formatRestoreSummary(summary), 'success', { timeout: 6000 });
  } catch (error) {
    console.error(error);
    if (error instanceof Error && error.message) {
      toast.push(error.message, 'error', { timeout: 6000 });
    } else {
      toast.push('Gagal memulihkan data.', 'error');
    }
  } finally {
    isRestoring.value = false;
    if (restoreInput.value) {
      restoreInput.value.value = '';
    }
  }
}
</script>
