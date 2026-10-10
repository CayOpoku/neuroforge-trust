# NeuroForge Trust — Report Format

Read this when writing a report, a single finding, the chat briefing or the closing verdict.

---

## The audit report

One file: `neuroforge/audit/trust-audit.md`. Write it without asking; it is this skill's own output. Every report this skill writes lives in `neuroforge/audit/` — a spot check goes to `trust-spot-<topic>.md` and a trust table to `trust-table-<endpoint>.md` in the same folder.

**The file is built during the audit, not after it.** Create it as a skeleton once the write surface is mapped, then add to it with targeted edits as findings are confirmed (see `references/scanning.md`). Findings accumulate across runs; the file is never rewritten from scratch.

The report lists exploitable flaws with their locations, so make sure `.gitignore` covers `neuroforge/` before writing it. After the last edit, confirm the file exists.

```markdown
# Trust Audit — <project name>

**Last run:** <date> · **Coverage:** <n> paths traced, <n> sampled, <n> not reached

## Write surface
| Path | Kind | Writes | Caller | Depth |
|---|---|---|---|---|
| POST /orders | HTTP | orders, stock | signed-in customer | Traced |
| payments webhook | webhook | payments, subscriptions | payment provider | Traced |
| PATCH /profile | HTTP | users | signed-in user | Sampled |
| nightly-expiry | scheduled job | subscriptions | system | Not reached |

## Protected values
Identity: … · Entitlement: … · Value: … · Lifecycle: …
Constraints present: <unique indexes and checks that matter to Replay, Concurrency and Scope>

## Shared machinery
<one line per conclusion that applies to many routes: how authorization defaults, where scope is
resolved, what is rate-limited, how secrets are loaded>

## Findings
<findings, most severe first>

## Open questions
<what could not be determined from this repo, phrased so the owner can answer in a line>

## Out of scope, worth a look
<at most a few one-line notes, or omit the section>

## Verdict
<the verdict block>
```

Three sections carry more weight than they appear to:

- **Write surface**, with its **Depth** column, shows exactly what was examined and how closely. It is what keeps an efficient audit from being mistaken for an exhaustive one, and it is what makes the parallel-paths check possible.
- **Protected values** turns the universal classes into this codebase's actual columns.
- **Shared machinery** records each cross-cutting conclusion once, so that neither this run nor the next has to work it out again.

---

## A finding

```markdown
### C1 — A new account can be created on any paid tier at no cost
**Critical · Traced · Check 1 (Origin)**

**Where:** `src/accounts/create.ts:140` — the handler reads `tier` from the request body and passes
it to provisioning, which stores it with no expiry.

**Exploit:** any signed-in user sends `{ "name": "x", "tier": "business" }`. Tier names are listed
on the public pricing endpoint. They receive the paid tier permanently.

**Fix:** remove `tier` from the create-account request and always provision the free tier. A paid
tier arrives only through the payment path.
**Files the fix changes:** `src/accounts/create.ts`, `src/accounts/create-account.request.ts`
```

Rules for findings:

- **Number them by severity:** `C1…`, `H1…`, `M1…`. The ids are how the developer refers back to them, including when they say which ones to fix.
- **Name the check by number and name** (`Check 4 (Bounds)`), using the table in `SKILL.md`. One finding can cite two checks when both apply.
- **Quote or describe the exact code or flow.** A reader should be able to open the file and see it.
- **Write the exploit as what is sent and what is gained.** Describe the request shape; do not write a ready-to-run attack script against a real host.
- **Give the smallest fix that makes the server the authority.** If a constraint in storage would do it, prefer that over an application check.
- **List the files the fix would change.** This is what "Proceed" approves, so it has to be stated before the developer says it. Say so when a fix needs a schema change.
- **One flaw, one finding.** Three endpoints sharing a root cause are one finding with three locations.
- **Write each finding once, in final form,** when it is confirmed. Do not draft it and restate it later.

### Confidence labels

| Label | Meaning | Say alongside it |
|---|---|---|
| **Confirmed** | You ran it, locally or in a test, and observed the result | What you ran and what you saw |
| **Traced** | You read the full path from the incoming value to the write | Nothing more is needed |
| **Conditional** | It depends on code or configuration you cannot see | The exact condition, and where to look |

When a finding rests on a third party's documented behaviour that you did not test, say so in the finding.

