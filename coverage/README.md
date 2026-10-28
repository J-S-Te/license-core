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
  side effect. Cron scheduling is new business; licensed production uses split
  API and Worker components. Keep both alive for history-only mode.
- CRM: bootstrap cookie/machine groups and every modules `RegisterRoutes`;
  individual worker `RunOnce` before claim/side effect; Temporal activities too.
  Import preview persists jobs and is not historical read. Existing audit
  delivery and identity lifecycle are not business notification generation.
- Customer Portal: `internal/portalbootstrap/router.go` browser and machine
  groups; report generation/request is business, existing report download grant
  and consumption are export. Project export worker is historical export.
- Project: HTTP/application service checks; SLA notifier scan is business,
  dispatcher requires destination/message-purpose review. Worker and notifier
  now use the shared background runtime and remain mixed execution; the finite
  taskKinds and named activity policies are checked separately.
- Settlement: explicit `ServeHTTP` and reviewed role helper dispatch predicates
  plus finite method/path policies and services;
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
audit. Classifying an export does not manufacture a working exporter. CRM
presale report export remains explicitly unavailable. Portal filing now has a
PDF renderer wired into bootstrap; its actual state/permission/file checks are
described in the [portal manual](../../sharedDocs/10_操作手册/客户自助门户操作手册.md),
and must not be inferred solely from a historical classification entry.

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

## Release operator procedure (2026-10-13)

Source reviewers inspect handler bodies, service persistence, worker dispatch and side effects before updating classification. The commands above are available checks, not tests run in this documentation pass. Candidate generation rereads the fixed registry and invokes Scan/Check, then hashes the reviewed application source set; old digests after source changes cannot prove coverage.

Successful classification is only a source gate. Release administrators separately verify actual images, deploy selected components, confirm current READY/ACK and validate business/expiry/offline behavior. Unknown/drifted entries require semantic review, not relabeling as essential. CRM presale report export remains unavailable even when classified EXPORT_HISTORY. Licensed production Contract runs split API/Worker; keep both alive during expiry.

As of source review 2026-10-22, legacy V1 uses approved image/coverage digests for
runtime enrollment; the built-in V2 directory uses stable component identity.
This distinction does not remove the source review, release integrity or actual
execution gates. The commands here were not rerun as business verification in
this documentation update.

Evidence: registry.go Scan/Check, reviewed.go, tasks.go → commercial-license-tool/cmd/runtime-approval/main.go → internal/runtimeapproval/approval.go CoverageDigest/Prepare. See the [approval guide](../../sharedDocs/09_商业授权/02_运行组件批准清单.md); runtime operation maps remain separate in each business repository.

## Source review 2026-10-28

The reviewed registry contains 673 source boundaries and 218 source digests.
The deleted Settlement operations.go was traced to its real service split,
without restoring a placeholder. Role helper routing keeps separate read/write/
export policies, and the new recovery/reminder tasks remain business mutation.
Project background activity registration with options and its real taskKinds
are included; unknown members and changed handlers cannot inherit reviewed
policies. The gate comparison itself is unchanged.

Audit evidence and local regression results are maintained in
`.ai/qa/source-registry-review-20261028/`. These source classifications do not
certify runtime operation mappings: CRM notification projection currently uses
a looser runtime operation, while Portal validation and project notification
execution are stricter than their semantic classifications. Settlement's project
reference client wiring also needs separate business validation. Those business
changes were not made in this source-registry task.
