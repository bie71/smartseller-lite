<template>
  <section class="space-y-6">
    <div class="card space-y-5">
      <header class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex items-center gap-3">
          <div class="h-12 w-12 rounded-full bg-primary/10 text-primary flex items-center justify-center">
            <ClipboardDocumentListIcon class="h-6 w-6" />
          </div>
          <div>
            <h2 class="text-xl font-semibold">Buat Order</h2>
            <p class="text-sm text-slate-500">Ringkas, fokus, dan siap cetak label dalam sekali klik.</p>
          </div>
        </div>
        <div class="flex flex-wrap gap-3 text-sm">
          <div class="stat-chip">
            <BanknotesIcon class="h-4 w-4 text-emerald-600" />
            Subtotal Rp {{ formatCurrency(subtotal) }}
          </div>
          <div class="stat-chip">
            <ChartBarIcon class="h-4 w-4 text-indigo-600" />
            Profit estimasi Rp {{ formatCurrency(estimatedProfit) }}
          </div>
        </div>
      </header>

      <form class="space-y-4" @submit.prevent="submitOrder">
        <div class="grid md:grid-cols-2 gap-4">
          <div>
            <div class="flex items-center justify-between gap-3">
              <label class="text-sm font-medium text-slate-600">Pemesan</label>
              <div class="inline-flex items-center gap-1 rounded-full bg-slate-100 p-1 text-xs font-medium">
                <button
                  type="button"
                  :class="[contactModeButtonClasses, buyerMode === 'existing' ? contactModeActiveClasses : contactModeInactiveClasses]"
                  @click="buyerMode = 'existing'"
                >
                  Daftar Kontak
                </button>
                <button
                  type="button"
                  :class="[contactModeButtonClasses, buyerMode === 'manual' ? contactModeActiveClasses : contactModeInactiveClasses]"
                  @click="buyerMode = 'manual'"
                >
                  Input Manual
                </button>
              </div>
            </div>
            <div v-if="buyerMode === 'existing'" class="mt-2">
              <select v-model="form.buyerId" class="input" :required="buyerMode === 'existing'">
                <option disabled value="">Pilih customer</option>
                <option v-for="customer in customers" :key="customer.id" :value="customer.id">
                  {{ customer.name }} ({{ labelType(customer.type) }})
                </option>
              </select>
              <div v-if="buyer" class="mt-2 space-y-1 text-xs text-slate-500">
                <p v-if="buyer.phone">Telp: {{ buyer.phone }}</p>
                <p v-if="buyer.address">{{ buyer.address }}</p>
              </div>
            </div>
            <div v-else class="mt-3 space-y-3">
              <div>
                <label class="text-xs uppercase text-slate-500">Nama</label>
                <input v-model="buyerCustom.name" type="text" class="input mt-1" placeholder="Nama pemesan" required />
              </div>
              <div class="grid gap-3 sm:grid-cols-2">
                <div>
                  <label class="text-xs uppercase text-slate-500">Telepon</label>
                  <input v-model="buyerCustom.phone" type="tel" class="input mt-1" placeholder="Contoh: 0812..." />
                </div>
                <div>
                  <label class="text-xs uppercase text-slate-500">Email</label>
                  <input v-model="buyerCustom.email" type="email" class="input mt-1" placeholder="opsional" />
                </div>
              </div>
              <div>
                <label class="text-xs uppercase text-slate-500">Alamat</label>
                <textarea v-model="buyerCustom.address" rows="2" class="input mt-1" placeholder="Alamat lengkap"></textarea>
              </div>
              <div class="grid gap-3 sm:grid-cols-3">
                <div>
                  <label class="text-xs uppercase text-slate-500">Kota</label>
                  <input v-model="buyerCustom.city" type="text" class="input mt-1" />
                </div>
                <div>
                  <label class="text-xs uppercase text-slate-500">Provinsi</label>
                  <input v-model="buyerCustom.province" type="text" class="input mt-1" />
                </div>
                <div>
                  <label class="text-xs uppercase text-slate-500">Kode Pos</label>
                  <input v-model="buyerCustom.postal" type="text" class="input mt-1" />
                </div>
              </div>
              <p class="text-xs text-slate-500">Kontak baru akan otomatis tersimpan ke daftar customer.</p>
            </div>
          </div>
          <div>
            <div class="flex items-center justify-between gap-3">
              <label class="text-sm font-medium text-slate-600">Penerima</label>
              <div class="inline-flex items-center gap-1 rounded-full bg-slate-100 p-1 text-xs font-medium">
                <button
                  type="button"
                  :class="[contactModeButtonClasses, recipientMode === 'existing' ? contactModeActiveClasses : contactModeInactiveClasses]"
                  @click="recipientMode = 'existing'"
                >
                  Daftar Kontak
                </button>
                <button
                  type="button"
                  :class="[contactModeButtonClasses, recipientMode === 'manual' ? contactModeActiveClasses : contactModeInactiveClasses]"
                  @click="recipientMode = 'manual'"
                >
                  Input Manual
                </button>
              </div>
            </div>
            <div v-if="recipientMode === 'existing'" class="mt-2">
              <select v-model="form.recipientId" class="input" :required="recipientMode === 'existing'">
                <option disabled value="">Pilih penerima</option>
                <option v-for="customer in customers" :key="customer.id" :value="customer.id">
                  {{ customer.name }} ({{ labelType(customer.type) }})
                </option>
              </select>
              <div v-if="recipient" class="mt-2 space-y-1 text-xs text-slate-500">
                <p v-if="recipient.phone">Telp: {{ recipient.phone }}</p>
                <p v-if="recipient.address">{{ recipient.address }}</p>
              </div>
            </div>
            <div v-else class="mt-3 space-y-3">
              <div>
                <label class="text-xs uppercase text-slate-500">Nama</label>
                <input v-model="recipientCustom.name" type="text" class="input mt-1" placeholder="Nama penerima" required />
              </div>
              <div class="grid gap-3 sm:grid-cols-2">
                <div>
                  <label class="text-xs uppercase text-slate-500">Telepon</label>
                  <input v-model="recipientCustom.phone" type="tel" class="input mt-1" placeholder="Contoh: 0812..." />
                </div>
                <div>
                  <label class="text-xs uppercase text-slate-500">Email</label>
                  <input v-model="recipientCustom.email" type="email" class="input mt-1" placeholder="opsional" />
                </div>
              </div>
              <div>
                <label class="text-xs uppercase text-slate-500">Alamat</label>
                <textarea
                  v-model="recipientCustom.address"
                  rows="3"
                  class="input mt-1"
                  placeholder="Alamat lengkap untuk pengiriman"
                  required
                ></textarea>
              </div>
              <div class="grid gap-3 sm:grid-cols-3">
                <div>
                  <label class="text-xs uppercase text-slate-500">Kota</label>
                  <input v-model="recipientCustom.city" type="text" class="input mt-1" required />
                </div>
                <div>
                  <label class="text-xs uppercase text-slate-500">Provinsi</label>
                  <input v-model="recipientCustom.province" type="text" class="input mt-1" required />
                </div>
                <div>
                  <label class="text-xs uppercase text-slate-500">Kode Pos</label>
                  <input v-model="recipientCustom.postal" type="text" class="input mt-1" required />
                </div>
              </div>
              <p class="text-xs text-slate-500">Kontak baru akan tersimpan dan bisa dipakai ulang di order berikutnya.</p>
            </div>
          </div>
        </div>

        <div class="border border-dashed border-slate-300 rounded-lg p-4 space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="font-semibold flex items-center gap-2">
              <Squares2X2Icon class="h-5 w-5 text-primary" />
              Item Order
            </h3>
            <button type="button" class="btn-ghost" @click="addItem">
              <PlusCircleIcon class="h-5 w-5" />
              Tambah Item
            </button>
          </div>

          <div v-if="!form.items.length" class="text-sm text-slate-500">Tambahkan produk minimal satu item.</div>

          <div v-for="(item, idx) in form.items" :key="idx" class="grid md:grid-cols-5 gap-3 items-end border border-slate-200 rounded-lg p-3 bg-slate-50/60">
            <div class="md:col-span-2">
              <label class="text-xs uppercase text-slate-500">Produk</label>
              <select v-model="item.productId" class="input mt-1" @change="bindProduct(item)" required>
                <option disabled value="">Pilih produk</option>
                <option v-for="product in products" :key="product.id" :value="product.id">
                  {{ product.name }} (stok {{ product.stock }})
                </option>
              </select>
            </div>
            <div>
              <label class="text-xs uppercase text-slate-500">Qty</label>
              <input v-model.number="item.quantity" type="number" min="1" class="input mt-1" required />
            </div>
            <div>
              <label class="text-xs uppercase text-slate-500">Harga</label>
              <input
                v-model="item.unitPriceDisplay"
                type="text"
                inputmode="numeric"
                class="input mt-1 font-mono text-left"
                required
                @input="onUnitPriceInput(item, $event)"
                @blur="syncUnitPriceDisplay(item)"
              />
            </div>
            <div>
              <label class="text-xs uppercase text-slate-500">Diskon Item</label>
              <input
                v-model="item.discountDisplay"
                type="text"
                inputmode="numeric"
                class="input mt-1 font-mono text-left"
                @input="onDiscountInput(item, $event)"
                @blur="syncDiscountDisplay(item)"
              />
            </div>
            <div class="md:col-span-5 flex items-center justify-between text-xs text-slate-500">
              <span>Profit @ {{ formatCurrency(lineProfit(item)) }} | Modal Rp {{ formatCurrency(lineCost(item)) }}</span>
              <button type="button" class="inline-flex items-center gap-1 text-red-500 hover:text-red-600" @click="removeItem(idx)">
                <TrashIcon class="h-4 w-4" />
                Hapus
              </button>
            </div>
          </div>
        </div>

        <div class="grid md:grid-cols-3 gap-4">
          <div>
            <label class="text-sm font-medium text-slate-600">Ekspedisi</label>
            <select v-model="form.courier" class="input mt-1">
              <option v-for="option in courierOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
            <div v-if="selectedCourier" class="mt-2 space-y-1 text-xs text-slate-500">
              <p class="flex items-center gap-1">
                <TruckIcon class="h-4 w-4 text-primary" />
                {{ selectedCourier.services || 'Layanan bebas' }}
              </p>
              <p v-if="selectedCourier.contact" class="flex items-center gap-1">
                <PhoneIcon class="h-4 w-4 text-primary" />
                {{ selectedCourier.contact }}
              </p>
              <p v-if="selectedCourier.trackingUrl" class="flex items-center gap-1">
                <LinkIcon class="h-4 w-4 text-primary" />
                <a :href="selectedCourier.trackingUrl" target="_blank" rel="noreferrer" class="underline">Lacak kiriman</a>
              </p>
            </div>
          </div>
          <div>
            <label class="text-sm font-medium text-slate-600">Layanan</label>
            <input v-model="form.serviceLevel" type="text" class="input mt-1" placeholder="REG, YES, dsb" />
          </div>
          <div>
            <label class="text-sm font-medium text-slate-600">Biaya Kirim</label>
            <input
              v-model="shippingCostDisplay"
              type="text"
              inputmode="numeric"
              class="input mt-1 font-mono text-left"
              @input="onShippingCostInput($event)"
              @blur="syncShippingCostDisplay"
            />
          </div>
        </div>

        <div class="grid md:grid-cols-3 gap-4">
          <div>
            <label class="text-sm font-medium text-slate-600">Nomor Resi</label>
            <input v-model="form.trackingCode" type="text" class="input mt-1" />
          </div>
          <div>
            <label class="text-sm font-medium text-slate-600">Diskon Order</label>
            <input
              v-model="orderDiscountDisplay"
              type="text"
              inputmode="numeric"
              class="input mt-1 font-mono text-left"
              @input="onOrderDiscountInput($event)"
              @blur="syncOrderDiscountDisplay"
            />
          </div>
          <div>
            <label class="text-sm font-medium text-slate-600">Catatan</label>
            <input v-model="form.notes" type="text" class="input mt-1" />
          </div>
        </div>

        <footer class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div class="text-sm text-slate-600 flex items-center gap-2">
            <BanknotesIcon class="h-5 w-5 text-emerald-600" />
            <span>Total bayar Rp {{ formatCurrency(orderTotal) }} · Profit Rp {{ formatCurrency(estimatedProfit) }}</span>
          </div>
          <button type="submit" class="btn-primary" :disabled="!canSubmit || savingOrder">
            <PrinterIcon class="h-5 w-5" />
            {{ savingOrder ? 'Menyimpan...' : 'Simpan & Buka Cetak' }}
          </button>
        </footer>
      </form>
    </div>

    <div class="card space-y-5">
      <header class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div class="flex items-center gap-2">
          <ArrowPathIcon class="h-5 w-5 text-primary" />
          <h3 class="text-lg font-semibold">Histori Order</h3>
        </div>
        <div class="flex w-full flex-col gap-3 lg:w-auto">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
            <div class="relative w-full sm:w-64">
              <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
              <input
                v-model="orderSearch"
                type="search"
                class="input pl-10"
                placeholder="Cari kode order, ekspedisi, produk, atau customer"
              />
            </div>
            <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <input v-model="orderDateStart" type="date" class="input" aria-label="Tanggal mulai order" />
              <input v-model="orderDateEnd" type="date" class="input" aria-label="Tanggal akhir order" />
            </div>
          </div>
          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end sm:gap-3">
            <select v-model="orderCourierFilter" class="input sm:w-48" aria-label="Filter ekspedisi">
              <option v-for="option in courierFilterOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
            <button
              v-if="hasOrderFilters"
              type="button"
              class="btn-ghost text-xs"
              @click="resetOrderFilters"
            >
              Bersihkan filter
            </button>
            <button class="btn-secondary text-xs" @click="exportOrders">
              <ArrowDownTrayIcon class="h-4 w-4" />
              Export CSV
            </button>
            <button class="btn-secondary text-xs" @click="loadOrders">
              <ArrowPathIcon class="h-4 w-4" />
              Refresh
            </button>
          </div>
        </div>
      </header>
      <div v-if="orderFilterChips.length" class="flex flex-wrap gap-2 text-xs text-slate-500">
        <span
          v-for="chip in orderFilterChips"
          :key="chip"
          class="inline-flex items-center gap-2 rounded-full bg-slate-100 px-3 py-1"
        >
          {{ chip }}
        </span>
      </div>
      <div v-if="filteredOrders.length" class="flex flex-wrap gap-2 text-xs text-slate-600 sm:text-sm">
        <div class="stat-chip">
          <ClipboardDocumentListIcon class="h-4 w-4 text-primary" />
          {{ orderInsights.count }} order
        </div>
        <div class="stat-chip">
          <BanknotesIcon class="h-4 w-4 text-emerald-600" />
          Rp {{ formatCurrency(orderInsights.revenue) }} omset
        </div>
        <div class="stat-chip">
          <ChartBarIcon class="h-4 w-4 text-indigo-600" />
          Profit Rp {{ formatCurrency(orderInsights.profit) }}
        </div>
        <div class="stat-chip">
          <TruckIcon class="h-4 w-4 text-orange-500" />
          Favorit: {{ orderInsights.topCourier }}
        </div>
        <div class="stat-chip">
          <Squares2X2Icon class="h-4 w-4 text-rose-500" />
          Produk terlaris: {{ orderInsights.topProduct }}
        </div>
      </div>
      <div v-if="!orders.length" class="text-sm text-slate-500">Belum ada order.</div>
      <div v-else-if="!filteredOrders.length" class="text-sm text-slate-500">Tidak ada order yang cocok dengan pencarian.</div>
      <div v-for="order in filteredOrders" :key="order.id" class="border border-slate-200 rounded-lg p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="font-semibold">{{ order.code }}</h4>
            <p class="text-xs text-slate-500">
              {{ formatDate(order.createdAt) }} · {{ order.shipment.courier }} ({{ order.shipment.serviceLevel || 'N/A' }})
            </p>
            <p class="text-xs text-slate-400" v-if="orderParticipants(order)">
              {{ orderParticipants(order) }}
            </p>
          </div>
          <div class="flex flex-col gap-2 sm:flex-row">
            <button
              type="button"
              class="btn-secondary text-xs"
              @click="openOrderDetail(order)"
            >
              <InformationCircleIcon class="h-4 w-4" />
              Detail
            </button>
            <button
              type="button"
              class="btn-ghost text-xs"
              @click="loadOrderIntoForm(order)"
            >
              <ArrowUturnLeftIcon class="h-4 w-4" />
              Muat ke Form
            </button>
            <button
              class="btn-secondary text-xs"
              :disabled="labelBusy === order.id"
              @click="printLabel(order)"
            >
              <PrinterIcon class="h-4 w-4" />
              {{ labelBusy === order.id ? 'Menyiapkan...' : 'Label PDF' }}
            </button>
          </div>
        </div>
        <ul class="text-sm text-slate-600 list-disc list-inside">
          <li v-for="item in order.items" :key="item.id">
            {{ productName(item.productId) }} × {{ item.quantity }} @ Rp {{ formatCurrency(item.unitPrice) }}
          </li>
        </ul>
        <div class="text-sm text-slate-500">
          Total Rp {{ formatCurrency(order.total) }} · Profit Rp {{ formatCurrency(order.profit) }}
        </div>
      </div>
    </div>
    <BaseModal v-model="orderDetailOpen" title="Detail Order">
      <div v-if="activeOrder" class="space-y-4 text-sm text-slate-600">
        <div class="grid gap-2 sm:grid-cols-2">
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Kode Order</p>
            <p class="text-base font-semibold text-slate-800">{{ activeOrder.code }}</p>
          </div>
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Tanggal</p>
            <p>{{ formatDate(activeOrder.createdAt) }}</p>
          </div>
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Pemesan</p>
            <p>{{ activeOrderBuyer?.name || '-' }}</p>
            <p v-if="activeOrderBuyer?.address" class="text-xs text-slate-500">{{ activeOrderBuyer.address }}</p>
          </div>
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Penerima</p>
            <p>{{ activeOrderRecipient?.name || '-' }}</p>
            <p v-if="activeOrderRecipient?.address" class="text-xs text-slate-500">{{ activeOrderRecipient.address }}</p>
          </div>
        </div>
        <div class="space-y-2">
          <p class="text-xs font-semibold uppercase text-slate-400">Item</p>
          <ul class="space-y-1">
            <li v-for="item in activeOrder.items" :key="item.id" class="flex justify-between gap-4">
              <span>{{ productName(item.productId) }} × {{ item.quantity }}</span>
              <span class="font-mono">Rp {{ formatCurrency(Math.max(item.unitPrice - item.discount, 0)) }}</span>
            </li>
          </ul>
        </div>
        <div class="grid gap-2 sm:grid-cols-2">
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Ekspedisi</p>
            <p>
              {{ activeOrder.shipment.courier || '-' }}
              <span v-if="activeOrder.shipment.serviceLevel">· {{ activeOrder.shipment.serviceLevel }}</span>
            </p>
            <p v-if="activeOrder.shipment.trackingCode" class="text-xs text-slate-500">
              Resi: {{ activeOrder.shipment.trackingCode }}
            </p>
          </div>
          <div>
            <p class="text-xs font-semibold uppercase text-slate-400">Pembayaran</p>
            <p>Subtotal Rp {{ formatCurrency(orderSubtotal(activeOrder)) }}</p>
            <p>Diskon Rp {{ formatCurrency(activeOrder.discount) }}</p>
            <p>Ongkir Rp {{ formatCurrency(activeOrder.shipment.shippingCost) }}</p>
            <p class="font-semibold text-slate-800">Total Rp {{ formatCurrency(activeOrder.total) }}</p>
          </div>
        </div>
        <div v-if="activeOrder.notes" class="rounded-lg bg-slate-50 p-3 text-xs text-slate-500">
          {{ activeOrder.notes }}
        </div>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="orderDetailOpen = false">Tutup</button>
      </template>
    </BaseModal>

    <BaseModal v-model="labelPreviewOpen" :title="labelPreviewTitle" :subtitle="labelPreviewSubtitle">
      <div class="space-y-3">
        <div v-if="labelPreviewUrl" class="aspect-[3/4] w-full overflow-hidden rounded-xl border border-slate-200 shadow-inner">
          <iframe :src="labelPreviewUrl" title="Pratinjau label" class="h-full w-full"></iframe>
        </div>
        <p v-else class="text-sm text-slate-500">Label sedang disiapkan. Silakan tunggu...</p>
      </div>
      <template #actions>
        <button type="button" class="btn-secondary" @click="labelPreviewOpen = false">Tutup</button>
        <button type="button" class="btn-primary" :disabled="!labelPreview" @click="downloadLabelPreview">
          <PrinterIcon class="h-4 w-4" />
          Unduh PDF
        </button>
      </template>
    </BaseModal>
  
    <!-- Panel Cetak Label -->
    <transition name="fade">
      <div v-if="showLabelPanel" class="fixed inset-0 z-50 bg-black/40 flex justify-center items-start p-4 overflow-y-auto h-screen w-screen">
        <div class="bg-white w-full max-w-7xl rounded-2xl shadow-lg my-auto flex flex-col max-h-[90vh]">
          <div class="flex items-center justify-between p-4 border-b flex-shrink-0">
            <div class="flex items-center gap-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-primary" viewBox="0 0 20 20" fill="currentColor"><path d="M6 2a2 2 0 00-2 2v3h12V4a2 2 0 00-2-2H6z" /><path fill-rule="evenodd" d="M4 9a2 2 0 012-2h8a2 2 0 012 2v6h-3v3H7v-3H4V9zm6 6a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd" /></svg>
              <h3 class="text-lg font-semibold">Cetak Label</h3>
            </div>
            <button class="text-slate-500 hover:text-slate-700" @click="showLabelPanel=false">Tutup</button>
          </div>
          <div class="p-4 overflow-y-auto">
            <LabelForm :auto-data="autoLabelData" />
          </div>
        </div>
      </div>
    </transition>
  </section>

