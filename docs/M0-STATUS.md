# M0 Status

Status: **FROZEN**

M0 establishes the AmiRender protocol and execution foundation.

## Accepted baseline

- versioned job and node schemas
- newline-delimited JSON control protocol
- controller and node executables
- pluggable engine contract
- deterministic null engine
- real TCP SUBMIT -> COMPLETE end-to-end test
- configurable node listen address
- CI gating

## M0 boundary

M0 deliberately excludes production render engines, asset transfer/cache, scheduling, discovery, AmigaOS client binaries and farm orchestration.

POV-Ray is the first production backend planned for M1.

Frozen after CI PASS.
