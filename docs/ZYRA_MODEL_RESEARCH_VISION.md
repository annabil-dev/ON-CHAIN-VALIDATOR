# Visi Riset Model ZYRA — Catatan Ide Founder

**Status:** VISI, dicatat 8 Oktober 2026. Bukan spec aktif, bukan janji rilis. Bahasa: Indonesia.

## Ide inti (kata founder, dirapikan tanpa diubah maknanya)

1. Task-task di jaringan bukan lemparan acak. Setiap task adalah **inference penelitian**:
   hasilnya berguna untuk perkembangan blockchain, sistem ZYRA, dan sedikit demi
   sedikit melatih model AI baru — **model ZYRA milik mythprotocol sendiri**.
2. Ide besarnya: **kompresi/minimisasi parameter AI** — model berparameter kecil
   yang bisa dijalankan ringan tetapi kualitasnya tidak jauh dari parameter aslinya.
3. Setiap task **wajib punya reason**: alasan kenapa task itu ada dan fungsi apa yang
   diharapkan dari hasilnya. Task tanpa reason sama saja dengan spam.

## Prinsip turunan yang disepakati

- **R1. Task tanpa reason ditolak secara konsep.** Setiap template katalog wajib
  menyatakan: research question, artefak yang diharapkan, dan kegunaan artefak
  (dataset latih, eval, atau upgrade sistem).
- **R2. Bootstrap mode.** Producer sintetik mulai melempar task **segera setelah
  chain siap**, bukan menunggu jadwal acak pertama. Kuota harian tetap berlaku
  setelah burst awal agar tidak membanjiri chain.
- **R3. Research-ready catalog.** Setiap template menghasilkan artefak yang bisa
  dipakai ulang + skor Judge sebagai label lemah. Task testnet otomatis menabung
  dataset untuk riset model nanti.
- **R4. Target kompresi yang terukur.** Klaim "kualitas tidak berbeda" diganti target
  terukur per domain, misal "X% kualitas pada Y% parameter di domain Z", via
  distilasi/kuantisasi/pruning. Tanpa angka, klaim tidak bisa diuji.

## Catatan jujur (dari review teknis)

- On-chain hanya menyimpan CID/proof + skor, **bukan artefak**. Flywheel riset butuh
  pipeline off-chain (kolektor CID → dataset → kurasi → training/eval) yang **belum ada**.
- Sinyal vote 3-of-4 itu noisy untuk training; cukup untuk payout, tipis untuk reward
  modeling. Skor rubrik berbobot adalah modal yang sudah ada.
- Scaling laws nyata: kompresi selalu ada trade-off. Domain sempit lebih realistis
  daripada model umum.
- Scope testnet tidak berubah: validasi L1 + peran + rejoint. Visi ini **tidak boleh**
  menunda genesis `mythchain-testnet-v2`.

## Komponen yang belum ada (daftar jujur)

1. Kolektor artefak off-chain (CID → dataset).
2. Kurasi + eval harness untuk dataset.
3. Katalog task v2 research-ready (template + reason per task).
4. Mode bootstrap pada producer sintetik (kode producer ada di sisi ZYRA, bukan repo ini).
5. Eksperimen distilasi/kuantisasi pertama.

## Pertanyaan terbuka untuk founder

1. Domain riset pertama apa (codegen? web-checks? reasoning?) — pilih SATU yang sempit.
2. Siapa kurator katalog v2 — founder sendiri atau Judge terpilih?
3. Dataset disimpan di mana (IPFS pin? repo terpisah? akses publik/privat)?