</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { listProducts } from '../../modules/product';
import type { Product } from '../../modules/product';
import { listCustomers, saveCustomer } from '../../modules/customer';
import type { Customer, CustomerType } from '../../modules/customer';
import { createOrder, downloadLabel, ensureLabelBlob, generateLabel, listOrders, type Order, type UiOrderItem } from '../../modules/order';
import { fetchOrdersCsv } from '../../modules/reports';
import { listCouriers, type Courier } from '../../modules/settings';
import BaseModal from '../components/BaseModal.vue';
import LabelForm from '../components/LabelForm.vue';
import { useToastStore } from '../stores/toast';
import {
  ArrowDownTrayIcon,
  ArrowPathIcon,
  ArrowUturnLeftIcon,
  BanknotesIcon,
  ChartBarIcon,
  ClipboardDocumentListIcon,
  InformationCircleIcon,
  LinkIcon,
  MagnifyingGlassIcon,
  PhoneIcon,
  PlusCircleIcon,
  PrinterIcon,
  Squares2X2Icon,
  TrashIcon,
  TruckIcon
} from '@heroicons/vue/24/outline';

type OrderFormItem = UiOrderItem & { unitPriceDisplay: string; discountDisplay: string };

const customers = ref<Customer[]>([]);
const products = ref<Product[]>([]);
const orders = ref<Order[]>([]);
const couriers = ref<Courier[]>([]);
const orderSearch = ref('');
const orderDateStart = ref('');
const orderDateEnd = ref('');
const orderCourierFilter = ref('all');
const savingOrder = ref(false);
const labelBusy = ref<string | null>(null);
const orderDetailOpen = ref(false);
const activeOrder = ref<Order | null>(null);
const labelPreviewOpen = ref(false);
const labelPreview = ref<{ order: Order; base64: string } | null>(null);
const labelPreviewUrl = ref('');
const autoLabelData = ref<any>(null);
const toast = useToastStore();

