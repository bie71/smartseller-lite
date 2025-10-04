<template>
  <div :id="`label-preview-${label.id}`" class="bg-white p-4 border border-slate-300 rounded-lg shadow-sm w-full relative font-sans text-xs text-slate-900 leading-tight flex flex-col">
    <div v-if="!isPrinting" class="absolute top-2 right-2 flex space-x-1 z-10">
      <button
        @click="$emit('print', label.id)"
        class="p-1 bg-white/50 backdrop-blur-sm text-gray-600 hover:bg-gray-100 hover:text-gray-800 rounded-full transition-colors"
        aria-label="Print label"
      >
        <PrinterIcon class="w-4 h-4" />
      </button>
      <button
        @click="$emit('edit', label.id)"
        class="p-1 bg-white/50 backdrop-blur-sm text-blue-500 hover:bg-blue-100 hover:text-blue-700 rounded-full transition-colors"
        aria-label="Edit label"
      >
        <PencilIcon class="w-4 h-4" />
      </button>
      <button
        @click="$emit('remove', label.id)"
        class="p-1 bg-white/50 backdrop-blur-sm text-red-500 hover:bg-red-100 hover:text-red-700 rounded-full transition-colors"
        aria-label="Remove label"
      >
        <XMarkIcon class="w-4 h-4" />
      </button>
    </div>

    <!-- Header -->
    <div class="flex justify-between items-start pb-2 border-b-2 border-black flex-shrink-0">
      <div class="flex-1">
        <p class="font-bold text-lg uppercase">{{ label.courier }}</p>
        <p class="font-mono tracking-wider text-slate-800">{{ label.orderCode }}</p>
        <p class="text-slate-600">Berat: {{ label.weight }} kg</p>
      </div>
      <div class="text-right">
        <div class="flex justify-end items-center gap-2 mb-1">
          <img v-if="logoSrc" :src="logoSrc" alt="Logo" class="max-h-8 w-auto rounded-xl" />
          <p class="font-bold text-sm">{{ appSettings?.brandName || 'SmartSeller' }}</p>
        </div>
        <div v-if="label.isCOD" class="px-2 py-1 bg-black text-white font-bold text-center justify-center text-sm rounded">
          COD: {{ formatCurrency(label.codAmount) }}
        </div>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-grow min-h-0">
      <!-- Addresses -->
      <div class="grid grid-cols-2 gap-2 py-2 border-b border-dashed border-slate-400">
        <div class="space-y-2">
          <div>
            <p class="font-semibold text-slate-500">PENGIRIM:</p>
            <p class="font-bold">{{ label.senderName }}</p>
            <p class="break-words">{{ label.senderPhone }}</p>
          </div>
        </div>
        <div>
          <p class="font-semibold text-slate-500">PENERIMA:</p>
          <p class="font-bold text-base">{{ label.recipientName }}</p>
          <p class="break-words">{{ label.recipientPhone }}</p>
          <p class="mt-1 break-words">{{ label.recipientAddress }}</p>
        </div>
      </div>

      <!-- Order Details -->
      <div class="py-2 space-y-2">
        <div>
          <p class="font-semibold text-slate-500">BARANG:</p>
          <ul class="list-disc list-inside">
            <li v-for="item in label.items" :key="item.productName">{{ item.quantity }}x {{ item.productName }}</li>
          </ul>
          <p v-if="label.notes" class="font-medium break-words mt-2">Catatan: {{ label.notes }}</p>
        </div>
      </div>
    </div>

    <!-- Barcode and Order ID (Footer) -->
    <div class="mt-auto pt-2 border-t border-dashed border-slate-400 flex-shrink-0">
      <BarcodePlaceholder :tracking-id="label.trackingCode" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue';
import { PencilIcon, XMarkIcon, PrinterIcon } from '@heroicons/vue/24/solid';
import type { LabelData } from '../../modules/label';
import type { AppSettings } from '../../modules/settings';
import BarcodePlaceholder from './BarcodePlaceholder.vue';

const props = defineProps<{
  label: LabelData;
  isPrinting?: boolean;
  appSettings?: AppSettings | null;
}>();

defineEmits(['edit', 'remove', 'print']);

const formatCurrency = (amount?: number) => {
  if (amount === undefined) return '';
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(amount);
};

function makeImageSrc(data?: string, mime?: string): string {
  if (!data) return '';
  const type = (mime && mime.length ? mime : 'image/png').trim() || 'image/png';
  return `data:${type};base64,${data}`;
}

const logoSrc = computed(() => {
  const candidate = props.appSettings;
  if (!candidate) return '';
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
});

</script>
