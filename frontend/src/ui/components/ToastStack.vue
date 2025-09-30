<template>
  <teleport to="body">
    <div class="fixed inset-x-0 top-4 z-[1000] flex flex-col items-center gap-3 px-4 sm:items-end sm:px-6">
      <TransitionGroup name="toast">
        <article
          v-for="toast in toasts"
          :key="toast.id"
          class="w-full max-w-sm rounded-xl border px-4 py-3 shadow-xl backdrop-blur-md"
          :class="toneClass(toast.tone)"
        >
          <div class="flex items-start gap-3">
            <div class="mt-0.5">
              <component :is="iconFor(toast.tone)" class="h-5 w-5" />
            </div>
            <p class="text-sm font-medium leading-5">{{ toast.message }}</p>
            <button type="button" class="ml-auto text-xs text-white/70" @click="dismiss(toast.id)">
              Tutup
            </button>
          </div>
        </article>
      </TransitionGroup>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useToastStore, type ToastTone } from '../stores/toast';
import { CheckCircleIcon, ExclamationCircleIcon, InformationCircleIcon } from '@heroicons/vue/24/solid';

const toastStore = useToastStore();
const toasts = computed(() => toastStore.items);

function toneClass(tone: ToastTone) {
  switch (tone) {
    case 'success':
      return 'border-emerald-400/30 bg-emerald-600/90 text-white';
    case 'error':
      return 'border-rose-400/30 bg-rose-600/90 text-white';
    default:
      return 'border-slate-400/30 bg-slate-800/90 text-white';
  }
}

function iconFor(tone: ToastTone) {
  switch (tone) {
    case 'success':
      return CheckCircleIcon;
    case 'error':
      return ExclamationCircleIcon;
    default:
      return InformationCircleIcon;
  }
}

function dismiss(id: number) {
  toastStore.dismiss(id);
}
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-10px) scale(0.98);
}
</style>