const shippingCostDisplay = ref('');
const orderDiscountDisplay = ref('');
const contactModeButtonClasses =
  'inline-flex items-center justify-center rounded-full px-3 py-1 transition focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/60';
const contactModeActiveClasses = 'bg-white text-primary shadow-sm';
const contactModeInactiveClasses = 'text-slate-500 hover:text-primary';

type ContactMode = 'existing' | 'manual';

interface ManualContact {
  name: string;
  phone: string;
  email: string;
  address: string;
  city: string;
  province: string;
  postal: string;
}

function createEmptyManualContact(): ManualContact {
  return {
    name: '',
    phone: '',
    email: '',
    address: '',
    city: '',
    province: '',
    postal: ''
  };
}

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

const buyerMode = ref<ContactMode>('existing');
const recipientMode = ref<ContactMode>('existing');
const buyerCustom = reactive(createEmptyManualContact());
const recipientCustom = reactive(createEmptyManualContact());

const customerMap = computed(() => {
  const map = new Map<string, Customer>();
  customers.value.forEach((customer) => {
    if (customer.id) {
      map.set(customer.id, customer);
    }
  });
  return map;
});

const productMap = computed(() => {
  const map = new Map<string, Product>();
  products.value.forEach((product) => {
    if (product.id) {
      map.set(product.id, product);
    }
  });
  return map;
});

