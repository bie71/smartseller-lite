import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';

type ProductForm = {
  id?: string;
  name: string;
  sku: string;
  costPrice: number;
  salePrice: number;
  stock: number;
  category: string;
  lowStockThreshold: number;
  description: string;
  imageData?: string;
  imageMime?: string;
  imageUrl?: string;
  thumbUrl?: string;
  imagePath?: string;
  thumbPath?: string;
  imageHash?: string;
  imageWidth?: number;
  imageHeight?: number;
  imageSizeBytes?: number;
  thumbWidth?: number;
  thumbHeight?: number;
  thumbSizeBytes?: number;
};

export const useProductFormStore = defineStore('product-form', () => {
  const form = reactive<ProductForm>({
    id: undefined,
    name: '',
    sku: '',
    costPrice: 0,
    salePrice: 0,
    stock: 0,
    category: '',
    lowStockThreshold: 5,
    description: '',
    imageData: '',
    imageMime: '',
    imageUrl: '',
    thumbUrl: '',
    imagePath: '',
    thumbPath: '',
    imageHash: '',
    imageWidth: 0,
    imageHeight: 0,
    imageSizeBytes: 0,
    thumbWidth: 0,
    thumbHeight: 0,
    thumbSizeBytes: 0
  });

  const editing = ref(false);

  function resetForm() {
    Object.assign(form, {
      id: undefined,
      name: '',
      sku: '',
      costPrice: 0,
      salePrice: 0,
      stock: 0,
      category: '',
      lowStockThreshold: 5,
      description: '',
      imageData: '',
      imageMime: '',
      imageUrl: '',
      thumbUrl: '',
      imagePath: '',
      thumbPath: '',
      imageHash: '',
      imageWidth: 0,
      imageHeight: 0,
      imageSizeBytes: 0,
      thumbWidth: 0,
      thumbHeight: 0,
      thumbSizeBytes: 0
    });
    editing.value = false;
  }

  return { form, editing, resetForm };
});
