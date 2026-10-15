# Building the AmigaOS/m68k client

The portable client code is tested by the normal Linux CI. The Amiga-specific
transport lives in `src/bsdsocket_amiga.c` and targets an AmigaOS-compatible
m68k toolchain with `bsdsocket.library` development headers.

Expected source set:

```text
src/submit.c
src/client.c
src/bsdsocket_amiga.c
```

The transport opens `bsdsocket.library` version 4, resolves the controller or
worker host with `gethostbyname()`, creates an IPv4 TCP socket and exposes it
through the portable `amirender_transport` interface.

The default AmiRender TCP port is 6800.

## Qualification

M1 qualification should compile this source with the Ploos `amiga-dev` Bebbo
toolchain, then execute the resulting m68k client in `amiga-runtime` against a
real AmiRender node. Runtime qualification must use an m68k Amiga environment;
AROS/i386 is not a substitute for this gate.

No Kickstart ROM, AmigaOS installation, or other proprietary system file is
stored in this repository.
