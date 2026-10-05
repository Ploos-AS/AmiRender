# Engine adapter contract

Every engine adapter exposes the same lifecycle:

1. probe
2. capabilities
3. prepare
4. render
5. progress
6. cancel
7. collect
8. cleanup

Adapters may represent:

- a native executable such as POV-Ray
- a translator into AmiRender IR
- Blender/Cycles or Eevee
- AmiVM plus an original m68k renderer
- AmiTerrain or Ami3D native render cores

The controller must not contain application-specific rendering logic.
