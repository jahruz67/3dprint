[CmdletBinding()]
param()
. "$PSScriptRoot/common.ps1"
Push-Location $script:ForgeRoot
try {
 if (-not (Test-Path -LiteralPath bin/ForgeSlicer.exe)) { throw 'Run scripts/build.ps1 before packaging.' }
 New-Item -ItemType Directory -Force artifacts | Out-Null
 [IO.File]::WriteAllText((Join-Path $script:ForgeRoot 'artifacts/go.mod'),"module forge-slicer-local-artifacts`n`ngo 1.26.0`n",[Text.UTF8Encoding]::new($false))
 $package = Join-Path $script:ForgeRoot ('artifacts/ForgeSlicer-windows-' + (Get-Date -Format 'yyyyMMdd-HHmmss'))
 New-Item -ItemType Directory -Force $package | Out-Null
 Copy-Item -LiteralPath bin/ForgeSlicer.exe -Destination $package
 foreach ($file in @('LICENSE','THIRD_PARTY_NOTICES.md','README.md','VERIFICATION.md')) { Copy-Item -LiteralPath $file -Destination $package }
 foreach ($dir in @('docs','schemas','examples')) { Copy-Item -LiteralPath $dir -Destination $package -Recurse }
 $source = Join-Path $package 'source'
 function Copy-ForgeSource([string]$From,[string]$To) {
  New-Item -ItemType Directory -Force $To | Out-Null
  foreach ($item in Get-ChildItem -LiteralPath $From -Force) {
   if ($item.Name -in @('tools','bin','artifacts','node_modules','dist','.git')) { continue }
   if ($item.PSIsContainer) { Copy-ForgeSource $item.FullName (Join-Path $To $item.Name) }
   else { Copy-Item -LiteralPath $item.FullName -Destination $To }
  }
 }
 Copy-ForgeSource $script:ForgeRoot $source
 $notices = Join-Path $package 'third-party-licenses'
 New-Item -ItemType Directory -Force $notices | Out-Null
 $modules = & $script:ForgeGo list -m -f '{{.Path}}|{{.Dir}}' all
 Assert-ForgeExit 'Go module notice inventory'
 foreach ($entry in $modules) {
  $parts = $entry.Split('|'); if ($parts.Length -ne 2 -or -not $parts[1]) { continue }
  $target = Join-Path $notices ('go-' + ($parts[0] -replace '[^a-zA-Z0-9.-]','_'))
  foreach ($file in Get-ChildItem -LiteralPath $parts[1] -File | Where-Object { $_.Name -match '^(LICENSE|NOTICE|COPYING)' }) { New-Item -ItemType Directory -Force $target | Out-Null; Copy-Item -LiteralPath $file.FullName -Destination $target }
 }
 foreach ($file in Get-ChildItem -LiteralPath frontend/node_modules/.pnpm -File -Recurse | Where-Object { $_.Name -match '^(LICENSE|NOTICE|COPYING)' }) {
  $relative = $file.FullName.Substring((Join-Path $script:ForgeRoot 'frontend/node_modules/.pnpm').Length).TrimStart('\')
  $target = Join-Path $notices ('npm-' + ($relative -replace '[^a-zA-Z0-9.-]','_'))
  Copy-Item -LiteralPath $file.FullName -Destination $target
 }
 Compress-Archive -LiteralPath $package -DestinationPath ($package + '.zip')
 Write-Output ($package + '.zip')
} finally { Pop-Location }
