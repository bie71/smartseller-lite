# SmartSeller Lite

Toolkit fulfilment untuk agen pengiriman yang berjalan sepenuhnya secara lokal. Backend ditenagai server HTTP Go + MySQL/MariaDB, sedangkan antarmuka dibangun menggunakan Vue 3 dan disajikan lewat browser favorit Anda.

## Fitur utama

- 🧾 Pembuatan order cepat lengkap dengan kalkulasi profit dan cetak label PDF.
- 📦 Manajemen produk, stok, dan histori penyesuaian lengkap dengan arsip produk.
- 👥 Direktori customer/marketer/reseller dengan informasi kontak lengkap.
- 🚚 Daftar ekspedisi yang dapat diedit sesuai kebutuhan agen.
- 🔍 Pencarian instan lengkap dengan filter tanggal & ekspedisi plus insight penjualan untuk histori order dan katalog ekspedisi.
- 🔁 Histori order bisa langsung mengunduh ulang label PDF atau mengisi ulang formulir untuk repeat order cepat.
- 📊 Stock opname terpadu yang menyimpan nama petugas, menyesuaikan stok otomatis, dan menyimpan riwayat audit.
- 🔔 Notifikasi stok menipis dengan ambang yang dapat diatur per produk sehingga restock bisa diprioritaskan.
- 🎨 Pengaturan branding aplikasi termasuk upload logo yang tercetak pada label PDF.

## Arsitektur singkat

- Server Go melayani API REST di `http://127.0.0.1:8787` dan menyajikan berkas statis hasil build Vue (`frontend/dist`).
- Seluruh data tersimpan di database MySQL/MariaDB yang ditentukan melalui DSN (default `smartseller:smartseller@tcp(127.0.0.1:3306)/smartseller`).
- Parameter `--open-browser` akan otomatis membuka aplikasi di browser default setelah server siap.

## Prasyarat

- Go 1.22+
- Node.js 18+ (untuk membangun frontend)
- MySQL/MariaDB 10.5+ (akses lokal atau jaringan) dan hak membuat schema/table
- (Opsional) utilitas `mysqldump` & `mysql` untuk backup/restore berkas `.sql` berperforma tinggi

## Konfigurasi lingkungan

1. Salin contoh environment lalu sesuaikan bila perlu:
   ```bash
   cp .env.example .env
   ```
2. Variabel yang tersedia:
   - `APP_BRAND_NAME` – nama brand default yang tampil di UI & label PDF (dapat diganti lewat menu Pengaturan).
   - `APP_ADDR` – alamat bind server (default `127.0.0.1:8787`).
   - `DATABASE_DSN` – kredensial koneksi MySQL/MariaDB (contoh `user:pass@tcp(127.0.0.1:3306)/smartseller`).

## Instalasi dependensi

```bash
# frontend (ikon & build tool)
npm -C frontend install

# backend (mengunduh dependency Go)
go mod tidy
```

## Menjalankan aplikasi saat pengembangan

### Mode embed cepat

```bash
# 1. Bangun aset frontend sekali
npm -C frontend run build

# 2. Jalankan server lokal dan otomatis buka browser
DATABASE_DSN="smartseller:smartseller@tcp(127.0.0.1:3306)/smartseller" go run . --open-browser
```

### Mode hot reload (frontend + backend)

Jalankan backend dan frontend pada terminal terpisah agar perubahan kode langsung terlihat:

```bash
# Terminal 1 – backend Go
DATABASE_DSN="smartseller:smartseller@tcp(127.0.0.1:3306)/smartseller" go run . --addr 127.0.0.1:8787

# Terminal 2 – frontend Vite (host agar bisa diakses browser lain)
VITE_API_BASE_URL=http://127.0.0.1:8787/api npm -C frontend run dev -- --host
```

Saat menggunakan hot reload, aset dikirim dari server Vite (biasanya `http://127.0.0.1:5173`). Build produksi (`npm -C frontend run build`) tetap perlu dijalankan sebelum kompilasi biner Go agar bundel terbaru ter-embed ke executable.

## Build rilis

```bash
# Bangun aset Vue terlebih dahulu
npm -C frontend run build

# Windows (AMD64)
GOOS=windows GOARCH=amd64 go build -o dist/windows/smartsellerlite.exe .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dist/macos/smartsellerlite .

# Linux (AMD64)
GOOS=linux GOARCH=amd64 go build -o dist/linux/smartsellerlite .
```

Setiap hasil build sudah menyertakan aset frontend dari `frontend/dist`. Sesuaikan target `GOARCH` bila perlu (misal `386` untuk Windows 32-bit). Untuk paket instalasi Windows, gunakan skrip NSIS/Inno Setup sesuai kebutuhan dan pastikan shortcut menjalankan `smartsellerlite.exe --open-browser`.

## Tips penggunaan UI

- Tab **Pengaturan** memungkinkan Anda mengunggah logo (PNG/JPG ≤1MB) yang langsung dipakai di dashboard serta label PDF.
- Menu **Pengaturan → Backup & Restore** menyediakan ekspor SQL penuh (schema + data) dengan opsi hanya data atau hanya schema, lengkap dengan fitur restore `.sql` yang otomatis menonaktifkan foreign key check bila dipilih.
- Halaman **Order** kini menyediakan tombol ekspor CSV untuk laporan transaksi yang dapat dibuka di spreadsheet favorit Anda.
- Tab **Ekspedisi** menyimpan daftar ekspedisi favorit. Data ini juga muncul sebagai pilihan saat membuat order.
- Ikon dan badge di setiap halaman membantu memantau subtotal, profit, serta status stok secara sekilas.
- Badge kuning/merah pada tab Produk menandakan stok menipis atau habis. Sesuaikan ambang per SKU dari formulir produk dan gunakan arsip untuk menyembunyikan item yang tidak lagi dijual tanpa menghapus histori order.
- Setiap sesi stock opname menyimpan nama petugas sehingga audit dapat ditelusuri kembali dengan mudah.

Selamat berjualan lebih cerdas! 🚀
