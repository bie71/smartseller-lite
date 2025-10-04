<template>
  <section class="space-y-6">
    <div class="space-y-6 xl:grid xl:grid-cols-5 xl:gap-6 xl:space-y-0">
      <div class="space-y-6 xl:col-span-3">
        <div class="card space-y-5 xl:sticky xl:top-28">
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
                  rows="2"
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
            <div class="mt-2 flex items-center gap-2">
              <input id="isBuyerPayingShipping" v-model="isBuyerPayingShipping" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary" />
              <label for="isBuyerPayingShipping" class="text-sm text-slate-600">Ongkir dibayar pembeli</label>
            </div>
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
            <span>Total bayar Rp {{ formatCurrency(orderPaymentTotal) }} · Profit Rp {{ formatCurrency(estimatedProfit) }}</span>
          </div>
          <button type="submit" class="btn-primary" :disabled="!canSubmit || savingOrder">
            <PrinterIcon class="h-5 w-5" />
            {{ savingOrder ? 'Menyimpan...' : 'Simpan & Cetak Label' }}
          </button>
        </footer>
      </form>
        </div>
      </div>
      <aside class="space-y-6 xl:col-span-2">
        <div class="card space-y-5">
      <header class="space-y-4">
        <div class="flex items-center gap-2">
          <ArrowPathIcon class="h-5 w-5 text-primary" />
          <h3 class="text-lg font-semibold">Histori Order</h3>
        </div>
        <div class="space-y-3">
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
              class="inline-flex items-center rounded-lg px-3 py-1 text-xs font-medium text-slate-500 transition hover:text-primary"
              @click="resetOrderFilters"
            >
              Bersihkan filter
            </button>
            <button :class="orderHistoryButtonClasses" @click="exportOrders">
              <ArrowDownTrayIcon class="h-4 w-4" />
              Export CSV
            </button>
            <button :class="orderHistoryButtonClasses" @click="loadOrders">
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
      <div v-if="orderTotal > 0" class="flex flex-wrap gap-2 text-xs text-slate-600 sm:text-sm">
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
      <div
        v-if="ordersError"
        class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600"
      >
        {{ ordersError }}
      </div>
      <div v-else-if="!orders.length" class="text-sm text-slate-500">Belum ada order.</div>
      <div v-else-if="orderTotal === 0" class="text-sm text-slate-500">Tidak ada order yang cocok dengan pencarian.</div>
      <template v-else>
        <div v-for="order in orders" :key="order.id" class="border-t border-slate-200 p-4 space-y-3 first:border-t-0">
          <div class="space-y-1">
            <h4 class="font-semibold">{{ order.code }}</h4>
            <p class="text-xs text-slate-500">
              {{ formatDate(order.createdAt) }} · {{ order.shipment.courier }} ({{ order.shipment.serviceLevel || 'N/A' }})
            </p>
            <p class="text-xs text-slate-400" v-if="orderParticipants(order)">
              {{ orderParticipants(order) }}
            </p>
          </div>
          <ul class="text-sm text-slate-600 list-disc list-inside">
            <li v-for="item in order.items" :key="item.id">
              {{ productName(item.productId) }} × {{ item.quantity }} @ Rp {{ formatCurrency(item.unitPrice) }}
            </li>
          </ul>
          <div class="text-sm text-slate-500">
            Total Rp {{ formatCurrency(order.total) }} · Profit Rp {{ formatCurrency(order.profit) }}
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              :class="orderHistoryActionDetailClasses"
              @click="openOrderDetail(order)"
            >
              <InformationCircleIcon class="h-4 w-4" />
              Detail
            </button>
            <button
              type="button"
              :class="orderHistoryActionLoadClasses"
              @click="loadOrderIntoForm(order)"
            >
              <ArrowUturnLeftIcon class="h-4 w-4" />
              Muat ke Form
            </button>
            <button
              type="button"
              :class="orderHistoryActionPrintClasses"
              :disabled="labelBusy === order.id"
              @click="printLabel(order)"
            >
              <PrinterIcon class="h-4 w-4" />
              {{ labelBusy === order.id ? 'Menyiapkan...' : 'Label PDF' }}
            </button>
            <button
              type="button"
              :class="orderHistoryActionDeleteClasses"
              @click="deleteOrderAction(order)"
            >
              <TrashIcon class="h-4 w-4" />
              Hapus
            </button>
          </div>
        </div>
        <footer
          v-if="orderTotal > orderPageSize"
          class="flex flex-col gap-3 border-t border-slate-100 p-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between"
        >
          <span>{{ orderRangeLabel }}</span>
          <div class="flex items-center gap-3">
            <span>Halaman {{ orderPage }} / {{ totalOrderPages }}</span>
            <div class="flex items-center gap-2">
              <button type="button" :class="paginationButtonClasses" :disabled="orderPage === 1" @click="previousOrderPage">
                <ChevronLeftIcon class="h-4 w-4" />
              </button>
              <button type="button" :class="paginationButtonClasses" :disabled="orderPage === totalOrderPages" @click="nextOrderPage">
                <ChevronRightIcon class="h-4 w-4" />
              </button>
            </div>
          </div>
        </footer>
      </template>
        </div>
      </aside>
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
            <li
              v-for="item in activeOrder.items"
              :key="item.id"
              class="flex flex-col border-b border-dotted border-slate-100 dark:border-slate-700 pb-1"
            >
              <!-- Baris 1: nama produk dan harga -->
              <div class="flex justify-between text-sm">
                <span>{{ productName(item.productId) }} × {{ item.quantity }}</span>
                <span class="font-mono font-semibold">Rp {{ formatCurrency(item.unitPrice) }}</span>
              </div>

              <!-- Baris 2: diskon -->
              <div
                v-if="item.discountItem && item.discountItem > 0"
                class="flex justify-between text-xs text-slate-500 mt-1"
              >
                <span>Diskon Item</span>
                <span class="font-mono">- Rp {{ formatCurrency(item.discountItem) }}</span>
              </div>
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
              <p class="mt-1">Subtotal Rp {{ formatCurrency(orderSubtotal(activeOrder)) }}</p>

              <p v-if="activeOrder && activeOrder.discountOrder > 0"
                class="text-xs font-semibold text-emerald-600 mt-1">
                Diskon Order - Rp {{ formatCurrency(activeOrder.discountOrder) }}
              </p>

              <!-- Jika ongkir ditanggung pembeli: tampilkan diskon ongkir -->
              <template v-if="activeOrder && activeOrder.shipment.shippingByBuyer">
                <p class="text-xs line-through text-slate-400 mt-1">
                  Ongkir Rp {{ formatCurrency(activeOrder.shipment.shippingCost) }}
                </p>
                <p class="text-xs font-semibold text-emerald-600 mt-1">
                  Ongkir dibayar pembeli - Rp {{ formatCurrency(activeOrder.shipment.shippingCost) }}
                </p>
              </template>
              <!-- Jika tidak, tampilkan ongkir biasa -->
              <p v-else class="mt-1">
                Ongkir Rp {{ formatCurrency(activeOrder?.shipment.shippingCost || 0) }}
              </p>

              <p class="font-semibold text-slate-800 mt-1">
                Total Rp {{ formatCurrency(activeOrder?.total || 0) }}
              </p>

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

    <!-- Panel Cetak Label -->
    <WideBaseModal v-model="showLabelPanel" title="Cetak Label">
      <div v-if="labelPreviewUrl" class="space-y-4">
        <div class="aspect-[3/4] w-full overflow-hidden rounded-xl border border-slate-200 shadow-inner">
          <iframe :src="labelPreviewUrl" title="Pratinjau label" class="h-full w-full"></iframe>
        </div>
        <div class="flex justify-end gap-3">
            <button type="button" class="btn-secondary" @click="showLabelPanel = false">Tutup</button>
            <button type="button" class="btn-primary" @click="downloadActiveLabel">
              <ArrowDownTrayIcon class="h-5 w-5" />
              Unduh PDF
            </button>
        </div>
      </div>
      <LabelForm v-else :auto-data="autoLabelData" @download-label="handleDownloadLabel" />
    </WideBaseModal>
  </section>

