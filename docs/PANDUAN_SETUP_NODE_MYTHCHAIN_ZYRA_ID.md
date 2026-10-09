# Panduan Setup Node Mythchain + ZYRA

- **Bahasa:** Indonesia
- **Snapshot awal panduan:** Mythchain `v0.1.6-myth-phase1` + ZYRA `2.1.65`
- **Pembaruan status:** 2026-10-04; PyPI ZYRA `2.1.66`, kandidat source lokal `2.1.67` belum dirilis
- **Chain ID yang dipilih untuk jaringan pengganti:** `myth-testnet-1` (testnet)
- **Status:** operator melaporkan node `myth-testnet-1` sudah berjalan; manifest genesis/checksum dan endpoint publik masih perlu dicatat.

Panduan ini menjelaskan setup empat role: **Validator Mythchain**, **Judge ZYRA**,
**Miner ZYRA**, dan **Client ZYRA**. Satu mesin dapat menjalankan lebih dari satu
role untuk development, tetapi untuk jaringan bersama setiap role harus menggunakan
identitas, Cosmos key, dan direktori data sendiri.

## Daftar isi

- [1. Status jaringan dan nilai yang harus disediakan operator](#1-status-jaringan-dan-nilai-yang-harus-disediakan-operator)
- [1.1 Delta sejak snapshot panduan](#11-delta-sejak-snapshot-panduan)
- [1.2 Snapshot kelancaran operasional](#12-snapshot-kelancaran-operasional)
- [2. Arsitektur dan pemisahan identitas](#2-arsitektur-dan-pemisahan-identitas)
- [3. Persiapan umum ZYRA](#3-persiapan-umum-zyra)
- [4. Setup Validator Mythchain](#4-setup-validator-mythchain)
- [5. Setup Judge ZYRA](#5-setup-judge-zyra)
- [6. Setup Miner ZYRA](#6-setup-miner-zyra)
- [7. Setup Client ZYRA](#7-setup-client-zyra)
- [8. Synthetic Client producer opsional](#8-synthetic-client-producer-opsional)
- [9. Pengujian alur end-to-end](#9-pengujian-alur-end-to-end)
- [10. Troubleshooting](#10-troubleshooting)
- [11. Keamanan operasional](#11-keamanan-operasional)
- [12. Dokumen dan rilis terkait](#12-dokumen-dan-rilis-terkait)

## 1. Status jaringan dan nilai yang harus disediakan operator

Rilis Mythchain MTC `v0.1.6-myth-phase1` dan paket Python ZYRA `2.1.66` tersedia.
Kandidat source ZYRA `2.1.67` masih lokal dan **belum** dipublish. Jangan anggap
fitur kandidat itu sudah tersedia dari `pip install zyra-network`.
Snapshot awal panduan mencatat Ardecserver di block 11 dan `zyra-syntheticd` inactive.
Log operator berikutnya (3 Oktober) menunjukkan chain mencapai setidaknya block 160 dan
producer sempat restart lalu mencatat submission synthetic task; ada juga retry ketika
registrasi belum terlihat committed. Status service/height saat panduan ini diedit belum
diverifikasi ulang. Ini belum membuktikan semua peer memakai genesis identik atau endpoint
publik sudah siap. **Jangan arahkan transaksi bernilai/mainnet ke placeholder di bawah.**

Sebelum menyiapkan node-node role, operator jaringan harus menerbitkan satu manifest
kanonis berisi:

| Nilai                    | Contoh/bentuk                          | Catatan                                                                       |
| ------------------------ | -------------------------------------- | ----------------------------------------------------------------------------- |
| Chain ID                 | `myth-testnet-1`                     | Testnet, bukan mainnet produksi                                               |
| Binary Mythchain         | `v0.1.6-myth-phase1`                 | `multi-node` memaksa cadence 60 detik dan CORS wildcard saat menulis config |
| Target block cadence     | 60 detik per block                     | Config lokal`timeout_commit`; selaraskan semua validator                    |
| Genesis JSON + SHA-256   | URL + 64 digit hash                    | Semua node harus mendapat bytes genesis yang sama                             |
| P2P seed/persistent peer | `<node-id>@<host>:26656`             | Jangan menebak node ID/IP                                                     |
| RPC Mythchain            | `tcp://<rpc-host>:26657`             | Endpoint harus bisa dijangkau role yang menggunakannya                        |
| Minimum gas price        | nilai`umtc` resmi                    | Gunakan nilai manifest, bukan contoh lama`umyth`                            |
| P2P ZYRA tracker/seed    | URL tracker atau`ws://<peer>:<port>` | Gunakan endpoint yang disetujui operator ZYRA                                 |
| Alamat fee account       | `myth1...` + saldo                   | Setiap transaksi role membayar fee sesuai aturan jaringan                     |

Isi variabel berikut dari manifest sebelum menjalankan contoh:

```text
MYTH_CHAIN_ID=myth-testnet-1
MYTH_RPC=tcp://<RPC_HOST>:26657
MYTH_GENESIS_SHA256=<SHA256_GENESIS_KANONIS>
ZYRA_TRACKER_URL=<TRACKER_URL_RESMI>
ZYRA_SEED_PEER=<OPSIONAL_WS_SEED_RESMI>
```

Alamat RPC, peer, tracker, atau hash contoh bukan endpoint aktif sampai diumumkan
operator. Apabila chain ID/manifest berubah, perbarui konfigurasi semua role sebelum
transaksi.

### 1.1 Delta sejak snapshot panduan

Catatan berikut merangkum perubahan dan tes sejak snapshot ZYRA `2.1.65`:

- **Status rilis:** PyPI `2.1.66` adalah versi publik. Source kerja lokal bernomor
  `2.1.67`, sudah dibuild dan diperiksa, tetapi belum commit/push/publish sesuai keputusan
  operator. Semua fitur baru di bawah ini adalah status source lokal sampai rilis disetujui.
- **Pemisahan EVM/Celo:** kandidat lokal menghapus Web3 dari dependency paket, memindahkan
  `web3_bridge.py` ke arsip, dan menghapus handler EVM dari runtime CLI. Data alamat
  MetaMask lama masih dapat ada di wallet JSON untuk kompatibilitas, tetapi tidak dipakai
  oleh alur aktif. Direktori Hardhat lama di repo belum dipindah/diarsipkan.
- **Wallet native:** `/wallet` membaca `umtc`/`uzyra` lewat Cosmos Bank dan posisi
  delegation/unbonding lewat staking query; saldo SQLite ditampilkan terpisah sebagai
  saldo lokal. `/send` memakai Cosmos Bank, sedangkan `/stake` dan `/unstake` mendelegasikan
  atau memulai undelegation MTC (`umtc`). Undelegation mengikuti unbonding period.
- **E2E wallet (laporan operator):** `tests/test_mythchain_wallet_e2e.py` dilaporkan
  lulus `1 passed` dalam 165,57 detik untuk transfer, delegate, dan undelegate. Simpan
  chain ID dan endpoint yang dipakai bersama hasil run; output pytest yang dibagikan belum
  mencatat kedua nilai tersebut. `.env` workspace saat ini menunjuk ke `myth-testnet-1`;
  pada versi test lama guard disposable belum diterapkan, jadi jangan anggap run tersebut
  memakai chain disposable tanpa mencatat targetnya. Source test saat ini diperketat agar
  menolak chain ID selain `mythprotocol-3val-test` dan endpoint non-loopback.
- **E2E task (laporan operator):** smoke Miner melaporkan Miner race, Docker delivery,
  result commit, dan tiga transaksi Judge sampai status L1 `APPROVED`. Tes itu memakai
  relay loopback/in-process; warning `0 peers` pada tahap seeding berarti tes tersebut
  belum membuktikan propagasi artefak lewat relay jaringan produksi. Log ini juga belum
  membuktikan payout `uzyra`.
- **Status reward:** approval task tidak otomatis membuktikan reward sudah dibayar.
  Query read-only ke endpoint `myth-testnet-1` pada `.env` workspace menunjukkan
  `enable_pouw_emissions=false` dan saldo `uzyra` account module `pouw` kosong. Ini
  konsisten dengan genesis MTC-only; jangan menjanjikan payout dari hasil `APPROVED` saja.
  `/submit` normal memakai `parse_submission` untuk membuat rubric bila belum diberikan,
  lalu `task_dispatch` menurunkan kategori reward dari difficulty/profile. Smoke task ID
  `local-smoke-...` memakai `make_contract()` langsung, yang tidak membuat rubric; task
  tersebut tercatat dengan `criteria_json` dan `task_category` kosong, sehingga memang
  bukan tes payout yang valid.
- **Reset/reward genesis:** reset script `reset_mythchain_testnet_from_zero.sh` kini
  mempertahankan `myth-testnet-1` sebagai MTC-only dan memvalidasi bahwa
  `.app_state.mythprotocol.params.enable_pouw_emissions=false`. Script itu tetap
  destruktif untuk home/data yang dipilih; **jangan jalankan pada L1 aktif**. Untuk tes
  payout, source ZYRA menyediakan `scripts/prepare_disposable_reward_genesis.sh`: script
  terpisah yang hanya membuat file genesis baru ber-chain ID
  `mythprotocol-3val-test`, memakai flag builder `--enable-pouw-emissions`, dan mengubah
  port CometBFT/gRPC agar tidak bentrok dengan L1. gRPC disposable hanya bind loopback;
  dari Windows gunakan SSH local tunnel ke port gRPC disposable, bukan membuka port
  Cosmos ke LAN. Script tidak menghapus home, tidak menghentikan proses, dan tidak
  men-start validator. Akun Client/Miner/Judge ditambahkan saldonya setelah chain
  disposable start melalui transfer dari akun validator test;
  ini menjaga supply genesis, tidak membuat saldo uzyra semu. Reviewer/operator tetap
  harus memeriksa genesis/ports sebelum start. Jangan seed saldo `uzyra` module secara manual;
  emisi per block yang mengisi accounting reward pool saat parameter aktif. Env
  `MYTHCHAIN_TEST_POUW_EMISSIONS=1` hanya guard tes, bukan pengaturan chain.
  Alamat Client boleh sama dengan satu alamat Judge pada fixture lokal, karena keeper chain
  menolak Client menilai task-nya sendiri; setidaknya tiga Judge lain harus independen dari
  Client dan Miner agar quorum 3-dari-4 bisa tercapai.
- **Quorum:** weighted task memakai maksimal empat Judge dengan quorum 3-dari-4 per
  criterion. Smoke scripts lokal sudah diselaraskan dengan jumlah/quorum tersebut dan
  menolak chain ID selain disposable `mythprotocol-3val-test` serta endpoint non-loopback.

Sebelum rilis `2.1.67`, masih perlu E2E payout `reward_settled` + delta saldo `uzyra` pada
genesis disposable yang memang dibuat dengan PoUW emissions aktif. Gunakan task yang
diregistrasi lewat dispatch normal agar rubric dan reward category ada di state kanonik;
smoke langsung memakai `make_contract()` tanpa rubric/kategori tidak membuktikan payout.
Setelah itu lakukan audit akhir atas workspace Hardhat/skrip legacy. Jangan mengarahkan
tes ini ke L1 aktif.

### 1.2 Snapshot kelancaran operasional

Gunakan status berikut sebagai bukti terbatas per log/screenshot operator, bukan klaim
bahwa semua komponen jaringan sudah sehat:

| Komponen                | Bukti terbaru yang dibagikan                                                                                                                              | Kesimpulan operasional                                                                                                                                                                    |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| L1 CometBFT             | Log memperlihatkan block 157, 158, 159, 160 ter-finalisasi dengan jarak kira-kira satu menit                                                              | L1 membuat block terus; pada cuplikan tersebut`num_txs=0`, jadi cuplikan itu sendiri tidak membuktikan transaksi task ikut masuk ke block tersebut.                                     |
| Peer CometBFT           | Log berulang menunjukkan`numInPeers=0`, `numOutPeers=0`, `No addresses to dial`, fallback ke seeds                                                  | Validator yang ditampilkan belum menunjukkan peer CometBFT aktif. Pastikan ini memang fixture satu-validator; untuk jaringan multi-validator, verifikasi peer dan voting power tiap node. |
| Synthetic producer      | Unit`zyra-syntheticd` restart dan log kemudian mencatat submission task; beberapa percobaan lebih awal gagal karena registrasi belum terlihat committed | Producer hidup dan pernah submit, tetapi setiap task harus dicocokkan lewat ID/query L1. Pesan`Submitted` bukan pengganti pemeriksaan task state/tx di chain.                           |
| Windows Miner/P2P       | ZYRA`2.1.66` tersambung ke WebSocket relay dan `/mine` berulang `Task is not registered on Mythchain`                                               | P2P dapat menemukan kandidat, tetapi claim canonical gagal untuk kandidat tersebut. Periksa task ID, chain ID, gRPC/REST endpoint, dan apakah task itu stale.                             |
| Relay Producer vs Miner | Log Producer dan Windows menunjukkan URL/transport relay yang berbeda (WSS domain vs WS alamat/port langsung)                                             | Belum terbukti keduanya menuju relay yang sama. Cocokkan konfigurasi/DNS sebelum menyimpulkan task current tersebar ke Miner Windows.                                                     |
| Smoke task disposable   | Miner race, Docker delivery, result commit, tiga vote Cosmos, dan status`APPROVED` berhasil pada harness lokal                                          | Ini membuktikan alur task lokal/chain pada fixture yang dipakai; relay loopback ditutup sebelum seeding CID, jadi belum membuktikan distribusi artefak lewat relay produksi.              |

Ringkasnya: **L1 terlihat terus membuat block dan task lifecycle berhasil di smoke terisolasi, tetapi jalur Producer → relay yang sama → Windows Miner canonical claim dan payout belum terbukti end-to-end.** Jangan samakan “block terus maju”, “task muncul di P2P”, “task `APPROVED`”, dan “reward sudah settle”; masing-masing perlu bukti query yang berbeda.

## 2. Arsitektur dan pemisahan identitas

- **Validator Mythchain** menjalankan daemon Cosmos/CometBFT dan mengamankan L1
  dengan consensus key. Ini berbeda dari Judge ZYRA.
- **Judge ZYRA** menerima artefak, menjalankan acceptance checks di Docker, lalu
  mengirim vote P2P dan (untuk task chain-required) transaksi vote Cosmos.
- **Miner ZYRA** mengambil task, mengklaim lease canonical ke Mythchain, menjalankan
  Planner/Coder dan task runtime Docker, lalu mengirim CID/proof.
- **Client ZYRA** mengirim prompt/acceptance contract, memantau task, dan mengunduh
  hasil. Client dapat memakai P2P saja atau mewajibkan registrasi Mythchain.
- `zyra-syntheticd` adalah Client headless opsional. Ia bukan bagian consensus dan
  sebaiknya hanya ada satu instance per jaringan.

Gunakan identitas terpisah untuk validator consensus key, Cosmos Client, Cosmos Miner,
setiap Cosmos Judge, dan identitas P2P ZYRA. Jangan salin private key antar-role.
Wallet internal/P2P ZYRA juga bukan otomatis alamat Cosmos atau MetaMask.

## 3. Persiapan umum ZYRA

### 3.1 Windows PowerShell (umumnya Client atau Miner)

Perlu Python 3.10+; Docker Desktop diperlukan untuk Miner/Judge runtime. Buka
PowerShell:

```powershell
py -3 --version
py -3 -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install --upgrade pip
python -m pip install --upgrade zyra-network==2.1.66
zyra
```

Jika PowerShell memblokir aktivasi venv, buka sesi PowerShell sesuai kebijakan mesin
atau jalankan entry point langsung:

```powershell
.\.venv\Scripts\python.exe -m zyra_cmd.zyra_cli
```

Gunakan versi package yang dipatok ke `2.1.66` untuk mengikuti perilaku adapter yang
didokumentasikan. Perintah `/update` dapat memasang versi terbaru; jangan pakai untuk
deployment terkendali sebelum versi baru diuji terhadap binary dan jaringan target.

### 3.2 Ubuntu/Linux (umumnya Validator, Judge, atau synthetic producer)

```sh
sudo apt-get update
sudo apt-get install -y python3 python3-venv python3-pip ca-certificates
mkdir -p "$HOME/zyra-node"
cd "$HOME/zyra-node"
python3 -m venv .venv
. .venv/bin/activate
python -m pip install --upgrade pip
python -m pip install --upgrade 'zyra-network==2.1.66'
zyra --help
```

Pasang Docker Engine + Docker Buildx dari sumber yang sesuai distro untuk Judge.
Buildx juga diperlukan untuk membangun task runtime Miner. Pastikan daemon Docker
berjalan dan user yang akan menjalankan ZYRA memiliki akses Docker; setelah perubahan
grup, login ulang.

```sh
docker version
docker buildx version
docker buildx inspect --bootstrap
```

### 3.3 Identitas P2P dan jaringan

Pada peluncuran pertama, ZYRA membuat identitas dan data lokal di direktori default
user. Untuk memisahkan beberapa role pada user OS yang sama, tetapkan direktori unik
sebelum menjalankan program.

PowerShell:

```powershell
$env:ZYRA_DATA_DIR = "$HOME\.zyra-client"
$env:TRACKER_URL = "<TRACKER_URL_RESMI>"
# Opsional bila operator memberi seed peer:
$env:ZYRA_SEED_PEER = "ws://<PEER_HOST>:<P2P_PORT>"
```

Ubuntu:

```sh
export ZYRA_DATA_DIR="$HOME/.local/share/zyra-judge"
export TRACKER_URL="<TRACKER_URL_RESMI>"
# Opsional bila operator memberi seed peer:
export ZYRA_SEED_PEER="ws://<PEER_HOST>:<P2P_PORT>"
```

Simpan direktori P2P masing-masing role. Hindari menjalankan beberapa node dengan
data directory sama karena mereka akan berbagi identitas dan state lokal.

### 3.4 Akses Mythchain dari ZYRA

Mode berikut menentukan integrasi task dengan chain:

- `ZYRA_MYTHCHAIN_MODE=off` (default): task manual berjalan lewat P2P; lease/vote
  P2P bersifat advisory dan tidak menjadikan state canonical Mythchain.
- `ZYRA_MYTHCHAIN_MODE=required`: task baru harus terdaftar di L1, Miner memerlukan
  lease canonical, dan Judge harus mengonfirmasi vote L1 sebelum meneruskan vote P2P.

PyPI `2.1.66` sudah memakai native Cosmos signing melalui CosmPy untuk task Mythchain.
Kandidat source lokal `2.1.67` memutus dependency/runtime EVM dari CLI dan menambahkan
query saldo Bank, transfer native, serta delegation MTC. Kandidat `2.1.67` belum dirilis;
perintah wallet-native tersebut baru tersedia dari source kandidat.

Untuk mode `required`, adapter ZYRA menandatangani Cosmos transaction secara native
di Python menggunakan CosmPy. Miner/Client/Judge tidak memerlukan binary
`mythprotocold` atau WSL pada mesin ZYRA. Setiap role tetap perlu Cosmos key sendiri,
dan alamat dari file key harus cocok dengan alamat role.

**Penting:** `tcp://<RPC_HOST>:26657` adalah CometBFT RPC, bukan Cosmos gRPC. Untuk
native signing, arahkan ZYRA ke Cosmos gRPC (`grpc+http://<GRPC_HOST>:9090`) atau Cosmos
REST (`rest+http://<REST_HOST>:1317`, jika API diaktifkan). Contoh Linux Miner:

```sh
export ZYRA_MYTHCHAIN_MODE=required
export MYTHCHAIN_CHAIN_ID="myth-testnet-1"
export MYTHCHAIN_GRPC_ENDPOINT="grpc+http://<GRPC_HOST>:9090"
export MYTHCHAIN_MINER_ADDRESS="myth1..."
export MYTHCHAIN_MINER_MNEMONIC_FILE="$HOME/.config/zyra/miner.mnemonic"
export MYTHCHAIN_TX_FEES="<FEE_AMOUNT>umtc"
export MYTHCHAIN_GAS_LIMIT=1000000
```

Windows Miner dapat memakai endpoint gRPC yang bisa dijangkau langsung dari Windows;
tidak perlu WSL. Buat file mnemonic/private-key lokal dengan ACL hanya untuk user Miner,
jangan menaruh nilainya di environment variable, argumen command line, P2P payload,
atau repository. Adapter memverifikasi alamat hasil derivasi agar key yang salah tidak
menandatangani transaksi.

`MYTHCHAIN_TX_FEES` mengikuti fee policy jaringan. Jika kosong, adapter mencoba transaksi
gasless; transaksi berbayar harus mengatur amount seperti `1000umtc`. MTC Phase 1
menerima fee `umtc`; gunakan denom lain hanya setelah perubahan fase diumumkan. Gagal
menjangkau gRPC/REST, chain ID yang tidak cocok, saldo fee kurang, atau response tx
gagal akan menahan task dalam status non-final.

#### Wallet native pada kandidat source `2.1.67`

Kandidat lokal menyediakan:

```text
/wallet
/send 1.25 MTC myth1recipient...
/stake 10 mythvaloper1validator...
/unstake 2 mythvaloper1validator...
```

Amount memakai satuan manusia dengan maksimum enam angka desimal. `/wallet` membaca
saldo `umtc`/`uzyra` dari Cosmos Bank dan menampilkan saldo SQLite secara terpisah;
saldo lokal bukan saldo on-chain. Query saldo bersifat read-only. Bila beberapa role
diatur dalam satu proses, pilih akun yang ditampilkan dengan
`MYTHCHAIN_WALLET_ROLE=client|miner|judge`. Transaksi `/send`, `/stake`, dan `/unstake`
memerlukan Cosmos key role-specific dan fee MTC. `/unstake` memulai masa unbonding,
bukan penarikan langsung. Task PoUW dibayar otomatis oleh module saat settlement aktif;
perintah EVM `/claim` bukan jalur claim reward Mythchain.

Fitur wallet-native ini belum ada di PyPI `2.1.66`; jangan ikuti contoh tersebut dengan
paket publik itu. Uji hanya dari checkout kandidat lokal yang sudah dikonfigurasi.

## 4. Setup Validator Mythchain

Validator L1 dapat menjadi host bagi Judge atau synthetic producer, tetapi ketiganya
tetap process dan identity terpisah. Gunakan Ubuntu Linux yang selalu online, SSD,
jam sistem tersinkronisasi, dan jaringan stabil. Validator ikut consensus hanya
setelah ada di validator set aktif.

### 4.1 Install binary MTC

Contoh untuk Ubuntu amd64:

```sh
VERSION=0.1.6-myth-phase1
TAG=v0.1.6-myth-phase1
wget "https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/download/${TAG}/mythprotocold_${VERSION}_amd64.deb"
wget "https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/download/${TAG}/mythprotocold_${VERSION}_amd64.deb.sha256"
sha256sum -c "mythprotocold_${VERSION}_amd64.deb.sha256"
sudo apt install "./mythprotocold_${VERSION}_amd64.deb"
mythprotocold version
```

Untuk ARM64 ganti akhiran `_amd64.deb` menjadi `_arm64.deb`. Verifikasi checksum
paket yang diunduh; jangan memakai hash genesis lama berdenom `umyth` untuk genesis
MTC.

### 4.2 Genesis validator pertama (hanya founder/operator genesis)

Lakukan ini hanya pada host yang ditunjuk menjadi validator awal, setelah operator
menetapkan chain ID, alokasi, genesis parameter, dan keputusan reset lama. Builder
berikut membuat keyring, akun, gentx, dan genesis terkoordinasi; **jangan** menjalankan
`init`, `add-genesis-account`, atau `gentx` terpisah untuk home yang sama.
Command di bawah sengaja tidak menyalakan PoUW emissions (MTC-only default). Untuk payout
E2E, buat genesis/home disposable terpisah dengan chain ID lain dan tambahkan flag
`--enable-pouw-emissions`; jangan ubah genesis L1 yang sedang dipakai.

```sh
export MYTH_GENESIS_DIR="$HOME/myth-testnet-1-genesis"
export MYTH_CHAIN_ID="myth-testnet-1"
export MYTH_PUBLIC_IP="<IP_PUBLIK_VALIDATOR_PERTAMA>"
export FOUNDER_SELF_BOND="800000000000" # 800.000 MTC dalam umtc

if [ -e "$MYTH_GENESIS_DIR" ]; then
  echo "Refusing to overwrite existing genesis directory: $MYTH_GENESIS_DIR" >&2
  exit 1
fi

mythprotocold multi-node --v 1 \
  --output-dir "$MYTH_GENESIS_DIR" \
  --chain-id "$MYTH_CHAIN_ID" \
  --validators-stake-amount "$FOUNDER_SELF_BOND" \
  --starting-ip-address "$MYTH_PUBLIC_IP" \
  --commission-rate 0.05 \
  --commission-max-rate 0.06 \
  --commission-max-change-rate 0.01 \
  --keyring-backend file \
  --minimum-gas-prices "0.001umtc,0.001uzyra"
```

Builder akan meminta passphrase untuk keyring `file`. Gunakan nilai minimum gas price
yang disetujui manifest jika berbeda dari contoh. Hasil validator awal berada di
`$MYTH_GENESIS_DIR/validator0`.

```sh
export MYTH_HOME="$MYTH_GENESIS_DIR/validator0"
mythprotocold genesis validate-genesis --home "$MYTH_HOME"
sha256sum "$MYTH_HOME/config/genesis.json"
```

Periksa chain ID, denom `umtc`, supply, Community Pool, metadata MTC, gentx, self-bond,
komisi, dan `enable_pouw_emissions=false`. Bagikan genesis **hanya setelah disetujui**;
semua validator harus menerima file byte-identik dan hash yang sama. Pertahankan
private key, keyring, node key, consensus key, dan file validator state di host.

### 4.3 Bergabung sebagai validator tambahan

Validator tambahan tidak membuat genesis sendiri. Dapatkan binary, chain ID, genesis
kanonis, seed/persistent peer, minimum gas price, dan RPC resmi dari operator.

```sh
export MYTH_HOME="$HOME/.mythprotocol-mtc"
export MYTH_CHAIN_ID="myth-testnet-1"

mythprotocold init "<NODE_MONIKER>" --chain-id "$MYTH_CHAIN_ID" --home "$MYTH_HOME"
cp ./genesis.json "$MYTH_HOME/config/genesis.json"
sha256sum "$MYTH_HOME/config/genesis.json"
mythprotocold genesis validate-genesis --home "$MYTH_HOME"
```

Hash harus sama persis dengan manifest. Atur `$MYTH_HOME/config/config.toml` untuk
node ID seed/persistent peer resmi dan RPC sesuai kebijakan operator; setel
`minimum-gas-prices` di `config/app.toml` ke nilai resmi.

Jalankan full node dan tunggu sampai sinkron:

```sh
mythprotocold start --home "$MYTH_HOME"
```

Di terminal lain:

```sh
mythprotocold status --node "<RPC_URL>"
mythprotocold query staking validators --status bonded --node "<RPC_URL>"
```

Buat/siapkan operator Cosmos key di home keyring, isi alamat dengan MTC untuk self-
bond dan fee, lalu buat gentx staking melalui transaksi on-chain. Konsensus key tetap
berada di host validator:

```sh
mythprotocold keys add operator-01 --keyring-backend file --home "$MYTH_HOME"
mythprotocold keys show operator-01 -a --keyring-backend file --home "$MYTH_HOME"
mythprotocold tendermint show-validator --home "$MYTH_HOME"
```

Simpan mnemonic yang hanya ditampilkan saat pembuatan key di tempat offline yang
aman. Minta operator mendanai alamat operator dengan self-bond dan fee yang disetujui.
Kemudian isi `validator.json` dengan pubkey hasil command tadi:

```json
{
  "pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"<CONSENSUS_PUBKEY_BASE64>"},
  "amount": "<SELF_BOND_UMTC>umtc",
  "moniker": "<NODE_MONIKER>",
  "identity": "",
  "website": "",
  "security": "",
  "details": "",
  "commission-rate": "0.05",
  "commission-max-rate": "0.06",
  "commission-max-change-rate": "0.01",
  "min-self-delegation": "1"
}
```

Kirim transaksi setelah memastikan gas price dan minimum self-delegation sesuai
genesis/manifest:

```sh
mythprotocold tx staking create-validator ./validator.json \
  --from operator-01 --chain-id "$MYTH_CHAIN_ID" \
  --node "<RPC_URL>" --gas auto --gas-prices "<MIN_GAS_PRICE_UMTC>" --yes \
  --home "$MYTH_HOME"
```

Gunakan self-bond yang disepakati dan komisi dalam batas genesis. Periksa status transaksi dan
validator set; berhasil mengirim transaksi belum berarti validator langsung masuk
set aktif.

### 4.4 Atur target block cadence satu menit

Target operasional yang disepakati adalah kira-kira **satu block per 60 detik**. Source
`multi-node` memaksa nilai-nilai ini pada titik akhir saat menulis setiap `config.toml`,
meski config masuknya dari home lama berisi timeout lain:

```toml
timeout_commit = "1m0s"
skip_timeout_commit = false
create_empty_blocks = true
create_empty_blocks_interval = "0s"
cors_allowed_origins = ["*", ]
```

`1m0s` berarti 60 detik; tanda koma terakhir pada array CORS valid dalam TOML.
Config yang ditulis ulang ini berlaku untuk setiap node yang dibuat oleh perintah
`multi-node` pada binary `v0.1.6`.

Pada node yang sudah dibuat, binary baru tidak otomatis menimpa
`config/config.toml`. Periksa home aktual dari unit systemd lalu edit konfigurasi setiap
validator:

```toml
[consensus]
timeout_commit = "60s"
create_empty_blocks = true
create_empty_blocks_interval = "0s"
```

Edit nilai di section `[consensus]` yang sudah ada; jangan menambahkan section duplikat.
Upgrade package tidak mengubah config node aktif yang sudah dibuat sebelumnya, jadi
home validator Ardecserver tetap perlu disetel manual untuk menerapkan cadence sekarang.

Cadangkan `config.toml` dahulu. Terapkan perubahan saat restart terkoordinasi: untuk
satu validator, chain berhenti sementara selama restart; pada set beberapa validator,
ikuti prosedur rolling restart operator agar quorum tetap tersedia. Setelah restart,
bandingkan waktu block dari beberapa block berturut-turut. 60 detik adalah target/minimum
cadence, bukan jaminan waktu dinding eksak—consensus, jaringan, dan node yang tidak
sehat dapat membuat block lebih lambat.

### 4.5 Jalankan sebagai service Linux (template)

Setelah sinkron dan konfigurasi sudah diverifikasi, buat user service khusus dan unit
systemd yang sesuai lokasi home. Contoh minimal:

```ini
[Unit]
Description=Mythchain MTC Validator
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=myth
Group=myth
ExecStart=/usr/bin/mythprotocold start --home /var/lib/mythprotocol
Restart=on-failure
RestartSec=10
LimitNOFILE=65535
UMask=0077

[Install]
WantedBy=multi-user.target
```

Pastikan user service memiliki home dan file yang dibutuhkan. Validasi file unit, lalu:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now mythprotocold
sudo systemctl status mythprotocold
sudo journalctl -u mythprotocold -f
```

Sesuaikan nama unit dan home dengan instalasi. Batasi RPC ke localhost/private
network atau reverse proxy yang diamankan; jangan membuka unsafe RPC ke internet.
Buka P2P CometBFT hanya sesuai kebutuhan dan aturan firewall operator.

### 4.6 Reset state lama `umyth` (operasi terkoordinasi)

Perubahan binary/denom tidak mengonversi genesis atau saldo. Sebelum reset: hentikan
seluruh validator lama, pastikan tidak ada node lama yang tetap signing, backup genesis
dan key secara aman, tetapkan chain ID baru, lalu ikuti prosedur reset di
`MYTH_PUBLIC_VALIDATOR_ONBOARDING.md`. `unsafe-reset-all` hanya membuang state/WAL lokal;
ia tidak membuat genesis baru. Jangan jalankan pada home aktif dan jangan menghidupkan
kembali home lama setelah migrasi.

**Peringatan:** reset script lokal tetap destruktif terhadap genesis/home/data target;
jangan jalankan terhadap home/L1 aktif. Script itu sengaja menjaga emissions mati untuk
reset `myth-testnet-1`. Untuk payout E2E gunakan fixture terpisah
`my_ai/scripts/prepare_disposable_reward_genesis.sh`, yang membuat genesis baru dengan
flag builder emissions dan port terpisah tanpa menghentikan atau menghapus node aktif.
Tinjau alamat, home, dan port hasilnya sebelum memulai fixture. Jangan menambahkan
saldo/supply `uzyra` manual ke genesis.

## 5. Setup Judge ZYRA

Judge adalah process ZYRA, bukan validator konsensus. Ubuntu direkomendasikan untuk
service yang selalu hidup; Docker + Buildx menjalankan checks secara terisolasi.

### 5.1 Install dan siapkan runtime

Ikuti [Persiapan umum ZYRA](#32-ubuntu-linux-umumnya-validator-judge-atau-synthetic-producer),
lalu pastikan:

```sh
docker version
docker buildx version
docker buildx inspect --bootstrap
```

Masuk ke CLI:

```sh
export ZYRA_DATA_DIR="$HOME/.local/share/zyra-judge-1"
export TRACKER_URL="<TRACKER_URL_RESMI>"
zyra
```

Di prompt ZYRA, jalankan `/runtime` untuk menyiapkan runtime Flask/Chromium yang
digunakan oleh task checks. Build pertama dapat memerlukan internet untuk mengunduh
base image dan dependencies yang dipatok. Bila Judge membutuhkan penilaian model
lokal, jalankan `/engine`, pilih model yang sesuai, lalu `/models` untuk mengecek model
tersedia. Setelah runtime/model siap, ketik `exit`; konfigurasi key dan environment
chain-required terlebih dahulu, lalu jalankan CLI lagi.

### 5.2 Aktifkan vote canonical Mythchain

Untuk menilai task dengan `lease_mode=mythchain`, sediakan Cosmos Judge key **unik**
dan alamat akun hasil funding fee. Jangan gunakan validator consensus key, client key,
atau miner key. Contoh konfigurasi Linux untuk satu Judge:

```sh
export ZYRA_MYTHCHAIN_MODE=required
export MYTHCHAIN_CHAIN_ID="myth-testnet-1"
export MYTHCHAIN_GRPC_ENDPOINT="grpc+http://<GRPC_HOST>:9090"
export MYTHCHAIN_JUDGE_ADDRESS="myth1..."
export MYTHCHAIN_JUDGE_PRIVATE_KEY_FILE="$HOME/.config/zyra/judge-01.key"
export MYTHCHAIN_TX_FEES="<FEE_AMOUNT>umtc"
```

File key harus berisi private key secp256k1 dalam 64 digit hex. Lindungi permission,
simpan di luar repo, lalu danai alamat yang sesuai dengan fee MTC. Native adapter
memverifikasi address turunan dari key file sebelum transaksi:

```sh
chmod 600 "$MYTHCHAIN_JUDGE_PRIVATE_KEY_FILE"
```

Jalankan proses ZYRA setelah environment native signing siap, lalu masuk mode Judge:

```text
zyra
```

Di prompt ZYRA:

```text
/judge
```

Judge mengambil task/artefak dari P2P, memverifikasi acceptance contract dan attempt,
menjalankan checks di Docker, lalu mengirim vote. Pada task Mythchain, ia harus melihat
attempt canonical dan transaksi Cosmos vote terkonfirmasi sebelum vote diteruskan ke
P2P. Untuk weighted task, konfigurasi chain saat ini memerlukan **3 dari maksimal 4
Judge independen yang selaras untuk setiap criterion**; siapkan akun dan instance
terpisah. Query L1 menjadi sumber status final, bukan tampilan quorum P2P lokal.

## 6. Setup Miner ZYRA

Miner dapat berjalan di Windows atau Linux. Windows setup memakai Docker Desktop
dan Ollama; native Cosmos signing tidak memerlukan Mythchain CLI atau WSL.

### 6.1 Install AI engine dan Docker

Ikuti [Persiapan umum ZYRA Windows](#31-windows-powershell-umumnya-client-atau-miner),
install/start Ollama, kemudian buka sesi persiapan dengan `zyra` dan jalankan:

```text
/engine
/models
/runtime
```

Setelah persiapan runtime/model selesai, ketik `exit`. Set environment Miner/chain
di bawah sebelum menjalankan sesi mining.

Pilih model Planner/Coder yang sesuai kapasitas RAM/GPU. Docker Desktop harus aktif;
Buildx diperlukan untuk base image runtime dan per task. Runtime task memakai platform
yang disepakati jaringan (default dokumentasi `linux/amd64`). First build memerlukan
akses internet; Miner dan Judge harus kompatibel pada platform/base image/runtime
fingerprint yang sama.

### 6.2 Pilih jaringan P2P dan data directory unik

PowerShell:

```powershell
$env:ZYRA_DATA_DIR = "$HOME\.zyra-miner-01"
$env:TRACKER_URL = "<TRACKER_URL_RESMI>"
# Opsional:
$env:ZYRA_SEED_PEER = "ws://<PEER_HOST>:<P2P_PORT>"
```

### 6.3 Konfigurasi claim dan result melalui Mythchain

Untuk task chain-required, buat Cosmos Miner key tersendiri dan isi akun dengan fee
MTC. GRPC endpoint harus mengarah ke Cosmos gRPC service, bukan port CometBFT 26657.
Atur environment sebelum menjalankan ZYRA:

```powershell
$env:ZYRA_MYTHCHAIN_MODE = "required"
$env:MYTHCHAIN_CHAIN_ID = "myth-testnet-1"
$env:MYTHCHAIN_GRPC_ENDPOINT = "grpc+http://<ARDESERVER_HOST>:9090"
$env:MYTHCHAIN_MINER_ADDRESS = "myth1..."
$env:MYTHCHAIN_MINER_PRIVATE_KEY_FILE = "$HOME\.zyra-secrets\miner-01.key"
$env:MYTHCHAIN_TX_FEES = "<FEE_AMOUNT>umtc"
```

File private key harus berisi private key secp256k1 dalam hex (64 digit) dan hanya
bisa dibaca user Windows Miner. Simpan backup terenkripsi terpisah dan danai alamat
yang berasal dari key itu. Jangan share isi file atau mnemonic.

Adapter akan membuat claim, menunggu state committed yang menunjukkan lease/attempt
milik miner ini, baru memulai kerja. Setelah checks Miner lolos, CID/proof dikirim ke
chain. Jangan anggap task P2P sebagai izin kerja ketika mode `required` gagal query
atau claim.

### 6.4 Mulai dan pantau Miner

Di PowerShell yang sama—setelah semua environment P2P dan Mythchain di atas disiapkan—
mulai ZYRA, lalu jalankan:

```powershell
zyra
```

Di prompt ZYRA:

```text
/mine
```

Miner memantau task pending. Untuk menghentikan loop dengan aman, gunakan `Ctrl+C`.
Pantau query task L1, claim/attempt, status runtime Docker, dan pengiriman CID/proof;
jangan mengklaim native ZYRA payout bila emisi/reward pool belum aktif.

## 7. Setup Client ZYRA

Client mengirim task dan menyimpan status/hasil. Ollama tidak diwajibkan untuk
pengiriman sederhana; tanpa model lokal, ZYRA menggunakan acceptance fallback yang
didukung. Hasil disimpan lokal pada folder output.

### 7.1 Install dan mulai

Gunakan Windows atau Linux dari bagian Persiapan umum. Atur direktori unik dan tracker
resmi. Contoh Windows:

```powershell
$env:ZYRA_DATA_DIR = "$HOME\.zyra-client-01"
$env:TRACKER_URL = "<TRACKER_URL_RESMI>"
```

### 7.2 Mode P2P biasa

P2P-only mode tidak membuat task canonical di Mythchain. Di PowerShell yang sama,
jalankan CLI lalu perintah task:

```powershell
zyra
```

```text
/output "D:\ZYRA Results"
/submit Buat aplikasi kalkulator Python dengan unit tests
/tasks
```

Client tetap memantau hasil ketika aktif, dan task/result history disimpan lokal untuk
proses versi ini. Prompt, source requirements, acceptance criteria, dan informasi task
yang direlay ke peer harus dianggap data yang dibagikan ke jaringan.

### 7.3 Mode chain-required

Untuk meregistrasi task dan acceptance hash di Mythchain, buat Cosmos Client key khusus
dan beri saldo MTC untuk transaksi. Pada Windows, adapter memakai Cosmos gRPC yang
bisa dijangkau langsung:

```powershell
$env:ZYRA_MYTHCHAIN_MODE = "required"
$env:MYTHCHAIN_CHAIN_ID = "myth-testnet-1"
$env:MYTHCHAIN_GRPC_ENDPOINT = "grpc+http://<ARDESERVER_HOST>:9090"
$env:MYTHCHAIN_CLIENT_ADDRESS = "myth1..."
$env:MYTHCHAIN_CLIENT_MNEMONIC_FILE = "$HOME\.zyra-secrets\client-01.mnemonic"
$env:MYTHCHAIN_TX_FEES = "<FEE_AMOUNT>umtc"
```

Mnemonic file hanya boleh dibaca user Windows Client. Jangan share file atau menaruh
mnemonic dalam command line/environment literal. Adapter memeriksa bahwa derived address
sesuai `MYTHCHAIN_CLIENT_ADDRESS`.

Setelah mengisi environment dan membuat key di atas, mulai CLI dari PowerShell yang
sama:

```powershell
zyra
```

Di prompt CLI, `/submit` menyusun/menampilkan acceptance criteria, mengunci contract
hash, meregistrasi task ke chain, baru mengiklankannya ke P2P setelah registrasi
committed. Simpan task ID dan tx reference;
gunakan `/tasks` untuk memantau hasil.

Jika pengiriman task harus berulang tanpa operator manusia, gunakan producer headless
di bagian berikut—jangan membuat banyak Client synthetic producer berjalan bersamaan.

## 8. Synthetic Client producer opsional

`zyra-syntheticd` tersedia pada ZYRA 2.1.65 dan dirancang untuk Linux. Gunakan satu
instance untuk satu jaringan, service Cosmos Client terpisah dari validator key, serta
data directory durable. Defaultnya maksimum 50 registrasi sukses/UTC day dengan jeda
acak 5–30 menit; registrasi memakai fee dan prompt task adalah data publik chain.

Siapkan user/service dan venv khusus. Contoh asumsi venv berada di `/opt/zyra-venv`:

```sh
sudo useradd --system --home /var/lib/zyra-synthetic \
  --create-home --shell /usr/sbin/nologin zyra-synthetic
sudo mkdir -p /etc/zyra /var/lib/zyra-synthetic
sudo chown -R zyra-synthetic:zyra-synthetic /var/lib/zyra-synthetic
sudo python3 -m venv /opt/zyra-venv
sudo /opt/zyra-venv/bin/pip install 'zyra-network==2.1.66'
```

Siapkan Cosmos Client private key khusus untuk producer, bukan validator consensus
key, dan minta operator mendanai alamatnya dengan fee MTC. Key file harus berisi 64
digit hex; jangan simpan plaintext di environment file atau repository. Bagian berikut
memasukkan key ke systemd encrypted credential.

Buat `/etc/zyra/synthetic.env` dengan permission ketat dan nilai manifest:

```text
ZYRA_MYTHCHAIN_MODE=required
ZYRA_DATA_DIR=/var/lib/zyra-synthetic
ZYRA_SYNTHETIC_MAX_PER_DAY=50
ZYRA_SYNTHETIC_MIN_INTERVAL_SECONDS=300
ZYRA_SYNTHETIC_MAX_INTERVAL_SECONDS=1800
TRACKER_URL=<TRACKER_URL_RESMI>
MYTHCHAIN_CHAIN_ID=myth-testnet-1
MYTHCHAIN_GRPC_ENDPOINT=grpc+http://127.0.0.1:9090
MYTHCHAIN_CLIENT_ADDRESS=myth1...
MYTHCHAIN_TX_FEES=<FEE_AMOUNT>umtc
MYTHCHAIN_GAS_LIMIT=1000000
```

Lindungi environment file dan SQLite dengan permission service user. Simpan private-key
file sebagai systemd encrypted credential; jangan menaruh secret plaintext dalam unit,
environment file, atau shell history. Pada host dengan `systemd-creds`/
`LoadCredentialEncrypted`:

```sh
sudo chmod 600 /etc/zyra/synthetic.env
sudo install -d -m 700 /etc/credstore.encrypted
sudo install -d -m 700 /run/zyra-setup
sudoedit /run/zyra-setup/cosmos-key
sudo systemd-creds encrypt --name=cosmos-key \
  /run/zyra-setup/cosmos-key \
  /etc/credstore.encrypted/zyra-cosmos-key
sudo rm -rf /run/zyra-setup
```

Sebelum mengaktifkan unit, pastikan credential terenkripsi dapat dibuka oleh systemd
dan nama credential sama dengan `LoadCredentialEncrypted` di unit. Bila versi systemd
belum menyediakan fitur ini, ikuti prosedur secret-management host sebelum menjalankan
daemon headless.

Contoh unit service:

```ini
[Unit]
Description=ZYRA Mythchain synthetic task producer
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=zyra-synthetic
Group=zyra-synthetic
WorkingDirectory=/var/lib/zyra-synthetic
EnvironmentFile=/etc/zyra/synthetic.env
LoadCredentialEncrypted=cosmos-key:/etc/credstore.encrypted/zyra-cosmos-key
Environment=MYTHCHAIN_CLIENT_PRIVATE_KEY_FILE=%d/cosmos-key
ExecStart=/opt/zyra-venv/bin/zyra-syntheticd
Restart=on-failure
RestartSec=30
UMask=0077

[Install]
WantedBy=multi-user.target
```

Nama service account/home harus cocok dengan deployment lokal. Setelah memvalidasi
permission dan konektivitas:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now zyra-syntheticd
sudo systemctl status zyra-syntheticd
sudo journalctl -u zyra-syntheticd -f
```

Untuk melihat katalog task tanpa submit:

```sh
/opt/zyra-venv/bin/python -m zyra_cmd.synthetic_tasks --list
```

Outbox dan counter quota tersimpan di SQLite pada `ZYRA_DATA_DIR`; backup sesuai
kebijakan, jangan menghapus saat daemon aktif. Task yang sudah diregistrasi tetap
pending sampai Miner mengklaimnya.

## 9. Pengujian alur end-to-end

Mulai dari testnet yang punya genesis, RPC, P2P tracker, fee policy, dan saldo akun
yang sudah diumumkan. Jangan mulai dengan task confidential.

1. **Validator:** pastikan daemon sinkron, `chain_id` cocok, block height bertambah,
   dan validator berstatus bonded/active sesuai target.
2. **Client:** buat task kecil dengan acceptance checks jelas; konfirmasi registrasi
   committed di L1 sebelum mencari task pada P2P.
3. **Miner:** konfirmasi canonical lease dimiliki alamat Miner sebelum Docker menjalankan
   task; cek CID/proof dan transaction result committed.
4. **Judge:** gunakan akun independen, pastikan attempt/CID/hash cocok, cek report
   Docker/evidence, lalu verifikasi vote masuk ke L1.
5. **Client:** query status canonical L1 dan ambil hasil. Status lokal/P2P bukan bukti
   bahwa reward chain sudah settlement.
6. **Synthetic producer (opsional):** mulai dengan quota rendah, cek journal, task ID,
   UTC counter, fee, dan task pending; pastikan tidak ada instance kedua.

Untuk weighted task pada binary MTC yang sesuai, sediakan sampai empat Judge independen
agar tersedia tiga suara selaras per criterion. Emisi PoUW pada genesis saat ini
`false`; persetujuan task tidak berarti ada payout ZYRA.

Operator melaporkan wallet E2E berhasil (`1 passed`, 165,57 detik) untuk bank send,
delegate, dan undelegate. Task Miner smoke juga melaporkan race claim, Docker delivery,
result commit, dan final `APPROVED` setelah quorum tiga Judge. Kedua hasil tersebut belum
menampilkan bukti `reward_settled=true` atau delta `uzyra`; catat chain ID, endpoint, task
ID, dan tx hash saat pengulangan. Source test wallet kini menolak transaksi kecuali chain
ID `mythprotocol-3val-test` dan endpoint gRPC loopback cocok. Run wallet yang dilaporkan
sebelumnya tidak mencatat endpoint/chain ID, jadi jangan anggap hasilnya sebagai bukti
bahwa guard disposable baru sudah lulus.

Task smoke memakai relay in-process/loopback. Warning `0 peers` saat CID seeding berarti
ia belum menguji relay produksi atau propagasi artefak ke peer eksternal. Status L1
`APPROVED` tetap bukti task/vote canonical, bukan bukti artifact availability atau payout.

## 10. Troubleshooting

| Gejala                                 | Periksa                                                                                                                          |
| -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Native adapter tidak menemukan chain   | `MYTHCHAIN_GRPC_ENDPOINT` scheme/host/port, chain ID, gRPC/REST service, TLS, DNS/firewall; port CometBFT `26657` bukan gRPC |
| Private key/address mismatch           | Pastikan mnemonic/private-key file milik role itu benar dan derived`myth1...` address sama dengan config                       |
| `insufficient fees` / tx gagal       | Saldo`umtc`, fee manifest, `MYTHCHAIN_TX_FEES`, minimum gas price, account sequence, dan fee denoms fase chain               |
| Miner terus menunggu task              | Client telah register + commit? P2P tracker/seed benar?`ZYRA_DATA_DIR` mengarah ke jaringan yang sama?                         |
| Miner melihat task tetapi gagal claim  | Task sudah punya lease? acceptance hash sama? endpoint RPC mengarah ke chain yang benar?                                         |
| Judge menolak mulai                    | Docker daemon/Buildx tersedia? jalankan`/runtime`; model Judge siap? acceptance/attempt/CID sinkron?                           |
| Judge vote hanya terlihat di P2P       | Mode`required` dan Judge Cosmos key/address benar? transaction committed/query L1 berhasil?                                    |
| Genesis/hash tidak cocok               | Bandingkan file bytes, checksum manifest, binary, chain ID; jangan ubah genesis sesudah hash diedarkan                           |
| `umyth`/`umtc` balance tidak cocok | Itu dua genesis/denom berbeda; binary baru tidak mengonversi saldo lama                                                          |
| Synthetic daemon langsung exit         | Pastikan Linux,`ZYRA_MYTHCHAIN_MODE=required`, interval 300–1800, service credential, keyring, RPC dan fee                    |
| Task Docker build gagal                | Docker Desktop/Engine aktif, Buildx terpasang, internet tersedia untuk first build, platform dan disk mencukupi                  |

Saat mengirim log untuk debugging, hapus mnemonic, password, private key, credential,
dan informasi task privat terlebih dahulu.

## 11. Keamanan operasional

- Genesis, endpoint, chain ID, fees, seed, serta upgrade harus berasal dari manifest
  resmi; contoh placeholder bukan konfigurasi live.
- Jangan pernah mengirim mnemonic, Cosmos private key, `priv_validator_key.json`,
  node key, keyring password, atau encrypted credential ke chat/repository.
- Jangan jalankan validator dengan key consensus yang sama pada dua host bersamaan.
- Jangan reset home validator aktif. Backup dahulu dan koordinasikan penghentian seluruh
  jaringan lama sebelum migrasi genesis/chain ID.
- Jangan membuka unsafe RPC/REST ke internet. Pakai firewall, endpoint privat, dan
  endpoint read-only/reverse proxy sesuai pedoman operator.
- Gunakan service Cosmos key terpisah untuk Client, Miner, setiap Judge, dan producer;
  siapkan saldo fee minimum yang dipantau.
- Perlakukan task prompt, acceptance contract, result CID, dan metadata publik sebagai
  data permanen yang dapat dibaca peer/jaringan.
- Local ZYRA credits atau P2P quorum bukan native settlement Mythchain.
- `enable_pouw_emissions=false` mematikan emisi, bukan registrasi/claim/result/vote.

## 12. Dokumen dan rilis terkait

- [Validator onboarding/reset](MYTH_PUBLIC_VALIDATOR_ONBOARDING.md)
- [Status keseluruhan Mythchain + ZYRA](MYTHCHAIN_ZYRA_STATUS_KESELURUHAN.md)
- [Kontrak adapter ZYRA–Mythchain](ZYRA_MYTHCHAIN_ADAPTER.md)
- [Protokol task lease](TASK_LEASE_PROTOCOL.md)
- [Rilis MTC `v0.1.6-myth-phase1`](https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/tag/v0.1.6-myth-phase1)
- [Paket ZYRA `2.1.66`](https://pypi.org/project/zyra-network/2.1.66/)

**Kesimpulan:** install software saja belum berarti node siap ikut jaringan. Validator
memerlukan genesis/network manifest yang kanonis; Judge, Miner, dan Client memerlukan
P2P endpoint resmi; mode chain-required juga memerlukan RPC, chain ID, Cosmos key
role-specific, dan saldo fee. Jalankan role pada testnet terlebih dahulu dan verifikasi
state canonical sebelum onboarding mainnet.
