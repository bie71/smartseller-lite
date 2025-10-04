import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';
import type { UiOrderItem } from '../../modules/order';

type OrderFormItem = UiOrderItem & { unitPriceDisplay: string; discountDisplay: string };

export const useOrderFormStore = defineStore('order-form', () => {
  const form = reactive({
    buyerId: '',
    recipientId: '',
    courier: '',
    serviceLevel: '',
    trackingCode: '',
    shippingCost: 0,
    discount: 0,
    notes: '',
    items: [] as OrderFormItem[]
  });

  const buyerMode = ref<'existing' | 'manual'>('existing');
  const recipientMode = ref<'existing' | 'manual'>('existing');

  const buyerCustom = reactive({
    name: '',
    phone: '',
    email: '',
    address: '',
    city: '',
    province: '',
    postal: ''
  });

  const recipientCustom = reactive({
    name: '',
    phone: '',
    email: '',
    address: '',
    city: '',
    province: '',
    postal: ''
  });

  const isBuyerPayingShipping = ref(false);

  function resetForm() {
    form.buyerId = '';
    form.recipientId = '';
    form.courier = '';
    form.serviceLevel = '';
    form.trackingCode = '';
    form.shippingCost = 0;
    form.discount = 0;
    form.notes = '';
    form.items = [];
    buyerMode.value = 'existing';
    recipientMode.value = 'existing';
    Object.assign(buyerCustom, {
      name: '',
      phone: '',
      email: '',
      address: '',
      city: '',
      province: '',
      postal: ''
    });
    Object.assign(recipientCustom, {
      name: '',
      phone: '',
      email: '',
      address: '',
      city: '',
      province: '',
      postal: ''
    });
    isBuyerPayingShipping.value = false;
  }

  return { form, buyerMode, recipientMode, buyerCustom, recipientCustom, isBuyerPayingShipping, resetForm };
});
