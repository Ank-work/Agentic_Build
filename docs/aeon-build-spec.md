# Aeon — master build specification

Working name: **Aeon** (keep it). Audience: a team that did not sit through the research. This document is the single publishable spec for Phase 0–1. Do not re-derive facts from chat history; implement from here.

---

## 1. Status and how to use this document

**This file is the source of truth.** The root `README.md` is a pointer only. `openclaw-office/` is an unrelated Office dispatch skill (see §19).

**Today this repository is documentation-only.** There is no Aeon product code, no Node/Python app, no AK mutation, and no toolchain to install from this repo. Phase 0–1 work specified here is *to be implemented later* on AK and in product repos — specify, do not execute, from this tree.

Use this spec to:

- Verify the Cursor worker on AK (M0) and gather host facts.
- Stand up the **separate Go executor daemon** and job JSON Schema.
- Point a stdio MCP shim at OpenClaw on the worker (not a tunnel).
- Register science-gateway accounts and pin runtimes.
- Refuse anti-patterns in §18.

Unverified items are marked **TODO** or **must-verify**. Do not overclaim license compliance in sales or docs.

---

## 2. Product definition

Aeon is a **full startup application** that fuses AI with physical science — not an MVP sketch and not a chatbot with a calculator.

The product lets a scientist (or a lab) state a question in natural language, then:

1. Retrieve literature and observational data from federated live sources.
2. Have the Cursor Cloud Agent write analysis code, numerical checks, and reports.
3. Run that work on local scientific runtimes on AK (the Windows PC), never by reopening OpenClaw `exec`/`spawn`/`shell`.
4. Return artifacts, citations, uncertainties, and provenance (RO-Crate + W3C PROV).
5. Keep approvals, Telegram, and long-term memory on OpenClaw.

Aeon is the control plane + science gateway + local executor around that loop. The flagship workflows are W1–W8 (§14). Correctness machinery (§13) ships **before** feature surface.

Honest data model: **federated live retrieval + on-demand granule fetch + selective local cache.** Aeon cannot and will not mirror CERN (~6.4 PB), Gaia DR4 (~400 TB official, 2 Dec 2026), LIGO O4, the OpenAlex snapshot, or AlphaFold (~23 TiB / 644M files).

---

## 3. Locked architecture

| Role | Where | What it does |
| --- | --- | --- |
| **BRAIN** | Cursor Cloud Agent | Writes **all** code. Reads data. Analysis and inference. Decides *what* to run. Never invents physical constants. |
| **HANDS / surface** | OpenClaw on Windows PC **AK** | Approvals, Telegram, memory, nodes, tasks, sandbox, secrets (SecretRef), audit. **Does not run science jobs.** |
| **EXECUTOR** | **Separate Go daemon** on AK | Actually runs jobs from argv arrays. Network default `none`. Uploads artifacts. Enforces capability tokens. |

Flow: Cursor says "run this" → job spec hits AK → executor runs → results (and artifacts) return.

Do **not** treat OpenClaw as the executor. Gateway `POST /tools/invoke` **hard-denies** `exec`, `spawn`, `shell`, `fs_write`, and related tools. Do not reopen those. Build a separate executor.

Do **not** conflate this AK science stack with `openclaw-office/` (Office Cursor-cloud dispatch for `Ank-work/Agentic_Build`).

---

## 4. Control plane: self-hosted Cursor worker on AK

Self-hosted Cloud Agent workers went GA **25 Mar 2026**.