</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue';
import { listProducts } from '../../modules/product';
import type { Product } from '../../modules/product';
import { listCustomers, saveCustomer } from '../../modules/customer';
import type { Customer, CustomerType } from '../../modules/customer';
import { createOrder, deleteOrder, downloadLabel, listOrders, type Order, type OrderListResponse, type OrderListSummary, type UiOrderItem } from '../../modules/order';
import { generateSingleLabelPdf, type LabelData } from '../../modules/label';
import { fetchOrdersCsv } from '../../modules/reports';
import { getSettings, listCouriers, type AppSettings, type Courier } from '../../modules/settings';
import BaseModal from '../components/BaseModal.vue';
import WideBaseModal from '../components/WideBaseModal.vue';
import LabelForm from '../components/LabelForm.vue';
import { useToastStore } from '../stores/toast';
import {
  ArrowDownTrayIcon,
  ArrowPathIcon,
  ArrowUturnLeftIcon,
  BanknotesIcon,
  ChartBarIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
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
const ordersLoading = ref(false);
const ordersError = ref('');
const orderTotal = ref(0);
const orderSummaryMeta = ref<OrderListSummary | null>(null);
const orderAvailableCouriers = ref<string[]>([]);
const couriers = ref<Courier[]>([]);
const orderSearch = ref('');
const orderDateStart = ref('');
const orderDateEnd = ref('');
const orderCourierFilter = ref('all');
const savingOrder = ref(false);
const labelBusy = ref<string | null>(null);
const orderDetailOpen = ref(false);
const activeOrder = ref<Order | null>(null);
const autoLabelData = ref<any>(null);
const toast = useToastStore();

const shippingCostDisplay = ref('');
const orderDiscountDisplay = ref('');
const isBuyerPayingShipping = ref(false);
const labelPreviewUrl = ref('');
const activeLabelOrder = ref<Order | null>(null);

const orderPage = ref(1);
const orderPageSize = 5;
const paginationButtonClasses =
  'inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 bg-white text-slate-600 transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-40';

const orderHistoryButtonClasses =
  'inline-flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1 text-xs font-medium text-slate-600 transition hover:border-primary/40 hover:bg-primary/5 disabled:cursor-not-allowed disabled:opacity-50';
const orderHistoryActionDetailClasses =
  'inline-flex items-center gap-1 rounded-lg bg-sky-600 px-3 py-1 text-xs font-semibold text-white shadow-sm transition hover:bg-sky-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 focus-visible:ring-sky-500';
const orderHistoryActionLoadClasses =
  'inline-flex items-center gap-1 rounded-lg bg-amber-500 px-3 py-1 text-xs font-semibold text-white shadow-sm transition hover:bg-amber-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 focus-visible:ring-amber-500';
const orderHistoryActionPrintClasses =
  'inline-flex items-center gap-1 rounded-lg bg-emerald-600 px-3 py-1 text-xs font-semibold text-white shadow-sm transition hover:bg-emerald-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 focus-visible:ring-emerald-500 disabled:cursor-not-allowed disabled:opacity-60';
const orderHistoryActionDeleteClasses =
  'inline-flex items-center gap-1 rounded-lg bg-rose-600 px-3 py-1 text-xs font-semibold text-white shadow-sm transition hover:bg-rose-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1 focus-visible:ring-rose-500';

let orderFetchId = 0;
let orderFetchTimer: ReturnType<typeof setTimeout> | null = null;

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
  shippingByBuyer: false,
  discountOrder: 0,
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
  () => form.discountOrder,
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
    const revenue = item.unitPrice * item.quantity - item.discountItem;
    return total + Math.max(revenue, 0);
  }, 0)
);

