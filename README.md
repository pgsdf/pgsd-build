# PGSD Build System Prototype

This repository is a prototype of the Pacific Grove Software Distribution (PGSD) build
and install system. It is not production ready, but it captures the intended structure:

- `pgsdbuild` (Go) – image and bootenv builder
- `pgsd-inst` (Go + Bubble Tea) – TUI installer
- `images/*.lua` – system image recipes (e.g., pgsd-desktop)
- `variants/*.lua` – boot environment recipes (e.g., pgsd-bootenv-arcan)
- `overlays/*` – filesystem overlays used at build time
- `docs/*` – design and architecture documents

For full details, see `docs/ARCHITECTURE.md`, `docs/DESIGN.md`, and `docs/BUILD_PIPELINE.md`.
