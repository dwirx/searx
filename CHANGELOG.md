# Changelog

All notable changes to this project are documented in this file.

## Unreleased

- No unreleased changes.

## v1.2.6 - 2026-03-20

- Added CLI-managed lifecycle commands:
  - `search update` now updates the CLI binary and checks Lightpanda status.
  - `search update --lightpanda-only` for Lightpanda-only updates.
  - `search uninstall` and `search uninstall --keep-lightpanda`.
- Fixed subcommand flag parsing so `search update --lightpanda-only` and `search uninstall --keep-lightpanda` are handled correctly.
- Added typo alias `search unistall`.
- Added tests for subcommand option parsing and installer action argument generation.
- Updated `README.md`, `docs/INSTALL.md`, and `docs/USAGE.md` for the new CLI update/uninstall flow and latest-version skip behavior.

## v1.2.5 - 2026-03-20

- Changed default search engine to `ddg`.
- Added automatic fallback for blocked responses (`202/403/429`) on `ddg` and `searx`.
- Added tests for default-engine and fallback behavior.

## v1.2.4 - 2026-03-20

- Added installer actions: `--update` and `--uninstall`.
- Added uninstall option `--keep-lightpanda` to keep Lightpanda files.
- Added automatic PATH profile update for user-directory installs (disable with `SEARX_AUTO_PATH=0`).
- Improved Lightpanda version detection to avoid unnecessary re-downloads when already up to date.
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
