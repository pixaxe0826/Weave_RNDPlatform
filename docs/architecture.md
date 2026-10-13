# Architecture

R&D Fabric separates decisions, execution, data transfer and administration.
The Controller coordinates metadata; it does not relay bulk artifacts or perform
workflow computation itself. A physical machine may host several logical roles.

## Logical components

```mermaid
flowchart TB
    User[Human client] --> API[Common API]
    API --> Controller[Controller: Windows service mode]
    Controller --> Scheduler[Deterministic Scheduler and Lease ledger]
    Controller --> Workflow[Workflow state and audit]
    Controller --> Agent[Node Agent]
    Controller --> Manager[Privileged Node Manager]
    Manager --> Admin[Approved typed administration]
    Agent --> Runtime[Unprivileged Runtime]
    Runtime --> Compute[Native CPU or qualified GPU computation]
    Agent <-->|Direct data transfer| Store[Authoritative Store: Windows service mode]
    Store -.->|Publication receipts| Controller
```

| Boundary | Role |
|---|---|
| Controller | Own durable control state, policy, registry, scheduling intent, Leases and audit |
| Store | Verify declared artifact bytes and publish immutable objects with receipts |
| Node Manager | Independently validate approved typed management requests |
| Agent | Report capabilities and observations; enforce local resource and execution constraints |
| Runtime | Start and observe real native processes under the granted allocation |
| Provider | Normalize platform/backend facts into versioned Fabric contracts |

Windows service mode is implemented for the Controller and Store; its further
physical deployment acceptance is in progress. This is distinct from the future
cross-platform deployment lifecycle milestone.

## Heterogeneous Node contexts

| Context | Bounded capability represented |
|---|---|
| Windows native execution | CPU execution and a qualified dedicated GPU Provider |
| Native Linux AMD64 | CPU execution, ordinary Agent/Runtime lifecycle and selected management packaging |
| Linux AMD64 with AMD GPU | A separately qualified ROCm execution environment |
| Android ARM64 with Termux and Linux PRoot | A constrained Linux execution context for ordinary workloads |

The Scheduler consumes normalized capability, capacity, approval, health,
freshness and reservation facts. It does not branch on a machine nickname or
hardware product name. Qualification belongs to an exact supported environment;
adding a compatible device does not automatically inherit that evidence.

## Job, Lease, receipt and ACK

The following is the configured artifact-delivery path, not a generic promise
that every legacy execution profile publishes to Store.

```mermaid
sequenceDiagram
    participant Client
    participant Controller
    participant Agent
    participant Runtime
    participant Store
    Client->>Controller: Submit Job and resource requirements
    Controller->>Controller: Validate Claim and reserve concrete resources
    Controller->>Agent: Dispatch Job with real Lease and generation
    Agent->>Agent: Recheck local allocation and generation
    Agent->>Runtime: Execute under the granted allocation
    Runtime-->>Agent: Process result and finalized output
    Agent->>Store: Direct upload under scoped authorization
    Store-->>Agent: Signed publication receipt
    Agent->>Controller: Report result and receipt
    Controller->>Controller: Validate publication evidence
    Controller-->>Agent: ACK verified delivery
    Agent->>Agent: Persist ACK and clean owned temporary output
    Agent-->>Controller: Terminal and cleanup evidence
    Controller->>Controller: Reconcile completion and release Lease
```

Process exit alone is not verified artifact delivery. Publication must match the
declared output identity and integrity metadata. The Agent persists the ACK
before removing owned source copies. Delivery recovery reconciles the same
execution rather than recomputing the task merely because a transfer was interrupted.

## Workflow composition

```mermaid
flowchart LR
    A[Job A on a compatible Node] --> SA[Immutable artifact in Store]
    SA --> B[Job B on a compatible Node]
    B --> SB[Immutable artifact in Store]
    SB --> C[Reducer Job on a compatible Node]
```

Each workflow task uses the same Scheduler, Claim, Lease, Agent and Runtime path.
The Controller advances dependencies using durable result and publication evidence.
Reducers execute on Nodes. Intermediates use Store, not a separate direct peer
execution protocol.

CPU, memory and GPU allocations describe resources, not guaranteed performance.
Current performance history and Shadow ETA are observation/prediction layers;
adaptive ETA-aware placement is planned, not activated by those observations.
