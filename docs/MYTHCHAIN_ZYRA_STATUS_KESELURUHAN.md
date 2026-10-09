# Rangkuman Status Keseluruhan Mythchain + Ekosistem ZYRA

**Snapshot:** 9 Oktober 2026
**Bahasa:** Indonesia
**Tujuan:** catatan tunggal untuk status kode, rilis, tokenomics, validator, ZYRA, pengujian, dan pekerjaan tersisa — disiapkan sebagai bahan review independen oleh AI lain.

> **Inti status:** Phase 0–2 selesai, treasury 2.9 selesai (governance pool), tooling Phase 3 selesai tetapi **belum ada stable release yang dipublish**. Candidate `mythchain-testnet-v2` berjalan di mini PC; joiner WSL tersambung lewat ngrok, initial sync selesai, dan block/app hash + validator set cocok pada height 2. Candidate belum diumumkan atau onboarding dibuka; backup key, faucet/reset policy, dan web explorer masih belum siap. **`mythchain-mainnet-v1` dicadangkan untuk mainnet**.

> **Catatan untuk reviewer:** bagian ZYRA (§6, §7-ZYRA, dan baris ZYRA di tabel status) dibawa dari snapshot 2 Oktober 2026 dan belum diverifikasi ulang pada pass ini. Semua bagian L1/Mythchain mencerminkan working tree per 9 Oktober 2026.

## Daftar isi

