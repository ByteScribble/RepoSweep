# Evidence

Every claim RepoSweep makes must be traceable to something a reader can check.
This is the project's central constraint, and most other design decisions follow
from it.

## Why this comes first

A tool that explains a codebase is only useful if its explanations can be
trusted, and trust in this category is easy to lose and hard to rebuild. One
confidently wrong architectural claim during a code review teaches an engineer
to stop reading the tool. It does not teach them to read it more carefully.

So RepoSweep never asserts something it cannot show. When it does not know, it
says so.

## The confidence vocabulary

Every event and every record carries exactly one class:

| Class | Meaning | Example |
|---|---|---|
| `observed` | RepoSweep measured this directly. | A file was written. A call expression names a concrete method. |
| `reported` | An external integration asserted it. | An agent posted its intent. A test runner reported a failure. |
| `inferred` | Deterministically derived from observed evidence, by a rule that can be stated. | An interface call resolved through a single registered implementation. |

Three rules govern the vocabulary:

1. **An inference never becomes an observation.** Adding a second heuristic that
   agrees with the first does not upgrade the class. Only direct measurement
   produces `observed`.
2. **`reported` is never treated as verified.** An agent stating that it fixed
   authentication is evidence about the agent, not about the code.
3. **The class survives to the interface.** The dashboard shows it. The API
   returns it. A projection that drops the distinction is a bug.

## Resolution chains

`inferred` alone is not enough for anything non-trivial. Dependency injection is
the clearest case: knowing that `PaymentService` reaches `StripeGateway` is only
useful if you can see why.

```text
PaymentService.charge  →  StripeGateway.charge

confidence: inferred
resolution:
  s.gateway has interface type payment.Gateway
  payment.Gateway has one registered implementation
  internal/app/wire.go:31 binds it to *StripeGateway
evidence:
  internal/payment/service.go:64   s.gateway.Charge(ctx, amount)
  internal/app/wire.go:31          wire.Bind(new(Gateway), new(*StripeGateway))
```

A reader who disagrees can now check each step. That is the difference between a
tool making an argument and a tool making an assertion.

Any record with `confidence: inferred` must carry a `resolution` chain. See
[RFC 0001](../rfcs/0001-repository-intelligence-record.md).

## Unknown is a valid answer

When an analyzer cannot resolve a target, it emits an unresolved relation with a
reason. It does not pick the most likely candidate.

This matters more than it first appears. A tool that guesses is right most of the
time, which is worse than being wrong consistently: the failures are silent,
scattered, and indistinguishable from the successes. Every projection built on
top inherits the error with no way to detect it.

Empty results are a legitimate output. "Three implementations, none determined"
is a useful thing to tell an engineer.

## Where a language model is allowed

The boundary is strict, and it is the most important design decision in the
project.

**The deterministic engine answers questions of fact:**

```text
What exists?               Who calls this?
Where is it?               What does it call?
What changed?              What tables does it touch?
Which events does it emit? What is the path from this entry point?
What is the blast radius?
```

**A model may answer questions of description:**

```text
Explain this flow in a sentence.
Summarise what changed today.
Explain this component to a new engineer.
Draft documentation for this module.
```

**A model never contributes topology.** It does not create a symbol, an entry
point, an edge, a flow, or a change record. It does not resolve an ambiguous
call. It does not decide a risk level. It receives the graph as context and
writes prose about it.

```text
        code
          │
          ▼
  deterministic engine
          │
          ├──────────────►  interface
          │
          └──►  context  ──►  model  ──►  description (advisory)
```

If the model is removed, RepoSweep still works and every claim still stands. The
descriptions get worse. Nothing becomes wrong.

This is also why RepoSweep runs without a model provider configured, and why it
will keep doing so.

## What this rules out

Some otherwise appealing features are excluded by the above, and it is worth
naming them so they are not proposed repeatedly:

- Asking a model to guess which implementation an interface call resolves to.
- Summarising a diff by feeding raw hunks to a model and reporting the result as
  a change record.
- Numeric confidence scores, which imply a precision no analyzer here has.
- Any view that shows a model-generated claim with the same weight as a measured
  one.
