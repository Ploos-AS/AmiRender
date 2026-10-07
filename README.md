# AmiRender

A render accelerator appliance, distributed Linux render farm, and preservation bridge for classic 3D workflows.

The Amiga is the workstation. AmiRender provides the modern compute.

## M0 goals

- Define the wire/job protocol.
- Establish the appliance/controller/node split.
- Provide a native Amiga-facing client contract.
- Make render engines pluggable.
- Start with a deterministic null backend, then POV-Ray.
- Keep ARM64 Linux as the reference appliance platform while remaining portable to amd64.

## Target application families

AmiRender is intended to support, over time:

- LightWave 3D
- Imagine
- Real3D
- Cinema 4D (Amiga)
- Aladdin 4D
- Tornado3D
- POV-Ray
- Vista/VistaPro and scenery tools through AmiTerrain
- AmiTerrain
- Ami3D
- Blender/Cycles/Eevee as appliance/farm backends
- Atari ST CAD-3D / Cyber Studio family
- Freescape / 3D Construction Kit scenes from C64, Atari ST, and Amiga

## Render modes

AmiRender keeps preservation and enhancement as separate, explicit goals:

- `original`: use the original renderer/software where practical, on original hardware or a qualified emulator/VM, preserving native scene and output formats.
- `authentic`: reproduce the original platform renderer, palette, resolution, shading, and relevant limitations in an automatable AmiRender backend.
- `enhanced`: preserve the source scene semantics in AmiRender IR and render with modern backends for the best practical result.

Importers must preserve original values and metadata alongside normalized IR data. Import must not destructively "improve" source assets.

Native formats are first-class preservation inputs/outputs. Planned examples include Atari CAD-3D/Cyber formats and Freescape/3D Construction Kit data. Modern interchange formats are additional outputs, not replacements for native-format support.

## Render farm

POV-Ray and Blender are first-class modern Linux farm backends. The scheduler is intended to support amd64 and ARM64 CPU workers and GPU-capable workers where the backend supports them. Retro scene import is independent of the renderer: a preserved classic scene may be rendered in original/authentic mode or sent through the same modern farm in enhanced mode.

## User experience

From AmigaOS, AmiRender should feel like an external render coprocessor, not like a Linux server.

## License

Software: MIT.
