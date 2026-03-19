# Changelog

All notable changes to this project are documented in this file.

## Unreleased

- Added installer actions: `--update` and `--uninstall`.
- Added uninstall option `--keep-lightpanda` to keep Lightpanda files.
- Added automatic PATH profile update for user-directory installs (can be disabled with `SEARX_AUTO_PATH=0`).
- Changed default search engine to `ddg` so `search "<query>"` works more reliably out of the box.
- Added automatic engine fallback for blocked responses (`202/403/429`) on `ddg` and `searx`.
- Updated docs with explicit update/uninstall commands.

## v1.2.3 - 2026-03-20

- Improved release installer (`install.sh`) with OS/arch detection, safer install paths, and version-aware binary install.
- Added automatic post-install setup check so `search setup` runs after install.
- Added Lightpanda version-aware setup/update flow: skip download when already latest, update when outdated.
- Moved Lightpanda default path to a stable user location (`~/.local/share/searx/lightpanda`) with optional `SEARX_LIGHTPANDA_PATH` override.
- Updated reader integration to use resolved Lightpanda binary path instead of fixed `./lightpanda`.
- Added `search --version` support and embedded release tag version in release binaries.
- Expanded GitHub Release workflow to publish Linux and macOS binaries.
- Added detailed documentation in `docs/INSTALL.md` and `docs/USAGE.md`.

## v1.2.2 - 2026-03-20

- Fixed one-line installer to use the correct repository (`dwirx/searx`).
- Fixed installer binary target to `search` so downloaded assets match release build names.
- Released updated CLI build and install script via GitHub Actions release pipeline.
