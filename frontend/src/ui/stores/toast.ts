import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

export type ToastTone = 'success' | 'error' | 'info';

export interface ToastPayload {
  id: number;
  message: string;
  tone: ToastTone;
  timeout: number;
}

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<ToastPayload[]>([]);
  let counter = 0;
  const timers = new Map<number, ReturnType<typeof setTimeout>>();

  function dismiss(id: number) {
    const index = toasts.value.findIndex((toast) => toast.id === id);
    if (index !== -1) {
      toasts.value.splice(index, 1);
    }
    const timer = timers.get(id);
    if (timer) {
      clearTimeout(timer);
      timers.delete(id);
    }
  }

  function push(message: string, tone: ToastTone = 'info', options?: { timeout?: number }) {
    const id = counter + 1;
    counter = id;
    const timeout = Number.isFinite(options?.timeout) ? Number(options?.timeout) : 3500;
    const toast: ToastPayload = { id, message, tone, timeout };
    toasts.value.push(toast);
    if (timeout > 0) {
      const timer = window.setTimeout(() => {
        dismiss(id);
      }, timeout);
      timers.set(id, timer);
    }
    return id;
  }

  const items = computed(() => toasts.value);

  return {
    items,
    push,
    dismiss
  };
});
