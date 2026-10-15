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


## UPLOAD

A client may stage one small render asset on a worker before submitting a job.
M1 uses base64 inside the existing newline-delimited JSON framing.

```json
{"type":"UPLOAD","name":"scene.pov","data":"Y2FtZXJhIHt9Cg=="}
```

M1 constraints:

- decoded payload size is limited to 1 MiB;
- `name` is a basename only; directory traversal and client-selected worker paths are rejected;
- the worker creates an isolated staging directory and chooses the filesystem path;
- staged files are private to the worker and are not a durable asset store;
- M1 initially targets a self-contained single-file POV-Ray scene. Includes, textures and multi-file bundles require a later asset-manifest protocol.

## STAGED

A successful upload returns the worker-selected asset reference:

```json
{"type":"STAGED","asset":"/worker-selected/staging/path/scene.pov"}
```

The path is an opaque M1 asset reference from the client's perspective. Clients must not construct, modify, or predict it. A following `SUBMIT` may use this returned reference as its scene value.

Invalid names, invalid base64, empty data, oversized data, or staging failures return `FAILED`.


## SUBMIT and remote output

A remote render job may use a staged asset reference as its scene. The client's requested output path is a client-side destination hint; it is not a filesystem path that a Linux worker may write directly.

For engines that produce files remotely, the worker creates its own staging directory and returns the staged result in `COMPLETE.output`.

Example:

```json
{"type":"SUBMIT","job":{"id":"frame-1","engine":"povray","scene":"<staged-scene>","output":"RAM:frame.png","width":320,"height":256}}
```

```json
{"type":"COMPLETE","job_id":"frame-1","engine":"povray","output":"<worker-output-asset>"}
```

Clients must treat `COMPLETE.output` as an opaque worker asset reference in remote mode. In particular, a worker must not interpret an Amiga path such as `RAM:frame.png` or `DH0:...` as a local Linux path.

## DOWNLOAD

A client retrieves a staged result by sending the asset reference returned by the worker:

```json
{"type":"DOWNLOAD","asset":"<worker-output-asset>"}
```

M1 uses the same bounded base64 transfer model as `UPLOAD`. The requested asset must be a worker-controlled staged asset; arbitrary filesystem paths are rejected.

## DATA

A successful download returns:

```json
{"type":"DATA","asset":"<worker-output-asset>","data":"<base64>"}
```

The client decodes `data` and writes it to its own requested local output path. This keeps Amiga filesystem semantics on the Amiga side and Linux filesystem semantics on the worker side.

## M1 end-to-end contract

The qualified single-file POV-Ray flow is:

```text
local scene
  -> UPLOAD
  <- STAGED
  -> SUBMIT
  <- COMPLETE
  -> DOWNLOAD
  <- DATA
  -> local output
```

CI exercises this sequence over one TCP connection with a real POV-Ray render and verifies that the downloaded result has a valid PNG signature.

## Known M1 transfer limitations

The current base64-in-JSON transfer is deliberately simple and bounded. It is suitable for qualification and small assets, not production-size scenes or render outputs.

A later protocol revision should add chunked upload/download, opaque asset IDs rather than filesystem paths, explicit asset sizes/hashes, multi-file manifests, and staged-asset lifetime/cleanup semantics.
