# AmiRender protocol notes

M0 uses newline-delimited UTF-8 JSON messages over TCP for ease of inspection and implementation on AmigaOS.

Initial message classes:

- HELLO
- CAPABILITIES
- SUBMIT
- ACCEPTED
- PROGRESS
- COMPLETE
- FAILED
- CANCEL
- PING
- PONG

Large scene assets are intentionally outside the control-message body. A later milestone will add content-addressed asset transfer using hashes.

Protocol version 1 favours simplicity over compactness. A compact binary transport may be added later without changing the job semantics.
