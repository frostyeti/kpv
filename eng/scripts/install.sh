#!/usr/bin/env bash
# Install script for kpv CLI tool
# Supports Linux and macOS with automatic platform detection
# Usage: ./install.sh [VERSION]
# Environment variables:
#   KPV_INSTALL_DIR - Installation directory (default: ~/.local/bin)
#   GITHUB_TOKEN    - GitHub token for API authentication (optional)

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

REPO="frostyeti/kpv"
BINARY_NAME="kpv"

get_default_install_dir() {
  if [[ -n "${KPV_INSTALL_DIR:-}" ]]; then
    echo "$KPV_INSTALL_DIR"
  else
    echo "$HOME/.local/bin"
  fi
}

detect_os() {
  local os
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    linux*) echo "linux" ;;
    darwin*) echo "darwin" ;;
    *) echo "unknown" ;;
  esac
}

detect_arch() {
  local arch
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unknown" ;;
  esac
}

get_latest_version() {
  local api_url="https://api.github.com/repos/${REPO}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
      curl -fsSL -H "Authorization: token $GITHUB_TOKEN" "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
      curl -fsSL "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    fi
  elif command -v wget >/dev/null 2>&1; then
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
      wget -q --header="Authorization: token $GITHUB_TOKEN" -O - "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
      wget -q -O - "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    fi
  else
    echo "Error: neither curl nor wget is installed" >&2
    exit 1
  fi
}

download_file() {
  local url="$1"
  local output="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$output" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$output" "$url"
  else
    echo -e "${RED}Error: neither curl nor wget is installed${NC}" >&2
    return 1
  fi
}

main() {
  local version="${1:-}"
  local install_dir
  local os
  local arch
  local archive_name
  local download_url
  local temp_dir
  local version_for_url
  local binary_path

  os=$(detect_os)
  arch=$(detect_arch)

  if [[ "$os" == "unknown" ]]; then
    echo -e "${RED}Error: unsupported operating system${NC}" >&2
    exit 1
  fi

  if [[ "$arch" == "unknown" ]]; then
    echo -e "${RED}Error: unsupported architecture: $(uname -m)${NC}" >&2
    exit 1
  fi

  echo "Detected platform: ${os}/${arch}"

  if [[ -z "$version" ]]; then
    echo "Detecting latest version..."
    version=$(get_latest_version)
    if [[ -z "$version" ]]; then
      echo -e "${RED}Error: could not detect latest version${NC}" >&2
      exit 1
    fi
    echo "Latest version: $version"
  fi

  version_for_url="${version#v}"
  install_dir=$(get_default_install_dir)
  echo "Install directory: $install_dir"

  if [[ ! -d "$install_dir" ]]; then
    mkdir -p "$install_dir"
  fi

  archive_name="${BINARY_NAME}-${os}-${arch}-v${version_for_url}.tar.gz"
  download_url="https://github.com/${REPO}/releases/download/v${version_for_url}/${archive_name}"

  temp_dir=$(mktemp -d)
  trap 'rm -rf "$temp_dir"' EXIT

  if ! download_file "$download_url" "$temp_dir/$archive_name"; then
    archive_name="${BINARY_NAME}-${os}-${arch}-${version}.tar.gz"
    download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    if ! download_file "$download_url" "$temp_dir/$archive_name"; then
      echo -e "${RED}Error: failed to download release archive${NC}" >&2
      exit 1
    fi
  fi

  (cd "$temp_dir" && tar -xzf "$archive_name")
  binary_path=$(find "$temp_dir" -name "$BINARY_NAME" -type f | head -1)
  if [[ -z "$binary_path" ]]; then
    echo -e "${RED}Error: could not find binary in archive${NC}" >&2
    exit 1
  fi

  chmod +x "$binary_path"

  local install_path="${install_dir}/${BINARY_NAME}"
  if [[ "$install_dir" == /usr/* || "$install_dir" == /opt/* ]]; then
    if [[ -w "$install_dir" ]]; then
      cp "$binary_path" "$install_path"
    else
      sudo cp "$binary_path" "$install_path"
    fi
  else
    cp "$binary_path" "$install_path"
  fi

  if [[ -x "$install_path" ]]; then
    echo -e "${GREEN}✓ ${BINARY_NAME} ${version} installed successfully!${NC}"
    if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
      echo ""
      echo -e "${YELLOW}Warning: ${install_dir} is not in your PATH${NC}"
      echo "Add the following to your shell profile:"
      echo "  export PATH=\"\$PATH:${install_dir}\""
    fi
    echo ""
    echo "Installed version:"
    "$install_path" --version 2>/dev/null || echo "(version command not available)"
  else
    echo -e "${RED}Error: installation failed${NC}" >&2
    exit 1
  fi
}

main "$@"
