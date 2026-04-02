#!/usr/bin/env pwsh
# Install script for kpv CLI tool
# Supports Windows with automatic platform detection
# Usage: ./install.ps1 [VERSION]
# Environment variables:
#   KPV_INSTALL_DIR - Installation directory (default: ~/AppData/Local/Programs/bin)

param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$Repo = "frostyeti/kpv"
$BinaryName = "kpv.exe"

function Get-DefaultInstallDir {
    if ($env:KPV_INSTALL_DIR) {
        return $env:KPV_INSTALL_DIR
    }
    return Join-Path $env:USERPROFILE "AppData\Local\Programs\bin"
}

function Get-LatestVersion {
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    $headers = @{}
    if ($env:GITHUB_TOKEN) {
        $headers["Authorization"] = "token $($env:GITHUB_TOKEN)"
    }
    $response = Invoke-RestMethod -Uri $apiUrl -Headers $headers
    return $response.tag_name
}

function Download-File($Url, $Output) {
    $headers = @{}
    if ($env:GITHUB_TOKEN) {
        $headers["Authorization"] = "token $($env:GITHUB_TOKEN)"
    }
    Invoke-WebRequest -Uri $Url -OutFile $Output -Headers $headers
}

$arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "unknown" }
if ($arch -eq "unknown") {
    throw "Unsupported architecture"
}

Write-Host "Detected platform: windows/$arch"

if (-not $Version) {
    Write-Host "Detecting latest version..."
    $Version = Get-LatestVersion
    Write-Host "Latest version: $Version"
}

$versionForUrl = $Version -replace '^v', ''
$installDir = Get-DefaultInstallDir
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$archiveName = "kpv-windows-$arch-v$versionForUrl.zip"
$downloadUrl = "https://github.com/$Repo/releases/download/v$versionForUrl/$archiveName"
$tempDir = Join-Path $env:TEMP ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    $archivePath = Join-Path $tempDir $archiveName
    try {
        Download-File $downloadUrl $archivePath
    }
    catch {
        $archiveName = "kpv-windows-$arch-$Version.zip"
        $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$archiveName"
        Download-File $downloadUrl $archivePath
    }

    Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force
    $binaryPath = Get-ChildItem -Path $tempDir -Filter $BinaryName -Recurse | Select-Object -First 1
    if (-not $binaryPath) {
        throw "Could not find binary in archive"
    }

    $installPath = Join-Path $installDir $BinaryName
    Copy-Item -Path $binaryPath.FullName -Destination $installPath -Force
    Write-Host "Installed $Version to $installPath"
    & $installPath --version
}
finally {
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force
    }
}