- Docs: [My Machines](https://cursor.com/docs/cloud-agent/self-hosted-guides/my-machines) and [pool](https://cursor.com/docs/cloud-agent/self-hosted-guides/pool).
- Blog: [cursor.com/blog/self-hosted-cloud-agents](https://cursor.com/blog/self-hosted-cloud-agents).

### 4.1 Network

Worker is **outbound-only HTTPS** to:

- `api2.cursor.sh`
- `api2direct.cursor.sh`
- `cloud-agent-artifacts.s3.us-east-1.amazonaws.com`

**No inbound ports.** Do not design NAT holes, reverse tunnels, or inbound webhooks for the worker.

### 4.2 Install (native Windows)

Native Windows installer (WSL is **not** required for the worker):

```powershell
irm 'https://cursor.com/install?win32=true' | iex
```

Run worker verify as user **`kulan`** (not `openclaw`). See M0 in §15.

### 4.3 My Machines vs pools

- Teams can run personal **My Machines** workers.
- Self-hosted **pools require Cursor Enterprise** (service account API key). Staff: "Enterprise is required at this time" for pools. Do not plan a Teams pool without Enterprise.

### 4.4 MCP placement

- **stdio MCP** runs **on the worker (AK)**.
- **HTTP/SSE MCP** runs on Cursor's backend.
- OpenClaw is reachable via a **stdio MCP shim on the worker**, not a tunnel.

### 4.5 Timeouts and persistence

- No documented max cloud-agent runtime.
- `--idle-release-timeout` default **3600s** (post-session).
- `workerReadyTimeoutSeconds` is **pool-only**.
- My Machines **does not survive sleep/reboot** without a custom controller. Treat that as a Day-1 operational constraint.

### 4.6 Secrets / OIDC — do not assume

OIDC token socket is documented only for **Cursor-managed VMs**, not self-hosted workers. **Do not design secrets around OIDC until verified on a worker.** Use Doppler machine identities (§7) until that is proven.

### 4.7 SDK pins

- `@cursor/sdk` and `cursor-sdk` both **1.0.30**.
- `CloudAgentOptions.env` accepts `{ type: "pool" | "machine", name }`.

### 4.8 Token rate

Teams/Enterprise pay **$0.25/M "Cursor Token Rate"** on third-party models **including BYOK**. First-party (Grok, Composer) are exempt.

### 4.9 Fallback if the plan cannot run a worker

If the user's Cursor plan cannot run a self-hosted worker, use a **git-mediated job queue** (see §6.3). That is the fallback, not the happy path.

---

## 5. Local runtime on AK: OpenClaw vs executor split

### 5.1 OpenClaw facts (do not reopen denied tools)

- OpenClaw **2026.7.1-2**.
- State: `C:\Users\kulan\.openclaw` (**not** `C:\Users\openclaw`).
- CLI: `C:\Users\kulan\.openclaw\bin\openclaw.ps1`.
- SSH user `ak\openclaw` — **cannot** control Tailscale ("already in use by AK\kulan") or reliably introspect the host (WMI / `systeminfo` denied; `wmic` gone on Win 10.0.26200). `Get-PSDrive` returns silent `freeGB=0`. Host facts for RAM must be gathered **elevated as kulan** (M0).
- Gateway on **18789**.
- `POST /tools/invoke` **hard-denies** `exec`, `spawn`, `shell`, `fs_write`, etc. **Do not reopen those. Build a separate executor.**

OpenClaw **does** have: `tasks`, `nodes`, `approvals`, `sandbox`, `secrets` (SecretRef), `audit`, `mcp serve`, plugins. `exec-approvals.json` is the existing approval gate. Skills live at `C:\Users\kulan\.openclaw\skills\<name>\SKILL.md`.

**Remediate** the plaintext GitHub Copilot token in `openclaw.json` via `openclaw secrets configure --apply` and rotate. **Do not print the token.** Do not copy it into this repo.

### 5.2 Executor (separate process)

A **Go 1.25 daemon** (see §7) consumes job specs (§6), runs argv-only processes, default `network.mode: none`, uploads artifacts to R2, and returns exit codes + logs. It is not an OpenClaw plugin that wraps `exec`.

### 5.3 What OpenClaw is for in Aeon

Approvals (including wiring to `exec-approvals.json` as a *human* gate, not as a shell), Telegram notifications, memory, node presence, SecretRef lookup (never plaintext in job specs), audit trail. Science binaries are the executor's problem.

---

## 6. Job protocol

Jobs are JSON documents. **argv arrays, never shell strings.** Default **no egress**. Two-stage **reader / executor** (CaMeL / Dual-LLM + capabilities): the reader may parse untrusted literature/data; only the executor stage may run with a capability token.

### 6.1 JSON Schema (inline)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://aeon.local/schemas/job-spec.v1.json",
  "title": "AeonJobSpec",
  "type": "object",
  "additionalProperties": false,
  "required": [
    "apiVersion",
    "kind",
    "metadata",
    "spec"
  ],
  "properties": {
    "apiVersion": {
      "type": "string",
      "const": "aeon.dev/v1"
    },
    "kind": {
      "type": "string",
      "const": "Job"
    },
    "metadata": {
      "type": "object",
      "additionalProperties": false,
      "required": ["id", "createdAt"],
      "properties": {
        "id": { "type": "string", "minLength": 1 },
        "createdAt": { "type": "string", "format": "date-time" },
        "workflow": { "type": "string", "description": "W1–W8 or other named workflow" },
        "labels": {
          "type": "object",
          "additionalProperties": { "type": "string" }
        }
      }
    },
    "spec": {
      "type": "object",
      "additionalProperties": false,
      "required": ["argv", "timeoutSeconds", "network", "stage"],
      "properties": {
        "stage": {
          "type": "string",
          "enum": ["reader", "executor"],
          "description": "Reader parses untrusted input; executor runs with a capability token."
        },
        "argv": {
          "type": "array",
          "minItems": 1,
          "items": { "type": "string" },
          "description": "Process argument vector. Never a shell string. argv[0] must be on the allowlist."
        },
        "cwd": {
          "type": "string",
          "description": "Working directory on AK; must be under an allowed root."
        },
        "env": {
          "type": "object",
          "additionalProperties": { "type": "string" },
          "description": "Explicit env keys only. No shell expansion. Secrets via SecretRef names, not values."
        },
        "timeoutSeconds": {
          "type": "integer",
          "minimum": 1,
          "maximum": 86400
        },
        "network": {
          "type": "object",
          "additionalProperties": false,
          "required": ["mode"],
          "properties": {
            "mode": {
              "type": "string",
              "enum": ["none", "allowlist"],
              "default": "none",
              "description": "Default none. allowlist requires destinations[] and a matching capability."
            },
            "destinations": {
              "type": "array",
              "items": {
                "type": "object",
                "additionalProperties": false,
                "required": ["host"],
                "properties": {
                  "host": { "type": "string" },
                  "port": { "type": "integer", "minimum": 1, "maximum": 65535 }
                }
              }
            }
          }
        },
        "capabilityToken": {
          "type": "string",
          "description": "Opaque capability for executor stage. Reader stage must omit or send empty. Never log the token."
        },
        "artifacts": {
          "type": "object",
          "additionalProperties": false,
          "properties": {
            "paths": {
              "type": "array",
              "items": { "type": "string" },
              "description": "Glob or relative paths under cwd to upload after exit."
            },
            "bucket": {
              "type": "string",
              "description": "R2 bucket name. PC→R2 direct."
            },
            "prefix": { "type": "string" }
          }
        }
      }
    }
  }
}
```

**Invariants the executor must enforce (not optional):**

- Reject any spec where a single string is treated as a shell command.
- `network.mode` defaults to `none` if omitted; treat missing `network` as `{ "mode": "none" }`.
- `stage: reader` must not receive a usable `capabilityToken`.
- `stage: executor` must present a valid capability token whose allowlist covers `argv[0]` and network destinations.
- Artifact upload is PC → Cloudflare R2 direct (see §7). Do not hairpin through Cursor's artifact bucket for science outputs.

### 6.2 Two-stage reader / executor

Untrusted content (papers, HTML, notebooks, simulator decks) is handled by the **reader** stage: schema-validate, extract, never execute. The **executor** stage runs allowlisted binaries with capabilities. This is the CaMeL / Dual-LLM + capabilities pattern: the model that *reads* untrusted text is not the process that *runs* code.

Never execute fetched notebooks. Schema-validate simulator decks before any run.

### 6.3 Fallback: git-mediated queue

If the Cursor plan cannot run a worker:

1. Brain writes a job spec file into a private git repo (or a designated branch/path).
2. A watcher on AK (still the Go executor, not OpenClaw `exec`) pulls and validates against the schema.
3. Results and artifact URIs are committed or pushed as a reply document.

This is **fallback only**. Happy path is Cursor worker → job on AK → results return.

---

## 7. Recommended stack

| Layer | Choice | Notes |
| --- | --- | --- |
| Control plane | Self-hosted Cursor worker on AK | My Machines until Enterprise pool is available |
| Executor | **Go 1.25** + **Temporal** | Separate daemon; argv jobs; timeouts |
| Object store | **Cloudflare R2** | PC→R2 direct; ~$0.015/GB-mo, **$0 egress** |
| Surface | OpenClaw | Telegram, approvals, memory — not execution |
| UI (later) | Next.js | Not Phase 0–1 |
| Auth (later) | WorkOS | Not Phase 0–1 |
| Secrets | **Doppler** | Machine identities free. Infisical is $20/identity — do not default to it |

Do not invent a Node/Python product app in *this* repo. The stack above is the target for implementation repos / AK, specified here.

Warm cache tier: R2. Local hot cache on `C:` with ~113 GB headroom reserved from the science spec. Science data layer cost is ~$15–65/mo; **real cost is compute and people**.

---

## 8. Science gateway

**One MCP server, ~10 tools, not 40 MCP servers.**

Adopt upstream where it exists; build the rest:

| Cluster | Action |
| --- | --- |
| NASA Earthdata | **Adopt** official MCP at `https://cmr.earthdata.nasa.gov/mcp/v1` (**7 tools**). Count as the Earthdata cluster, not seven Aeon-owned tools. |
| Materials Project | **Adopt** official `mpmcp`. |
| Literature / astro / GW / chem | **Build ourselves.** |

Proposed **Aeon surface (~10 tools)** — Earthdata's 7 stay adopted-upstream so *our* tool count stays ~10:

1. `earthdata_*` — adopted NASA Earthdata MCP (7 upstream tools; one cluster in our gateway).
2. `mpmcp` — Materials Project, official.
3. Literature search — OpenAlex `corpus=all` (~477M works) + Semantic Scholar (`utm_source=api` + name/logo) + ADS **query-time only**.
4. PMC/OA fetch — **new** inventory/ESearch path; **not** the deleted `oa_comm` / `oa_noncomm` bulk architecture.
5. GWOSC / GWTC-5.0 granule + strain.
6. Horizons / SPICE / `de442.bsp`.
7. lightkurve / MAST — **Gaia gated** (CC BY-NC; commercial needs ESA written auth).
8. PubChem (governor **5/s**) + ChEMBL (**isolate SA fields**).
9. HITRAN / ExoMol (**SA warning** on ExoMol).
10. STAC — **Earth Search / CDSE primary**; Planetary Computer is Preview-licensed ("not meant for production") — not primary.

Pin-and-diff MCP: treat upstream MCP schemas as pinned artifacts; fail closed on unexpected tool/schema drift.

---

## 9. Source catalog

Condensed high-leverage + corrected rows. **commercial-safety** is `ok` / `restricted` / `stop` / `query-time-only`. Not every research row. **Do not overclaim.**

| Source | Access pattern | License / terms | Commercial-safety | Notes |
| --- | --- | --- | --- | --- |
| NASA Earthdata / CMR | Official MCP `https://cmr.earthdata.nasa.gov/mcp/v1` | NASA Earthdata ToS; account required | restricted | Adopt 7-tool MCP. Need Earthdata login. |
| Materials Project | Official `mpmcp` | MP ToS; API key | restricted | Account required. |
| OpenAlex | Free API key; usage-based (~$1/day free) | Snapshot still **CC0** | restricted | **Polite pool retired 13 Feb 2026.** Use `corpus=all` (~477M works). |
| Semantic Scholar | API + bulk (licensed) | **Not ODC-BY.** Commercial needs AI2 licence | restricted | `utm_source=api` + name/logo required. |
| ADS | Query API | Bans systematic download "for any purpose, whether commercial or not" | query-time-only | No bulk harvest. |
| PMC OA | ESearch licence filters **or** daily inventory CSV → `s3://pmc-oa-opendata/PMC<id>.<ver>/` | Per-article OA terms | restricted | **Deleted** OA bulk architecture Aug 2026 (`oa_comm` / `oa_noncomm` gone). Rebuild on the new path. |
| bioRxiv TDM | Query / TDM bucket | No redistribution/re-host | query-time-only | Do not re-host. |
| Gaia / ESA SSA | Archive / TAP | **CC BY-NC 3.0 IGO** (not CC BY 4.0) | restricted | Commercial use needs **written ESA auth** (`data.licences@esa.int`). Applies to ESA Space Science Archive generally. |
| GWOSC | Granule / strain | **CC BY 4.0** (not CC0) | ok | Attribution + acknowledgement required. Pin **GWTC-5.0** (May 2026, O4b + Virgo). Decimated strain hard ceiling **1700 Hz**. |
| CDS / SIMBAD / VizieR | Per-dataset | No blanket commercial right; SIMBAD **ODbL**; logos banned commercially without consent | restricted | Check dataset, not portal. |
| Rubin / LSST | DACs / brokers | Gate is **non-profit affiliation**; no purchase path for a company | stop | Even post-proprietary, not via Rubin DACs for non-rights-holders. Alerts via community brokers only. DR1 ~end June 2028. |
| Horizons / SPICE | JPL APIs + kernels | NASA / NAIF terms | restricted | Prefer **`de442.bsp`** over de440. |
| MAST / lightkurve | MAST API | Mix; Gaia subset NC | restricted | Gate Gaia-derived products. |
| PDG | Tables | **PDG 2026.0**, CC BY 4.0 | ok | Pin release. |
| HITRAN | API / files | HITRAN terms | restricted | Account; cite. |
| ExoMol | Files | **CC BY-SA 4.0** | restricted | Share-alike contamination risk. Isolate artifacts. |
| PubChem | PUG/REST | NCBI terms | restricted | Governor **5 requests/s**. NCBI account. |
| ChEMBL | API | Mix; isolate SA fields | restricted | Do not leak SA into non-SA artifacts. |
| Open-Meteo | API | Data CC BY 4.0; **free-tier ToS is non-commercial** | restricted | Paid plan required for a commercial product. |
| Microsoft Planetary Computer | STAC | **Preview** ("not meant for production") | stop | Live but not production-primary. Fallback: **Earth Search / CDSE**. |
| EarthScope FDSN **event** | — | Retired | stop | HTTP **410** (retired Jun 2026). Use **USGS NEiC**. Station/dataselect remain on `service.earthscope.org`. |
| CERN / LHC open data | — | Cannot mirror ~6.4 PB | query-time-only | Federated fetch only; no full mirror. |
| Gaia DR4 bulk | — | ~400 TB official, 2 Dec 2026 | restricted | No full mirror; NC license. |
| LIGO O4 bulk | — | Cannot mirror | query-time-only | Granules via GWOSC. |
| AlphaFold DB | — | ~23 TiB / 644M files | query-time-only | No full mirror. |
| MACE-OMAT-0 / MACE-MATPES-0 | Checkpoints | **ASL non-commercial** | stop | Do not ship. |
| MACE-MP-0 / MPA-0 | Checkpoints | **MIT** | ok | Allowed MLIP funnel. |
| Orb | Checkpoints | **Apache-2.0** (cleanest) | ok | Prefer. |
| DrugBank | Files / API | **CC BY-NC** | stop | Do not install for commercial Aeon. |
| ORCA | Binary | Academic-only | stop | Do not install. |
| GDB | Files | Patent-use ban | stop | Do not install. |
| Wolfram Engine (free) | Binary | Free = **no production** | stop | Do not install for production. |
| ESM3 | Weights | Licence **self-contradictory** | stop | Do not use. |
| GROMACS | Local binary | LGPL-2.1 | ok | argv/files/exit codes. |
| LAMMPS | Local binary | GPL-2.0 (LGPL-2.1 on request) | restricted | FSF mere-aggregation / pipes+sockets if we stay argv/files/exit codes. |
| Quantum ESPRESSO | Local binary | GPL-2.0 | restricted | Same aggregation caution. |
| OpenFOAM | Local binary | **GPL-3.0** (not GPL-2.0) | restricted | Same. Do not relicense into proprietary blobs. |

---

## 10. Hardware and runtime pins on AK

Observed on AK (do not "fix" these by SSH from this repo):

| Item | Pin / fact |
| --- | --- |
| GPU | RTX 5060, 8151 MiB, **sm_120** |
| Driver | **610.88** (Data Center R610 → CUDA **13.3**). NFB EOL **Aug 2026**. **Do not pair with cu128.** |
| CPU | AMD Zen 3, 8C/16T |
| RAM | **unknown** — gather via elevated `host-facts.json` as **kulan** |
| Disk | C: 763 GB free / 930.5 GB. E: 135.3 / 476.7 GB. Cache on C: with **~113 GB** science headroom |
| Python | **None** installed. Floor **≥3.12** (numpy 2.5.2 / scipy 1.18.1). Prefer **3.12** for science-wheel long tail; 3.13 well-covered |
| `uv` | Standalone **0.12.8** as of 31 Aug 2026; no bootstrap Python |
| Docker | CLI present, **engine down**. Leave Docker stopped on Day-1 |
| WSL | WSL2 enabled, **no distro**. Worker does **not** require WSL. PySCF and JAX-GPU still **WSL-only** |
| Node | 24.19 / npm 11.17 / Git 2.55 |

### 10.1 PyTorch / CUDA

- PyTorch stable **2.13.0**.
- Windows official CUDA indexes: **cu126 / cu130 / cu132**.
- **cu128 index tops out at 2.11.0.** Do not use cu128 with driver 610.88.
- Driver 610.88 vs CUDA 12.x is minor-compat only for drivers **<580** — **must re-verify** GeForce driver + `torch.cuda.get_arch_list()` contains **sm_120**.
- `torch.load` `weights_only=True` since 2.6.

### 10.2 Science wheels

- **gwpy 4.0.2** native Windows (LALSuite optional/Unix).
- **OpenMM 8.6.0** native Windows + `openmm[cuda12]`.
- **PySCF** and **JAX-GPU** still WSL-only.
- **poliastro archived**; successor is **boinor 0.20.0**, not hapsira.
- Prefer `de442.bsp` over de440.
- SciPy **1.18.1**; CODATA 2022 since SciPy 1.15.0 — pin and record release.

---

## 11. Legal / license constraints

**Do not overclaim.** The sentence "local execution is always license-compliant" is **unsafe**. Strike it from any pitch.

Defensible pitch: local execution avoids *distribution / sublicensing / service-bureau* problems. It does **not** cure automation, benchmark-disclosure, or notification clauses. **Get counsel on MATLAB SLA §3.19 before sales language.**

### 11.1 Commercial scientific codes

| Product | Cite | Constraint (summary — not legal advice) |
| --- | --- | --- |
| **MATLAB** | **SLA §3.19** (not the Limited License) | Prohibits mechanizing/automating license checkout. A **job queue that starts MATLAB sessions is in the blast radius**. §3.16 bars access to temporary intermediate files. **§3.5** is service bureau (**not** §3.3). §3.8: no web/network interface except as permitted by the License Offering. |
| **Schrödinger** | Nov 2025 EULA §2(ii), §2(v) | No service bureau / time-sharing / use for benefit of third parties. No benchmarks. |
| **VASP** | Public statements; underlying agreement **not public** | "An AI provider is a third party." No training on source / PP-database. |
| **Gaussian** | Nature 429:231; competitor ban | No publishing performance data. |
| **Dassault** | CLOSA §14.3, §2.2(b), §2.2(f) | Third-party hosting **expressly permitted** (competitor-exclusion). Service bureau banned. No public benchmarks. |
| **COMSOL 6.4** | Hosting rules; **§11** | Hosting/time-share banned except **CSL** (COMSOL Server), which may be sublicensed for hosting apps. FNL allows remote/cluster. **§11 notification duty** — customer must name third parties with access. |
| **ANSYS** | Core Mechanical/Fluent EULA **not public** | AGI/STK SLA (different product) bars service bureau and remote cloud tasking. Ansys **sells** Elastic Licensing + Gateway on AWS — **"Ansys forbids the cloud" is false.** |

### 11.2 Open-source engines

GROMACS LGPL-2.1; LAMMPS GPL-2.0 (LGPL-2.1 on request); QE GPL-2.0; **OpenFOAM is GPL-3.0 not GPL-2.0**. FSF mere-aggregation / pipes+sockets is favorable **if we stay argv / files / exit codes** (matches the job protocol).

### 11.3 Product licensing (Aeon itself)

See §16. Dual-license the executor; keep the control plane proprietary; Apache-2.0 SDKs.

---

## 12. Security model

1. **Two-stage CaMeL / Dual-LLM + capabilities.** Reader vs executor (§6.2). Untrusted text never becomes a shell string.
2. **argv allowlist.** `argv[0]` must be on a signed allowlist. No `/bin/sh -c`, no `cmd.exe /c`, no PowerShell `-Command`.
3. **Default no-egress.** `network.mode: none`. Allowlist destinations only with a matching capability.
4. **Pin-and-diff MCP.** Upstream Earthdata / `mpmcp` schemas pinned; drift fails closed.
5. **Capability tokens.** Executor stage only. Do not log tokens. Do not put secrets in job JSON (SecretRef / Doppler).
6. **Do not assume OIDC on the self-hosted worker** until verified (§4.6).
7. **Code-signing notes (Windows worker/executor installers):** Azure Artifact Signing is unavailable if the entity is **<3 years** old. Use OV from a CA + FIPS token; **no EV**. Certum OS currently **out of stock**. WiX v7 OSMF is on the *WiX binary release* at $10k revenue, waivable by self-compile; need `<AcceptEula>wix7</AcceptEula>`.
8. **OpenClaw secrets:** remediate plaintext Copilot token (`openclaw secrets configure --apply` + rotate). Never print it.
9. **stdio MCP shim on the worker** to reach OpenClaw — not a public tunnel to :18789.

---

## 13. Correctness (built before features)

Ship this before expanding W1–W8 feature surface:

| Rule | Implementation intent |
| --- | --- |
| Constants never from the LLM | `scipy.constants` / `astropy.constants` / CODATA. CI lint on float literals in analysis code. |
| Units | `pint` / `astropy.units`. |
| Uncertainty | Required on numeric claims. |
| Golden-value suite | Regression tests against pinned values. **TODO: values themselves still need a primary-source pull** — do not invent the goldens in this spec. |
| Statistical guardrails | No p-hacking defaults; pre-register analysis choices in the job metadata where applicable. |
| Citation discipline | Every retrieved fact carries a source URI and license tag from §9. |
| Provenance | RO-Crate + W3C PROV on artifacts. |
| Notebooks | **Never execute fetched notebooks.** |
| Simulator decks | Schema-validate before run. |
| SciPy / CODATA | SciPy 1.18.1; CODATA 2022 since 1.15.0 — pin and record the release in every crate. |

---

## 14. Flagship workflows W1–W8

| ID | Workflow | Pins / cautions |
| --- | --- | --- |
| **W1** | Literature → hypothesis → numerical check | OpenAlex `corpus=all` + S2 (utm + logo) + ADS query-time. PMC via new inventory/ESearch, not deleted bulk. No notebook execution. Constants from CODATA, not the LLM. |
| **W2** | GW event reanalysis | Pin **GWTC-5.0**. GWOSC **CC BY 4.0** (attribution + acknowledgement). Use **16 kHz** if high-frequency content matters. Decimated strain hard ceiling **1700 Hz**. gwpy 4.0.2 native Windows. |
| **W3** | Solar-system trajectory | SPICE + Horizons. Kernel **`de442.bsp`** (not de440). **boinor 0.20.0** (poliastro archived; not hapsira). |
| **W4** | Exoplanet transit (lightkurve) | MAST/lightkurve. **If using Gaia: CC BY-NC 3.0 IGO — commercial needs ESA written auth.** Do not present Gaia as CC BY 4.0. |
| **W5** | Materials MLIP funnel | **MIT / Apache checkpoints only** (MACE-MP-0 / MPA-0, Orb). **Stop:** MACE-OMAT-0 / MACE-MATPES-0 (ASL NC), ESM3. Adopt `mpmcp`. |
| **W6** | Molecular property | PubChem governor **5/s**. **Isolate ChEMBL SA fields.** Stop: DrugBank NC, ORCA academic, GDB patent-use, free Wolfram in production. |
| **W7** | Spectroscopy | HITRAN + ExoMol. **ExoMol CC BY-SA 4.0 — share-alike contamination risk.** Isolate SA outputs. |
| **W8** | Earth observation | STAC via **Earth Search / CDSE**. **Not Planetary Computer as primary** (Preview, not meant for production). Earthdata MCP adopted. EarthScope FDSN **event** is 410 — use USGS NEiC. |

---

## 15. Phased build

Merge **platform ~60 engineer-weeks** + **science ~6 months**. Do not pretend this is a weekend MVP.

### 15.1 M0 — user actions on AK (before platform code)

The implementing team cannot complete M0 from this Mac repo. **The user (kulan) must:**

1. Install the native Windows worker if needed: `irm 'https://cursor.com/install?win32=true' | iex`.
2. Run as **kulan** (not `openclaw`):
   - `agent login`
   - `agent worker debug --json`
3. Confirm outbound-only HTTPS to the three Cursor endpoints (§4.1).
4. Confirm My Machines comes up; do not assume a pool without Enterprise.
5. Note sleep/reboot: My Machines will not survive without a custom controller.

**Day-1 unblock list** (specify only; do not run from here):

- Elevated **`host-facts.json`** including **RAM** (SSH `openclaw` cannot see it).
- `uv` + Python **≥3.12** (prefer 3.12).
- PyTorch pin **after** hardware check: driver 610.88, `torch.cuda.get_arch_list()` contains **sm_120**, indexes **cu126 / cu130 / cu132** — **not cu128**.
- Accounts: Earthdata, Materials Project, NCBI, NASA, ADS, Semantic Scholar (plus OpenAlex API key; paid Open-Meteo if weather is in scope).
- Leave **Docker stopped**.
- Do **not** install ORCA, DrugBank, GDB, MACE-ASL checkpoints, or Wolfram-free.

**Phase 0–1 implementable slice** (specify, do not implement in this repo):

1. M0 worker verify as kulan.
2. `host-facts.json`.
3. `uv` + Python 3.12.
4. Go 1.25 **executor daemon skeleton** + job JSON Schema (§6.1) + argv allowlist + `network.mode: none`.
5. **stdio MCP shim** on the worker toward OpenClaw (approvals/Telegram/memory only).
6. R2 bucket + PC→R2 artifact path.
7. Accounts list above.
8. Git-mediated queue only if worker is unavailable.

### 15.2 Platform (~60 engineer-weeks)

Worker controller (sleep/reboot), Temporal + Go executor, capability tokens, Doppler, R2, two-stage reader/executor, pin-and-diff MCP, signing/WiX notes, Next.js/WorkOS later.

### 15.3 Science (~6 months)

Gateway tools 3–10, W1–W8, golden-value suite (after primary-source pull), cache policy (~113 GB on C: + R2 warm).

---

## 16. Licensing and commercial (Aeon the product)

| Component | License |
| --- | --- |
| Executor (Go daemon) | **AGPL-3.0 + commercial dual** |
| Control plane | **Proprietary** |
| SDKs | **Apache-2.0** |

- **Never meter local compute.**
- Lab tier under **$15k micro-purchase**; proposed **$9,600/yr**.
- Sales language must not claim local = license-safe (§11).
- Gaia/ESA SSA commercial use needs written ESA authorization before any paid Gaia-backed feature.

---

## 17. Open questions / must-verify

Mark these closed only with evidence (ticket + date). Until then they are **open**.

1. **OIDC on self-hosted worker** — documented for Cursor-managed VMs only. Verify or keep Doppler.
2. **GeForce driver 610.88 + PyTorch 2.13.0** — confirm `torch.cuda.get_arch_list()` contains **sm_120**; confirm which of cu126 / cu130 / cu132 is valid. **cu128 is out.**
3. **AK RAM** — unknown until elevated `host-facts.json` as kulan.
4. **MATLAB SLA §3.19** — counsel before any job-queue that starts MATLAB, and before sales language.
5. **VASP** underlying agreement not public; **ANSYS** Mechanical/Fluent EULA not public — do not claim either way beyond §11.
6. **Golden-value suite** — primary-source pull still **TODO**.
7. **ESA written auth** for commercial Gaia / SSA (`data.licences@esa.int`) if W4 or any SSA product is sold.
8. **AI2 commercial licence** for Semantic Scholar bulk, if bulk is needed.
9. **Open-Meteo paid plan** before shipping weather in a commercial product (free-tier ToS is non-commercial).
10. **Cursor plan vs worker** — if My Machines is unavailable, fall back to git-mediated queue; confirm plan entitlements.
11. **Pools** — Enterprise required; confirm when/if the org has it.
12. **Azure Artifact Signing / Certum OS / WiX OSMF** — entity age <3 years; Certum OS OOS; WiX EULA acceptance.
13. **ESM3** — remain **STOP** unless a non-contradictory licence is obtained.
14. **COMSOL §11** notification list if any COMSOL access is ever granted through Aeon.
15. **Max cloud-agent runtime** — undocumented; measure under load.

---

## 18. Anti-patterns

**Do not:**

- Submit **shell-string** jobs (always argv arrays).
- **Reopen** OpenClaw `exec` / `spawn` / `shell` / `fs_write` (or treat OpenClaw as the science executor).
- Assume **OIDC on the self-hosted worker**.
- Require **WSL for the Cursor worker** (native Windows installer exists).
- Plan a **Teams pool without Cursor Enterprise**.
- **Mirror** CERN, Gaia DR4, LIGO O4, OpenAlex snapshot, or AlphaFold.
- Use **Planetary Computer as production primary**.
- Ship **MACE ASL** checkpoints (OMAT-0 / MATPES-0).
- Use **ESM3**.
- Use **free Wolfram Engine in production**.
- Put **MATLAB on a job queue without counsel** (SLA §3.19).
- Say **"local = license-safe"** or "local execution is always license-compliant".
- Let the **LLM invent constants** (use scipy/astropy/CODATA; lint float literals).
- **Execute fetched notebooks**.
- Pair **cu128** with driver **610.88**.
- Present **Gaia as CC BY 4.0** (it is **CC BY-NC 3.0 IGO**).
- Present **GWOSC as CC0** (it is **CC BY 4.0**).
- Cite MATLAB **Limited License** instead of the **SLA** (especially §3.19, §3.16, §3.5, §3.8).
- Call OpenFOAM **GPL-2.0** (it is **GPL-3.0**).
- Use OpenAlex **polite pool** (retired 13 Feb 2026).
- Use PMC **`oa_comm` / `oa_noncomm` bulk** (deleted Aug 2026).
- Hit EarthScope FDSN **event** (HTTP 410; use USGS NEiC).
- Design secrets around printing or committing the Copilot token.
- Conflate **`openclaw-office/`** with the AK science executor.
- SSH to AK or mutate OpenClaw from this documentation repo.
- Treat **poliastro** or **hapsira** as the orbit library (use **boinor 0.20.0**).
- Use SSH user `openclaw` as if it were `kulan` for worker login, Tailscale, or host facts.

---

## 19. Pointers

| Path | What it is | What it is not |
| --- | --- | --- |
| [docs/aeon-build-spec.md](aeon-build-spec.md) (this file) | Aeon master build spec | — |
| [README.md](../README.md) | Short product pointer | Not a second spec |
| [`openclaw-office/`](../openclaw-office/) | Office Cursor-cloud dispatch skill for **Ank-work/Agentic_Build** (`SKILL.md`, `AGENTS-OFFICE.md`, `install-office.ps1`) | **Not** the AK science executor. Do not rewrite those files for Aeon. Do not send science jobs to Office. |

OpenClaw on AK remains the Telegram / approvals / memory surface for Aeon. The **separate Go executor** runs jobs. Cursor Cloud Agent is the brain.
