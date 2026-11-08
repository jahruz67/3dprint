[CmdletBinding()]
param([switch]$Integration)
. "$PSScriptRoot/common.ps1"
Push-Location $script:ForgeRoot
try {
 New-Item -ItemType Directory -Force artifacts | Out-Null
 & $script:ForgeGo test ./... -count=1 -v 2>&1 | Tee-Object artifacts/unit-tests.txt
 Assert-ForgeExit 'Unit tests'
 if ($Integration) {
	  if (-not $env:FORGE_ENGINE) { $candidate = Join-Path $script:ForgeRoot 'tools/prusa/PrusaSlicer-2.9.6/prusa-slicer-console.exe'; if (Test-Path -LiteralPath $candidate) { $env:FORGE_ENGINE = $candidate } }
  $env:FORGE_INTEGRATION = '1'
  & $script:ForgeGo test ./internal/service -run TestNativeEndToEnd -count=1 -v 2>&1 | Tee-Object artifacts/native-integration.txt
  Assert-ForgeExit 'Real native slicing integration'
  Remove-Item Env:FORGE_INTEGRATION
 }
 & $script:ForgeGo test ./internal/optimizer ./internal/geometry -run '^$' -bench . -benchmem -count=1 2>&1 | Tee-Object artifacts/benchmarks.txt
 Assert-ForgeExit 'Benchmarks'
} finally { Pop-Location }
