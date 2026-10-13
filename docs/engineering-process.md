# Engineering process and roadmap

The project advances through contract-led vertical slices, not a list of machines
to support. Every milestone identifies a capability, its normalized contract,
Provider responsibilities and a bounded acceptance matrix.

## Milestone progression

| Stage | Engineering focus | Current interpretation |
|---|---|---|
| M0 | Standalone Controller, local Agent, durable state and basic Jobs | Implemented |
| M1 | Approved multi-Node identity, transport and reconciliation | Implemented with bounded acceptance |
| M1.5 | Independent Runtime protocol and execution fencing | Implemented with bounded acceptance |
| M2 and integration increments | Native execution, remote Jobs and verified input/output handoff | Implemented with bounded acceptance |
| M2.5 | Resource Providers, GPU execution, telemetry and large-data transfer/cache | Implemented with qualified environment records |
| M2.6 | Deterministic, explainable placement | Implemented |
| M3 | Distributed DAG workflows using ordinary Jobs | Implemented with bounded cross-platform acceptance |
| M3.5 | Resource envelopes, performance history, Shadow ETA, renewable Leases and bounded management | Current maturity phase; delivered increments and further work coexist |
| M4 and M4.5 | Persistent Workspace platform and source/environment/execution binding | Planned main integration |
| M5 and M5.5 | Persistent deployment lifecycle and visual workflows | Planned |
| M6 and M6.5 | Operations portal and platform hardening/recovery | Planned |
| M7 | External Chat or local AI orchestration through the Common API | Planned |

The original broad milestone labels evolved into these delivery boundaries.
Implementation of a basic Workspace does not make the persistent Workspace
platform complete. A Windows service implementation does not complete the broader
deployment milestone. Shadow ETA does not enable adaptive scheduling.

M3.5 management has an operator-approved package-definition closure for a bounded
Linux scope. That is not a full qualification verdict. Lifecycle/recovery work
continues on the roadmap, while C1, fixed Linux Agent/Runtime pair maintenance,
and Windows service mode have merged implementations with further physical
acceptance in progress. The bounded pair maintenance flow is not completion of
the broader persistent deployment platform.

## Evidence before verdict

Acceptance uses distinct layers: generic contract, platform/Provider, backend,
actual device/environment, and multi-Provider integration. A passing build or
isolated fixture cannot substitute for an actual device gate.

Representative recorded campaigns exercised native execution, cancellation,
restart reconciliation, exclusive allocation, verified input staging, artifact
publication/ACK/cleanup and cross-platform workflow completion. Where a result
requires real hardware, the campaign records that execution separately from
controlled metadata and simulated negative cases.

| Evidence label | Meaning |
|---|---|
| PASS | The declared check passed within its recorded source and environment |
| FAIL | The declared check failed; its historical record remains retained |
| NOT RUN | The required action was not executed |
| NOT VERIFIED | Available evidence does not establish the claim |

Independent evaluation checks results, ownership and delivery evidence rather
than accepting a success banner. Fixes append new source-bound results; they do
not turn an earlier failure into a historical pass. Merged code, installed code,
unit tests and physical qualification are separate facts.

One recorded same-host Go test/subtest sweep for a continuity change reported
**838 passed, 0 failed and 7 skipped**. Those are snapshot test results, not a
performance benchmark or a device-qualification total.

## AI-assisted development

AI coding agents work inside the same specification and approval boundaries as
other contributors. Repository instructions and a document-routing index keep
recurring prompts short without asking the agent to reload every historical file.

```mermaid
flowchart LR
    Goal[Operator goal and scope] --> Spec[Relevant contracts and ADRs]
    Spec --> Change[Smallest coherent change]
    Change --> Review[Diff review and boundary checks]
    Review --> Test[Proportionate validation and selected acceptance]
    Test --> Report[Evidence-labelled report and PR review]
```

The operator retains decisions about shared contracts, privilege, publication and
deployment. An agent must preserve unrelated work, inspect actual state before
editing, and separate implementation from installation and qualification.
The development assistant is not the product's resource Scheduler.

### Failure handling: A / B / C

These labels summarize the repository's existing work rules for this showcase;
they are not a new runtime protocol or additional execution permissions.

| Policy | Situation | Response |
|---|---|---|
| A — Correct within scope | A defect introduced by the approved change can be repaired without changing its authority or shared contract | Make the smallest correction and rerun the affected validation; preserve the earlier result |
| B — Request a decision | The proposed remedy changes a shared contract, product behavior, data, cost, privilege or irreversible scope | Explain the evidence and proposed boundary; obtain a decision before the affected action while continuing independent approved work |
| C — Stop the affected action | Required authority, access or environment evidence is unavailable | Report the exact blocker and NOT RUN/NOT VERIFIED boundary; do not invent access, repeat an unchanged failing method or infer success |

The eventual product goal is for humans and external AI systems to describe
workflows and constraints while Fabric resolves resources, execution and evidence
through deterministic contracts. That orchestration layer remains planned.