const priceFormatter = new Intl.NumberFormat('id-ID');

function normalisePrice(value: number): { value: number; display: string } {
  const numeric = Number.isFinite(value) ? Math.max(0, Math.round(value)) : 0;
  return { value: numeric, display: numeric > 0 ? priceFormatter.format(numeric) : '' };
}

function parsePriceInputValue(raw: string): { value: number; display: string } {
  const cleaned = raw.replace(/[^0-9]/g, '');
  if (!cleaned) {
    return { value: 0, display: '' };
  }
  const numeric = Number(cleaned);
  return { value: numeric, display: numeric > 0 ? priceFormatter.format(numeric) : '' };
}

watch(
  () => form.shippingCost,
  (value) => {
    const formatted = normalisePrice(value).display;
    if (shippingCostDisplay.value !== formatted) {
      shippingCostDisplay.value = formatted;
    }
  },
  { immediate: true }
);

watch(
  () => form.discount,
  (value) => {
    const formatted = normalisePrice(value).display;
    if (orderDiscountDisplay.value !== formatted) {
      orderDiscountDisplay.value = formatted;
    }
  },
  { immediate: true }
);

const buyer = computed(() => (form.buyerId ? customerMap.value.get(form.buyerId) : undefined));
const recipient = computed(() => (form.recipientId ? customerMap.value.get(form.recipientId) : undefined));

