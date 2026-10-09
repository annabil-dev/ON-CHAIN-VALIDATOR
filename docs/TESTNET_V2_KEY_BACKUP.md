# Backup Kunci Founder `mythchain-testnet-v2` (mini PC → Windows)

Yang tersimpan di chat/WA (public key, Node ID, address) **bukan** backup. Backup
harus menyimpan file private key validator + P2P, state penandatanganan, dan keyring
wallet terenkripsi. Jangan `cat` atau kirim isi file-file itu.

## 1. Persiapan di mini PC

Karena ini satu-satunya validator, hentikan proses `mythprotocold start` secara
graceful dengan `Ctrl+C` di terminal tempat ia berjalan. Node boleh dimatikan
sebentar untuk membuat backup konsisten. Jangan menghentikan node dengan menghapus
`data/` atau membuat genesis baru.

```sh
command -v gpg
NODE="$HOME/mythchain-testnet-v2-20261008-180035/validator0"
ls -ld "$NODE/keyring-file" "$NODE/config/priv_validator_key.json" \
  "$NODE/config/node_key.json" "$NODE/data/priv_validator_state.json"
```

Siapkan passphrase backup **baru** yang berbeda dari passphrase keyring wallet.
Jangan tempelkan passphrase di shell history, file TXT, atau chat.

## 2. Buat arsip terenkripsi di mini PC (tanpa USB)

Jalankan seluruh blok ini di terminal mini PC **setelah node dihentikan**. File
tar sementara hanya dibuat di `/dev/shm` (RAM) dan dihapus otomatis ketika blok
selesai. Di `~/mythchain-encrypted-backups` hanya tersisa arsip `.gpg` terenkripsi.

```sh
(
  set -euo pipefail
  umask 077
  NODE="$HOME/mythchain-testnet-v2-20261008-180035/validator0"
  ROOT="$(dirname "$NODE")"
  NAME="$(basename "$NODE")"
  BACKUP_DIR="$HOME/mythchain-encrypted-backups"
  mkdir -p -m 700 "$BACKUP_DIR"
  STAGE="$(mktemp -d /dev/shm/mythchain-key-backup.XXXXXX)"
  trap 'rm -rf -- "$STAGE"' EXIT

  tar -C "$ROOT" -czf "$STAGE/founder-keys.tar.gz" \
    "$NAME/config/priv_validator_key.json" \
    "$NAME/config/node_key.json" \
    "$NAME/data/priv_validator_state.json" \
    "$NAME/keyring-file" \
    "$NAME/config/genesis.json" \
    "$NAME/config/genesis.sha256" \
    "$NAME/config/config.toml" \
    "$NAME/config/app.toml" \
    founder-manifest.json

  BACKUP="$BACKUP_DIR/mythchain-testnet-v2-founder-$(date +%Y%m%d-%H%M%S).tar.gz.gpg"
  gpg --pinentry-mode loopback --symmetric --cipher-algo AES256 \
    --output "$BACKUP" "$STAGE/founder-keys.tar.gz"
  test -s "$BACKUP"
  sha256sum "$BACKUP"
  printf 'Backup terenkripsi dibuat: %s\n' "$BACKUP"
)
```

`gpg` akan meminta passphrase backup; ketik hanya di mini PC. Untuk pemulihan,
kamu juga membutuhkan passphrase keyring wallet yang dipakai saat `founder-init`.
Simpan kedua passphrase di password manager/offline yang terpisah dari file `.gpg`.

## 3. Salin hanya file `.gpg` ke Windows

Di PowerShell Windows, ganti `NAMA_BACKUP_BENER` dengan nama file `.gpg` dari
output langkah 2. Jangan salin folder `keyring-file` atau file `.json` private
secara terpisah.

```powershell
$kh = "$env:LOCALAPPDATA\Temp\opencode\mythchain_ssh_known_hosts"
$backupDir = Join-Path $HOME 'MythChainEncryptedBackups'
New-Item -ItemType Directory -Force -Path $backupDir | Out-Null

scp.exe -o StrictHostKeyChecking=yes -o "UserKnownHostsFile=$kh" `
  "annabilardec@192.168.1.22:/home/annabilardec/mythchain-encrypted-backups/NAMA_BACKUP_BENER.tar.gz.gpg" `
  "$backupDir\"

(Get-FileHash (Join-Path $backupDir 'NAMA_BACKUP_BENER.tar.gz.gpg') -Algorithm SHA256).Hash
```

Hash dari `Get-FileHash` harus sama dengan `sha256sum` arsip `.gpg` pada mini PC.
Hash ini untuk memverifikasi transfer arsip terenkripsi, **bukan** hash genesis.

## 4. Uji pemulihan di Windows WSL TANPA menjalankan node lain

WSL Ubuntu pada workstation sudah mempunyai `gpg`, `tar`, dan `/dev/shm`.
Masukkan nama arsip `.gpg` yang sama; ekstraksi hanya di RAM WSL dan folder
hasil tes dihapus otomatis saat blok selesai.

```sh
BACKUP="/mnt/c/Users/anabi/MythChainEncryptedBackups/NAMA_BACKUP_BENER.tar.gz.gpg"
(
  set -euo pipefail
  umask 077
  RESTORE="$(mktemp -d /dev/shm/mythchain-key-restore.XXXXXX)"
  trap 'rm -rf -- "$RESTORE"' EXIT
  gpg --pinentry-mode loopback --decrypt "$BACKUP" | tar -xz -C "$RESTORE"
  NODE="$RESTORE/validator0"
  BIN=/mnt/c/Users/anabi/AppData/Local/Temp/opencode/mythprotocold-testnet-v2-linux-amd64
  "$BIN" comet show-node-id --home "$NODE"
  "$BIN" comet show-validator --home "$NODE"
  "$BIN" keys list --keyring-backend file --home "$NODE"
)
```

Node ID yang benar: `39e6c1180ab191c363b085b47505754506af1f86`.
Alamat wallet yang benar: `myth12zfy420wyx7qc2lpl2yllhuwjngdfrkdkc9tqd`.
Pastikan output `keys list` menunjukkan akun `validator0` dengan alamat itu.
**Jangan start node dari folder restore**: dua signer dengan validator key yang sama
tidak boleh aktif sekaligus, dan state lama bisa menyebabkan double-sign.

Setelah uji berhasil, jalankan ulang proses `mythprotocold start` dari node home
asli (atau service aslinya jika nanti sudah ada) dan periksa tinggi blok naik.

## Status

Backup dua-perangkat dianggap selesai setelah file `.gpg` sudah ada di Windows,
checksum transfer cocok, dan uji pemulihan berhasil. Salinan Windows yang selalu
online **belum sama dengan media offline**; pindahkan arsip terenkripsi tersebut
ke drive eksternal bila nanti tersedia. Output `stat` permission 600 dan
screenshot public key **tidak** menggantikan backup.
