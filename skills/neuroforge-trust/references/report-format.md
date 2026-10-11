# NeuroForge Trust — Report Format

Read this when writing a report, a single finding, the chat briefing or the closing verdict.

---

## The one rule: report problems, not the system

A report contains **what is wrong, what was not examined, and the verdict.** Nothing else.

The developer already knows their own system. An inventory of their endpoints, a list of their protected columns, or a description of how their authentication works tells them nothing new, costs tokens to write and to read, and buries the findings. Leave all of it out:

- No table of routes or write paths.
- No section describing protected values, constraints, middleware or "shared machinery".
- No account of what you read or how you went about it.
- No sections beyond the ones in the template below. Do not add "cleanup", "tests" or "background" sections.

What is working gets two or three sentences in the verdict, and only there.

You still build the map of write paths while scanning, because that is how the forgotten path is found. It stays in your working context. It does not go in the report.

---

## The audit report

One file: `neuroforge/audit/trust-audit.md`. Write it without asking; it is this skill's own output. Every report this skill writes lives in `neuroforge/audit/` — a spot check goes to `trust-spot-<topic>.md` and a trust table to `trust-table-<endpoint>.md`.

```markdown
# Trust Audit — <project name>

**Last run:** <date> · **Coverage:** <n> write paths — <n> traced, <n> sampled, <n> not reached

## Findings
<findings, most severe first — or the single line "None on these axes.">

## Not fully examined
<only the paths that were sampled or not reached, one line each with the reason.
Omit this section when every path was traced.>

## Open questions
<only questions whose answer would change a finding or the verdict, one line each.
Omit this section when there are none.>

## Verdict
<the verdict block>
```

The **Coverage** line and the **Not fully examined** section are what keep an efficient audit honest: they say how much was looked at and name the gaps, in a line or two, without listing everything that was fine.

---

## A finding

Keep each finding to this shape and about **80 words**. A finding that needs more is usually two findings, or is explaining the system.

```markdown
### C1 — A paid tier can be set at signup
**Critical · Traced · Check 1 (Origin)**
**Where:** `src/accounts/create.ts:140` — `tier` is read from the request body and stored with no expiry.
**Exploit:** any signed-in user sends `"tier": "business"` and keeps the paid tier permanently.
**Fix:** drop `tier` from the request and always provision the free tier; a paid tier arrives only through the payment path.
**Files:** `src/accounts/create.ts`, `src/accounts/create-account.request.ts`
```

Rules for findings:

- **Number them by severity:** `C1…`, `H1…`, `M1…`. The ids are how the developer refers back to them, including when they say which ones to fix.
- **Name the check by number and name**, using the table in `SKILL.md`. Cite two when both apply.
- **Where:** the file and line, and one clause on what the code does. Do not paste the code block.
- **Exploit:** one sentence — what is sent, what is gained. Describe the request; never write a ready-to-run attack script.
- **Fix:** one or two sentences, the smallest change that makes the server the authority. Prefer a storage constraint to an application check. No code.
- **Files:** what the fix would change. This is the scope "Proceed" approves. Say so when it needs a schema change.
- **One flaw, one finding.** Several endpoints sharing a root cause are one finding with several locations. Two checks failing on the same line are one finding citing both checks.

### Confidence labels

| Label | Meaning |
|---|---|
| **Confirmed** | You ran it, locally or in a test, and observed the result. Say what you ran |
| **Traced** | You read the full path from the incoming value to the write |
| **Conditional** | It depends on code or configuration you cannot see. State the condition |

### Severity

| Level | Threshold |
|---|---|
| **Critical** | Money, entitlement or another party's data can be taken or forged, by anyone with an account or with none |
| **High** | Account takeover, privilege gain inside a scope, or durable damage to another user's data |
| **Medium** | Needs a hard-to-get precondition (a guessed id, an insider), or the damage is limited and recoverable |

Anything below Medium is one line in a "Low-severity notes" list at the end of Findings, at most five lines, or is left out. Out-of-scope observations get the same treatment: one line each, at most three, or nothing.

---

## A clean result

When nothing is wrong, the report is short. Do not fill the space with notes, caveats and essays to look thorough.

```markdown
# Trust Audit — <project name>

**Last run:** <date> · **Coverage:** 4 write paths — 4 traced

## Findings
None on these axes.

## Open questions
- Is `Routes` always mounted behind the session middleware? It is not in this project.

## Verdict
**Findings:** none on these axes
The server is the authority on every write path examined: the owner comes from the session, the
ownership check is in the same statement as each write, and unknown request fields are rejected.
**Residual risks:** the session layer is outside this project and was not examined.
```

A clean report should fit in about 150 words.

---

## The chat briefing

When the report is finished, brief the developer in chat. **Do not paste the report.**

```markdown
## Trust audit — <project name>

**3 Critical · 4 High · 5 Medium.** 42 write paths: 14 traced, 22 sampled, 6 not reached (the reporting jobs).

**Critical**
- **C1** — a paid tier can be set at signup · `src/accounts/create.ts:140`
- **C2** — payments are applied without checking the amount paid · `src/payments/success.ts:429`

**High**
- **H1** — … · `path:line`

**Medium:** M1–M5, in the report.

**Fix plan**
1. C1 — `src/accounts/create.ts`, `src/accounts/create-account.request.ts`
2. C2 — `src/payments/success.ts`, plus a new migration adding a unique index

**Written:** `neuroforge/audit/trust-audit.md`

Say **Proceed** and I will make these fixes, or name the findings you want. I have not changed any
of your code.
```

- One line per finding at Critical and High. Summarise Medium by id unless there are only a few.
- State the coverage, including what was not reached.
- The fix plan names files, because that is the scope "Proceed" approves.
- List what you wrote, including a `.gitignore` line if you added one.
- End on the "Proceed" line and stop. When nothing was found, end on the verdict; there is nothing to proceed with.

---

## Re-running an audit

When `neuroforge/audit/trust-audit.md` exists, read it before the code and re-verify each earlier finding at its recorded location. Update the file in place:

| Status | When |
|---|---|
| **Open** | The flaw is still present |
| **Resolved** | The fix is in the code and you traced it. Note the date |
| **Changed** | The code moved; re-state the finding against the new location |
| **Withdrawn** | The earlier finding was wrong. Say why in one line |

Keep the original finding ids. Collapse each resolved finding to a single line under a "Resolved" heading — the id, the title and the date — so the history stays without the bulk. Never delete the file.

---

## A client-only repository

Most entries are **Conditional**, and each ends with the question for the server's owner instead of a fix.

```markdown
### H1 — The booking form sends its own end time
**High · Conditional · Check 1 (Origin)**
**Where:** `app/pages/bookings/index.vue:212` — computes `endAt` locally and submits it with `startAt`.
**If the server trusts it:** a caller books a one-minute slot to slip past the conflict check, or a multi-day slot to block a calendar.
**Ask the server's owner:** does `POST /bookings` compute the end time itself and ignore the submitted `endAt`?
```

---

## The verdict

```markdown
## Verdict
**Findings:** N Critical · N High · N Medium
[Two or three sentences: where this codebase holds the line, and where it hands decisions to the
client. Specific, not generic.]
**Fix first:** [one to three findings, by id]
**Residual risks:** [what you could not rule out from the code you could see]
```

Be honest in both directions. Do not invent a finding to look thorough, and do not soften a Critical to be kind. If coverage was partial, the Coverage line and the "Not fully examined" section must say so, so that a short report never implies a complete audit.
