# NeuroForge Trust — Abuse Tests

Read this when proving a finding or turning findings into regression tests.

An audit finds a flaw once. A test keeps it from returning, and it does not depend on anyone, human or model, remembering the rule. These shapes are parameterised by endpoint, so the same list applies to any codebase: fill in the route, the protected field and the invariant.

**Ground rules**

- **Tests are the developer's code.** Put them in the fix plan and write them only after "Proceed". Never add a test file during a scan.
- **Ask before running anything.** Say exactly what will run and against what, and proceed only on an explicit yes. "Proceed" approves writing the tests, not running them.
- **Local or test environments only.** Never a deployed, shared or third-party system.
- Write the tests in the project's existing test runner and match its conventions. If the project has no tests, describe the cases in plain language and let the developer choose the tooling.
- Every test asserts on **stored state**, not just on the status code. A success response that ignored the tampered field is a pass; a rejection that still wrote the row is a failure.

---

## The nine request-shaped tests

### 1. Tamper — Origin, Transition

Send a legitimate request with one server-owned field set to a value the attacker would want.

```
Given  a valid request to <endpoint>
When   it also carries <protected field> = <attacker's value>
Then   the stored <protected field> equals the server-derived value
       (the request is rejected or the field is ignored — pick one and assert it)
```

Run one case per protected field the endpoint could plausibly receive: the price, the role, the owner, the status, the timestamp. Include fields the request shape does *not* declare; that is what proves unknown fields are dropped.

### 2. Wrong caller — Permission

Call the endpoint as someone who should not be able to.

```
Given  <endpoint> is meant for <role>
When   it is called with no session, and then as each lesser role
Then   every call is refused and nothing is written
```

Cover the siblings as well as the main route: the batch variant, the export, the write beside a guarded read.

### 3. Cross scope — Scope, Response

Two actors in different scopes. One uses the other's identifiers.

```
Given  actor A and actor B in different scopes, and a record owned by B
When   A calls <endpoint> with B's record id
       (as the path id, as a body reference, inside a bulk array)
Then   the response is "not found", B's record is unchanged,
       and nothing about B's record appears in the response
```

Cover every route that takes an id: read, update, delete, nested children, and any foreign key accepted in a body. Add the tenant-selector case where one exists: A sends B's tenant in the header or subdomain.

### 4. Out of range — Bounds

Send values the client is allowed to choose, chosen badly.

```
Given  <endpoint> accepts <field> from the client
When   <field> is negative, zero, enormous, fractional, far in the past or future,
       or a real option that is not on offer to this caller
Then   the request is refused and nothing is written
```

One case per boundary. The negative-amount case is the one that most often turns out to move money backwards.

### 5. Illegal transition — Transition

Every move the state machine should refuse.

```
Given  a record in state <S>
When   <actor> attempts to move it to <T>, where S → T is not permitted for that actor
Then   the request is refused and the record is still in <S>
```

Generate the cases from the transition table: every (from, to, actor) triple that is not explicitly allowed. Include creating a record directly in a later state.

### 6. Replay — Replay

The same request, or the same external event, twice in a row.

```
Given  a request to <endpoint> that grants <effect>
When   the identical request is sent a second time
Then   <effect> has been applied exactly once,
       and the second response reports the original outcome
```

For events with two entry points (a webhook and a return page), send one through each.

### 7. Concurrent — Concurrency

Many requests at the same moment, against the real storage engine.

```
Given  <invariant> holds  (balance ≥ 0 · at most N rows · applied once)
When   M requests that each individually pass are sent simultaneously
Then   <invariant> still holds, and exactly the permitted number succeeded
```

This test is worthless against mocks or an in-memory fake: the bug lives in how the real database interleaves the statements. Run it against the same engine as production, and fire the requests genuinely in parallel rather than awaiting them one by one.

### 8. Redirected reach — Reach

Name a destination the server should refuse.

```
Given  <endpoint> accepts a destination from the client (a URL, a file name, a redirect, a recipient)
When   it is given a loopback or private address, a path that climbs out of its directory,
       an off-site redirect, or a recipient the caller has no relationship with
Then   the request is refused and the server contacted, opened or sent nothing
```

Use a harmless local stand-in for the forbidden target and assert that it received no request. Never aim this at a real internal service.

### 9. Flood — Abuse rate

Exceed the limit, then try to step around it.

```
Given  a limit of N per window on <endpoint>
When   one actor sends N + 1 requests
Then   the last is refused

When   a different actor (new address, new session) continues against the same target
Then   it is refused too
```

The second half is the one that matters. A limit keyed only on the caller's address stops nobody who can change addresses. For secret-guessing endpoints, also assert that the secret is invalidated after the permitted wrong guesses, across every endpoint that can check it.

---

## The two lifecycle tests

### 10. Revocation — Revocation

```
Given  a credential that has been rotated, revoked, logged out, or superseded by a password change
When   the old credential is presented
Then   it is refused

Given  an entitlement that has expired or been removed
When   the holder uses a feature it granted
Then   the feature is refused
```

One case per credential type (refresh token, session, key, reset link, invitation) and per entitlement (plan, trial, membership).

### 11. Fail closed — Fail closed

```
Given  the application is started without <required secret>
Then   it refuses to boot

Given  a route with no declared authorization rule
When   an unauthenticated request arrives
Then   it is refused
```

---

## Two checks that are about coverage, not shape

**Enforcement** needs no shape of its own, because every test here already sends its request directly to the server and never through the interface. What it adds is a list: for each rule the interface enforces, there must be a test above proving the server enforces it too. A rule with no such test is unproven.

**Parallel paths** is a matrix. Run **tamper**, **wrong caller** and **cross scope** against *every* path that writes the same column, not only the main one.

| Column | Path | Tamper | Wrong caller | Cross scope |
|---|---|---|---|---|
| organisation name | settings route | ✓ | ✓ | ✓ |
| organisation name | profile route | ? | ? | ? |
| organisation name | bulk import | ? | ? | ? |

Build the table from the audit's write-surface map. Every `?` is either a missing test or an undiscovered finding.

---

## Mapping findings to tests

When proposing tests in the fix plan, give each finding its test in one line, so the developer can see the coverage before anything is written:

```markdown
| Finding | Test shape | Parameters |
|---|---|---|
| C1 — tier accepted at signup | Tamper | POST /accounts · `tier` = paid value |
| C3 — reference applied repeatedly | Replay + Concurrent | verify endpoint · one reference × 10 |
| H2 — negative amount accepted | Out of range | POST /transfers · `amount` = −50 |
| M1 — role loaded by id only | Cross scope | GET/PATCH/DELETE /roles/:id |
```
