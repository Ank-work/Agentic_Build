# M0 blockers — human-only on AK

These are user actions from §15.1. **Specify, do not execute.** Cloud Agents must not SSH to AK or execute these steps.

## Worker verify as `kulan`

Run as **kulan** (not `openclaw`):

- `agent login`
- `agent worker debug --json`

## Outbound HTTPS (§4.1)

Confirm outbound-only HTTPS to the three Cursor endpoints:

- `api2.cursor.sh`
- `api2direct.cursor.sh`
- `cloud-agent-artifacts.s3.us-east-1.amazonaws.com`

No inbound ports.

## My Machines vs pools (§4.3)

Confirm My Machines comes up. Do not plan a Teams pool without Cursor Enterprise.

## Sleep/reboot (§4.5)

Sleep/reboot needs a custom controller; My Machines will not survive without one.

## Day-1 unblock list (specify only)

- Elevated `host-facts.json` including RAM (SSH `openclaw` cannot see it).
- `uv` + Python ≥3.12 (prefer 3.12).
- PyTorch pin **after** hardware check: driver 610.88, `torch.cuda.get_arch_list()` contains `sm_120`, indexes cu126 / cu130 / cu132 — **not cu128**.
- Accounts: Earthdata, Materials Project, NCBI, NASA, ADS, Semantic Scholar, OpenAlex API key; paid Open-Meteo if weather is in scope.
- Leave Docker stopped.
- Do **not** install ORCA, DrugBank, GDB, MACE-ASL checkpoints, or Wolfram-free.
