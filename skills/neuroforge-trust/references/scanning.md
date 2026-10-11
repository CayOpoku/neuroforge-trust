# NeuroForge Trust — Scanning Procedure

Read this before auditing a codebase that is too big to read in one sitting — roughly, more than a dozen write paths or more than a handful of source files. For a small project the three habits in `SKILL.md` are enough; skip this file.

The goal is an audit that is deep where the damage would be and honest about where it is shallow. Reading every file is neither: it spends most of the budget on code that cannot hurt anyone. **Efficiency here means choosing what to read, never reading carelessly.**

Everything in the phases below is working method. None of it is written into the report: the report holds findings, gaps and the verdict, and nothing about the system or how you examined it.

---

## Phase 1 — Orient

Cost: a handful of small reads.

Read the dependency manifest, the top-level directory listing, and the entry point or routing file. You are answering four questions: what language and framework, where routes are declared, where the schema or models live, and where authentication is wired in. Do not open feature code yet.

If `neuroforge/audit/trust-audit.md` exists, read it now (see "Re-running", below).

## Phase 2 — Map the write surface by searching

Cost: a few searches, no file reads.

Search for the constructs that declare an entry point in this stack, and for the calls that write to storage. Each hit is a `path:line`. Typical targets, whatever the framework calls them:

- Route and handler declarations for methods that change state.
- RPC and GraphQL mutation definitions, server actions, form handlers.
- Websocket and message handlers, webhook receivers, queue consumers, scheduled jobs.
- Import, bulk, sync and offline-replay endpoints. Admin and internal tools.
- Storage writes: insert, update, delete, save, create, upsert, raw statements.

Two lists that should reconcile: entry points, and storage writes. A storage write that no listed entry point reaches means you missed a path; go and find it.

The search results are your map. Keep it in your working context; do not write it out.

## Phase 3 — Learn what is protected

Cost: one read of the schema or models.

Read the schema once. Note the columns in the Identity, Entitlement, Value and Lifecycle classes, and which **constraints already exist**: unique indexes, foreign keys, check constraints. Several checks (Replay, Concurrency, Scope) are decided by whether a constraint is present.

## Phase 4 — Read the shared machinery once

Cost: a few targeted reads.

Read the code that many routes depend on: the authentication and authorization layer, tenant or scope resolution, the validation setup, rate-limit configuration, secret and config loading, and any base repository or query helper.

This is the best-value reading in the audit. One read answers Permission, Scope, Abuse rate and Fail closed for every route that relies on it. Apply each conclusion to the later paths instead of working it out again. A flaw found here is a finding like any other; a mechanism that works is not written down.

## Phase 5 — Sort the paths into tiers

Cost: none. This is a decision made from the map and the schema.

| Tier | A path belongs here if it… | Depth |
|---|---|---|
| **A** | writes a protected column; issues, checks or revokes a credential; receives a third party's claim; accepts a write with no session; fetches or redirects to a client-named destination | **Traced.** Every one, end to end |
| **B** | is an ordinary create, update or delete inside the caller's own scope | **Sampled.** One per pattern, the rest confirmed by search |
| **C** | only reads | **Searched** for Scope and Response patterns |

Tier A is never sampled. If you are unsure which tier a path belongs to, it is Tier A.

## Phase 6 — Trace Tier A, worst first

Cost: the bulk of the reading, spent where it matters.

Order by blast radius: money and entitlement, then identity and credentials, then scope, then lifecycle. For each path, start at the storage write and walk backwards: the write, the function that calls it, the handler, the request shape. Read those functions and nothing around them. Run the fifteen checks.

## Phase 7 — Sample Tier B and hunt by pattern

Cost: one read per pattern plus searches.

Tier B code is usually repetitive. Read one representative of each pattern in full, then confirm the others share it with a search instead of reading them: for example, search for queries by id on scoped tables and check each hit carries the scope predicate. A hit that breaks the pattern is promoted and read.

Then run the pattern hunts from the "How to find it" notes in `references/checks.md`. These are searches across the whole codebase, each aimed at one check: a request body spread into a write; a secret with a fallback; a read followed by an insert; an outbound request or redirect whose destination is a variable; a raw entity returned to the caller. Read a narrow window around each hit.

## Phase 8 — Parallel paths

Cost: none beyond the map you already have.

Group the write sites by the column they write and compare the rules each enforces. Because the surface is already mapped and the Tier A paths already traced, this needs little or no new reading.

## Phase 9 — Close

Finish the report: the coverage line, the paths not fully examined, the verdict. Confirm the file exists. Then brief the developer in chat (see `references/report-format.md`) and stop.

---

## Reading rules

- **Read a function, not a file.** Use the line number from the search hit and read the surrounding function. Widen only if the function calls something you need.
- **Read each thing once.**
- **Follow calls only as far as the write.** Stop when you reach storage or a third party. Framework internals and library code are not yours to audit.
- **Batch independent searches and reads** instead of doing them one at a time. Every separate step re-sends everything read so far, so ten single reads cost far more than one batch of ten.

**Skip entirely:** dependency and vendor directories, build output, generated code, lockfiles, assets, styles, documentation, and view or template files unless they compute a value that is submitted. Read tests only to learn what rule the authors intended. Read migration history only if there is no current schema file.

---

## Writing the report in batches

Writing the report is itself a cost: each edit is another step. Write in batches, not finding by finding.

- **Write the first batch when Tier A is done**: the header, the Tier A findings, and under "Not fully examined" the names of the paths still to do.
- **Write once more per tier**, adding that tier's findings and trimming the "Not fully examined" list as paths are covered.
- **Write each finding once, in final form.** Do not draft and restate.

An audit interrupted after Tier A therefore still leaves the findings that matter most, and an honest list of what was not reached. What remains under "Not fully examined" at the end is exactly the paths that were sampled or skipped.

---

## When the codebase is too large for one pass

Do not spread the effort evenly and thin. Finish Tier A completely, in blast-radius order, then go as far into Tier B as you can. Leave the rest under "Not fully examined", say so in the coverage line and in the chat briefing, and name what the next run should pick up. A complete audit of the dangerous third is worth more than a shallow audit of everything.

---

## Re-running

1. Re-verify each open finding at its recorded location. This is a few narrow reads.
2. Rebuild the map by search and re-trace Tier A. Start with any path the developer says has changed, and with paths listed under "Not fully examined" last time.
3. Re-read the shared machinery only if it changed, because a change there alters the answer for every route.
4. Update statuses in place, keep the finding ids, and add new findings below.

---

## Spot checks and trust tables

These are small by nature; keep them that way.

- **Spot check:** read the handler in question, the functions it calls down to the write, the relevant part of the schema, and the shared machinery it relies on. Load only the sections of `references/checks.md` for the checks that apply. Do not map the whole surface.
- **Trust table:** read the neighbouring endpoints and the schema for the tables involved, so the table reflects how this codebase already resolves identity and scope. Search for other paths that write the same columns, because that row is the one most often left blank. Save the table and nothing else: no essay around it.
