# Windows Founder Package — Installation

**Production status: BLOCKED until the founder genesis workflow and production
manifest are approved.**

1. Download the `windows-amd64` or `windows-arm64` ZIP and `SHA256SUMS` from the
   official stable release.
2. Verify the archive before extraction:

   ```powershell
   $archive = 'mythprotocold-<version>-windows-<arch>.zip'
   $line = Get-Content .\SHA256SUMS | Where-Object { $_ -match [regex]::Escape($archive) } | Select-Object -First 1
   $expected = ([regex]::Match($line, '^[0-9a-fA-F]{64}')).Value.ToLowerInvariant()
   $actual = (Get-FileHash $archive -Algorithm SHA256).Hash.ToLowerInvariant()
   if ($expected -ne $actual) { throw 'Release checksum mismatch' }
   ```

3. Extract the ZIP, then install with the versioned wizard:

   ```powershell
   Expand-Archive $archive -DestinationPath .\mythchain-release
   .\mythchain-release\install_mythprotocold.ps1 -Version <version>
   ```

After the production manifest and ceremony gates are approved, run the candidate
workflow from an elevated-free PowerShell session using a fresh path:

```powershell
$ChainId = '<approved-chain-id>'
$External = '<approved-founder-host>:26656'
& .\mythchain-release\mythprotocold.exe founder-init `
  --chain-id $ChainId --confirm-chain-id $ChainId `
  --output-dir "$HOME\mythchain-founder" --external-address $External
```

This creates a candidate; rebuild/use the binary pinned to its exact genesis hash
before start. Never put production keys in the ZIP or source tree.
