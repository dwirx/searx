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

### 3. Filter Berdasarkan Tahun dan Status
Gunakan `-law-year` untuk tahun tertentu dan `-law-status` untuk memantau status hukum.

**Contoh:**
```bash
search -pasal -law-year 2024         # Semua hukum tahun 2024
search -pasal -law-status dicabut    # Cari hukum yang sudah dicabut
search -pasal -law-type UU -law-year 2023 -law-status berlaku
```

**Kode Jenis Peraturan yang Didukung:**
- `UUD`: Undang-Undang Dasar
- `TAP_MPR`: Ketetapan MPR
- `UU`: Undang-Undang
- `PP`: Peraturan Pemerintah
- `PERPPU`: Perpu
- `PERPRES`: Peraturan Presiden
- `KEPPRES`: Keputusan Presiden
- `INPRES`: Instruksi Presiden
- `PENPRES`: Penetapan Presiden
- `PERMEN`: Peraturan Menteri
- `PERMENKUMHAM`: Peraturan Menkumham
- `PERMENKUM`: Peraturan Menkum
- `PERBAN`: Peraturan Badan
- `PERDA`: Peraturan Daerah
- `PERDA_PROV`: Perda Provinsi
- `PERDA_KAB`: Perda Kab/Kota
- `KEPMEN`: Keputusan Menteri
- `SE`: Surat Edaran
- `PERMA`: Peraturan MA
- `PBI`: Peraturan Bank Indonesia
- `UUDRT`: UU Darurat
- `UUDS`: UUDS 1950

**Nilai Status yang Didukung:**
- `berlaku`
- `dicabut`
- `diubah`

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
