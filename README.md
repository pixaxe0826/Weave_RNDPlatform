# Weave | R&D Fabric

## At a glance

- **Implementation project:** 39 merged PRs and 43 ADRs.
- **4 execution contexts:** Windows native; Linux AMD64; AMD GPU/ROCm; Android ARM64 Termux/PRoot.
- **Latest recorded Go sweep (PR #42):** 876 passed, 0 failed, 7 skipped.

**A personal, local-first R&D execution fabric that lets heterogeneous machines
work together through governed Jobs, resource Leases, and verified artifacts.**

## Why build this?

A personal R&D setup can contain a Windows workstation, Linux compute, a GPU,
and a small Android device. They offer different capabilities, but using each
independently means coordinating execution and moving results between environments.

R&D Fabric brings those capabilities into one execution model. A machine is not
a special product type: it is a Node exposing resources and compatible Providers.
The aim is to reuse available hardware without requiring a homogeneous cluster.

## What makes it interesting?

- **Explainable placement:** deterministic eligibility and ranking, not an LLM
  choosing machines or overriding resource policy.
- **Real execution:** native processes run through Agent and Runtime, with normal
  Claim, Lease, generation, cancellation, and completion accounting.
- **Verified data movement:** artifacts travel directly between a Node and Store;
  signed receipts and acknowledgments govern publication and owned cleanup.
- **Composable workflows:** a bounded DAG turns each task into an ordinary Job;
  intermediate results become immutable inputs for downstream tasks.
- **Continuity across Controllers:** signed, audited C1 identity transfer preserves
  the enrolled Node identity and advances a monotonic ContinuityFloor.
- **Separate administration:** approved typed Node Manager actions are distinct
  from ordinary workload execution.

## Components at a glance

| Component | Responsibility |
|---|---|
| Controller | Policy, Node registry, deterministic scheduling, Leases, workflow state and audit |
| Store | Authoritative artifact publication and direct data transfer |
| Node Manager | Approved typed administration within an explicit policy |
| Agent | Node observations, authenticated control and independent local enforcement |
| Runtime | Actual native execution, owned process lifetime and execution evidence |

Controller and Store have native Windows service implementations with an
operating service cutover PASS record. Further unattended-boot and crash-recovery
acceptance is in progress. Compute evidence includes selected Windows native,
Linux AMD64, AMD GPU/ROCm, and Android
ARM64 Termux/PRoot environments. These are bounded qualification contexts, not
universal platform or hardware support.

## Current state

The project is a research prototype in the **M3.5 resource, performance and
management maturity phase**.

| Area | State |
|---|---|
| Native Jobs, deterministic scheduling, Store-backed workflows and large-data transfer/cache | Implemented, with actual-device acceptance records |
| Resource envelopes, performance observations, passive task history and Shadow ETA | Implemented; Shadow ETA does not change current placement |
| Renewable Leases | Implemented, with long-running CPU acceptance records |
| Linux privileged-management package | Delivered within an operator-approved, bounded inspection/policy scope; package closure is not full qualification |
| C1 identity transfer | C1: implemented, and passed actual-device acceptance on one Android ARM64 Node. Identity transfer for other Nodes is in progress. |
| Windows Controller and Store services | Operating service cutover passed deployment checks; further unattended-boot and crash-recovery acceptance is in progress |
| Fixed Linux Agent/Runtime pair maintenance | Implemented and tested; physical acceptance is in progress |
| Adaptive ETA scheduling and extended accelerator memory/performance models | Planned |
| Persistent Workspace platform, deployment lifecycle, visual workflow, operations portal and AI orchestration | Planned |

The intended future AI role is to express goals and workflow constraints through
the Common API. Fabric remains the deterministic execution and resource layer.

## Technology

Go · SQLite · gRPC / Protocol Buffers · HTTPS / mutual TLS · native process
runtimes · CUDA and ROCm Providers · Windows Service Control Manager · Linux
systemd · Python acceptance tooling.

## Code samples

![Go samples CI](https://github.com/pixaxe0826/Weave_RNDPlatform/actions/workflows/ci.yml/badge.svg)

- [Continuity](go/continuity): checked generation allocation, a monotonic state
  guard, and explicit rejection of unknown high-water evidence.
  See [Monotonic ContinuityFloor](docs/trust-model.md#monotonic-continuityfloor).
- [Placement](go/placement): Node/resource eligibility checks and deterministic
  lexicographic ranking with a stable tie-break. Minimal normalized inputs replace
  private dependencies; score construction and execution authority are excluded.
  See [placement and execution](docs/architecture.md#workflow-composition).

Both packages share a standard-library-only Go module. From the `go` directory,
run `go vet ./...` and `go test ./...`. These are standalone excerpts, not the
production Scheduler, persistence layer or identity-transfer protocol.

## Read more

- [Architecture and execution flow](docs/architecture.md)
- [Trust and continuity design](docs/trust-model.md)
- [Engineering process and roadmap](docs/engineering-process.md)

This repository is a documentation-only hackathon showcase. It is not an
installation package or a copy of the implementation repository. Diagrams are
Mermaid source; operational configuration and evidence bundles are not included.