### Severity

| Level | Threshold |
|---|---|
| **Critical** | Money, entitlement or another party's data can be taken or forged, by anyone with an account or with none |
| **High** | Account takeover, privilege gain inside a scope, or durable damage to another user's data |
| **Medium** | Needs a hard-to-get precondition (a guessed id, an insider), or the damage is limited and recoverable |

Anything below Medium goes in a short "Low-severity notes" list, one line each, or is left out.

---

## The chat briefing

When the report is finished, brief the developer in chat. **Do not paste the report.** The detail is in the file; the briefing is what they need in order to decide what happens next.

```markdown
## Trust audit — <project name>

**3 Critical · 4 High · 5 Medium.** 14 paths traced, 22 sampled, 6 not reached (the reporting jobs).

**Critical**
- **C1** — a paid tier can be set at signup · `src/accounts/create.ts:140`
- **C2** — payments are applied without checking the amount paid · `src/payments/success.ts:429`
- **C3** — one payment can be applied several times · `src/payments/success.ts:450`

**High**
- **H1** — … · `path:line`

**Medium:** M1–M5, listed in the report.

**Fix plan** — in this order:
1. C1 — `src/accounts/create.ts`, `src/accounts/create-account.request.ts`
2. C2, C3 — `src/payments/success.ts`, plus a new migration adding a unique index on the reference
3. H1 — …

**Written:** `neuroforge/audit/trust-audit.md` · added `neuroforge/` to `.gitignore`

Say **Proceed** and I will make these fixes, or name the findings you want. I have not changed any
of your code.
```

Rules for the briefing:

- One line per finding at Critical and High. Summarise Medium by id unless there are only a few.
- State the coverage, including what was not reached. A short briefing must not imply a complete audit.
- The fix plan names the files, because that is the scope "Proceed" will approve.
- List what you wrote, including the `.gitignore` line if you added it.
- End on the "Proceed" line and stop. When nothing was found, end on the verdict instead; there is nothing to proceed with.

---

## Re-running an audit

When `neuroforge/audit/trust-audit.md` already exists, read it before the code and verify each earlier finding against the current source. Then update the file in place:

| Status | When |
|---|---|
| **Open** | The flaw is still present |
| **Resolved** | The fix is in the code and you traced it. Note the date |
| **Changed** | The code moved; re-state the finding against the new location |
| **Withdrawn** | The earlier finding was wrong. Say why, and keep the entry so the reasoning is not lost |

Keep the original finding ids. A resolved finding stays in the file under a "Resolved" heading; that history is what tells the next audit which mistakes this codebase tends to make. Update statuses and add findings freely, but never delete the file or discard its history.

Mark findings **Resolved** yourself after a fix you made on "Proceed".

---

## A client-only repository

Most entries are **Conditional**, and each ends with the question for the server's owner instead of a fix.

```markdown
### H1 — The booking form sends its own end time
**High · Conditional · Check 1 (Origin)**

**Where:** `app/pages/bookings/index.vue:212` — computes `endAt` from a locally cached duration and
submits it with `startAt`.

**If the server trusts it:** a caller books a one-minute slot to slip past the conflict check, or a
multi-day slot to block a calendar.

**Ask the server's owner:** does `POST /bookings` compute the end time from the service's stored
duration and ignore the submitted `endAt`?
```

---

## The verdict

Close every report with this block, and lead the chat briefing with its first two lines.

```markdown
## Trust Verdict

**Findings:** N Critical · N High · N Medium
**Coverage:** <n> paths traced, <n> sampled — <what was not reached>

[Two or three sentences: where this codebase holds the line, and where it hands decisions to the
client. Specific, not generic.]

**Fix first:** [the one to three findings that matter most, by id]
**Residual risks:** [what you could not rule out from the code you could see]
```

When nothing was found:

```markdown
## Trust Verdict

**Findings:** none on these axes
**Coverage:** <n> paths traced, <n> sampled — <what was not reached>

The server is the authority on every write path examined: [the specific things it gets right].

**Residual risks:** [what could not be verified from this repository — proxy and gateway
configuration, the behaviour of code outside it, limits that live in infrastructure]
```

Be honest in both directions. Do not invent a finding to look thorough, and do not soften a Critical to be kind. If coverage was partial, say so in the Coverage line rather than letting "none" imply the whole system was checked.
