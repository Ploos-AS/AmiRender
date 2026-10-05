# AmiRender farm model

AmiRender uses one engine-agnostic farm for all render backends.

A node advertises capabilities such as:

- `povray`
- `blender-cycles`
- `blender-eevee`
- future `ami3d`, `amiterrain` and translated classic engines

The controller schedules a job only to a node advertising the requested engine.

## POV-Ray

POV-Ray is a first-class farm engine, not a local-only appliance backend.

M1 targets frame-level distribution. Each animation frame is an independent render job and may execute on any capable ARM64 or AMD64 node. A single AmiRender appliance may act as both controller and worker.

Region/tile distribution for large still images is intentionally deferred until the frame scheduler is stable.

## Blender

Blender follows the same farm contract. CPU/GPU-specific capabilities may be advertised separately so the scheduler can prefer appropriate workers without exposing those implementation details to Amiga clients.

## Scheduler evolution

M1 uses deterministic first-capable-node selection as a testable baseline. Later milestones may add measured performance weights, current load, asset locality, architecture, memory and GPU capability.
