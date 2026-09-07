# Architecture

RepoSweep v0.1 is a local event pipeline compiled into one executable:

```text
OS file notifications        agent hook / CLI
         |                         |
         v                         v
 recursive watcher           POST /api/events
         |                         |
         +------> evidence event <-+
                       |
                 bounded event hub
                       |
               snapshot API + SSE
                       |
                embedded dashboard
```

## Truth contract

Every event carries one confidence class:

- `observed`: directly measured by RepoSweep, such as a filesystem write.
- `reported`: asserted by an agent integration, such as intent or test status.
- `inferred`: a deterministic heuristic derived from observed evidence.

The UI must preserve that distinction. Future AST, Git, runtime, and model adapters may add evidence but must not silently upgrade an inference into an observation. See [docs/concepts/evidence.md](docs/concepts/evidence.md) for the full contract, including resolution chains and the boundary a language model may not cross.

## Why Go

The prototype needs a native, cross-platform executable, cheap concurrency, a mature networking standard library, and simple contributor onboarding. The only runtime dependency outside the standard library is `fsnotify`, which provides direct operating-system file notification adapters. The web UI uses no framework and is embedded at compile time.

## Public contracts

The first integration contract is the agent event API:

```json
{
  "agent": "codex",
  "kind": "intent",
  "message": "Refactor the authorization flow",
  "path": "internal/auth/policy.go",
  "status": "active"
}
```

The next stable contract is the Repository Intelligence Record (RIR), covering symbols, entry points, edges, flows, and changes, each carrying its own evidence and confidence. It is specified in [RFC 0001](docs/rfcs/0001-repository-intelligence-record.md) and the edge vocabulary in [RFC 0002](docs/rfcs/0002-graph-model.md). Both are Draft: the format is unstable and will change before analyzer plugins are accepted.

## Reliability boundaries

- The HTTP server binds to loopback by default.
- Request bodies are bounded and unknown JSON fields are rejected.
- Event history and subscriber buffers are bounded.
- Filesystem bursts are coalesced per path.
- Generated, dependency, VCS, and temporary paths are excluded.
- Slow dashboard clients cannot stall repository watching.

## Known v0.1 limits

- Active-agent attribution is temporal and explicit, not process-level provenance.
- Events are not persisted across restarts.
- Classifiers use path evidence; they do not parse behavior yet.
- `.gitignore` rules are not yet evaluated beyond built-in directory exclusions.
- A rename may appear as separate operations depending on editor/OS behavior.
