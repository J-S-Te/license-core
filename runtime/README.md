# Local runtime license gate

This package is a standard-library foundation for Linux and macOS services. It
does not perform network synchronization or wire itself into Platform, APIs or
Workers. Authenticate the caller and apply the existing business permissions
before calling this commercial operation gate. `ESSENTIAL_SERVICE` is reserved
for explicitly classified service recovery/readiness work; never classify a
user's business mutation as essential to make it pass.

## Signed control contract

Source checked 2026-10-22: snapshot protocol remains `1`. Component registration
V2/catalog version `1` is a separate contract handled by catalog/consumer/syncclient;
it does not change these signed fields or relax verification and durable-state
rules. See [syncclient](../syncclient/README.md).

`PlatformSnapshot` has exact, required JSON fields: `protocol`, `instanceID`,
`environment`, `application`, `serviceID`, `revision`, `enforcementState`,
`migrationEligible`, `currentLicenseJWS`, `pendingLicenseJWS`,
`highestLicenseVersion`, and `snapshotIssuedAt`. Empty license strings explicitly
mean absence. `snapshotIssuedAt` and all license validity times are Unix seconds.

`SignSnapshot` emits compact Ed25519 JWS with `typ` equal to
`platform-runtime-snapshot+jws`. The runtime receives platform and independently
trusted vendor public keys from its caller, never the token. The shared production
consumer deploys platform trust read-only and compiles vendor trust into the
reviewed release. A platform snapshot cannot stand in for a
commercial license; the outer snapshot and both embedded commercial licenses
are verified independently. Duplicate fields, unknown fields, aliases, missing
or null fields, noncanonical base64url and wrong bindings are rejected. Signing
a snapshot validates control fields; embedded licenses are authenticated when
the consumer calls `VerifySnapshot` or `ApplySnapshot`.

Bindings are exact for the installation, environment, application and service.
The commercial licenses themselves bind installation and environment, and the
gate evaluates the bound application. Current and pending licenses must have
the same commercial customer/product binding and strictly increasing versions.
A pending license is the entire next license, with a later `not_before`; from
that absolute instant it replaces the entire old license, including removal of
applications. Application-level dates are still evaluated within that license.
It is not merged with the old entitlements.

## Durable use

Construct `NewFileRuntime(path, binding, platformPublicKeys, vendorPublicKeys)`.
Supply a received platform JWS to `ApplySnapshot(ctx, token, now)`; the package
does not fetch one. Call `Evaluate(ctx, operation, now)` at the point where the
operation is about to happen, including Worker execution rather than only task
submission. `ReadState` supplies an authenticated local diagnostics view.

New/equal/older snapshots never overwrite a newer revision. Highest observed
license versions cannot decrease; previously observed versions cannot receive
different license contents, current cannot decrease, and pending cannot vanish
without being superseded/promoted. Only signed `PENDING_ENFORCEMENT` with
`migrationEligible=true` grants transitional operation access. After `APPLYING`,
the state cannot return to pending. After `ENFORCED`, it cannot return to either
earlier state. These rules survive recreation of the runtime and process
restart. No unsigned setter/reset method exists.

Missing, corrupt, incorrectly permissioned or unavailable durable state denies
business operations; essential operations remain available. An initial valid
signed control snapshot can initialize a missing file. Deployments must protect
the state directory and treat a missing formerly established state as an
operational incident: file-only storage cannot detect a privileged restoration
of an older complete state/volume or prove an installation's history after
deletion. Restoring state requires obtaining the latest authenticated platform
revision; do not reuse a historical snapshot as a recovery bootstrap. No claim
of hardware-backed anti-tamper or resistance to service-owner/OS compromise is
made.

Every transaction takes a separate stable 0600 lock file using Unix `flock`,
including concurrent processes, and supports cancellation while waiting. State
reads refuse symlinks and nonregular/insecure files. Writes use a 0600 temporary
file, file `fsync`, atomic rename and parent-directory `fsync`. Keep the file and
lock in one service-owned directory on a local filesystem that supports flock,
atomic rename and sync semantics. Do not replace/remove the lock file while
processes are running. Public keys are defensively copied at construction.

## Clock behavior and recovery

The file persists the maximum observed wall time. License decisions use the
maximum of that time and the supplied current time, so tolerated clock jitter
cannot resurrect an expired license or reactivate the old whole pending
license. An observation more than 300 seconds behind the high-water mark
persists a blocked latch. Moving the clock forward, receiving a normal snapshot
or restarting does not clear it. Synchronization failures cannot slide an
absolute expiration forward.

`RecoverClock` accepts only the separate vendor-signed
`commercial-clock-recovery+jws` credential from the core package. It must bind
the highest observed current/pending license, be within its validity window and
within 300 seconds of its signed recovery anchor. It atomically consumes the
recovery ID and clears the clock latch while applying that anchor; consumed IDs
survive restart and cannot be replayed. The local ID ledger has a fixed maximum
of 1024 credentials; exhaustion fails closed and requires an explicit future
state-format migration. There is no boolean recovery override. Recovery changes
neither enforcement state nor license version/revision or expiration times.

The caller must use the actual observation clock, not a user-selected timestamp.
No network time or trusted time infrastructure is simulated by this package.

## Verification

Run from the `license-core` module:

```sh
go test ./runtime
go test -race ./runtime
go vet ./runtime
go build ./runtime
GOOS=linux GOARCH=amd64 go build ./runtime
```

This subpackage does not change the root commercial-license source files. The
existing mirror scripts already include reviewed runtime/syncclient/consumer
sources for business consumers; preserve their exact allowlist and digest
manifest checks. Do not copy issuer tools or private keys.

## Operator checks (2026-10-13)

Deployment administrators protect one durable volume per component, deliver the approved binding/public key and follow the [activation guide](../../sharedDocs/09_商业授权/03_运行时协调与激活说明.md). Confirm latest signed revision and independent ACK; state-file presence is insufficient. Do not edit JSON/delete the lock as recovery. Diagnose missing/corrupt state and obtain the latest authenticated snapshot.

History-only expiry is not a time-limited grace period and does not permit absent/future applications or clock-blocked state. Outages preserve state and absolute dates. Library RecoverClock exists, but normal consumer synchronization does not call it; a Platform recovery event cannot prove local recovery. Workers stay alive for synchronization and retry timers.

Evidence: consumer/consumer.go FromLookup/Check → syncclient/client.go SyncOnce → state.go ApplySnapshot/load/save/Evaluate and snapshot.go VerifySnapshot. Business mirror scripts already include reviewed runtime/syncclient/consumer sources. This pass did not execute tests, inspect live volumes or exercise recovery.
