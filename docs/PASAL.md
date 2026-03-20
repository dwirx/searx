# ⚖️ RI Law Search (Pasal.id) Guide

Fitur ini memungkinkan Anda mencari dan membaca naskah hukum Indonesia (Undang-Undang, Peraturan Pemerintah, dll) langsung dari terminal menggunakan API resmi dari [Pasal.id](https://pasal.id).

---

## 🔍 Cara Mencari Hukum

Gunakan flag `-pasal` diikuti dengan kata kunci pencarian Anda.

### 1. Pencarian Umum
Mencari kata kunci di seluruh database hukum:
```bash
search -pasal "upah minimum"
search -pasal "hak cipta"
```

### 2. Filter Berdasarkan Jenis Peraturan
Gunakan flag `-law-type` untuk mempersempit hasil pencarian (misal: hanya Undang-Undang atau Peraturan Pemerintah).

**Contoh:**
```bash
search -pasal -law-type UU "ketenagakerjaan"
search -pasal -law-type PP "pajak"
search -pasal -law-type PERPRES "gaji"
```

**Kode Jenis Peraturan yang Didukung:**
- `UUD`: Undang-Undang Dasar
- `UU`: Undang-Undang
- `PP`: Peraturan Pemerintah
- `PERPPU`: Perpu
- `PERPRES`: Peraturan Presiden
- `PERMEN`: Peraturan Menteri
- `INPRES`: Instruksi Presiden
- `PERDA`: Peraturan Daerah

---

## 📖 Membaca Naskah Lengkap

Setelah mendapatkan hasil pencarian, Anda dapat menyalin link URL `pasal.id` dan membacanya dalam format bersih (Reader Mode).

```bash
search -read "https://pasal.id/akn/id/act/uu/2020/11"
```

**Fitur Reader Mode Khusus Hukum:**
- **Status Otomatis**: Menampilkan apakah peraturan masih `BERLAKU`, `DICABUT`, atau `DIUBAH`.
- **Relasi Hukum**: Menampilkan daftar peraturan yang mengubah atau mencabut peraturan tersebut.
- **Struktur Rapi**: Otomatis membagi naskah berdasarkan **BAB** dan **Pasal**.

---

## 🎨 Penjelasan Tampilan (UI)

- **Label Magenta `[UU]`**: Menunjukkan jenis peraturan.
- **Teks Hijau `Pasal X`**: Menunjukkan nomor pasal yang relevan dengan keyword Anda.
- **Status Berwarna**:
    - 🟢 **BERLAKU**: Peraturan masih aktif.
    - 🔴 **DICABUT**: Peraturan sudah tidak berlaku.
    - 🟡 **DIUBAH**: Peraturan sudah mengalami amandemen.

---

## ⚙️ Integrasi API
Pencarian ini menggunakan API v1 `pasal.id`.
- **Base URL**: `https://pasal.id/api/v1`
- **Rate Limit**: 60 req/min (Sangat cukup untuk penggunaan CLI personal).

---
*Gunakan fitur ini untuk riset hukum yang cepat, akurat, dan bebas gangguan iklan.* 🇮🇩⚖️