const totalCost = computed(() =>
  form.items.reduce((total, item) => {
    const costPrice = item.product?.costPrice ?? productMap.value.get(item.productId)?.costPrice ?? 0;
    return total + costPrice * item.quantity;
  }, 0)
);

const orderPaymentTotal = computed(() => subtotal.value - form.discountOrder + form.shippingCost);

const estimatedProfit = computed(() => {
  const shippingCostToSubtract = isBuyerPayingShipping.value ? 0 : form.shippingCost;
  const value = subtotal.value - form.discountOrder - totalCost.value - shippingCostToSubtract;
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
  const options: Array<{ value: string; label: string }> = [{ value: 'all', label: 'Semua ekspedisi' }];
  const seen = new Set<string>();
  orderAvailableCouriers.value.forEach((raw) => {
    const key = raw ?? '';
    if (seen.has(key)) {
      return;
    }
    seen.add(key);
    const match = couriers.value.find((item) => {
      const candidate = item.code || item.name;
      return candidate === key || (key === '' && candidate === '');
    });
    const label = match
      ? `${match.code || match.name}${match.code && match.name ? ` · ${match.name}` : ''}`
      : key !== ''
        ? key
        : 'Lainnya';
    options.push({ value: key, label });
  });
  return options;
});

const totalOrderPages = computed(() => (orderTotal.value > 0 ? Math.ceil(orderTotal.value / orderPageSize) : 1));

