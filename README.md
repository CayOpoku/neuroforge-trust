# NeuroForge Trust

**How far can you trust your system?**

Finds every place your server believes something it should have decided for itself.

A system can be trusted only as far as the values it accepts without checking. NeuroForge Trust is a stack-agnostic audit skill that measures exactly that: tampered fields, cross-tenant access by id, illegal status changes, replayed and concurrent requests, unverified payments and webhooks, missing abuse limits, credentials that survive rotation, checks that fail open.

It gets there by asking one question of every write path — *who decided this value, and should they have been allowed to?* — and it works on any language or framework, because the checks are about where a value came from, not about how the code is written.

It answers that one question well and nothing else. It is not a full security scan: XSS, injection and dependency vulnerabilities are left to the tools built for them, so a clean result here means "the server is the authority", not "the system is secure".

**It works on its own.** No other skill, plugin or tool is required — install it into any project and run it. It is part of the NeuroForge family and sits comfortably beside [neuroforge-nest](https://github.com/cayopoku/neuroforge-nest), [neuroforge-nuxt](https://github.com/cayopoku/neuroforge-nuxt) and [neuroforge-qa](https://github.com/cayopoku/neuroforge-qa), but it depends on none of them.

---

## Installation

Install using [skills](https://github.com/vercel-labs/skills):

```bash
# GitHub shorthand
npx skills add cayopoku/neuroforge-trust

# Install globally (available across all projects)
npx skills add cayopoku/neuroforge-trust --global

# Install for specific agents
npx skills add cayopoku/neuroforge-trust -a claude-code -a cursor
```

---

## Three modes

| Mode | Ask for it with | You get |
|---|---|---|
| **Full audit** | `/neuroforge-trust`, "audit this for client trust", "can users cheat this?" | A report at `neuroforge/audit/trust-audit.md`, a short briefing in chat, and a fix plan waiting for your "Proceed" |
| **Spot check** | "can someone change another user's order here?", a diff, one handler | A direct answer with evidence, saved to `neuroforge/audit/trust-spot-<topic>.md` |
| **Plan gate** | "plan the redeem endpoint", any new write endpoint | A trust table at `neuroforge/audit/trust-table-<endpoint>.md`, to approve before code exists |

---

## It writes its reports. It does not touch your code.

The skill reads your project and writes its own reports into `neuroforge/audit/` without asking. Your code is never changed until you say **Proceed**.

| The skill wants to… | What happens |
|---|---|
| Read files in your project | It reads |
| Create or update a report in `neuroforge/audit/`, and add `neuroforge/` to `.gitignore` | It does so as part of the audit, and tells you what it wrote |
| Change your source, tests, schema or config | It presents the findings and a fix plan, then waits for "Proceed" |
| Run a command — tests, scripts, a server, a package manager | It asks and waits for a yes, even after "Proceed" |
| Send a request to a deployed or third-party system | It never does |

Reports list exploitable flaws and where they are, which is why `neuroforge/` is kept out of version control.

---

## Thorough without reading everything

A full audit does not read every file. It builds a map of every path that changes state by searching, then spends its reading where the damage would be:

- Paths that touch money, entitlement, identity or credentials, receive a third party's claim, or accept writes without a session are **traced end to end**, every one.
- Ordinary scoped paths are **sampled** by pattern and confirmed by search.
- The report holds **problems, not a description of your system**: findings, the paths that were not fully examined, and a verdict. No endpoint inventory, no tour of your middleware. A finding is about 80 words; a clean result is about 150 in total.

The report's coverage line says how many write paths were traced, sampled and not reached, and names any that were not fully traced, so a short report is never mistaken for a complete audit.

---

## What it checks

Every value a server stores is classed by where it may come from — identity, entitlement, value, lifecycle state, derived, external fact, or intent — and only intent may be decided by the client. Fifteen checks then run against every path that changes state:

1. **Origin** — does a non-intent value arrive from the request?
2. **Permission** — may this caller perform this action at all?
3. **Scope** — is every record reached by id filtered by the caller's scope?
4. **Bounds** — is every client-chosen value held to a sensible range?
5. **Transition** — is the change legal from the current state, for this caller?
6. **Replay** — does the same request twice take effect once?
7. **Concurrency** — do simultaneous requests leave the invariant intact?
8. **External claim** — is a third party's claim checked against a stored expectation?
9. **Reach** — is the server confined to destinations it should reach?
10. **Parallel paths** — does every route to the same change enforce the same rule?
11. **Enforcement** — is every rule the client enforces also enforced by the server?
12. **Abuse rate** — is guessing, sending and spending capped per actor and per target?
13. **Revocation** — does a revoked credential or expired entitlement stop working?
14. **Fail closed** — does missing config or a failed check refuse?
15. **Response** — does the caller see only what they are entitled to, and only when?

---

## What it finds, and what it leaves to other tools

It finds the places where your server trusts something it shouldn't. They are how people get free plans, forge payments and read other tenants' data, and they map onto most of the [OWASP API Security Top 10](https://owasp.org/API-Security/editions/2023/en/0x11-t10/). They are also the kind of flaw ordinary scanners tend to miss, because the code looks correct.

**What it does not find:**

- cross-site scripting and SQL injection
- vulnerable or outdated dependencies
- leaked passwords and API keys in the code
- weak encryption or TLS settings
- server, cloud and network misconfiguration

It leaves these out on purpose, because other tools already cover them well and adding them would dilute the focus.

### For fuller coverage

Pair it with a dependency scanner, a secret scanner, and a general code security review. Claude Code's built-in `/security-review` covers the injection-style issues this skill skips.

- **In Claude Code there is nothing to install.** Type `/security-review` and it reviews the pending changes on your current branch. It is listed in the [Claude Code command reference](https://code.claude.com/docs/en/commands).
- **To customise it, or to run it on every pull request,** use [anthropics/claude-code-security-review](https://github.com/anthropics/claude-code-security-review). That repository holds the command's source, which you can copy into your project's `.claude/commands/` folder and edit, and a GitHub Action that runs the same review on pull requests.

The two complement each other. `/security-review` looks at a change for a broad range of vulnerabilities; NeuroForge Trust looks across the codebase for one kind. That review also filters out rate-limiting and open-redirect findings by design, and this skill checks both (**Abuse rate** and **Reach**).

---

## Token Budget

| Layer | Size | When it loads |
|---|---|---|
| `SKILL.md` | ~20KB | On skill activation — the classes, the checks, the modes |
| `references/scanning.md` | ~9KB | Before a full audit — the scanning procedure |
| `references/checks.md` | ~23KB | Before a full audit; per-check sections for a spot check |
| `references/report-format.md` | ~10KB | When writing a report, the briefing or the verdict |
| `references/abuse-tests.md` | ~8KB | When proving a finding or writing regression tests |

---

## Structure

- `skills/neuroforge-trust/` — the installable skill. `skills add` copies this directory; nothing at the repo root ships.
  - `SKILL.md` — the entry point
  - `references/` — `scanning.md`, `checks.md`, `report-format.md`, `abuse-tests.md` (loaded on demand)
- `evals/` — test prompts and fixtures for developing the skill. Not shipped.
- `metadata.json` — version, organization, abstract

No build step, no dependencies. Markdown all the way down.

---

## Evals

`evals/evals.json` holds test prompts for checking that a change to the skill changes behaviour. The fixtures under `evals/fixtures/` are small services in different languages and unrelated domains (a game server in JavaScript, a wallet API in Python, a quiz platform in Ruby, a notes API in Go), so a pass means the skill generalises rather than recognising one app.

**The `game-save`, `wallet` and `quiz-app` fixtures are vulnerable on purpose.** They exist to be audited. Do not copy code from them. The `notes-clean` fixture is intentionally sound and checks that the skill reports a clean result instead of inventing findings.

---

## Contributing

1. Keep every check and example domain-neutral. If a sentence only makes sense for one kind of app, it belongs in that app's audit file, not in the skill.
2. A new check must be a question that applies to any write path, with a detection method that does not depend on a framework.
3. When a real audit turns up a flaw none of the fifteen checks would have caught, add the pattern to the right check in `references/checks.md`, and add a fixture that reproduces it.
