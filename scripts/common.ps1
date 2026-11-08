$ErrorActionPreference = 'Stop'
$script:ForgeRoot = Split-Path -Parent $PSScriptRoot
$script:ForgeGo = Join-Path $script:ForgeRoot 'tools/go/bin/go.exe'
if (-not (Test-Path -LiteralPath $script:ForgeGo)) { $script:ForgeGo = (Get-Command go -ErrorAction Stop).Source }
$script:ForgeNode = (Get-Command node -ErrorAction SilentlyContinue).Source
if (-not $script:ForgeNode) { $script:ForgeNode = Join-Path $env:USERPROFILE '.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin/node.exe' }
if (-not (Test-Path -LiteralPath $script:ForgeNode)) { throw 'Install Node 24 LTS and pnpm 11 first.' }
$script:ForgePnpm = Join-Path (Split-Path -Parent (Split-Path -Parent $script:ForgeNode)) 'node_modules/pnpm/bin/pnpm.cjs'
$env:PATH = (Split-Path -Parent $script:ForgeNode) + ';' + (Split-Path -Parent $script:ForgeGo) + ';' + $env:PATH
$gitDir = Join-Path $env:USERPROFILE '.cache/codex-runtimes/codex-primary-runtime/dependencies/native/git/cmd'
if (Test-Path -LiteralPath $gitDir) { $env:PATH = $gitDir + ';' + $env:PATH }
function Invoke-ForgePnpm {
 param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Arguments)
 if (Test-Path -LiteralPath $script:ForgePnpm) { & $script:ForgeNode $script:ForgePnpm @Arguments }
 else { & pnpm @Arguments }
 if ($LASTEXITCODE -ne 0) { throw "pnpm failed: $Arguments" }
}
function Assert-ForgeExit([string]$Operation) { if ($LASTEXITCODE -ne 0) { throw "$Operation failed with exit code $LASTEXITCODE" } }
