# Runtime synchronization client

`New(Options)` accepts an explicitly configured platform origin, service ID,
registration protocol/catalog version, a machine token source and a durable
runtime. Legacy V1 additionally requires approved coverage/config-image SHA-256
digests. V2 accepts optional canonical digests without using them as component
identity. The token must be for an independently registered service with
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

Current platform and business source mirrors include the reviewed adapter and
the shared component catalog. Mirror contents are governed by the existing
sync scripts, exact allowlists and SHA-256 manifests. The AST
coverage tool is development-only and is not a production dependency.

## Registration contract, checked 2026-10-22

Unset `RegistrationProtocol` and `CatalogVersion` select V1 for compatibility;
explicit V1 uses catalog version `0`. V2 requires `RegistrationProtocol=2` and
`CatalogVersion=1`; unsupported combinations fail. V2 READY includes these fields
and is checked against the platform's compiled component directory. Commercial
license JWS and signed runtime snapshots remain protocol `1`.

V2 credentials bind installation/application/environment/service and credential
generation, so ordinary image changes do not create a new component identity.
This is not a license or deployment bypass: controlled deployment, immutable
release verification, business permissions, current READY and natural ACK remain
separate checks. Source: `client.go` registration/New/SyncOnce,
`../consumer/consumer.go` registrationConfiguration, and `../catalog/catalog.go`.

## Troubleshooting and acceptance, 2026-10-13

After approved controlled enrollment, deployment administrators follow the [activation guide](../../sharedDocs/09_商业授权/03_运行时协调与激活说明.md). consumer/consumer.go obtains each independent license.runtime token with client_credentials/HTTP Basic at /oauth2/token. Machine routes bind OAuth application/environment/component before changing state.

Success requires the same signed bytes durably read back and accepted by ACK. Initial APPLYING→ENFORCED creates a new revision; confirm that too. Renewal can keep ENFORCED while invalidating old confirmation: check snapshot_current and every issued/ACK revision against application revision. Platform refuses ACK with READY older than60seconds.

401/403: check controlled client/scope/binding; revision conflict: re-fetch signed state; I/O failure: check independent volume; outage: fix transport without clearing cache. Clock latch needs separate recovery, never a deadline extension. client.go SyncOnce → runtime ApplySnapshot/ReadState → Platform coordination is the inspected chain. Current business mirrors already include this adapter; this pass did not execute sync or dynamically verify ACKs.
