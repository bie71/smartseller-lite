<template>
  <div>
    <svg width="100%" height="40" class="text-black">
      <rect x="0" y="5" width="2" height="30" fill="currentColor" />
      <rect x="4" y="5" width="4" height="30" fill="currentColor" />
      <rect x="10" y="5" width="2" height="30" fill="currentColor" />
      <rect x="14" y="5" width="6" height="30" fill="currentColor" />
      <rect x="22" y="5" width="2" height="30" fill="currentColor" />
      <rect x="28" y="5" width="4" height="30" fill="currentColor" />
      <rect x="34" y="5" width="2" height="30" fill="currentColor" />
      <rect x="38" y="5" width="2" height="30" fill="currentColor" />
      <rect x="44" y="5" width="4" height="30" fill="currentColor" />
      <rect x="52" y="5" width="2" height="30" fill="currentColor" />
      <rect x="56" y="5" width="4" height="30" fill="currentColor" />
      <rect x="64" y="5" width="2" height="30" fill="currentColor" />
      <rect x="70" y="5" width="6" height="30" fill="currentColor" />
      <rect x="78" y="5" width="2" height="30" fill="currentColor" />
      <rect x="82" y="5" width="4" height="30" fill="currentColor" />
      <rect x="90" y="5" width="2" height="30" fill="currentColor" />
      <rect x="94" y="5" width="2" height="30" fill="currentColor" />
      <rect x="100" y="5" width="4" height="30" fill="currentColor" />
      <rect x="106" y="5" width="2" height="30" fill="currentColor" />
      <rect x="110" y="5" width="2" height="30" fill="currentColor" />
      <rect x="116" y="5" width="6" height="30" fill="currentColor" />
      <text x="50%" y="4" text-anchor="middle" font-size="8" font-family="monospace" fill="currentColor">RESI PENGIRIMAN</text>
    </svg>
    <p class="text-center font-mono font-bold tracking-widest text-sm mt-1">{{ trackingId }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ trackingId?: string }>();
</script>
