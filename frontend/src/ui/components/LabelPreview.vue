<template>
  <div :id="`label-preview-${label.id}`" class="bg-white p-4 border border-slate-300 rounded-lg shadow-sm w-full relative font-sans text-xs text-slate-900 leading-tight flex flex-col">
    <div class="absolute top-2 right-2 flex space-x-1 z-10 print-hidden">
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
        <p class="text-slate-600">Berat: {{ label.weight }} kg</p>
      </div>
      <div class="text-right">
        <p class="font-bold text-sm">SmartSeller</p>
        <div v-if="label.isCOD" class="mt-1 px-2 py-1 bg-black text-white font-bold text-sm rounded">
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
          <div>
            <p class="font-semibold text-slate-500">ID PESANAN:</p>
            <p class="font-mono tracking-wider">{{ label.orderCode }}</p>
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
          <p class="font-medium break-words">{{ label.notes }}</p>
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
import { PencilIcon, XMarkIcon } from '@heroicons/vue/24/solid';
import type { LabelData } from '../../modules/label';
import BarcodePlaceholder from './BarcodePlaceholder.vue';

defineProps<{
  label: LabelData;
}>();

defineEmits(['edit', 'remove']);

const formatCurrency = (amount?: number) => {
  if (amount === undefined) return '';
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(amount);
};
</script>
