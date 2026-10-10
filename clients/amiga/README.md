# AmiRender Amiga client

This directory contains the first AmigaOS/m68k submission client components.

M1 deliberately separates the portable wire-protocol builder from networking. The
portable C layer builds the same newline-delimited JSON `SUBMIT` message accepted
by `amirender-node`. The Amiga transport layer will use `bsdsocket.library`.

The protocol builder rejects control characters, quotes and backslashes in text
fields rather than attempting incomplete JSON escaping. A later protocol utility
may add full escaping.

Target integration:

```text
Amiga application / CLI / ARexx
             |
     portable submit builder
             |
       bsdsocket.library
             |
       AmiRender node
```

No proprietary AmigaOS or Kickstart files belong in this repository.
