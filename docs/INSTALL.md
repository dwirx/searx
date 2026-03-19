# Install Guide

This guide installs `search` from GitHub Releases so it can be used immediately.

## Quick Install (Recommended)

Linux and macOS:

```bash
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash
```

After install:

```bash
search --version
```

The installer will:
- detect OS and CPU architecture,
- download the correct release binary,
- install to `/usr/local/bin` (or `~/.local/bin` fallback),
- run `search setup` to check/update Lightpanda.

## Manual Install

Linux x86_64:

```bash
curl -fLO https://github.com/dwirx/searx/releases/latest/download/search-x86_64-linux
chmod +x search-x86_64-linux
sudo mv search-x86_64-linux /usr/local/bin/search
search --version
```

macOS Apple Silicon:

```bash
curl -fLO https://github.com/dwirx/searx/releases/latest/download/search-aarch64-darwin
chmod +x search-aarch64-darwin
sudo mv search-aarch64-darwin /usr/local/bin/search
search --version
```

## Build From Source

```bash
git clone https://github.com/dwirx/searx
cd searx
go build -o search ./cmd/search
sudo mv search /usr/local/bin/
search --version
```

## Lightpanda Notes

- `search setup` ensures Lightpanda exists and updates when newer version is available.
- If latest Lightpanda is already installed, no re-download occurs.
- Optional custom location:

```bash
export SEARX_LIGHTPANDA_PATH="$HOME/.local/share/searx/lightpanda"
```
