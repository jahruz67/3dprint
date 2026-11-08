[CmdletBinding()]
param([switch]$SkipInstall)
. "$PSScriptRoot/common.ps1"
Push-Location $script:ForgeRoot
try {
 if (Test-Path -LiteralPath tools) { [IO.File]::WriteAllText((Join-Path $script:ForgeRoot 'tools/go.mod'),"module forge-slicer-local-tools`n`ngo 1.26.0`n",[Text.UTF8Encoding]::new($false)) }
 Push-Location frontend
 try {
  if (-not $SkipInstall) { Invoke-ForgePnpm install --frozen-lockfile --ignore-scripts }
  Invoke-ForgePnpm check
  Invoke-ForgePnpm build
 } finally { Pop-Location }
 & $script:ForgeGo mod tidy
 Assert-ForgeExit 'Go module resolution'
 New-Item -ItemType Directory -Force bin | Out-Null
 & $script:ForgeGo build -trimpath -ldflags '-H windowsgui -s -w' -o bin/ForgeSlicer.exe ./cmd/forge
 Assert-ForgeExit 'Native desktop build'
} finally { Pop-Location }
