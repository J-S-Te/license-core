# Offline commercial license core

Module `github.com/J-S-Te/license-core`, package `licensecore`, Go 1.25.4.
This package uses only the standard library. It does not automatically integrate
any platform, HTTP route, database, worker, or Temporal activity.

The protocol is version 1, product `unified-business-suite`, compact JWS with
`alg=EdDSA`, `typ=commercial-license+jws`, and a vendor `kid`. Verification only
uses caller-supplied trusted public keys. Tokens may not provide keys or alternate
algorithms. JSON duplicate keys, unknown fields, field-name aliases, trailing
JSON, noncanonical base64url, and tokens larger than 256 KiB are rejected.
Both License and Recovery require the fixed signed payload field
`issuer=unified-business-suite-vendor`; absent or different issuers are rejected.
The issuer name does not replace verification against trusted vendor public keys.

Applications are `customer_and_opportunity`, `customer_portal`,
`contract_management`, `project_management`, `settlement`, and `data_analysis`.
Kind is `FULL` or `TRIAL`. IDs, customer, instance, environment, and key IDs are
1–128 ASCII letters/digits or `. _ : -` with no spaces. Version must be positive.
Times are integer UTC Unix seconds in `(0,253402300799]`. IssuedAt cannot be later
than global NotBefore. Application dates may precede the new global NotBefore
when renewal preserves earlier grants; Evaluate enforces both activation times.
ExpiresAt must be strictly later than application NotBefore.

`ValidateRequest` validates application requests. `ParseRequest` also applies the
strict JSON format. Requests are unsigned identification inputs; parsing does not
authenticate a machine, customer, or requester.

`Sign` validates and signs a License; `Verify` verifies the signature and strict
schema. `Evaluate` must only receive a successfully verified License. The License
value itself carries no proof of verification; callers must not populate it from
an untrusted JSON object and call Evaluate directly.

For business operations, application presence, instance/environment binding, and
activation time are required. Expiry is exclusive: at `now == ExpiresAt`,
MUTATE_BUSINESS is denied, but READ_HISTORY and EXPORT_HISTORY remain available.
FULL and TRIAL have the same expiry policy. ESSENTIAL_SERVICE bypasses commercial
bindings and time restrictions after schema and operation validation. Essential
services should remain available when no usable license exists; integrations
must explicitly route them without inventing or accepting malformed licenses.
An unknown operation never grants access.

`CompareVersions(next,current)` rejects changes of customer, instance,
environment, or product and requires a strictly greater version. The caller must
verify both signatures and persist the accepted version atomically. This helper
does not create a durable antirollback store.

`Digest(raw,trusted)` verifies a license token before returning the lowercase hex
SHA256 of its exact compact JWS bytes. `Diff(next,current)` validates the same
binding and strict version increase, then returns application changes sorted by
code with `Kind` ADDED/REMOVED/UPDATED and copied Old/New snapshots. Unchanged
applications do not appear. Diff itself does not verify signatures and callers
must supply authenticated licenses.

`TimeObserver.Observe` optionally tracks a monotonic UTC high-water mark, with a
caller-chosen rollback tolerance. Its TimeStore must implement durable Load and
atomic SaveMax across processes. Storage errors fail closed; this interface alone
does not prevent clock/OS/storage tampering. No service currently calls it merely
by importing this module.

Validation: `go test ./...`, `go test -race ./...`, `go vet ./...`, `go build ./...`.

Clock recovery uses a separate `Recovery` schema and
`typ=commercial-clock-recovery+jws`, signed by the same trusted vendor key IDs.
`SignRecovery`, `VerifyRecovery`, and `ValidateRecovery(recovery,license,now)`
are separate from license verification; neither token type accepts the other.
The recovery is bound to product, instance, environment, license ID, and exact
license version. IssuedAt is inclusive and ExpiresAt exclusive for use; AnchorAt
must be between IssuedAt and ExpiresAt (inclusive) and all timestamps obey the
same protocol bounds. Validation requires a trusted current time: a compromised
local clock cannot itself establish recovery eligibility.

The core does not consume recovery IDs, lower or reset a persistent high-water
mark, or implement replay prevention. A future platform transaction must verify
both signatures, validate binding/time, reject an already consumed recovery ID,
persist consumption and apply the authorized anchor atomically, with audit and
appropriate exclusive access. Calling VerifyRecovery or ValidateRecovery alone
does not perform recovery or prevent replay.