const orderRangeLabel = computed(() => {
  if (orderTotal.value === 0) {
    return 'Menampilkan 0 dari 0 order';
  }
  const start = (orderPage.value - 1) * orderPageSize + 1;
  const end = start + orders.value.length - 1;
  return `Menampilkan ${start}-${end} dari ${orderTotal.value} order`;
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
      (orderCourierFilter.value === '' ? 'Lainnya' : orderCourierFilter.value);
    chips.push(`Ekspedisi: ${label}`);
  }
  return chips;
});

const orderInsights = computed(() => {
  const summary = orderSummaryMeta.value;
  if (!summary || summary.count <= 0) {
    return { count: 0, revenue: 0, profit: 0, topCourier: '-', topProduct: '-' };
  }
  const courierLabel = summary.topCourierHits > 0 ? `${summary.topCourier || 'Lainnya'} (${summary.topCourierHits})` : '-';
  const productLabelBase = summary.topProductName || (summary.topProductId ? productName(summary.topProductId) : '');
  const productLabel = summary.topProductQty > 0 ? `${productLabelBase || 'Produk'} (${summary.topProductQty})` : '-';
  return {
    count: summary.count,
    revenue: summary.revenue,
    profit: summary.profit,
    topCourier: courierLabel,
    topProduct: productLabel
  };
});

const activeOrderBuyer = computed(() => (activeOrder.value ? customerMap.value.get(activeOrder.value.buyerId) : undefined));
const activeOrderRecipient = computed(() =>
  activeOrder.value ? customerMap.value.get(activeOrder.value.recipientId) : undefined
);

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
    discountItem: 0,
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
  item.discountItem = normalised.value;
  item.discountDisplay = normalised.display;
}

