# Runtime synchronization client

`New(Options)` accepts an explicitly configured platform origin, service ID,
approved coverage/config-image SHA-256 digests, a machine token source and a
durable runtime. The token must be for an independently registered service with
`license.runtime` scope. No caller-supplied instance/application/environment
field is sent to replace the authenticated platform binding.

`Run(ctx, onError)` synchronizes immediately and every 30 seconds:

1. POST `/api/v1/internal/licenses/runtime/:service_id/ready`.
2. GET `/api/v1/internal/licenses/runtime/:service_id/snapshot`.
3. Verify and durably apply the signed snapshot through the runtime.
4. POST `/api/v1/internal/licenses/runtime/:service_id/ack` only for the exact
   revision and signed raw content actually stored locally.

Network errors do not delete or reset local state, change expiration times, or
acknowledge an unapplied revision. Business callers must still call `Evaluate`
at each execution boundary. This client alone does not enforce a business API
or Worker. ACK loss is retried on the next synchronization; it does not restore
an earlier local state. No redirect is followed; requests have a 10-second
timeout and bounded responses. HTTP requires explicit deployment approval via
`AllowHTTP`; HTTPS is the default for the adapter. Existing system HTTP default
is not changed by this library.

The adapter is not currently copied into the platform's production core mirror
because platform does not consume it. It must enter each business repository's
reviewed build allowlist when that service is actually integrated. The AST
coverage tool is development-only and is not a production dependency.
