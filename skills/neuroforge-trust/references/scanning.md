# NeuroForge Trust — Scanning Procedure

Read this before a full audit. It describes how to cover a codebase thoroughly without reading all of it, and how to write the report while you work.

The goal is a report that is deep where the damage would be and honest about where it is shallow. Reading every file is neither: it spends most of the budget on code that cannot hurt anyone and leaves little attention for the code that can. **Efficiency here means choosing what to read, never reading carelessly.**

---

## The three habits

1. **Search first, read second.** Build the picture from search hits — a path and a line number — and open a file only when a hit earns it. Then read the function, not the file.
2. **Spend depth where the damage is.** High-consequence paths are traced end to end, every one. Ordinary paths are sampled by pattern and confirmed by search.
3. **Write as you go.** The report is created early and grows as findings are confirmed. It is your working memory, not a summary written at the end.

---

## Phase 1 — Orient

Cost: a handful of small reads.

Read the dependency manifest, the top-level directory listing, and the entry point or routing file. You are answering four questions: what language and framework, where routes are declared, where the schema or models live, and where authentication is wired in. Do not open feature code yet.

If `neuroforge/audit/trust-audit.md` exists, read it now. It tells you what the last audit covered and found, so this run verifies and extends rather than starting over (see "Re-running", below).

## Phase 2 — Map the write surface by searching

Cost: a few searches, no file reads.

Search for the constructs that declare an entry point in this stack, and for the calls that write to storage. Record each hit as `path:line` with the kind of entry point. Typical targets, whatever the framework calls them:

- Route and handler declarations for methods that change state.
- RPC and GraphQL mutation definitions, server actions, form handlers.
- Websocket and message handlers, webhook receivers, queue consumers, scheduled jobs.
- Import, bulk, sync and offline-replay endpoints. Admin and internal tools.
- Storage writes: insert, update, delete, save, create, upsert, raw statements.

Two lists that should reconcile: entry points, and storage writes. A storage write that no listed entry point reaches means you missed a path; go and find it.

**Now create the report.** Write the header, the write-surface table with every path at depth *Not reached*, and nothing else. From here on the file is updated as you learn things.

## Phase 3 — Learn what is protected

Cost: one read of the schema or models.

Read the schema once. Mark the columns in the Identity, Entitlement, Value and Lifecycle classes, and note which **constraints already exist**: unique indexes, foreign keys, check constraints. Several checks (Replay, Concurrency, Scope) are decided by whether a constraint is present, and you will not need to open the schema again.

Add the protected values and the relevant constraints to the report.

## Phase 4 — Read the shared machinery once

Cost: a few targeted reads.

Read the code that many routes depend on: the authentication and authorization layer, tenant or scope resolution, the validation setup, rate-limit configuration, secret and config loading, and any base repository or query helper.

This is the best-value reading in the audit. One read answers Permission, Scope, Abuse rate and Fail closed for every route that relies on it. **Record each conclusion in the report as soon as you reach it** ("authorization defaults to allow when a route declares no rule"), so that it is applied to later paths instead of being worked out again.

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

After each path, update the report: set its depth to *Traced*, and append any finding the moment it is confirmed. Do not hold findings back for a final pass.

## Phase 7 — Sample Tier B and hunt by pattern

Cost: one read per pattern plus searches.

Tier B code is usually repetitive. Read one representative of each pattern in full, then confirm the others share it with a search instead of reading them: for example, search for queries by id on scoped tables and check each hit carries the scope predicate. A hit that breaks the pattern is promoted and read.

Then run the pattern hunts from the "How to find it" notes in `references/checks.md`. These are searches across the whole codebase, each aimed at one check: a request body spread into a write; a secret with a fallback; a read followed by an insert; an outbound request or redirect whose destination is a variable; a raw entity returned to the caller. Read a narrow window around each hit.

Mark sampled paths *Sampled* in the report.

## Phase 8 — Parallel paths

Cost: none beyond the map you already have.

Group the write sites by the column they write, using the map and the tracing notes, and compare the rules each enforces. Because the surface is already mapped and the Tier A paths already traced, this needs little or no new reading.

## Phase 9 — Close

Finish the report: coverage line, open questions, verdict. Confirm the file exists. Then brief the developer in chat (see `references/report-format.md`) and stop.

---

## Reading rules

- **Read a function, not a file.** Use the line number from the search hit and read the surrounding function. Widen only if the function calls something you need.
- **Read each thing once.** If you need a fact again, it should already be in the report.
- **Follow calls only as far as the write.** Stop when you reach storage or a third party. Framework internals and library code are not yours to audit.
- **Batch independent searches and reads** instead of doing them one at a time.

**Skip entirely:** dependency and vendor directories, build output, generated code, lockfiles, assets, styles, documentation, and view or template files unless they compute a value that is submitted. Read tests only to learn what rule the authors intended. Read migration history only if there is no current schema file.

---

## Writing the report as you go

- **Create it at the end of Phase 2**, as a skeleton. An audit interrupted at any later point still leaves a usable file.
- **Append and amend; do not rewrite.** Add a finding or change one row with a targeted edit. Rewriting the whole file each time wastes effort and risks dropping earlier entries.
- **Write findings once, in final form.** Compose the finding when it is confirmed, with its evidence, and do not restate it later.
- **Keep the depth column truthful.** Every path ends as *Traced*, *Sampled* or *Not reached*. This column is what stops an efficient audit from being mistaken for an exhaustive one.

---

## When the codebase is too large for one pass

Do not spread the effort evenly and thin. Finish Tier A completely, in blast-radius order, then go as far into Tier B as you can. Leave everything else at *Not reached*, say so in the coverage line and in the chat briefing, and name what the next run should pick up. A complete audit of the dangerous third is worth more than a shallow audit of everything.

---

## Re-running

An existing report makes the next audit cheap:

1. Re-verify each open finding at its recorded location. This is a few narrow reads.
2. Rebuild the write-surface map by search and compare it with the table in the report. New or moved paths are the work; unchanged *Traced* paths need a re-read only if the files around them changed or the developer says they did.
3. Re-read the shared machinery only if it changed, because a change there alters the answer for every route.
4. Update statuses in place, keep the finding ids, and add new findings below.

---

## Spot checks and trust tables

These are small by nature; keep them that way.

- **Spot check:** read the handler in question, the functions it calls down to the write, the relevant part of the schema, and the shared machinery it relies on. Load only the sections of `references/checks.md` for the checks that apply. Do not map the whole surface.
- **Trust table:** read the neighbouring endpoints and the schema for the tables involved, so the table reflects how this codebase already resolves identity and scope. Search for other paths that write the same columns, because that row is the one most often left blank.