function bindProduct(item: OrderFormItem) {
  const product = productMap.value.get(item.productId);
  if (product) {
    item.product = product;
    applyUnitPrice(item, product.salePrice);
    applyDiscount(item, item.discountItem);
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
  if (item.discountItem !== value) {
    item.discountItem = value;
  }
  item.discountDisplay = display;
}

function syncDiscountDisplay(item: OrderFormItem) {
  const { value, display } = parsePriceInputValue(item.discountDisplay);
  item.discountItem = value;
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
  if (form.discountOrder !== value) {
    form.discountOrder = value;
  }
  orderDiscountDisplay.value = display;
}

function syncOrderDiscountDisplay() {
  const { value, display } = parsePriceInputValue(orderDiscountDisplay.value);
  form.discountOrder = value;
  orderDiscountDisplay.value = display;
}

function lineRevenue(item: OrderFormItem) {
  return Math.max(item.unitPrice * item.quantity - item.discountItem, 0);
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
  return order.items.reduce((sum, item) => sum + Math.max(item.unitPrice * item.quantity - item.discountItem, 0), 0);
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

async function loadInitial() {
  try {
    const [customerRes, productRes, courierRes] = await Promise.all([
      listCustomers({ page: 1, pageSize: 500 }),
      listProducts({ page: 1, pageSize: 500 }),
      listCouriers({ page: 1, pageSize: 200 })
    ]);
    customers.value = customerRes.items;
    products.value = productRes.items;
    couriers.value = courierRes.items;
    if (!form.courier && couriers.value.length) {
      form.courier = couriers.value[0].code || couriers.value[0].name;
    }
    if (!form.items.length) addItem();
    await loadOrders();
  } catch (error) {
    console.error(error);
    toast.push('Gagal memuat data awal order.', 'error');
  }
}

async function loadOrders() {
  const requestId = ++orderFetchId;
  ordersLoading.value = true;
  ordersError.value = '';
  try {
    const response: OrderListResponse = await listOrders({
      page: orderPage.value,
      pageSize: orderPageSize,
      query: orderSearch.value.trim() || undefined,
      courier: orderCourierFilter.value !== 'all' ? orderCourierFilter.value : undefined,
      dateStart: orderDateStart.value || undefined,
      dateEnd: orderDateEnd.value || undefined
    });

    if (requestId !== orderFetchId) {
      return;
    }

    const totalPages = Math.max(1, Math.ceil(response.total / orderPageSize));
    if (orderPage.value > totalPages) {
      orderPage.value = totalPages;
      return;
    }

    orders.value = response.items;
    orderTotal.value = response.total;
    orderSummaryMeta.value = response.summary;
    orderAvailableCouriers.value = response.couriers;
    ordersError.value = '';
  } catch (error) {
    console.error(error);
    if (requestId === orderFetchId) {
      orders.value = [];
      orderTotal.value = 0;
      orderSummaryMeta.value = null;
      orderAvailableCouriers.value = [];
      ordersError.value =
        error instanceof Error && error.message ? error.message : 'Gagal memuat histori order.';
      toast.push('Gagal memuat histori order.', 'error');
    }
  } finally {
    if (requestId === orderFetchId) {
      ordersLoading.value = false;
    }
  }
}

function scheduleLoadOrders(delay = 250) {
  if (orderFetchTimer) {
    clearTimeout(orderFetchTimer);
  }
  orderFetchTimer = setTimeout(() => {
    orderFetchTimer = null;
    void loadOrders();
  }, delay);
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
  form.discountOrder = 0;
  form.notes = '';
  form.items = [] as OrderFormItem[];
  addItem();
  buyerMode.value = 'existing';
  recipientMode.value = 'existing';
  resetManualContact(buyerCustom);
  resetManualContact(recipientCustom);
  isBuyerPayingShipping.value = false;
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
      isBuyerPayingShipping: isBuyerPayingShipping.value,
      discountOrder: form.discountOrder,
      notes: form.notes,
      items: form.items.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        discountItem: item.discountItem
      }))
    };
    createdOrder = await createOrder(payload);
    orders.value = [createdOrder, ...orders.value];
    resetForm();
  } catch (error) {
    console.error(error);
    const message = error instanceof Error ? error.message : 'Gagal menyimpan order.';
    toast.push(message, 'error', { timeout: 5000 });
  } finally {
    savingOrder.value = false;
  }

  if (!createdOrder) {
    return;
  }

  if (createdOrder) {
    toast.push('Order tersimpan. Menyiapkan pratinjau label...', 'success');
    try {
      labelBusy.value = createdOrder.id;
      const appSettings = await getSettings();
      const buyer = customerMap.value.get(createdOrder.buyerId);
      const recipient = customerMap.value.get(createdOrder.recipientId);

      if (!buyer || !recipient) {
        throw new Error('Data pemesan atau penerima tidak lengkap');
      }

      const labelData: LabelData = {
        id: createdOrder.id,
        senderName: buyer.name,
        senderPhone: buyer.phone,
        senderAddress: buyer.address,
        recipientName: recipient.name,
        recipientPhone: recipient.phone,
        recipientAddress: recipient.address,
        courier: createdOrder.shipment.courier,
        service: createdOrder.shipment.serviceLevel,
        trackingCode: createdOrder.shipment.trackingCode,
        orderCode: createdOrder.code,
        notes: createdOrder.notes,
        items: createdOrder.items.map(item => ({
          productName: productName(item.productId),
          quantity: item.quantity,
          unitPrice: item.unitPrice,
          discountItem: item.discountItem
        }))
      };

      const blob = await generateSingleLabelPdf(labelData, appSettings);
      
      if (labelPreviewUrl.value) {
        URL.revokeObjectURL(labelPreviewUrl.value);
      }
      
      labelPreviewUrl.value = URL.createObjectURL(blob);
      activeLabelOrder.value = createdOrder;
      showLabelPanel.value = true;
    } catch (error) {
      console.error(error);
      toast.push('Order tersimpan, tapi pratinjau label gagal dibuat.', 'error');
    } finally {
      labelBusy.value = null;
    }
  }

  await Promise.all([loadOrders(), loadInitial()]);
}

