# DOX.md — SQL analyzer dependency

## Purpose

Own the focused SQL and catalog corrections used by the Dolt dependency fork.

## Ownership

Root owns the pinned upstream base, adaptation policy and attribution in FORK.md. Upstream remains dolthub/go-mysql-server.

## Local Contracts

- Base this dependency on upstream 38a1fb50b434dc88df4a6e6dd3a3b50598dc9446 with Go 1.27.3. Preserve the Apache-2.0 license, public module path and current upstream MIN/MAX and MOD fixes.
- Projected aliases require unique column IDs independently of sibling-reference visibility.
- Preserve predicates and NULL/cardinality semantics through optimizer rewrites; use full-scan or unsupported-reordering fallbacks when exact index bounds or complete relation properties cannot be proved. Keyed FULLTEXT deduplication must preserve separate physical keyless rows.
- Full-Text costed scans require a MATCH predicate bound to that exact index. Other AND predicates remain residual filters; OR branches without a complete lookup use the ordinary fallback. Ordinary indexes cannot consume MATCH predicates.
- Missing databases during catalog enumeration may be skipped only through typed missing-database errors; propagate unrelated failures and preserve routine visibility.
- MOD exposes operand-appropriate integer/float/decimal metadata; exact remainders retain dividend signedness and zero normalization.
- Keep contributor regressions and query-plan updates with the fix. No Beads-specific configuration, schema or private data.
- A consumer must pin the replacement in its own main module; dependency replacements are not transitive.

- Join reordering rejects alternatives that leave a required connecting predicate unapplied. Omit a predicate only when existing functional dependencies prove its equality redundant; preserve non-equality filters and unsupported-plan fallback.

- GREATEST/LEAST choose comparison and return types separately for mixed numeric, text and temporal operands. Preserve exact decimals, unsigned range, NULL propagation and conversion warnings; comparison errors must propagate. Shared decimal precision helpers retain the existing COALESCE correction.

- Historical bind parameters defer only during explicit metadata preparation. Execution rebinds the original AST, resolves the supplied revision and rejects NULL or missing values; current-schema preparation does not promise access to historical-only tables or columns.

- Correlated lateral filters remain inside the non-preserved join input; preserve NULL-extended rows. Session select limits apply to simple SELECT and DISTINCT without overriding explicit LIMIT.
- Error-handler operations follow interpreter branch/loop counters, drain and close results, and propagate failures. Procedure result-set multiplicity remains unchanged.
- SQL assignments to GLOBAL, PERSIST and PERSIST_ONLY require SYSTEM_VARIABLES_ADMIN or SUPER. Check the complete SET before evaluation and recheck current privileges on bound-plan execution; expression rewrites retain authorization. Preserve session/user/row assignments and trusted no-auth engine initialization.

## Work Guidance

Keep the patch focused on the reported SQL behavior; qualify the complete consumer dependency tuple before installation.

## Verification

Run view and reported SQL queries in normal and prepared mode, query-plan suites, expression/types, catalog and planbuilder tests. Compare broader failures against the unchanged base; pure-Go regex behavior is distinct from ICU expectations.

Run TestSetGlobalAuthorization and TestUserPrivileges, including grant/revoke on the same session, copied bound plans, persistence and rejection before partial assignment. Qualify the Dolt SQL server with disposable users and persisted configuration before consumer installation.

## Child DOX Index

