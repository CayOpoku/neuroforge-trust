# NeuroForge Trust — The Fifteen Checks

Detection patterns, search strategy and fix shape for each check. The examples deliberately span unrelated domains: the pattern is what transfers, not the nouns.

**Contents:** [1 Origin](#1-origin) · [2 Permission](#2-permission) · [3 Scope](#3-scope) · [4 Bounds](#4-bounds) · [5 Transition](#5-transition) · [6 Replay](#6-replay) · [7 Concurrency](#7-concurrency) · [8 External claim](#8-external-claim) · [9 Reach](#9-reach) · [10 Parallel paths](#10-parallel-paths) · [11 Enforcement](#11-enforcement) · [12 Abuse rate](#12-abuse-rate) · [13 Revocation](#13-revocation) · [14 Fail closed](#14-fail-closed) · [15 Response](#15-response)

The general method for all of them: **start at the write and walk backwards.** Find where state is persisted, list the values stored, and follow each to its source. Walking forwards from the request tempts you to trust a value because a validator touched it.

---

## 1. Origin

**Question:** does any value outside *Intent* reach the write from the request?

**Looks like**

- A handler passes the request body, or a spread of it, straight into a create or update.
- The request schema declares a field whose name is a protected column: `price`, `total`, `role`, `plan`, `status`, `ownerId`, `tenantId`, `createdAt`, `endsAt`, `score`, `balance`.
- A header, cookie, subdomain or path segment selects the tenant or user, with no membership check behind it.
- Headers that a gateway or proxy is meant to set (forwarded address, host, a user id) are trusted, while the application can be reached without that gateway or the gateway does not strip them.
- A hidden form field carries an amount. A client-generated id or timestamp is stored as received. The client's clock decides when something happened.
- A token or metadata blob that the client assembled is treated as trusted because it came back through a third party.
- A save or sync endpoint accepts a whole object and persists it.

**How to find it.** Search request schemas, form models and serializers for the names of the protected columns you marked in the schema. Search for body spreads and bulk assignment into the storage layer. Check what an allow-list actually allows: stripping unknown fields does nothing about a known field that should never have been there.

**Fix shape.** Remove the field from the request shape entirely, then derive it: identity from the session, values from the server's own records, derived fields by computation. Where the client legitimately chooses (which plan to buy, which slot to book), accept the choice as an identifier and resolve everything that follows from it (price, duration, eligibility) on the server, through the path that enforces the requirement.

**Same flaw, different domains:** a game accepts the player's coin total in a save file · a storefront accepts the cart total · a signup form accepts the subscription tier · a booking form accepts the appointment's end time.

---

## 2. Permission

**Question:** is this caller allowed to perform this action at all, checked on the server for this route?

**Looks like**

- An administrative, staff or internal route that requires a session and nothing more.
- The permission exists only in the interface: the menu item is hidden, but the handler accepts anyone.
- The role is read from the request, or from a token claim the client can alter, instead of from the server's record.
- Routes protected by being unlinked: `/internal/`, debug, export and maintenance endpoints.
- The main action is checked but a sibling is not: the single-item route has the check, the batch variant does not; reading is guarded, writing is not.
- A route that should need a session and has none.

**How to find it.** List every route with the caller it is meant for. Any route with no stated requirement is an open question, not an assumption. Then compare against what the interface shows to which roles: a control only some users see implies a rule the server must hold.

**Fix shape.** Every route declares who may call it; the default is deny. The permission is resolved on the server from the session's stored role and permissions, never from the request.

**Same flaw, different domains:** any member calls the owner-only "cancel subscription" endpoint · a student calls the grade-edit endpoint the teacher's screen uses · a visitor with no account downloads a data export.

---

## 3. Scope

**Question:** is every record reached by a client-supplied id filtered by the caller's scope in the same query?

**Looks like**

- Load, update or delete by primary key alone.
- Load by id, then compare the owner afterwards, with a code path that skips the comparison.
- A parent is checked but the child id is not tied to that parent (`/projects/:p/tasks/:t` verifies the project, then loads the task by id).
- A reference in the body points at another scope's record: attach this role, this address, this payment method.
- Bulk endpoints that take an array of ids.
- Tenant resolved from a header or subdomain with no check that the caller belongs to it.
- Records that should be out of reach by state, not by owner: soft-deleted, archived or suspended rows that an id still retrieves.
- A cache key without the scope in it.

**How to find it.** List every query by id on a scoped table and check whether the scope predicate is in the same query. Then list every foreign key accepted from a request and check that the referenced record is in scope too. The second list is the one people forget.

**Fix shape.** Put the scope predicate in the query itself, in one data-layer place that every read and write goes through. Return "not found" rather than "forbidden", because the latter confirms the record exists. Validate referenced ids against the same scope.

**Same flaw, different domains:** editing another clinic's role by its id · transferring from an account number the caller does not own · attaching another shop's coupon to an order.

---

## 4. Bounds

**Question:** is every client-chosen value held to the range that makes sense for this caller?

The client is allowed to choose *Intent* values, which is exactly why they need limits. "The client may set it" never means any value is acceptable.

**Looks like**

- A negative quantity or amount, which reverses the direction of a transfer or reduces a total.
- Zero, where zero makes an operation free. A refund larger than the payment.
- Very large numbers, overflow, or a fraction where a whole number is required.
- Money held in floating point, or rounding that can be exploited by repeating a small operation.
- No ceiling on page size, array length, string length, file size or nesting depth.
- A choice that is well-formed but not on offer to this caller: a retired or internal plan id, a time slot outside availability, a discount meant for another group.
- Dates in the past or the far future.
- A file's declared type, size or name taken on trust.

**How to find it.** For every value the client is permitted to choose, state the allowed set in one sentence and then look for where the server enforces it. For numbers: sign, minimum, maximum, step. For choices: resolved against what the server currently offers *this* caller, not merely against the set of ids that exist.

**Fix shape.** Validate on the server against server-owned ranges and lists. Hold money in whole minor units or a decimal type. Cap every size.

**Same flaw, different domains:** a negative transfer pulls money from the recipient · a quantity of −1 lowers a cart total · a booking ten years ahead blocks a calendar · a page size of one million takes the service down.

---

## 5. Transition

**Question:** is this change legal from the record's current state, for this caller?

**Looks like**

- A generic status setter that accepts any member of the enum.
- A create endpoint that accepts the initial status.
- Flags such as `isPaid`, `verified`, `approved`, `published` inside a general update shape.
- A step that can be reached without the step before it: confirm without pay, publish without review, ship a cancelled order.
- The current state is read, checked in application code, and written in a separate statement.

**How to find it.** For each lifecycle column, list every write to it and the precondition each enforces. Draw the transitions the code actually permits and compare them with the ones the business intends.

**Fix shape.** The server sets the initial state. Each transition is either a named operation or a row in a transition table of (from, to, who may). The precondition is checked against stored state inside the same atomic write, for example an update conditional on the current status, so that it cannot be raced.

**Same flaw, different domains:** a customer creates a booking as already approved · a seller marks their own listing verified · a refund is issued on an order that was never paid.

---

## 6. Replay

**Question:** does the same request sent twice take effect once, enforced by a uniqueness constraint rather than a read-then-write?

**Looks like**

- "Have we seen this reference?" as a read, followed by an insert.
- No unique constraint on an external reference, event id or idempotency key.
- Redeem, claim, apply, confirm and verify endpoints.
- Two handlers applying the same event, typically a webhook and a return page.
- A retryable job with a side effect and no record of having run.
- A concurrency token (`updatedAt`, version) that the client can read and send back, used as if it proved freshness.

**How to find it.** For every operation that grants something, ask what stops it running twice and look for that thing in the schema, not in the handler. If the answer lives only in application code, it is a race.

**Fix shape.** A uniqueness constraint in storage. Insert first and treat a conflict as "already done"; apply the effect in the same transaction as the insert; return the original outcome on a repeat.

**Same flaw, different domains:** one payment reference verified ten times at once extends a subscription ten terms · one coupon code credited twice · a level-complete reward claimed on every retry.

---

## 7. Concurrency

**Question:** do N simultaneous requests leave the invariant intact?

**Looks like**

- Read a balance, compute, write it back.
- Count rows, compare with a limit, then insert: a booking cap, one vote per user, the last seat.
- A consumable (a credit, a first-purchase discount) granted to every open session, because it is only deducted on completion.
- A check on one row and a write on another, with no lock tying them together.
- A multi-step write outside a transaction.

**How to find it.** State the invariant in one sentence ("balance never goes negative", "at most one redemption per user") and ask what, in the storage engine, prevents two requests violating it together. Sequential tests cannot show this bug.

**Fix shape.** An atomic conditional update (decrement only where the balance is sufficient), a uniqueness constraint, or a row lock inside a transaction. Reserve scarce things when the flow opens, not when it completes, and release them on expiry.

**Same flaw, different domains:** two withdrawals each see the full balance · ten checkouts each carry the same one-time discount · two buyers both get the last item.

---

## 8. External claim

**Question:** is a third party's claim checked against what the server stored beforehand, not against data echoed back?

**Looks like**

- A payment-success handler that reads the plan, the user or the quantity from the provider's metadata.
- No comparison of the amount and currency actually paid with what was expected.
- A webhook with no signature check, a signature computed over the parsed body instead of the raw bytes, a non-constant-time comparison, or a secret that defaults to empty.
- A redirect parameter such as `?status=success` treated as proof.
- A signed token whose signature, algorithm, issuer, audience or expiry is not verified, or whose claims (email verified, group membership) are believed without that verification.
- A flag such as "scanned" or "approved" arriving from the client about work a third party was meant to do.

**How to find it.** Find every place the system hands off to a third party and every place it hears back. For each return path, list which fields it believes and where each one originated. Anything the client could have set before the handoff is client input, however many servers it passed through.

**Fix shape.** Record the expectation before the handoff: who, what, how much, under which reference. On return, load that record by reference, confirm with the third party server-to-server, compare against the expectation, and ignore anything echoed back. Verify signatures over the raw bytes with a constant-time comparison, and refuse to boot without the secret.

**Same flaw, different domains:** a tiny payment carrying forged metadata buys a large credit pack · a forged "delivered" webhook releases escrow · an unverified sign-in token claims an administrator's email.

---

## 9. Reach

**Question:** when the client names a destination, is the server confined to destinations it should reach?

Here the client decides not a stored value but *where the server goes*: what it fetches, which file it opens, where it sends someone, whom it writes to.

**Looks like**

- The server fetches a URL the client supplied: an avatar by link, a link preview, "import from URL", a configurable webhook target, a document renderer. The address can point at internal services or a cloud metadata endpoint.
- A client-supplied file name or path is joined onto a directory, allowing a read or overwrite outside it.
- A redirect target after sign-in or payment is taken from a query parameter.
- The recipient or sender of a system-sent email or text message is chosen by the client, turning the feature into a relay.
- A template name, storage bucket, queue or table name comes from the request.
- A callback address is accepted at the start of a payment or sign-in flow without being checked against a registered list.

**How to find it.** Search for outbound requests, file operations, redirects and message sends. For each, trace the destination back to its source. Any destination with a client in its ancestry needs a restriction.

**Fix shape.** An allow-list of destinations. For URLs, check the resolved address rather than the name, block private and loopback ranges, and re-check after every redirect. Use server-generated file names and ids instead of client paths. Keep redirect targets relative, or pick them from a registered list. Derive recipients from the server's own records.

**Same flaw, different domains:** "import from URL" reads the host's cloud credentials · an upload named with `../` overwrites a config file · a sign-in link forwards the user to an attacker's page · "share by email" sends anyone's text to any address from your domain.

---

## 10. Parallel paths

**Question:** does every route to the same change enforce the same rule?

**Looks like**

- Two endpoints write the same column with different checks: the settings route requires a permission, and the profile route changes the same thing without one.
- Create validates what update does not, or the reverse.
- Sibling operations where only one carries a check: inviting with a role versus changing a member's role.
- Side doors: bulk import, CSV upload, an older API version, a GraphQL mutation beside the REST route, a mobile-only endpoint, an admin tool, an internal job, an offline-sync replay endpoint.

**How to find it.** Group every write site by the column it writes, then compare the checks side by side. This is the only reliable method, and it requires the whole write surface to be mapped first. Run it last.

**Fix shape.** The rule lives with the write, in one function or domain method that every path calls. A rule that lives in a handler is missing from the next handler.

**Same flaw, different domains:** a receptionist renames the business through the profile form · a bulk import sets prices the single-item form would reject · the v1 endpoint still accepts the field v2 removed.

---

## 11. Enforcement

**Question:** is every rule the client enforces also enforced by the server, at the moment of the write?

Check 1 asks about values; this asks about *rules*. A client can only suggest a rule. Anyone who sends the request by hand has skipped it.

**Looks like**

- A disabled button, a hidden menu or a client-side route guard as the only gate.
- Form validation — required, maximum length, format — that exists only in the interface.
- A paywall or feature gate that hides content the client has already received.
- Cooldowns, timers, trial expiry or attempt counters kept in the client, or computed from the client's clock.
- A multi-step flow where the current step, or a flag such as "code verified", is held by the client.
- Game rules — movement speed, hit detection, cooldowns — evaluated only on the player's device.
- An offline queue whose entries are accepted at sync time without re-running the rules that applied when they were made.

**How to find it.** In a client repository, list every rule the interface enforces and ask where its counterpart is on the server. In a server repository, take the rules the product evidently has (from the interface, the tests, the documentation, the error messages) and locate each one in a handler. A rule you cannot find on the server is a finding or a question.

**Fix shape.** The server re-checks every rule at the time of the write, using its own clock and its own records. The client is sent only what it is entitled to hold. Client-side checks stay as convenience for honest users, and are never the gate.

**Same flaw, different domains:** premium articles hidden by a stylesheet · a quiz that ships the answers and marks itself · a trial that never ends for anyone who changes their device clock · the third step of a checkout callable without the second.

---

## 12. Abuse rate

**Question:** is anything that guesses a secret, sends a message, costs money or consumes a quota capped per actor *and* per target?

**Looks like**

- Login, one-time code, password reset and verification endpoints protected only by a global limit.
- One secret checkable through two endpoints with different limits; the attacker uses the looser one.
- Unlimited fresh challenges, each with its own allowance of guesses.
- Endpoints that send email or text messages, or call a metered service, with no cap.
- An unauthenticated write that consumes a tenant's quota or fills someone's calendar.
- Limits keyed only on the address, which the attacker rotates, or only on the account, which lets them lock the victim out.

**How to find it.** List every endpoint that checks a secret, triggers a send, or spends a limited resource. For each, **do the arithmetic**: size of the secret space divided by attempts allowed per window, across every endpoint that can check it. If the answer is days, it is a finding.

**Fix shape.** Limit per actor and per target (account, phone number, tenant). Count wrong guesses against the secret itself and invalidate it after a few. Allow one open challenge per target. Put a hard cap on spend.

**Same flaw, different domains:** a six-digit reset code guessable in a day through the endpoint nobody throttled · a contact form that drains the day's email allowance · a public form that uses up a free account's monthly quota.

---

## 13. Revocation

**Question:** after rotate, revoke, logout or expiry, does the old credential or entitlement actually stop working?

**Looks like**

- A refresh token that still works after it has been rotated.
- A stored hash from a function that truncates its input, so that different long tokens produce the same hash. Password-hashing functions with an input limit are the usual cause.
- Logout that only clears the client's copy.
- A password change that leaves existing sessions alive.
- A removed role or permission that persists in a long-lived token.
- Credentials of a deleted or suspended account that still authenticate.
- Invitation, reset and magic links that can be used more than once, or signed links with no expiry.
- An entitlement checked only when it was granted: an expired plan, a lapsed trial or a removed member that the system keeps serving because nothing looks again.

**How to find it.** For each credential and each entitlement, find where it is issued, where it is checked and where it is meant to end. Then ask what, on the server, differs between the moment before revocation and the moment after. If nothing differs, it was not revoked.

**Fix shape.** Keep a server-side record per credential, compare the full value (hash the entire token with a fast cryptographic hash; high-entropy tokens do not need a password hash), and delete or mark it on use. For claims carried in tokens, check a version counter on each request. Check entitlements at the time of use, not only at sign-in.

**Same flaw, different domains:** a stolen token outlives every refresh · a dismissed employee's key keeps working · a reset link is reused a week later · a cancelled subscription keeps its paid features.

---

## 14. Fail closed

**Question:** does a missing secret, config value, permission or failed check refuse, rather than allow?

**Looks like**

- A secret with a fallback: an empty string, a development default, a literal in the code.
- A guard that allows the request when no rule was declared for the route.
- A `catch` around authentication or verification that carries on.
- A lookup failure (flag service, permission store) interpreted as "enabled" or "allowed".
- An unknown role or status that falls through to the permissive branch.
- Verification that is skipped under a debug or environment flag.

**How to find it.** Read every authorization and verification function for its failure branches: what happens when the thing it needs is absent, malformed or throws? Search for defaults on secret-bearing config.

**Fix shape.** Required secrets are validated at startup and the process refuses to boot without them. Authorization defaults to deny. An error inside a check is a denial, not a pass.

**Same flaw, different domains:** a webhook secret that defaults to empty accepts any signature · a route someone forgot to annotate is public · an outage in the permission store lets everyone in.

---

## 15. Response

**Question:** does the response contain only what this caller is entitled to see, and only once they are entitled to see it?

**Looks like**

- A raw third-party response or a whole stored entity returned to the caller.
- Fields that were never meant to leave: hashes, tokens, internal notes, other people's contact details, payment details.
- A lookup by arbitrary reference that returns details to whoever asks.
- Different errors for "does not exist" and "exists but is not yours", which lets a caller enumerate.
- A scoped list whose joined or included records ignore the scope.
- Data sent ahead of entitlement: answers delivered with the questions, another player's hidden state, unpublished or paid content, a full dataset that the client is expected to filter.

**How to find it.** For each response, list its fields and ask which ones this caller is entitled to right now. Follow includes and joins; the top-level record is often scoped correctly while its relations are not.

**Fix shape.** An explicit response shape that names its fields. Lookups happen within the caller's scope, existence is not revealed across scopes, and anything the caller has not yet earned stays on the server.

**Same flaw, different domains:** a payment-status endpoint returns the payer's card details to any reference holder · a member list includes every member's reset token · a card game sends each player the whole deck.