async function printLabel(order: Order) {
  if (labelPreviewUrl.value) {
    URL.revokeObjectURL(labelPreviewUrl.value);
    labelPreviewUrl.value = '';
  }
  activeLabelOrder.value = null;
  showLabelPanel.value = true;
  nextTick(() => {
    autoLabelData.value = buildAutoLabelData(order) as any;
  });
}

async function deleteOrderAction(order: Order) {
    if (!order.id) return;
    const confirmed = window.confirm(`Hapus order ${order.code}? Tindakan ini tidak bisa dibatalkan.`);
    if (!confirmed) return;

    try {
        await deleteOrder(order.id);
        orders.value = orders.value.filter(o => o.id !== order.id);
        toast.push(`Order ${order.code} telah dihapus.`, 'success');
    } catch (error) {
        console.error(error);
        const message = error instanceof Error ? error.message : 'Gagal menghapus order.';
        toast.push(message, 'error');
    }
}

async function downloadActiveLabel() {
  if (!activeLabelOrder.value) return;
  if (labelBusy.value) return;
  try {
    labelBusy.value = activeLabelOrder.value.id;
    const appSettings = await getSettings();
    const buyer = customerMap.value.get(activeLabelOrder.value.buyerId);
    const recipient = customerMap.value.get(activeLabelOrder.value.recipientId);
    if (!buyer || !recipient) throw new Error('Data pemesan atau penerima tidak lengkap');

    const labelData: LabelData = {
      id: activeLabelOrder.value.id,
      senderName: buyer.name,
      senderPhone: buyer.phone,
      senderAddress: buyer.address,
      recipientName: recipient.name,
      recipientPhone: recipient.phone,
      recipientAddress: recipient.address,
      courier: activeLabelOrder.value.shipment.courier,
      service: activeLabelOrder.value.shipment.serviceLevel,
      trackingCode: activeLabelOrder.value.shipment.trackingCode,
      orderCode: activeLabelOrder.value.code,
      notes: activeLabelOrder.value.notes,
      items: activeLabelOrder.value.items.map(item => ({
        productName: productName(item.productId),
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        discountItem: item.discountItem
      }))
    };
    const blob = await generateSingleLabelPdf(labelData, appSettings);
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `label-${activeLabelOrder.value.code}.pdf`;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
    toast.push(`Label untuk order ${activeLabelOrder.value.code} telah diunduh.`, 'success');
  } catch (error) {
    console.error(error);
    toast.push('Gagal mengunduh label.', 'error');
  } finally {
    labelBusy.value = null;
  }
}

