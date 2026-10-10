# Commercial entry coverage gate

Run from `license-core`:

```sh
go test -race ./coverage/...
go vet ./coverage/...
go run ./coverage/cmd/entry-gate -root ..
```

`-inventory` prints a deterministic JSON inventory for review. Regeneration is
not approval: unresolved entries remain `UNCLASSIFIED` and the gate exits 1.
The checked-in registry is `.ai/qa/commercial-license/entry-registry.json`.
File SHA-256 includes handler bodies, while registration identities capture
receiver, method/path expression and handler. Changes require renewed semantic
review. No GET-wide or notification/approval-wide essential policy exists.

The denominator is **source registration expressions and executable/task
boundaries**, not expanded running routes or individual Temporal activities.
Conditional registrations, symbolic loop paths, imported registration helpers,
implicit HEAD/OPTIONS and struct activities require separate live route and
activity registration tests. The module does not enforce runtime licensing and
cannot, on its own, certify deployment or full entry coverage. New routing styles
or split registration files must be added to the inventory scope explicitly.

Known attachment boundaries:

- Contract: `internal/transport/httpapi/router.go` after identity; application
  service methods before persistent changes; `workflows.Activities` before each
  side effect. Cron scheduling is new business; do not stop the API's embedded
  worker container to implement history-only mode.
- CRM: bootstrap cookie/machine groups and every modules `RegisterRoutes`;
  individual worker `RunOnce` before claim/side effect; Temporal activities too.
  Import preview persists jobs and is not historical read. Existing audit
  delivery and identity lifecycle are not business notification generation.
- Customer Portal: `internal/portalbootstrap/router.go` browser and machine
  groups; report generation/request is business, existing report download grant
  and consumption are export. Project export worker is historical export.
- Project: HTTP/application service checks; SLA notifier scan is business,
  dispatcher requires destination/message-purpose review. Empty Temporal poller
  has no business activities today, not proof no old cluster tasks exist.
- Settlement: explicit `ServeHTTP` dispatch predicates plus services;
  task-level split inside the worker. Report export and its job bookkeeping stay
  available for licensed history mode; dunning/tax reconcile/new tax commands
  pause before claim. PLATFORM_AUDIT delivery remains independent.
- Data analysis: bootstrap API plus **both** embed issue and token proxy,
  including wildcard resources; aggregation queue/API/DSN passes and alert
  worker checks before claim and each side effect. Historical dashboards do not
  imply permission to compute new aggregate projections.

Mixed worker executables use `MIXED_EXECUTION`, never a whole-process essential
exemption. Their finite task policies and actual Run/RunOnce dispatch and outbox
destination declarations are reconciled; unknown calls/destinations remain
unclassified. Rejecting a task must preserve pending/retry state; a denied license
is not a terminal business failure. In-flight completion and notification
delivery policy need explicit tests rather than a process-wide exemption.

## Reviewed policy scope (2026-10-09)

`reviewed.go` holds finite source-scoped handler allowlists. Stored dashboard
snapshots and SQL aggregation over existing facts are historical queries; the
separate aggregation worker creates new projections and is business mutation.
Contact-phone historical read requires its existing authorization and mandatory
audit. Unsupported presale/filing exports stay explicitly unavailable; tagging
the intended operation does not manufacture a working export.

Executable `ESSENTIAL_SERVICE` applies **only to bootstrap/approved maintenance**,
not every request or goroutine started by that process. Each route/activity/task
must still be authorized independently. Security session/account revocation and
personal read acknowledgements do not exempt new business notifications.

Previously unresolved groups have been reviewed (2026-10-09):

1. CRM notification delivery/owner-change/assignment/progress workers: existing
   delivery is essential; owner-change/assignment/progress notification projection
   is business mutation, not automatically exempt merely because it consumes outbox.
2. Portal invitation `Verify`: writes expired state and audit despite the name;
   identity verification/expiry/audit is essential; consumption/provisioning
   remains business mutation.
3. BI embedding, old-token proxy and wildcard are prohibited on expiry by explicit
   user decision; ordinary history reads/exports remain separately classified.
4. Project SLA scan is business; existing immutable notification dispatch essential.
5. Settlement export is historical export; existing audit/notification delivery
   essential; dunning, tax reconcile and tax command issuance are business mutation.

Source classification can now pass independently. This is NOT a full release
pass: actual runtime gate attachment/expanded route and activity comparison,
worker expiry behavior, installation approval and browser acceptance remain
separate required evidence. Registered activity
methods are additionally inventoried, but runtime registration and installation
evidence remain separate gates, even after all source entries are classified.
