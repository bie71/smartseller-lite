<template>
  <transition name="modal-fade">
    <div v-if="modelValue" class="fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="close"></div>
      <div class="relative z-10 w-full max-w-4xl overflow-hidden rounded-2xl bg-white shadow-2xl" role="dialog" aria-modal="true">
        <header class="flex items-start justify-between gap-4 border-b border-slate-100 bg-slate-50 px-6 py-4">
          <div>
            <h2 class="text-lg font-semibold text-slate-800">{{ title }}</h2>
            <p v-if="subtitle" class="text-sm text-slate-500">{{ subtitle }}</p>
          </div>
          <button type="button" class="icon-btn" @click="close" aria-label="Tutup">
            <XMarkIcon class="h-5 w-5" />
          </button>
        </header>
        <div class="max-h-[70vh] overflow-y-auto px-6 py-4">
          <slot />
        </div>
        <footer v-if="$slots.actions" class="flex items-center justify-end gap-3 border-t border-slate-100 bg-slate-50 px-6 py-4">
          <slot name="actions" />
        </footer>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { onUnmounted, watch } from 'vue';
import { XMarkIcon } from '@heroicons/vue/24/outline';

const props = defineProps<{
  modelValue: boolean;
  title: string;
  subtitle?: string;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
}>();

function close() {
  emit('update:modelValue', false);
}

watch(
  () => props.modelValue,
  (open) => {
    document.body.style.overflow = open ? 'hidden' : '';
  },
  { immediate: true }
);

onUnmounted(() => {
  document.body.style.overflow = '';
});
</script>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.98);
}
</style>
