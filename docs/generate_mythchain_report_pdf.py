"""Generate the polished Indonesian Mythchain progress report PDF."""

from pathlib import Path

from playwright.sync_api import sync_playwright


OUTPUT = Path(__file__).with_name("MYTHCHAIN_RANGKUMAN_PERUBAHAN_PROGRESS_ID.pdf")

HTML = r"""<!doctype html>
<html lang="id">
<head>
<meta charset="utf-8">
<title>Mythchain — Rangkuman Perubahan &amp; Progress</title>
<style>
  @page { size: A4; margin: 18mm 17mm 18mm; }
  * { box-sizing: border-box; }
  body { margin: 0; color: #172033; font: 10.3pt/1.52 "Segoe UI", Arial, sans-serif; }
  h1, h2, h3, p { margin-top: 0; }
  h1 { font-size: 29pt; line-height: 1.13; letter-spacing: -0.6pt; margin: 0 0 12pt; }
  h2 { color: #123b58; font-size: 17pt; margin: 0 0 12pt; padding-bottom: 6pt; border-bottom: 2px solid #24a89a; }
  h3 { color: #176a72; font-size: 11.5pt; margin: 12pt 0 4pt; }
  p { margin-bottom: 7pt; }
  ul, ol { padding-left: 18pt; margin: 4pt 0 10pt; }
  li { margin: 0 0 4pt; }
  code { font: 8.5pt "Consolas", "Courier New", monospace; color: #174b62; }
  a { color: #087b80; text-decoration: none; }
  .cover { min-height: 245mm; color: white; padding: 20mm 15mm; background: linear-gradient(145deg,#102b46,#135b68 68%,#168e83); position: relative; page-break-after: always; }
  .eyebrow { color: #a8eee0; text-transform: uppercase; letter-spacing: 2px; font-size: 9pt; font-weight: 700; margin-bottom: 22mm; }
  .cover h1 { max-width: 145mm; font-size: 34pt; }
  .cover .subtitle { color: #dbf5f1; font-size: 15pt; max-width: 145mm; margin-bottom: 20mm; }
  .cover .rule { width: 35mm; border-top: 4px solid #54dbc0; margin: 11mm 0; }
  .cover .covermeta { color: #d0e8ee; position: absolute; left: 15mm; bottom: 24mm; font-size: 10pt; }
  .cover .version { display: inline-block; border: 1px solid #8bded1; border-radius: 20px; padding: 5px 12px; color: #c7fff0; font-size: 9pt; font-weight: 700; }
  .page { page-break-before: always; }
  .lead { color: #46566c; font-size: 11.5pt; }
  .status { border-left: 4px solid #df9e27; background: #fff7e7; padding: 10pt 12pt; margin: 12pt 0; }
  .note { border-left: 4px solid #218f86; background: #eaf7f5; padding: 9pt 12pt; margin: 10pt 0; }
  .alert { border-left: 4px solid #d06554; background: #fff0ed; padding: 9pt 12pt; margin: 10pt 0; }
  .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8pt; margin: 12pt 0 16pt; }
  .card { border: 1px solid #d8e2e8; border-radius: 8px; padding: 10pt; background: #fbfdfe; }
  .card .value { color: #0e7778; font-size: 18pt; line-height: 1.15; font-weight: 700; }
  .card .label { color: #536579; font-size: 8.5pt; margin-top: 3pt; }
  table { width: 100%; border-collapse: collapse; margin: 8pt 0 13pt; font-size: 9pt; }
  th { background: #153e5a; color: white; text-align: left; padding: 7pt; }
  td { border-bottom: 1px solid #dbe4e8; padding: 6.5pt 7pt; vertical-align: top; }
  tr:nth-child(even) td { background: #f3f8fa; }
  .tag { white-space: nowrap; color: #0b7777; font-weight: 700; }
  .small { color: #5f6f7f; font-size: 8.6pt; }
  .hash { overflow-wrap: anywhere; font: 8.2pt "Consolas", monospace; background: #f1f5f7; padding: 7pt; border-radius: 4px; }
  .step { display: grid; grid-template-columns: 8mm 1fr; gap: 8pt; margin: 9pt 0; }
  .stepno { width: 7mm; height: 7mm; border-radius: 50%; background: #1c8d87; color: white; text-align: center; line-height: 7mm; font-weight: 700; }
  .footer { font-size: 8pt; color: #667788; }
  @media print { .cover { -webkit-print-color-adjust: exact; print-color-adjust: exact; } }
</style>
</head>
<body>
<section class="cover">
  <div class="eyebrow">Laporan proyek · 01 Oktober 2026</div>
  <span class="version">MTC CANDIDATE · v0.1.3-myth-phase1</span>
  <div style="height:20mm"></div>
  <h1>Mythchain<br>Rangkuman Perubahan<br>&amp; Progress</h1>
  <div class="rule"></div>
  <p class="subtitle">Riwayat implementasi, rilis validator, tokenomics MTC/ZYRA, dan rencana reset genesis ke denom umtc.</p>
  <div class="covermeta">
    <strong>Status:</strong> chain lama memakai umyth; reset ke genesis MTC baru disiapkan.<br>
    <strong>Chain ID baru:</strong> perlu dipilih; jangan gunakan ulang myth-mainnet-1<br>
    <strong>Validator genesis:</strong> mini-PC founder · 114.10.44.157
  </div>
</section>

<section>
  <h2>Ringkasan eksekutif</h2>
  <p class="lead">Source Mythchain diubah dari base denom umyth ke umtc (display MTC). Rilis v0.1.3 menyiapkan binary untuk genesis baru; state chain lama tidak dapat dipakai sebagai genesis MTC.</p>
  <div class="status"><strong>Progress terkini:</strong> alokasi founder tetap 1.000.000 MTC, dengan self-bond 800.000 MTC dan 200.000 MTC cair. Genesis MTC production dan chain ID baru menunggu reset terkoordinasi.</div>
  <div class="grid">
    <div class="card"><div class="value">21 juta</div><div class="label">Batas suplai MTC · dicetak saat genesis</div></div>
    <div class="card"><div class="value">20 juta</div><div class="label">MTC Community Pool / Treasury</div></div>
    <div class="card"><div class="value">800 ribu</div><div class="label">MTC self-bond dari alokasi founder 1 juta</div></div>
    <div class="card"><div class="value">5% / 6%</div><div class="label">Komisi awal / batas maksimum · max-change 1 poin</div></div>
  </div>
  <p class="small">Sisa sekitar 200.000 MTC menjadi saldo cair founder setelah gentx diterapkan. Genesis 1 juta MTC tersebut tetap menyumbang ke suplai total; self-bond menentukan bagian yang langsung di-stake.</p>

  <h3>Versi dan rilis</h3>
  <table>
    <thead><tr><th>Versi</th><th>Perubahan / catatan</th></tr></thead>
    <tbody>
      <tr><td class="tag">v0.1.0</td><td>Rilis pengembangan awal. Workflow paket pertamanya gagal; kemudian alur rilis diperbaiki.</td></tr>
      <tr><td class="tag">v0.1.1</td><td>Paket Ubuntu amd64/arm64 dan tarball Linux tersedia. Builder saat itu mengunci komisi pada 100%.</td></tr>
      <tr><td class="tag">v0.1.2</td><td>Rilis commission-configurable; base denom lama umyth.</td></tr>
      <tr><td class="tag">v0.1.3</td><td>Kandidat rilis MTC: bond/base fee umtc, denom metadata MTC, dan tetap memakai komisi 5%/6%/1%.</td></tr>
    </tbody>
  </table>
</section>

<section class="page">
  <h2>Riwayat perubahan proyek</h2>
  <p>Commit utama pada branch <code>myth-first-phase</code> membentuk urutan kerja berikut.</p>
  <table>
    <thead><tr><th>Commit</th><th>Area</th><th>Ringkasan</th></tr></thead>
    <tbody>
      <tr><td><code>a4824f4</code></td><td>MYTH-first launch</td><td>Persiapan fase peluncuran yang berfokus pada MYTH dan konfigurasi chain.</td></tr>
      <tr><td><code>eb5cfa3</code></td><td>Rilis</td><td>Workflow GitHub Actions untuk membangun paket validator Linux.</td></tr>
      <tr><td><code>14b4857</code></td><td>Dokumentasi</td><td>Panduan pemasangan validator MYTH Phase 1.</td></tr>
      <tr><td><code>ebb2173</code></td><td>Tokenomics</td><td>Rencana tokenomics MYTH/ZYRA dan tahap aktivasi jaringan.</td></tr>
      <tr><td><code>91c4e2d</code></td><td>Genesis</td><td>Alur persiapan validator genesis founder pada mini-PC.</td></tr>
      <tr><td><code>745bf5e</code></td><td>Komisi &amp; rilis</td><td>Flag komisi validator, validasi nilai komisi, script rilis Python tanpa shell WSL manual, dan tes pada workflow rilis.</td></tr>
    </tbody>
  </table>

  <h3>Tokenomics MTC</h3>
  <ul>
    <li><strong>MTC:</strong> denom dasar <code>umtc</code>; 1 MTC = 1.000.000 umtc.</li>
    <li>Batas suplai 21.000.000 MTC, seluruhnya dialokasikan saat genesis; tidak ada inflasi/halving MTC berkelanjutan.</li>
    <li>Alokasi genesis: 1.000.000 MTC untuk akun founder dan 20.000.000 MTC ke Community Pool/Treasury.</li>
    <li>Keputusan founder: dari alokasi genesis 1.000.000 MTC, 800.000 MTC di-self-bond dan sekitar 200.000 MTC disisakan cair.</li>
    <li>Biaya Phase 1 menggunakan MTC. Distribusi Treasury memerlukan governance on-chain.</li>
  </ul>

  <h3>Tokenomics ZYRA dan PoUW</h3>
  <ul>
    <li>ZYRA menggunakan <code>uzyra</code>, suplai genesis nol, dan batas 21 juta ZYRA.</li>
    <li>Emisi PoUW dimulai dalam keadaan mati; aktivasi memerlukan koordinasi upgrade/governance.</li>
    <li>Parameter implementasi: emisi 0,2 ZYRA per committed block; penurunan 25% setiap 3 juta ZYRA kumulatif; batas akhir 21 juta ZYRA.</li>
    <li>Payout tugas yang lolos: 60% Miner dan 40% Judge pool. Quorum 3-dari-4 harus selaras untuk setiap kriteria.</li>
    <li>Setelah aktivasi, biaya dapat memakai MTC atau ZYRA; denom bond staking tetap umtc.</li>
  </ul>
</section>

<section class="page">
  <h2>Implementasi dan pengujian</h2>
  <h3>Reset genesis MTC mini-PC</h3>
  <ul>
    <li>Chain ID baru belum dipilih. Chain lama <code>myth-mainnet-1</code> memakai umyth dan jangan dipakai ulang untuk state MTC.</li>
    <li>IP publik validator founder: <code>114.10.44.157</code>.</li>
    <li>Builder <code>mythprotocold multi-node --v 1</code> akan membuat node/key baru, alokasi MTC, gentx bertanda tangan, denom metadata MTC, dan genesis bersih.</li>
    <li>Mode non-test tidak membuat saldo <code>testtoken</code> dan tidak menulis mnemonic ke <code>key_seed.json</code>.</li>
    <li>Komisi gentx: rate 5%, maksimum 6%, perubahan maksimum 1 poin persentase. Bond/base fee menggunakan umtc.</li>
  </ul>

  <h3>Hasil pengujian yang sudah tercatat</h3>
  <ul>
    <li>Source MTC lulus <code>go test ./...</code>; v0.1.2 sebelumnya menjalankan test suite yang sama sebelum perubahan denom.</li>
    <li>Genesis test MTC lokal berhasil divalidasi: suplai 21 juta umtc, Community Pool 20 juta umtc, metadata display MTC, dan gentx 800 miliar umtc.</li>
    <li>Genesis lama berdenom umyth tidak dapat digunakan pada binary MTC; validator state/genesis harus dibuat ulang dan didistribusikan konsisten.</li>
    <li>Pengujian fitur ZYRA mencakup quorum 3/4, kasus 2–2 tanpa quorum, batas gasless, dan fallback biaya adapter.</li>
    <li>Target founder tetap 1 juta MTC alokasi, 800 ribu MTC self-bond, dan sekitar 200 ribu MTC saldo cair.</li>
  </ul>

  <div class="alert"><strong>Reset data diperlukan:</strong> ganti binary saja tidak mengubah genesis atau saldo lama. Hentikan semua validator, simpan backup kunci yang diperlukan, buat genesis MTC baru, dan gunakan chain ID baru. Seluruh node harus memakai file genesis/checksum yang identik.</div>

  <h3>Checksum</h3>
  <div class="hash">Checksum baru umtc: belum dibuat. Hitung ulang setelah MTC genesis divalidasi.</div>
  <p class="small">Checksum <code>54884b60…</code> yang pernah diberikan operator adalah untuk file genesis umyth lama; checksum itu tidak berlaku untuk genesis umtc.</p>

  <h3>Verifikasi final yang disarankan</h3>
  <ol>
    <li>Pastikan gentx MTC menunjukkan nilai <code>800000000000umtc</code>, memo IP yang benar, serta komisi <code>0.05 / 0.06 / 0.01</code>.</li>
    <li>Pastikan genesis berisi suplai <code>21000000000000umtc</code>, Community Pool <code>20000000000000umtc</code>, metadata MTC, dan <code>enable_pouw_emissions=false</code>.</li>
    <li>Jalankan <code>mythprotocold genesis validate-genesis --home "$MYTH_HOME"</code>, lalu hitung ulang SHA-256 setelah semua edit selesai.</li>
    <li>Distribusikan genesis identik beserta checksum kepada validator awal; jangan distribusikan keyring atau private consensus key.</li>
  </ol>
</section>

<section class="page">
  <h2>Rilis software dan penggunaan</h2>
  <h3>Rilis v0.1.3-myth-phase1</h3>
  <p>Rilis ini menukar bond/base-fee denom menjadi umtc dan menampilkan simbol MTC. Paket adalah software daemon; genesis dan chain ID MTC baru dibuat terpisah.</p>
  <p><a href="https://github.com/annabil-dev/ON-CHAIN-VALIDATOR/releases/tag/v0.1.3-myth-phase1">Buka halaman rilis v0.1.3-myth-phase1 di GitHub</a></p>

  <h3>Script rilis Windows</h3>
  <p><code>release_mythchain.py</code> dijalankan dari PowerShell dengan Python, Git, dan GitHub CLI yang sudah terautentikasi. Script memeriksa working tree/tag, mendorong branch dan tag, menunggu GitHub Actions, lalu memeriksa aset rilis. Build dilakukan oleh runner GitHub—tidak perlu membuka shell WSL.</p>
  <p>Windows: jalankan <code>python release_mythchain.py v0.1.3-myth-phase1</code>. GitHub Actions menguji source dan membuat paket amd64/arm64.</p>

  <h3>Status mainnet</h3>
  <div class="status"><strong>Chain MTC baru belum diluncurkan.</strong> Node yang berjalan memakai umyth lama; reset lintas semua validator diperlukan sebelum boot ulang dengan umtc.</div>
  <p>Parameter genesis tambahan—termasuk governance Treasury, unbonding, slashing, daftar validator awal, waktu peluncuran, seed, dan RPC—perlu disepakati serta dipublikasikan sebelum onboarding eksternal.</p>

  <h3>File dan artefak</h3>
  <table>
    <tbody>
      <tr><th>Repository</th><td><code>annabil-dev/ON-CHAIN-VALIDATOR</code>, branch <code>myth-first-phase</code></td></tr>
      <tr><th>Dokumentasi onboarding</th><td><code>docs/MYTH_PUBLIC_VALIDATOR_ONBOARDING.md</code></td></tr>
      <tr><th>Rencana tokenomics</th><td><code>docs/MYTH_TOKENOMICS_LAUNCH_PLAN.md</code></td></tr>
      <tr><th>Generator paket</th><td><code>build_validator_release.sh</code></td></tr>
      <tr><th>Otomasi rilis Windows</th><td><code>release_mythchain.py</code></td></tr>
    </tbody>
  </table>
  <p class="footer">Laporan ini merangkum riwayat repository serta keputusan dan artefak genesis yang dilaporkan operator hingga 1 Oktober 2026. Cocokkan gentx 800K dan checksum tepat sebelum distribusi final.</p>
</section>
</body>
</html>"""


def main() -> None:
    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        page = browser.new_page()
        page.set_content(HTML, wait_until="load")
        page.pdf(
            path=str(OUTPUT),
            format="A4",
            print_background=True,
            display_header_footer=True,
            header_template="<div style='font:8px Segoe UI,Arial;color:#718096;width:100%;padding:0 17mm'>MYTHCHAIN · LAPORAN PROGRESS</div>",
            footer_template="<div style='font:8px Segoe UI,Arial;color:#718096;width:100%;padding:0 17mm;display:flex;justify-content:space-between'><span>Rangkuman proyek · 01 Okt 2026</span><span><span class='pageNumber'></span> / <span class='totalPages'></span></span></div>",
            margin={"top": "19mm", "right": "17mm", "bottom": "19mm", "left": "17mm"},
        )
        browser.close()
    print(OUTPUT)


if __name__ == "__main__":
    main()
