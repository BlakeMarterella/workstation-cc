# install.ps1 — workstation-cc entry point (Windows).
#
# Thin entry: downloads the verified worker binary for Windows and hands off to
# it. All real logic lives in the Go worker. Any arguments are forwarded.
#
# Usage:
#   ./install.ps1
#   ./install.ps1 install --profile slim --dry-run

[CmdletBinding()]
param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]] $WorkerArgs
)

$ErrorActionPreference = 'Stop'

# Defaults (override via environment before running).
$Repo    = if ($env:WORKSTATION_REPO) { $env:WORKSTATION_REPO } else { 'BlakeMarterella/workstation-cc' }
$Version = if ($env:WORKSTATION_VERSION) { $env:WORKSTATION_VERSION } else { 'latest' }
$BinDir  = if ($env:WORKSTATION_BIN_DIR) { $env:WORKSTATION_BIN_DIR } else { Join-Path $env:LOCALAPPDATA 'workstation\bin' }

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    default { throw "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)" }
}

$asset = "workstation_windows_$arch.exe"
$base  = if ($Version -eq 'latest') {
    "https://github.com/$Repo/releases/latest/download"
} else {
    "https://github.com/$Repo/releases/download/$Version"
}

$tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP ([System.Guid]::NewGuid())) -Force
try {
    Write-Host "[preflight] Downloading $asset ($Version)"
    $binPath = Join-Path $tmp $asset
    Invoke-WebRequest -Uri "$base/$asset" -OutFile $binPath
    Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

    $expected = (Get-Content (Join-Path $tmp 'checksums.txt') |
        Where-Object { $_ -match "\s$([regex]::Escape($asset))$" } |
        ForEach-Object { ($_ -split '\s+')[0] }) | Select-Object -First 1
    if (-not $expected) { throw "No checksum listed for $asset" }

    $actual = (Get-FileHash -Algorithm SHA256 -Path $binPath).Hash.ToLower()
    if ($expected.ToLower() -ne $actual) {
        throw "Checksum mismatch for $asset (expected $expected, got $actual)"
    }

    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    $dest = Join-Path $BinDir 'workstation.exe'
    Copy-Item -Path $binPath -Destination $dest -Force
    Write-Host "[preflight] Installed worker to $dest"

    # Point the worker at this checkout so it can symlink dotfiles/ + app-configs/.
    if (-not $env:WORKSTATION_ROOT) {
        $env:WORKSTATION_ROOT = $PSScriptRoot
    }

    & $dest @WorkerArgs
    exit $LASTEXITCODE
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
