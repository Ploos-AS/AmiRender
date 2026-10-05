# AmiRender

A render accelerator appliance and distributed rendering system for classic Amiga workflows.

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

## User experience

From AmigaOS, AmiRender should feel like an external render coprocessor, not like a Linux server.

## License

Software: MIT.
