# Aeon

This repository holds the **Aeon** Phase 0–1 specification and skeletons. It is **not** a science executor.

**Source of truth:** [`docs/aeon-build-spec.md`](docs/aeon-build-spec.md) — the only spec. This README is a pointer, not a second spec.

OpenClaw is the approvals / Telegram / memory **surface**, not the executor. `openclaw-office/` is an unrelated Office dispatch skill and **must not** receive science jobs.

Aeon will never meter local compute.

| Component | License |
| --- | --- |
| Executor (Go daemon) | **AGPL-3.0 + commercial dual** (`executor/`) |
| Control plane | **Proprietary** |
| SDKs | **Apache-2.0** |

This tree is spec + skeletons. Do not treat any app in this repo as the science runtime.