function resetManualContact(target: ManualContact) {
  target.name = '';
  target.phone = '';
  target.email = '';
  target.address = '';
  target.city = '';
  target.province = '';
  target.postal = '';
}

function manualContactToPayload(manual: ManualContact): Customer {
  return {
    type: 'customer',
    name: manual.name.trim(),
    phone: manual.phone.trim() || undefined,
    email: manual.email.trim() || undefined,
    address: manual.address.trim() || undefined,
    city: manual.city.trim() || undefined,
    province: manual.province.trim() || undefined,
    postal: manual.postal.trim() || undefined
  };
}

function upsertCustomer(record: Customer) {
  if (!record.id) {
    return;
  }
  const next = [record, ...customers.value.filter((customer) => customer.id !== record.id)];
  next.sort((a, b) => a.name.localeCompare(b.name, 'id-ID'));
  customers.value = next;
}

const subtotal = computed(() =>
  form.items.reduce((total, item) => {
    const revenue = item.unitPrice * item.quantity - item.discount;
    return total + Math.max(revenue, 0);
  }, 0)
);

const totalCost = computed(() =>
  form.items.reduce((total, item) => {
    const costPrice = item.product?.costPrice ?? productMap.value.get(item.productId)?.costPrice ?? 0;
    return total + costPrice * item.quantity;
  }, 0)
);

const orderTotal = computed(() => subtotal.value - form.discount + form.shippingCost);

const estimatedProfit = computed(() => {
  const value = subtotal.value - form.discount - totalCost.value - form.shippingCost;
  return value < 0 ? 0 : value;
});

const canSubmit = computed(() => {
  const hasItems =
    form.items.length > 0 &&
    form.items.every((item) => item.productId && item.quantity > 0 && Number.isFinite(item.unitPrice));
  const buyerReady =
    buyerMode.value === 'existing'
      ? !!form.buyerId
      : buyerCustom.name.trim().length > 0;
  const recipientReady =
    recipientMode.value === 'existing'
      ? !!form.recipientId
      : recipientCustom.name.trim().length > 0 &&
        recipientCustom.address.trim().length > 0 &&
        recipientCustom.city.trim().length > 0 &&
        recipientCustom.province.trim().length > 0 &&
        recipientCustom.postal.trim().length > 0;
  return hasItems && buyerReady && recipientReady;
});

const courierOptions = computed(() =>
  couriers.value.map((courier) => ({
    value: courier.code || courier.name,
    label: courier.name ? `${courier.code} · ${courier.name}` : courier.code,
    courier
  }))
);

const selectedCourier = computed(() =>
  courierOptions.value.find((item) => item.value === form.courier)?.courier
);

const courierFilterOptions = computed(() => {
  const seen = new Set<string>();
  const options: Array<{ value: string; label: string }> = [{ value: 'all', label: 'Semua ekspedisi' }];
  orders.value.forEach((order) => {
    const value = order.shipment.courier || 'Lainnya';
    if (!seen.has(value)) {
      seen.add(value);
      options.push({ value, label: value });
    }
  });
  courierOptions.value.forEach((option) => {
    const value = option.value;
    if (!seen.has(value)) {
      seen.add(value);
      options.push({ value, label: option.label });
    }
  });
  return options;
});

const filteredOrders = computed(() => {
  let list = orders.value;
  if (orderCourierFilter.value !== 'all') {
    list = list.filter((order) => (order.shipment.courier || 'Lainnya') === orderCourierFilter.value);
  }
  if (orderDateStart.value) {
    const start = new Date(`${orderDateStart.value}T00:00:00`);
    list = list.filter((order) => new Date(order.createdAt) >= start);
  }
  if (orderDateEnd.value) {
    const end = new Date(`${orderDateEnd.value}T23:59:59`);
    list = list.filter((order) => new Date(order.createdAt) <= end);
  }
  const query = orderSearch.value.trim().toLowerCase();
  if (!query) {
    return list;
  }
  return list.filter((order) => {
    const candidateFields = [
      order.code,
      order.shipment.courier,
      order.shipment.serviceLevel,
      order.shipment.trackingCode,
      order.notes,
      customerMap.value.get(order.buyerId)?.name,
      customerMap.value.get(order.recipientId)?.name
    ];
    const productMatches = order.items.some((item) => productName(item.productId).toLowerCase().includes(query));
    return (
      candidateFields
        .filter(Boolean)
        .some((field) => (field as string).toLowerCase().includes(query)) || productMatches
    );
  });
});

const hasOrderFilters = computed(
  () =>
    orderCourierFilter.value !== 'all' ||
    !!orderDateStart.value ||
    !!orderDateEnd.value ||
    !!orderSearch.value.trim()
);

const orderFilterChips = computed(() => {
  const chips: string[] = [];
  if (orderSearch.value.trim()) {
    chips.push(`Kata kunci: “${orderSearch.value.trim()}”`);
  }
  if (orderDateStart.value && orderDateEnd.value) {
    chips.push(`Rentang: ${formatDateOnly(orderDateStart.value)} – ${formatDateOnly(orderDateEnd.value)}`);
  } else if (orderDateStart.value) {
    chips.push(`Sejak ${formatDateOnly(orderDateStart.value)}`);
  } else if (orderDateEnd.value) {
    chips.push(`Hingga ${formatDateOnly(orderDateEnd.value)}`);
  }
  if (orderCourierFilter.value !== 'all') {
    const label =
      courierFilterOptions.value.find((option) => option.value === orderCourierFilter.value)?.label ||
      orderCourierFilter.value;
    chips.push(`Ekspedisi: ${label}`);
  }
  return chips;
});

