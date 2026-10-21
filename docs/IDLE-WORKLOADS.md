# Optional idle workloads

AmiRender may use otherwise idle ARM64/AMD64 worker capacity for explicitly enabled background workloads.

## Policy

- disabled by default
- explicit opt-in per node
- rendering always has absolute priority
- idle work is stopped before render dispatch
- configurable idle delay and resource limits
- credentials, wallet addresses and pool configuration are supplied by the node owner; AmiRender ships none
- idle and render accounting remain separate
- m68k workers do not enable idle mining by default

## Initial adapters

Planned initial adapters:

- Monero / RandomX
- Verus / VerusHash

Mining software is an external dependency. AmiRender does not vendor miners and does not contain default wallets or mining-pool credentials.

## Generality

The subsystem is intentionally named idle workloads rather than mining. Future adapters may include CI, transcoding or other owner-authorized batch work.

## Node lifecycle

IDLE -> IDLE_DELAY -> IDLE_WORKLOAD

When a render job arrives:

IDLE_WORKLOAD -> PREEMPTING -> RENDERING

After the render queue drains:

RENDERING -> IDLE -> IDLE_DELAY -> IDLE_WORKLOAD
