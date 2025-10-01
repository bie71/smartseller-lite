<template>
  <div class="flex flex-col items-center">
    <p class="font-semibold text-center mb-1">Resi Pengiriman</p>
    <svg ref="barcode" class="w-48 h-12"></svg>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
// @ts-ignore
import JsBarcode from "jsbarcode";

const props = defineProps<{ trackingId?: string }>();
const barcode = ref<SVGSVGElement | null>(null);

onMounted(() => {
  if (barcode.value && props.trackingId) {
    JsBarcode(barcode.value, props.trackingId, {
      format: "CODE128",
      displayValue: true,
      fontSize: 10,
      height: 40,   // tinggi garis barcode (px)
      width: 1,     // ketebalan garis
      margin: 0     // hilangkan margin default
    });
  }
});
</script>