const orderInsights = computed(() => {
  const list = filteredOrders.value;
  if (!list.length) {
    return { count: 0, revenue: 0, profit: 0, topCourier: '-', topProduct: '-' };
  }
  const revenue = list.reduce((sum, order) => sum + order.total, 0);
  const profit = list.reduce((sum, order) => sum + order.profit, 0);
  const courierCounter = new Map<string, number>();
  const productCounter = new Map<string, number>();
  list.forEach((order) => {
    const courierName = order.shipment.courier || 'Lainnya';
    courierCounter.set(courierName, (courierCounter.get(courierName) || 0) + 1);
    order.items.forEach((item) => {
      productCounter.set(item.productId, (productCounter.get(item.productId) || 0) + item.quantity);
    });
  });
  const topCourierEntry = Array.from(courierCounter.entries()).sort((a, b) => b[1] - a[1])[0];
  const topProductEntry = Array.from(productCounter.entries()).sort((a, b) => b[1] - a[1])[0];
  return {
    count: list.length,
    revenue,
    profit,
    topCourier: topCourierEntry ? `${topCourierEntry[0]} (${topCourierEntry[1]})` : '-',
    topProduct: topProductEntry ? `${productName(topProductEntry[0])} (${topProductEntry[1]})` : '-'
  };
});

const activeOrderBuyer = computed(() => (activeOrder.value ? customerMap.value.get(activeOrder.value.buyerId) : undefined));
const activeOrderRecipient = computed(() =>
  activeOrder.value ? customerMap.value.get(activeOrder.value.recipientId) : undefined
);

const labelPreviewTitle = 'Pratinjau Label';
const labelPreviewSubtitle = computed(() => (labelPreview.value ? `Order ${labelPreview.value.order.code}` : ''));

function labelType(type: CustomerType | undefined) {
  if (!type) return 'Customer';
  switch (type) {
    case 'marketer':
      return 'Marketer';
    case 'reseller':
      return 'Reseller';
    default:
      return 'Customer';
  }
}

function addItem() {
  form.items.push({
    productId: '',
    quantity: 1,
    unitPrice: 0,
    discount: 0,
    unitPriceDisplay: '',
    discountDisplay: ''
  });
}

function removeItem(index: number) {
  form.items.splice(index, 1);
}

function applyUnitPrice(item: OrderFormItem, value: number) {
  const normalised = normalisePrice(value);
  item.unitPrice = normalised.value;
  item.unitPriceDisplay = normalised.display;
}

function applyDiscount(item: OrderFormItem, value: number) {
  const normalised = normalisePrice(value);
  item.discount = normalised.value;
  item.discountDisplay = normalised.display;
}

function bindProduct(item: OrderFormItem) {
  const product = productMap.value.get(item.productId);
  if (product) {
    item.product = product;
    applyUnitPrice(item, product.salePrice);
    applyDiscount(item, item.discount);
  }
}

function onUnitPriceInput(item: OrderFormItem, event: Event) {
  const target = event.target as HTMLInputElement;
  const { value, display } = parsePriceInputValue(target.value);
  if (item.unitPrice !== value) {
    item.unitPrice = value;
  }
  item.unitPriceDisplay = display;
}

function syncUnitPriceDisplay(item: OrderFormItem) {
  const { display } = parsePriceInputValue(item.unitPriceDisplay);
  item.unitPriceDisplay = display;
}

function onDiscountInput(item: OrderFormItem, event: Event) {
  const target = event.target as HTMLInputElement;
  const { value, display } = parsePriceInputValue(target.value);
  if (item.discount !== value) {
    item.discount = value;
  }
  item.discountDisplay = display;
}

function syncDiscountDisplay(item: OrderFormItem) {
  const { value, display } = parsePriceInputValue(item.discountDisplay);
  item.discount = value;
  item.discountDisplay = display;
}

function onShippingCostInput(event: Event) {
  const target = event.target as HTMLInputElement;
  const { value, display } = parsePriceInputValue(target.value);
  if (form.shippingCost !== value) {
    form.shippingCost = value;
  }
  shippingCostDisplay.value = display;
}

function syncShippingCostDisplay() {
  const { value, display } = parsePriceInputValue(shippingCostDisplay.value);
  form.shippingCost = value;
  shippingCostDisplay.value = display;
}

function onOrderDiscountInput(event: Event) {
  const target = event.target as HTMLInputElement;
  const { value, display } = parsePriceInputValue(target.value);
  if (form.discount !== value) {
    form.discount = value;
  }
  orderDiscountDisplay.value = display;
}

function syncOrderDiscountDisplay() {
  const { value, display } = parsePriceInputValue(orderDiscountDisplay.value);
  form.discount = value;
  orderDiscountDisplay.value = display;
}

function lineRevenue(item: OrderFormItem) {
  return Math.max(item.unitPrice * item.quantity - item.discount, 0);
}

function lineCost(item: OrderFormItem) {
  const costPrice = item.product?.costPrice ?? productMap.value.get(item.productId)?.costPrice ?? 0;
  return costPrice * item.quantity;
}

function lineProfit(item: OrderFormItem) {
  const value = lineRevenue(item) - lineCost(item);
  return value < 0 ? 0 : value;
}

function orderSubtotal(order: Order | null | undefined) {
  if (!order) {
    return 0;
  }
  return order.items.reduce((sum, item) => sum + Math.max(item.unitPrice * item.quantity - item.discount, 0), 0);
}

function formatCurrency(value: number) {
  return Number(value || 0).toLocaleString('id-ID');
}

