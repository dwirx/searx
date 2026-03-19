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
- run `search setup` to check/update Lightpanda,
- auto-add PATH to your shell profile when installed in a user directory.

## Update

Update `search` to latest release:

```bash
search update
```

Script form:

```bash
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash -s -- --update
```

`search update` and installer update both skip binary download if the installed version is already the latest release.
Lightpanda update check also skips download when already up to date.

## Uninstall

Remove `search` and Lightpanda:

```bash
search uninstall
```

Script form:

```bash
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash -s -- --uninstall
```

Remove `search` only (keep Lightpanda):

```bash
search uninstall --keep-lightpanda
```

Script form:

```bash
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash -s -- --uninstall --keep-lightpanda
```

## Advanced Install Control

Install to custom directory:

```bash
SEARX_INSTALL_DIR="$HOME/bin" \
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash
```

Skip automatic setup step:

```bash
SEARX_SKIP_SETUP=1 \
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash
```

Disable automatic PATH profile update:

```bash
SEARX_AUTO_PATH=0 \
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash
```

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
