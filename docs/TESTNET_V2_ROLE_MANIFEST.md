# Manifest Pemisahan Peran — Public Testnet `mythchain-testnet-v2`

**Status:** TEMPLATE — alamat diisi saat genesis testnet dibuat. Token testnet tak bernilai; jaringan boleh di-reset.

## Tabel akun peran

| Peran | Alamat | Fungsi |
|---|---|---|
| Founder wallet | `myth1… (isi saat genesis)` | Simpan alokasi testnet, funding awal transparan |
| Validator operator | `mythvaloper1…` + consensus key | Signing blok, komisi 5% |
| Task producer/client | `myth1… (isi saat pendaftaran)` | `MsgRegisterTask` + rubrik kriteria |
| Miner | `myth1…` | Claim + submit hasil |
| Judge (maks 4/task) | `myth1…` | `MsgVoteTask` + skor rubrik |

## Aturan keras fase testnet

1. Satu task: `client ≠ miner ≠ judge`. `client == miner` **ditolak kode** (`ClaimTask`); `miner/client == judge` **ditolak kode** (`VoteTask`).
2. Akun founder/validator testnet **dilarang mine dan judge**; hanya operasikan validator + funding awal transparan.
3. Kunci founder wallet, kunci konsensus, dan kunci P2P **tidak boleh berbagi secret**.
4. Satu alamat = satu vote per attempt; maks 4 vote per task.
5. Pelanggaran = hasil task dibatalkan + alamat dicoret dari testnet (tanpa penalti dana pada fase ini).

## Registrasi Judge/Miner (ringan, tanpa stake)

1. Pendaftaran via allowlist alamat di manifest jaringan publik.
2. Syarat: alamat `myth1…`, kontak publik, setuju kuorum 3-of-4 dan payout 60% Miner / 40% Judge selaras.
3. Misbehavior (kolusi, spam, vote asal): hapus dari allowlist + catatan publik.
4. Stake/slashing/reputasi on-chain **ditunda ke mainnet** via parameter governance.

## Faucet-gated onboarding (anti-Sybil)

1. Pesan PoUW (`Register/Claim/Submit/Vote/Release`) **wajib bayar fee** — tidak ada jalur
   gasless untuknya (`isPoUWTaskMessage` di `app/ante.go`). Satu-satunya sumber dana
   pertama adalah faucet founder (manual, rate-limit, catat alamat + tanggal).
2. Laju spam maksimum = laju drip faucet, bukan laju pembuatan alamat gratis.
3. Gasless satu-trial tetap berlaku untuk pesan non-PoUW (onboarding dasar akun baru).
4. Monitoring: pola 3 vote satu cluster pada `AttemptId` yang sama, atau miner+judge
   satu geng → batalkan task + coret alamat + catatan publik.

## Determinisme kriteria (anti-vote-asal)

1. Semua tipe check (`application_runs`, `stdout_contains`, `browser_contains`,
   `browser_fetch`, `http`) berbentuk executable; tipe di luar daftar ditolak saat
   registrasi, dan hard gate wajib bertipe executable.
2. Vote PASS pada check substring wajib mengutip output yang diobservasi di evidence
   (`BindEvidenceToChecks`): klaim "berisi X" tanpa X di evidence ditolak kode.
   Vote FAIL dikecualikan karena evidence-nya menjelaskan kegagalan.
3. Katalog testnet menargetkan 100% bobot executable; tipe subjektif (mis.
   opini Judge) tidak diterima sampai ada aturan cap + batas bobotnya.

## Faucet dan distribusi token tester (manual)

1. Satu-satunya sumber dana pertama adalah transfer manual founder dari alokasi
   testnet. Tidak ada faucet bot pada fase awal.
2. Drip standar: **1 MTC per alamat tester** (≈10.000 tx pada min gas `0.001umtc`).
   Satu drip per alamat; alamat + tanggal dicatat publik.
3. ZYRA tidak didistribusikan manual (supply genesis nol) — tester memperolehnya
   hanya lewat payout Miner/Judge setelah emisi dinyalakan.

## Kebijakan reset testnet

1. Testnet boleh di-reset kapan saja; token tak bernilai dan tidak ada kompensasi.
2. Reset = genesis + chain baru; alamat/key lama tidak berlaku otomatis.
3. Setiap reset diumumkan di kanal resmi dengan hash genesis dan peer baru.

## Publikasi vs privat

- **Publikasi:** chain ID, genesis SHA-256, peer/seed, manifest + allowlist ini, parameter komisi.
- **Privat:** mnemonic/private key, keyring, validator state, `.env`.