function formatDate(value: string) {
  try {
    return new Intl.DateTimeFormat('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  } catch (error) {
    return value;
  }
}

function productName(productId: string) {
  return productMap.value.get(productId)?.name ?? productId;
}

function customerName(customerId: string) {
  return customerMap.value.get(customerId)?.name ?? '';
}

function orderParticipants(order: Order) {
  const buyerName = customerName(order.buyerId);
  const recipientName = customerName(order.recipientId);
  if (buyerName && recipientName) {
    return `${buyerName} → ${recipientName}`;
  }
  if (buyerName) {
    return `Pembeli: ${buyerName}`;
  }
  if (recipientName) {
    return `Penerima: ${recipientName}`;
  }
  return '';
}

function formatDateOnly(value: string) {
  try {
    return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(new Date(`${value}T00:00:00`));
  } catch (error) {
    return value;
  }
}

async function resolveBuyerId(): Promise<string | null> {
  if (buyerMode.value === 'existing') {
    if (!form.buyerId) {
      toast.push('Pilih pemesan dari daftar atau isi manual.', 'error');
      return null;
    }
    return form.buyerId;
  }
  const name = buyerCustom.name.trim();
  if (!name) {
    toast.push('Nama pemesan wajib diisi untuk kontak baru.', 'error');
    return null;
  }
  try {
    const payload = manualContactToPayload(buyerCustom);
    payload.name = name;
    const created = await saveCustomer(payload);
    if (!created.id) {
      toast.push('Kontak pemesan baru tidak valid.', 'error');
      return null;
    }
    upsertCustomer(created);
    form.buyerId = created.id;
    buyerMode.value = 'existing';
    resetManualContact(buyerCustom);
    toast.push(`Pemesan ${created.name} disimpan ke daftar kontak.`, 'success');
    return created.id;
  } catch (error) {
    console.error(error);
    const message = error instanceof Error && error.message ? error.message : 'Gagal menyimpan kontak pemesan baru.';
    toast.push(message, 'error', { timeout: 5000 });
    return null;
  }
}

async function resolveRecipientId(): Promise<string | null> {
  if (recipientMode.value === 'existing') {
    if (!form.recipientId) {
      toast.push('Pilih penerima dari daftar atau isi manual.', 'error');
      return null;
    }
    return form.recipientId;
  }
  const name = recipientCustom.name.trim();
  if (!name) {
    toast.push('Nama penerima wajib diisi untuk kontak baru.', 'error');
    return null;
  }
  if (!recipientCustom.address.trim() || !recipientCustom.city.trim() || !recipientCustom.province.trim() || !recipientCustom.postal.trim()) {
    toast.push('Lengkapi alamat penerima sebelum menyimpan.', 'error');
    return null;
  }
  try {
    const payload = manualContactToPayload(recipientCustom);
    payload.name = name;
    const created = await saveCustomer(payload);
    if (!created.id) {
      toast.push('Kontak penerima baru tidak valid.', 'error');
      return null;
    }
    upsertCustomer(created);
    form.recipientId = created.id;
    recipientMode.value = 'existing';
    resetManualContact(recipientCustom);
    toast.push(`Penerima ${created.name} disimpan ke daftar kontak.`, 'success');
    return created.id;
  } catch (error) {
    console.error(error);
    const message = error instanceof Error && error.message ? error.message : 'Gagal menyimpan kontak penerima baru.';
    toast.push(message, 'error', { timeout: 5000 });
    return null;
  }
}

function openOrderDetail(order: Order) {
  activeOrder.value = order;
  orderDetailOpen.value = true;
}

function revokeLabelPreviewUrl() {
  if (labelPreviewUrl.value) {
    URL.revokeObjectURL(labelPreviewUrl.value);
    labelPreviewUrl.value = '';
  }
}

function showLabelPreview(order: Order, base64: string) {
  try {
    const blob = ensureLabelBlob(base64);
    revokeLabelPreviewUrl();
    labelPreview.value = { order, base64 };
    labelPreviewUrl.value = URL.createObjectURL(blob);
    labelPreviewOpen.value = true;
  } catch (error) {
    console.error(error);
    toast.push('Label tidak dapat ditampilkan.', 'error');
  }
}

function downloadLabelPreview() {
  if (!labelPreview.value) {
    return;
  }
  downloadLabel(labelPreview.value.order, labelPreview.value.base64);
  toast.push(`Label ${labelPreview.value.order.code} diunduh.`, 'success');
}

async function loadInitial() {
  try {
    const [cust, prod, logistic] = await Promise.all([listCustomers(), listProducts(), listCouriers()]);
    customers.value = cust;
    products.value = prod;
    couriers.value = logistic;
    if (!form.courier && logistic.length) {
      form.courier = logistic[0].code || logistic[0].name;
    }
    if (!form.items.length) addItem();
  } catch (error) {
    console.error(error);
    toast.push('Gagal memuat data awal order.', 'error');
  }
}

async function loadOrders() {
  try {
    orders.value = await listOrders(50);
  } catch (error) {
    console.error(error);
    toast.push('Gagal memuat histori order.', 'error');
  }
}

async function exportOrders() {
  try {
    const blob = await fetchOrdersCsv({
      search: orderSearch.value.trim() || undefined,
      start: orderDateStart.value || undefined,
      end: orderDateEnd.value || undefined,
      courier: orderCourierFilter.value !== 'all' ? orderCourierFilter.value : undefined
    });
    const url = URL.createObjectURL(blob);
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `orders-${timestamp}.csv`;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
    toast.push('CSV laporan order berhasil diunduh.', 'success', { timeout: 4000 });
  } catch (error) {
    console.error(error);
    const message = error instanceof Error && error.message ? error.message : 'Gagal mengekspor CSV.';
    toast.push(message, 'error', { timeout: 5000 });
  }
}

function resetForm() {
  form.buyerId = '';
  form.recipientId = '';
  form.courier = couriers.value[0] ? couriers.value[0].code || couriers.value[0].name : '';
  form.serviceLevel = '';
  form.trackingCode = '';
  form.shippingCost = 0;
  form.discount = 0;
  form.notes = '';
  form.items = [] as OrderFormItem[];
  addItem();
  buyerMode.value = 'existing';
  recipientMode.value = 'existing';
  resetManualContact(buyerCustom);
  resetManualContact(recipientCustom);
}

function resetOrderFilters() {
  orderSearch.value = '';
  orderDateStart.value = '';
  orderDateEnd.value = '';
  orderCourierFilter.value = 'all';
}

async function submitOrder() {
  if (!canSubmit.value || savingOrder.value) return;
  savingOrder.value = true;
  let createdOrder: Order | null = null;
  try {
    const buyerId = await resolveBuyerId();
    if (!buyerId) {
      return;
    }
    const recipientId = await resolveRecipientId();
    if (!recipientId) {
      return;
    }
    const payload = {
      buyerId,
      recipientId,
      courier: form.courier,
      serviceLevel: form.serviceLevel,
      trackingCode: form.trackingCode,
      shippingCost: form.shippingCost,
      discount: form.discount,
      notes: form.notes,
      items: form.items.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        discount: item.discount
      }))
    };
    createdOrder = await createOrder(payload);
    orders.value = [createdOrder, ...orders.value];
    resetForm();
  } catch (error) {
    console.error(error);
    toast.push('Gagal menyimpan order. Mohon periksa data yang diisi.', 'error', { timeout: 5000 });
  } finally {
    savingOrder.value = false;
  }

  if (!createdOrder) {
    return;
  }

  try {
    labelBusy.value = createdOrder.id;
    const pdf = await generateLabel(createdOrder.id);
    showLabelPreview(createdOrder, pdf);
    toast.push('Order tersimpan. Pratinjau label siap diunduh.', 'success', { timeout: 5000 });
  } catch (error) {
    console.error(error);
    toast.push(
      'Order tersimpan, tetapi label gagal dibuat. Gunakan tombol Label PDF di histori order.',
      'info',
      { timeout: 6000 }
    );
  } finally {
    labelBusy.value = null;
    if (createdOrder) { showLabelPanel.value = true; nextTick(() => { autoLabelData.value = buildAutoLabelData(createdOrder) as any; }); }
  }

  await Promise.all([loadOrders(), loadInitial()]);
}