- [Pengenalan proyek](#pengenalan-proyek)
- [Latar belakang](#latar-belakang)
- [Fundamental dan prinsip desain](#fundamental-dan-prinsip-desain)
- [Tujuan](#tujuan)
- [Cara kerja end-to-end](#cara-kerja-end-to-end)
- [Permintaan review](#permintaan-review)
- [1. Status cepat](#1-status-cepat)
- [2. Riwayat versi dan rilis](#2-riwayat-versi-dan-rilis)
- [3. Arsitektur sistem](#3-arsitektur-sistem)
- [4. Tokenomics dan state genesis](#4-tokenomics-dan-state-genesis)
- [5. Genesis, chain ID, dan endpoint jaringan](#5-genesis-chain-id-dan-endpoint-jaringan)
- [6. Ekosistem ZYRA dan synthetic producer](#6-ekosistem-zyra-dan-synthetic-producer)
- [7. Riwayat rilis dan commit](#7-riwayat-rilis-dan-commit)
- [8. Pengujian dan batas klaim](#8-pengujian-dan-batas-klaim)
- [9. Pekerjaan berikutnya](#9-pekerjaan-berikutnya)
- [10. File dan artefak utama](#10-file-dan-artefak-utama)
- [Kesimpulan](#kesimpulan)

## Pengenalan proyek

**Mythchain** adalah Layer-1 (L1) aplikasi khusus berbasis Cosmos SDK (`v0.55.0`) dan CometBFT (`v0.40.0`), modul `mythprotocol`. Di sistem ini blockchain menjadi tempat pencatatan canonical untuk staking MTC, biaya transaksi, registrasi task PoUW, klaim Miner, pengiriman bukti hasil, serta keputusan Judge. Chain menyimpan state final; relay P2P membantu pesan dan artefak berpindah antarpeserta.

**ZYRA** adalah lapisan aplikasi decentralized-AI yang mengoordinasikan Client, Miner, dan Judge. Client menyusun dan mengirim task, Miner mengerjakan task dengan Planner/Coder lokal, lalu Judge menjalankan acceptance checks secara independen. ZYRA menggunakan P2P untuk penemuan dan penyebaran task; Mythchain menetapkan state canonical untuk task yang memakai mode chain-required.

Peran token dipisahkan: **MTC (`umtc`)** adalah aset dasar untuk staking dan fee L1, sedangkan **ZYRA (`uzyra`)** adalah token reward PoUW yang emisinya dikendalikan parameter chain dan ketersediaan reward pool.

## Latar belakang

Alur task yang hanya mengandalkan gossip/P2P dapat berbeda antar-peer ketika claim datang terlambat atau jaringan terpartisi. Dua Miner dapat sama-sama mengira mereka memenangkan task; hasil dan vote juga dapat tidak seragam. Dokumen protokol ZYRA / Mythchain merancang Mythchain sebagai otoritas untuk registrasi task, acceptance hash, lease berdasarkan urutan transaksi di block, attempt ID, CID/proof commitment, dan vote.

PoUW kemudian menghubungkan kerja AI yang bisa diperiksa—bukan sekadar transaksi token—dengan reward. Agar network economics tidak langsung aktif sebelum tooling, validator, dan prosedur upgrade siap, MTC Phase 1 tetap menjadi fokus lebih dahulu dan emisi ZYRA diset nonaktif pada genesis.

## Fundamental dan prinsip desain

1. **State chain adalah sumber kebenaran.** P2P menyebarkan task dan hasil, tetapi tidak menentukan lease atau persetujuan canonical.
2. **Acceptance dikunci sebelum mining.** Task memiliki acceptance hash; weighted task menyimpan criteria JSON beserta checks executable.
3. **Claim mendahului kerja.** Miner harus mendapatkan lease/attempt yang committed dari chain sebelum Swarm mengerjakan task.
4. **Judge memverifikasi artefak secara terpisah.** ZIP/CID diambil dan diuji dalam Docker runtime. Vote P2P bersifat informasi sampai transaksi Cosmos Judge tercatat.
5. **Denom dan tokenomics eksplisit.** MTC adalah bond/base-fee denom; ZYRA memiliki aturan emisi dan payout tersendiri. Emisi dimatikan pada genesis MTC.
6. **Identitas dipisahkan.** Validator consensus key tidak dipakai sebagai Client, Miner, Judge, atau synthetic producer account. Founder wallet, validator consensus key, node P2P key, dan kontrol governance adalah peran berbeda.
7. **Synthetic source tidak berada di consensus loop.** `zyra-syntheticd` berjalan sebagai satu sidecar task producer dan mendaftarkan task lewat transaksi; tiap validator tidak membuat task acak sendiri-sendiri.
8. **Genesis hanya dibuat founder; node lain join dengan genesis resmi terverifikasi.** Perintah generik Cosmos `init` dimatikan di binary; `multi-node` hanya untuk testnet lokal disposable.
9. **Treasury milik protokol, bukan perorangan.** 20 juta MTC berada di akun modul `distribution`/Community Pool yang dikontrol governance, tanpa private key founder.

## Tujuan

- Menyediakan L1 PoS berdenom MTC dengan suplai genesis tetap dan Treasury yang dikendalikan governance.
- Mengurangi klaim Miner ganda dengan lease dan urutan block Mythchain.
- Mengikat kerja ke acceptance contract yang eksplisit, artefak, dan bukti hasil.
- Memungkinkan beberapa Judge menilai task dengan evidence dan weighted criteria.
- Mengoperasikan synthetic task stream yang bervariasi untuk menguji Client, P2P, Miner, Judge, dan L1 dalam kondisi berulang.
- Mengaktifkan emisi/payout ZYRA hanya setelah stabilitas, ekonomi, tata kelola, dan prosedur upgrade dinyatakan siap.
- **Target berurutan:** public testnet `mythchain-testnet-v2` dulu untuk mencari user/tester (token tak bernilai, boleh reset), lalu mainnet `mythchain-mainnet-v1`.

## Cara kerja end-to-end

### A. Alur task ZYRA (tidak berubah dari desain)

1. **Task source:** Client manusia mengirim `/submit`, atau daemon `zyra-syntheticd` memilih template dari katalog 24 task yang telah ditinjau.
2. **Registrasi:** Client/producer menghitung acceptance hash, menetapkan difficulty dan kategori reward, lalu mengirim registrasi ke Mythchain. Pada mode `ZYRA_MYTHCHAIN_MODE=required`, task baru dipublikasikan ke P2P setelah query L1 mengonfirmasi registrasi committed.
3. **Penyebaran:** P2P relay meneruskan payload prompt/criteria kepada Miner.
4. **Mining:** Miner memantau task, mengirim claim ke L1, menunggu canonical lease, lalu menjalankan Planner/Coder Swarm di Docker; CID/proof commitment dikirim ke chain.
5. **Judging:** Judge mengambil artefak/CID, menjalankan checks terisolasi di Docker/Chromium, lalu menandatangani vote Cosmos (weighted task: 3 vote selaras dari maksimal 4 Judge).
6. **Finalitas/reward:** Query L1 menentukan APPROVED/REJECTED dan skor. Pembayaran ZYRA memerlukan emisi/reward pool; `enable_pouw_emissions=false` berarti task dapat diuji tanpa mencetak ZYRA baru.

### B. Alur node Mythchain (state per 8 Oktober 2026)

```text
Founder (mini PC, Ubuntu):
  founder-init --chain-id X --confirm-chain-id X
    --output-dir <fresh-dir> --external-address <host:26656>
      → candidate genesis + file keyring + validator key
      → node identity + config + genesis/checksum + manifest
      → treasury-audit → build binary yang di-pin ke hash genesis
      → start founder

Joiner (Linux/Windows/Debian/macOS):
  installer (sh/ps1/deb, versi eksplisit atau `latest`)
      → init-node --genesis <official genesis>
      → join --chain-id X --persistent-peers ID@host:26656
      → start
```

Setiap command menutup dengan perintah next-step yang eksplisit, jadi joiner tinggal copy-paste (`init-node` → `join` → `start`).

## Permintaan review

Kepada AI reviewer, founder meminta penilaian independen atas:

1. **Kesiapan public testnet** `mythchain-testnet-v2`: apakah urutan genesis → pinned release → manifest → joiner sudah aman dan lengkap?
2. **Desain ceremony founder**: apakah larangan overwrite, validasi chain ID ganda, audit treasury, dan pin hash binary cukup sebagai pengaman human-error?
3. **Risiko operasional tunnel ngrok free** (endpoint berubah tiap restart, bandwidth/limit, relay pihak ketiga) untuk testnet publik sementara.
4. **UX joiner**: apakah installer + next-step output + panduan per-OS cukup sederhana untuk user non-teknis?
5. **Konsistensi tokenomics**: 21jt supply, 1jt founder (800rb bond + 200rb cair), 20jt treasury governance — apakah penegakannya di kode/genesis/test sudah konsisten?
6. **Hal yang masih terbuka** (§9): mana yang blocker peluncuran vs yang bisa menyusul?

## 1. Status cepat

| Bagian | Status terkini |
|---|---|
| Fase roadmap | Phase 0–2 selesai; 2.9 treasury selesai (governance pool); tooling Phase 3 selesai, **belum publish stable release** |
| Binary / denom | `umtc` (display `mtc`, simbol `MTC`); toolchain Go di WSL |
| Chain ID aktif | `mythchain-testnet-v2` = candidate berjalan; founder→WSL join/sync lulus; onboarding publik belum dibuka. `mythchain-mainnet-v1` = dicadangkan |
| Genesis di repo | `release/genesis.json` = `chain-rlkh6n`, height awal 1, SHA-256 `55f420c9…fb41f` — artefak dev/rilis lama, **bukan** genesis testnet/mainnet |
| Founder tercatat di genesis repo | `myth1csxr7p5tc4e6lvlh7tfeyud8ncjfyrtlhq8guu` (1jt MTC) — perlu konfirmasi ulang untuk genesis baru |
| Treasury | `myth1jv65s3grqf6v6jl3dp4t6c9t9rk99cd86qepld` (akun modul `distribution`), 20jt MTC + Community Pool 20jt MTC, otoritas `x/gov` |
| Validator pertama (rencana) | Mini-PC Ardecserver, Ubuntu; P2P publik via tunnel ngrok sementara (direct inbound tidak tersedia: router kos tanpa akses admin, UPnP tidak ada) |
| Target block cadence | Satu block per menit (`timeout_commit` 60 detik; test lokal memakai nilai pendek) |
| Self-bond yang diinginkan | 800.000 MTC (`800000000000umtc`); 200.000 MTC cair |
| Komisi gentx | 5% rate, 6% maksimum, perubahan maksimum 1 poin persentase |
| Min gas default | `0.001umtc` (dirender nyata di `app.toml`, bukan template) |
| ZYRA client/package | PyPI `zyra-network==2.1.65` (snapshot 2 Okt; belum diverifikasi ulang) |
| Synthetic task producer | `zyra-syntheticd` (snapshot 2 Okt; belum diverifikasi ulang) |
| Emisi PoUW | Tetap `false` pada genesis. Task bisa diuji, tetapi jangan harapkan emisi ZYRA baru |
| Installer | `install_mythprotocold.sh` (Linux/macOS) + `.ps1` (Windows): verifikasi checksum, dukung versi eksplisit/`latest`, wizard init/join, next-step output |
| Paket per-OS/role | `release-packages/{linux,windows,debian,macos}/{founder,joiner}/INSTALL.md` = panduan wizard-first; bundle fisik menunggu genesis testnet |

## 2. Riwayat versi dan rilis

### Mythchain

| Versi | Perubahan utama |
|---|---|
| `v0.1.0-myth-phase1` s.d. `v0.1.2-myth-phase1` | Rilis pengembangan awal; komisi gentx menjadi 5%/6%/1%; denom masih `umyth`. |
| `v0.1.3-myth-phase1` | Denom bond/base fee `umtc`, metadata display `mtc`/simbol `MTC`; genesis fresh wajib. |
| `v0.1.4-myth-phase1` | CometBFT `timeout_commit=60s`, empty blocks aktif. |
| `v0.1.5-myth-phase1` | Tag ada, CI gagal assertion TOML; aset tidak dipublikasikan. |
| `v0.1.6-myth-phase1` | `multi-node` menulis timeout 60 detik + CORS `*`; rilis CI lulus. |
| working tree Okt 2026 (belum ditag) | `founder-init` kandidat genesis; `init` generik dimatikan; filter peer-ID ABCI; manifest local-testnet; `treasury-audit`; pipeline stable multi-OS + pin hash genesis; installer `latest` + next-step; paket per-OS/role; perbaikan `app.toml` init-node. **Belum ada tag stabil** — public testnet menunggu tag + genesis `mythchain-testnet-v2`. |

Rilis lama: [GitHub releases](https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases). Helper rilis stabil: `python release_mythchain.py vX.Y.Z` (worktree bersih, `gh` terautentikasi, konfirmasi sebelum push tag; workflow menjalankan verify+test+vet, build 6 target, `SHA256SUMS`, publish stable bukan prerelease).

### ZYRA (snapshot 2 Okt; belum diverifikasi ulang)

| Versi | Perubahan utama |
|---|---|
| `zyra-network 2.1.62` | Adapter Mythchain weighted criteria, canonical claim/result/vote, peran Client/Miner/Judge, Miner polling. |
| `2.1.63` | Dispatch/claim-release, metadata reward category, alur task berbasis Mythchain. |
| `2.1.64` | Feeder synthetic manual sisi Client; digantikan sidecar di v2.1.65. |
| `2.1.65` | Daemon `zyra-syntheticd` terpisah, outbox/quota persisten, keyring file headless, pesan Docker Buildx lebih jelas. Terbit di PyPI. |

## 3. Arsitektur sistem

```text
Founder (mini PC Ubuntu, dial-out via ngrok TCP sementara)
  mythprotocold founder validator (chain mythchain-testnet-v2, setelah dibuat)
  zyra-syntheticd (akun Client Cosmos khusus, bukan consensus key)
        │ register task di L1 + broadcast via P2P
        ▼
P2P relay (CometBFT; peer allowlist = persistent_peers/seeds terkonfigurasi)
        │
        ├── Miner (Windows): claim → Swarm/Ollama/Docker → CID/proof ke chain
        └── Judge (Ubuntu): checks terisolasi → vote Cosmos ke L1
                 ▼
Mythchain L1: lease, result CID/proof, weighted criteria, status Judge canonical
```

- **Mythchain adalah sumber canonical** untuk task registration, lease/claim, attempt ID, result, Judge vote, dan status akhir.
- **P2P adalah transport/discovery**, bukan penentu final. Node joiner yang di-init menolak peer ID di luar allowlist.
- **Identitas dipisahkan**: founder wallet, validator consensus key, node P2P key, dan kontrol governance berbeda. Jangan salin validator consensus key sebagai key aplikasi.
- Prompt/task di L1 adalah data publik/permanen. Hindari rahasia atau data pribadi.
- **Tidak ada relay bawaan di repo.** Relay yang dipakai rehearsal adalah layanan ngrok (tunnel keluar dari mini PC); repo tidak menyimpan server/authtoken relay.

## 4. Tokenomics dan state genesis

### Base token

- Simbol: **MTC**; denom base/bond/fee: **`umtc`**; presisi **1 MTC = 1.000.000 umtc**.
- Genesis supply: **21.000.000 MTC** (`21000000000000umtc`), tanpa emisi MTC (inflasi nol, max supply ditegakkan).
- Alokasi founder: **1.000.000 MTC** (800.000 self-bond + 200.000 cair).
- Community Pool/Treasury: **20.000.000 MTC** di akun modul `distribution`, dibelanjakan via `MsgCommunityPoolSpend` otoritas `x/gov`.
- Metadata: base `umtc`, display `mtc`, simbol `MTC`.
- Min gas default builder/app: `0.001umtc` (+ `0.001uzyra` tertera; ante Phase 1 menerima MTC, fee ZYRA di-gate sampai aktivasi PoUW).
- Komisi founder default: 5% / 6% / 1pp.
- Penegakan: konstanta di `x/mythprotocol/types/keys.go`, guard supply di builder genesis, `treasury-audit`, dan unit test alokasi + otoritas spend.

### ZYRA / PoUW (snapshot 2 Okt; belum diverifikasi ulang)

- ZYRA: `uzyra`; suplai genesis nol; cap **21.000.000 ZYRA**; `enable_pouw_emissions=false` mematikan emisi, bukan registrasi/claim/vote.
- Aturan saat aktif: 0,2 ZYRA per committed block; −25% tiap 3 juta kumulatif; stop di cap.
- Base reward: Light 0,10; Medium 0,25; Heavy 0,50; Very Heavy 1,00. Split 60% Miner / 40% pool Judge selaras; gagal = 0 dan boleh retry.
- Gasless: maks 1 tx/menit saat saldo nol di semua denom fee aktif; saldo positif mencabut permanen.

## 5. Genesis, chain ID, dan endpoint jaringan

- `chain-rlkh6n` (`release/genesis.json`, hash `55f420c9…fb41f`): artefak dev lama. Jangan dipakai untuk testnet/mainnet baru.
- Kandidat lama `mythchain-testnet-v2` di `/home/annabilardec/mythchain-testnet-v2-20261008-170203` memakai `TEMP-NGROK-HOST:TEMP-NGROK-PORT` dan binary probe lama; jangan start, publish, atau pakai untuk explorer.
- Kandidat baru `mythchain-testnet-v2` di `/home/annabilardec/mythchain-testnet-v2-20261008-180035`: SHA-256 `0e9c91fb…d6e3cf`, founder `myth12zfy…fvg80c4`, node ID `39e6c118…35486`, external P2P `0.tcp.ap.ngrok.io:29687`, min gas `0.001umtc,0.001uzyra`. Mini PC lulus genesis validation + treasury audit; binary hash-pinned dipindah dan founder berjalan. Joiner WSL tersambung via ngrok dan initial sync selesai di height 25; block/app hash serta validator set cocok pada height 2. Candidate belum diumumkan/disetujui; backup key dan faucet/reset policy masih pending.
- `mythchain-mainnet-v1`: dicadangkan; tidak dipakai sampai ceremony mainnet.
- Mini PC di kos: LAN `192.168.1.22` reachable; WAN direct `114.12.14.144:26656` timeout; router tanpa akses admin; UPnP tidak ada (`No IGD UPnP Device found`).
- Keputusan: tanpa budget VPS → rehearsal dan testnet awal lewat **ngrok TCP free**. Terbukti: dummy HTTP 200 + handshake P2P CometBFT asli dan sync selesai. Keterbatasan: endpoint berubah tiap restart agent, bandwidth/limit free tier, relay pihak ketiga — tidak cocok untuk mainnet.
- IP Tailscale (`100.103.162.72`) privat dan butuh tailnet. **Tailscale Funnel** berbeda: URL `*.ts.net` HTTPS publik bisa diakses siapapun tanpa join Tailnet; tidak cocok untuk P2P raw TCP `26656` tetapi bisa digunakan sebagai endpoint baca-blok website melalui proxy RPC baca-saja.
- Web `mythchain.pages.dev`: source `mythchain_web/` sebelumnya hardcode domain `trycloudflare.com` yang berubah-ubah. Sekarang UI memakai `/api/chain/status` dan `/api/chain/block`, Pages Function hanya meneruskan request GET baca-saja ke `MYTHCHAIN_RPC_ORIGIN` HTTPS stabil (`*.ts.net` via Funnel → proxy loopback `127.0.0.1:8765` → CometBFT RPC lokal); chain ID `mythchain-testnet-v2` diperiksa. **Belum dideploy atau ditautkan ke RPC testnet live.** P2P joiner tetap lewat ngrok terpisah.
- Rehearsal disposable `mythchain-ngrok-smoke-20261008` (BUKAN testnet publik): genesis `9add40e7…793b2`, founder Node ID `5c36184e…35436`, joiner WSL `9dd89b84…745e`; sync selesai di height 259; hash height 10 identik di kedua sisi (block `601ACAB5…`, app `14F640FA…`, validator `D042DA65…`, power 800000). Founder + ngrok sudah dihentikan setelah verifikasi.

## 6. Ekosistem ZYRA dan synthetic producer (snapshot 2 Okt; belum diverifikasi ulang)

- `zyra-syntheticd` sidecar terpisah (bukan consensus): katalog 24 task (6 per kategori), maks 50 registrasi/hari UTC, jeda acak 5–30 menit, SQLite quota/outbox, satu instance per network.
- Operasi role: Judge Ubuntu (Docker Engine + Buildx, RPC `tcp://127.0.0.1:26657`, key Judge unik), Miner Windows (Docker Desktop + Buildx, Ollama, RPC reachable, key Miner funded), producer (keyring file terpisah, fee `umtc`).
- Alur: `zyra-syntheticd` → register di L1 (fee MTC) → broadcast P2P → Miner claim lease → Swarm kerjakan di Docker → CID/proof ke chain → Judges checks terpisah → votes L1 → status/score L1.
- Dengan `enable_pouw_emissions=false`, alur task tetap jalan tetapi tidak ada emisi ZYRA baru.

## 7. Riwayat rilis dan commit

### Mythchain

Riwayat `v0.1.0`–`v0.1.6-myth-phase1` tercatat di §2. Setelah itu (working tree, belum ditag): implementasi `founder-init` kandidat, penonaktifan `init` generik, restricted `genesis validate`, manifest local-testnet, `treasury-audit` + test otoritas/signature, pipeline rilis stabil multi-OS dengan pin hash genesis, installer `latest` + next-step output, panduan per-OS/role, dan perbaikan render `app.toml`.

### ZYRA

`2.1.62`–`2.1.65` tercatat di §2 (snapshot 2 Okt; belum diverifikasi ulang).

## 8. Pengujian dan batas klaim

- `go test ./...`, `go vet ./...`, `go mod verify`: lulus (WSL).
- CLI: `init-node` → `join` smoke lulus; `genesis validate` lulus; `treasury-audit` lulus pada genesis kanonik.
- Build: 6 target OS/arsitektur + `.deb` amd64/arm64; `sha256sum -c SHA256SUMS` lulus; installer Linux diuji pada fixture lokal; parse PS lolos; pin-hash linker diuji menolak genesis salah sebelum membuat home.
- Jaringan: two-node lokal sinkron ke height 100 (hash/state/validator set sama, restart identitas stabil); rehearsal ngrok lintas mesin sinkron selesai + hash height 10 identik dua sisi.
- **Batas klaim (belum dilakukan):** candidate `mythchain-testnet-v2` sudah start dan WSL joiner initial-sync lulus via ngrok, tetapi backup keyring/validator offline belum rehearsal; uji dari network eksternal terpisah, stable tag, dan public web Funnel belum dilakukan; bundle fisik per-OS/role belum dirakit; macOS signing/notarization belum ada; ZYRA e2e Miner↔Judge pasca-perubahan belum diuji ulang.

## 9. Pekerjaan berikutnya

1. Backup encrypted founder keyring + validator/node keys offline, lalu umumkan hanya setelah policy/reset/faucet endpoint ditentukan.
2. Founder→WSL sync dan hash/validator comparison pada height bersama sudah lulus; monitor dengan endpoint ngrok yang aktif.
3. Setelah backup + policy testnet lulus, rakit bundle per-OS/role + publish manifest (genesis URL/hash, peers/seeds, RPC, fee, reset policy).
4. Aktifkan RPC read-only via Funnel dan verifikasi chain ID/height/block hash dari browser eksternal sebelum membuka pendaftaran tester.
5. Operasional: `systemd` untuk node + ngrok saat testnet live; prosedur refresh manifest tiap endpoint ngrok berganti; monitoring height/peer/signing; kebijakan reset testnet.
6. Setelah testnet stabil: ceremony `mythchain-mainnet-v1` dengan endpoint permanen (bukan ngrok free).

## 10. File dan artefak utama

- Source: repo `annabil-dev/ON-CHAIN-VALIDATOR` (`D:\Semester 5\AI\mythchain\mythprotocol`), modul `mythprotocol`.
- Status roadmap: `A Progres Project/MythChain Progress Documentation/MYTHCHAIN_ROADMAP_PROGRESS.md`.
- Ceremony: `docs/PRODUCTION_GENESIS_CEREMONY_CHECKLIST.md`; panduan: `docs/PRODUCTION_{USER,VALIDATOR,FOUNDER}_GUIDE.md`; keamanan: `docs/SECURITY_POLICY.md`.
- Tokenomics: `docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md`; onboarding: `docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md`; adapter: `docs/ZYRA_MYTHCHAIN_ADAPTER.md`; setup ZYRA: `docs/PANDUAN_SETUP_NODE_MYTHCHAIN_ZYRA_ID.md`.
- Rilis: `build_validator_release.sh`, `release_mythchain.py`, `.github/workflows/release.yml`; installer: `install_mythprotocold.sh`, `install_mythprotocold.ps1`; paket: `release-packages/{linux,windows,debian,macos}/{founder,joiner}/`.
- Website explorer: repo `mythchain_web/`, Cloudflare Pages Functions `functions/api/chain/[endpoint].js`, upstream HTTPS publik diatur via `MYTHCHAIN_RPC_ORIGIN`; proxy loopback pada mini PC: `explorer_readonly_proxy.py`.
- Perintah kunci: `founder-init`, `init-node`, `join`, `start`, `treasury-audit`, `genesis validate`, `comet show-node-id/show-validator`.

---

## Kesimpulan

Mythchain sekarang punya fondasi L1 lengkap dan rehearsal lintas mesin yang lolos. Candidate `mythchain-testnet-v2` terbaru memakai endpoint ngrok numerik tetapi belum release-pinned, dijalankan, atau disetujui. Jalur web ke RPC stabil sudah disiapkan tapi belum aktif. Belum ada artefak publik yang disetujui (stable tag + manifest + endpoint explorer). Bagian ZYRA perlu snapshot ulang sebelum review end-to-end dianggap final.