async function handleDownloadLabel(orderId: string) {
  if (!orderId) {
    toast.push('ID Order tidak ditemukan.', 'error');
    return;
  }
  const order = orders.value.find(o => o.id === orderId) || activeLabelOrder.value;
  if (!order) {
    toast.push('Order tidak ditemukan untuk diunduh.', 'error');
    return;
  }

  if (labelBusy.value) return;
  try {
    labelBusy.value = order.id;
    const appSettings = await getSettings();
    const buyer = customerMap.value.get(order.buyerId);
    const recipient = customerMap.value.get(order.recipientId);
    if (!buyer || !recipient) throw new Error('Data pemesan atau penerima tidak lengkap');

    const labelData: LabelData = {
      id: order.id,
      senderName: buyer.name,
      senderPhone: buyer.phone,
      senderAddress: buyer.address,
      recipientName: recipient.name,
      recipientPhone: recipient.phone,
      recipientAddress: recipient.address,
      courier: order.shipment.courier,
      service: order.shipment.serviceLevel,
      trackingCode: order.shipment.trackingCode,
      orderCode: order.code,
      notes: order.notes,
      items: order.items.map(item => ({
        productName: productName(item.productId),
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        discountItem: item.discountItem
      }))
    };
    const blob = await generateSingleLabelPdf(labelData, appSettings);
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `label-${order.code}.pdf`;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
    toast.push(`Label untuk order ${order.code} telah diunduh.`, 'success');
  } catch (error) {
    console.error(error);
    toast.push('Gagal mengunduh label.', 'error');
  } finally {
    labelBusy.value = null;
  }
}

function loadOrderIntoForm(order: Order) {
  form.buyerId = order.buyerId;
  form.recipientId = order.recipientId;
  form.courier = order.shipment.courier || '';
  form.serviceLevel = order.shipment.serviceLevel || '';
  form.trackingCode = '';
  form.shippingCost = order.shipment.shippingCost;
  form.discountOrder = order.discountOrder;
  form.notes = order.notes;
  form.items = order.items.map((item) => {
    const normalised = normalisePrice(item.unitPrice);
    const normalisedDiscount = normalisePrice(item.discountItem);
    return {
      productId: item.productId,
      quantity: item.quantity,
      unitPrice: normalised.value,
      discountItem: normalisedDiscount.value,
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
  isBuyerPayingShipping.value = order.shipment.shippingByBuyer ;
  toast.push(`Form diisi ulang dari order ${order.code}. Periksa sebelum menyimpan.`, 'info', { timeout: 6000 });
}

function previousOrderPage() {
  if (orderPage.value > 1) {
    orderPage.value -= 1;
  }
}

function nextOrderPage() {
  if (orderPage.value < totalOrderPages.value) {
    orderPage.value += 1;
  }
}

onMounted(async () => {
  await loadInitial();
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

watch(orderSearch, () => {
  orderPage.value = 1;
  scheduleLoadOrders();
});

watch(orderCourierFilter, () => {
  orderPage.value = 1;
  void loadOrders();
});

watch([orderDateStart, orderDateEnd], () => {
  orderPage.value = 1;
  void loadOrders();
});

watch(orderPage, (value, oldValue) => {
  if (value !== oldValue) {
    void loadOrders();
  }
});


const showLabelPanel = ref(false);

function buildAutoLabelData(order?: Order | null) {
  const buyer = customers.value.find(c => c.id === (order?.buyerId || form?.buyerId));
  const recipient = customers.value.find(c => c.id === (order?.recipientId || form?.recipientId));
  const items = order ? order.items : form.items;

  return {
    orderId: order?.id || '',
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
    notes: order?.notes || form.notes || '',
    items: items.map(item => ({
      productName: productName(item.productId),
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      discountItem: item.discountItem
    }))
  };
}

watch(showLabelPanel, (isOpen) => {
  if (isOpen) {
    document.body.classList.add("overflow-hidden");
  } else {
    document.body.classList.remove("overflow-hidden");
    if (labelPreviewUrl.value) {
      URL.revokeObjectURL(labelPreviewUrl.value);
      labelPreviewUrl.value = '';
    }
    activeLabelOrder.value = null;
  }
});
</script>