async function printLabel(order: Order) {
  showLabelPanel.value = true;
  nextTick(() => { autoLabelData.value = buildAutoLabelData(order) as any; });
}

function loadOrderIntoForm(order: Order) {
  form.buyerId = order.buyerId;
  form.recipientId = order.recipientId;
  form.courier = order.shipment.courier || '';
  form.serviceLevel = order.shipment.serviceLevel || '';
  form.trackingCode = '';
  form.shippingCost = order.shipment.shippingCost;
  form.discount = order.discount;
  form.notes = order.notes;
  form.items = order.items.map((item) => {
    const normalised = normalisePrice(item.unitPrice);
    const normalisedDiscount = normalisePrice(item.discount);
    return {
      productId: item.productId,
      quantity: item.quantity,
      unitPrice: normalised.value,
      discount: normalisedDiscount.value,
      product: productMap.value.get(item.productId),
      unitPriceDisplay: normalised.display,
      discountDisplay: normalisedDiscount.display
    };
  });
  if (!form.items.length) {
    addItem();
  }
  buyerMode.value = 'existing';
  recipientMode.value = 'existing';
  resetManualContact(buyerCustom);
  resetManualContact(recipientCustom);
  toast.push(`Form diisi ulang dari order ${order.code}. Periksa sebelum menyimpan.`, 'info', { timeout: 6000 });
}

onMounted(async () => {
  await loadInitial();
  await loadOrders();
});

watch(
  courierOptions,
  (options) => {
    if (!options.length) {
      form.courier = '';
      return;
    }
    if (!options.some((option) => option.value === form.courier)) {
      form.courier = options[0].value;
    }
  },
  { immediate: true }
);

watch(
  courierFilterOptions,
  (options) => {
    if (!options.some((option) => option.value === orderCourierFilter.value)) {
      orderCourierFilter.value = 'all';
    }
  },
  { immediate: true }
);

watch(orderDetailOpen, (open) => {
  if (!open) {
    activeOrder.value = null;
  }
});

watch(buyerMode, (mode) => {
  if (mode === 'manual') {
    form.buyerId = '';
    resetManualContact(buyerCustom);
  }
});

watch(recipientMode, (mode) => {
  if (mode === 'manual') {
    form.recipientId = '';
    resetManualContact(recipientCustom);
  }
});

watch(labelPreviewOpen, (open) => {
  if (!open) {
    revokeLabelPreviewUrl();
    labelPreview.value = null;
  }
});

watch(orderDateStart, (value) => {
  if (value && orderDateEnd.value && value > orderDateEnd.value) {
    orderDateEnd.value = value;
  }
});

watch(orderDateEnd, (value) => {
  if (value && orderDateStart.value && value < orderDateStart.value) {
    orderDateStart.value = value;
  }
});


const LABEL_APP_URL: string = (import.meta as any).env?.VITE_LABEL_APP_URL || 'http://localhost:3000';

function getCustomerById(id?: string) {
  if (!id) return undefined;
  return (customers.value || []).find((c: any) => c.id === id);
}

function buildLabelAppUrl(order: Order) {
  const buyer = getCustomerById(order.buyerId);
  const recipient = getCustomerById(order.recipientId);
  const params = new URLSearchParams();
  if (buyer?.name) params.set('buyerName', buyer.name);
  if (buyer?.phone) params.set('buyerPhone', buyer.phone);
  if (buyer?.address) params.set('buyerAddress', buyer.address);
  if (recipient?.name) params.set('recipientName', recipient.name);
  if (recipient?.phone) params.set('recipientPhone', recipient.phone);
  if (recipient?.address) params.set('recipientAddress', recipient.address);
  if (order.shipment?.courier) params.set('courier', order.shipment.courier);
  if (order.shipment?.serviceLevel) params.set('service', order.shipment.serviceLevel);
  if (order.shipment?.trackingCode) params.set('tracking', order.shipment.trackingCode);
  params.set('orderCode', order.code);
  if (order.notes) params.set('notes', order.notes);
  const url = `${LABEL_APP_URL}?tab=auto&${params.toString()}`;
  return url;
}

function openLabelApp(order: Order) {
  const url = buildLabelAppUrl(order);
  window.open(url, '_blank', 'noopener,noreferrer');
}



const showLabelPanel = ref(false);

function buildAutoLabelData(order?: Order | null) {
  const buyer = customers.value.find(c => c.id === (order?.buyerId || form.buyer?.id));
  const recipient = customers.value.find(c => c.id === (order?.recipientId || form.recipient?.id));
  return {
    senderName: buyer?.name || '',
    senderPhone: buyer?.phone || '',
    senderAddress: buyer?.address || '',
    recipientName: recipient?.name || '',
    recipientPhone: recipient?.phone || '',
    recipientAddress: recipient?.address || '',
    courier: order?.shipment?.courier || form.courier || '',
    service: order?.shipment?.serviceLevel || form.serviceLevel || '',
    trackingCode: order?.shipment?.trackingCode || form.trackingCode || '',
    orderCode: order?.code || '',
    notes: order?.notes || form.notes || ''
  };
}

onBeforeUnmount(() => {
  revokeLabelPreviewUrl();
});
</script>
