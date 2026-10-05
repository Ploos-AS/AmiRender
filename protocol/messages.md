# Control messages

AmiRender control messages are newline-delimited JSON.

## REGISTER

A worker registers itself with the controller and advertises render capabilities.

Example:

```json
{
  "type": "REGISTER",
  "node": {
    "version": 1,
    "id": "rock5-01",
    "arch": "arm64",
    "os": "linux",
    "memory_mb": 16384,
    "capabilities": [
      {"engine": "povray"}
    ]
  }
}
```

Re-registering the same node ID replaces its previous capability record. This allows a worker to update capabilities after configuration or runtime changes.

## REGISTERED

The controller acknowledges a valid registration:

```json
{"type":"REGISTERED","node_id":"rock5-01"}
```

Registration is the M1 discovery baseline. Network service discovery such as mDNS may locate controllers/workers later, but capability truth comes from the registration protocol.
