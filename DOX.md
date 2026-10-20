# DOX.md — SQL analyzer dependency

## Purpose

Own the focused SQL and catalog corrections used by the Dolt dependency fork.

## Ownership

Root owns the pinned upstream base, adaptation policy and attribution in FORK.md. Upstream remains dolthub/go-mysql-server.

## Local Contracts

- Preserve the Apache-2.0 license, public module path and current upstream MIN/MAX and MOD fixes.
- Projected aliases require unique column IDs independently of sibling-reference visibility.
- Preserve predicates and NULL/cardinality semantics through optimizer rewrites; use full-scan or unsupported-reordering fallbacks when exact index bounds or complete relation properties cannot be proved. Keyed FULLTEXT deduplication must preserve separate physical keyless rows.
- Missing databases during catalog enumeration may be skipped only through typed missing-database errors; propagate unrelated failures and preserve routine visibility.
- MOD exposes operand-appropriate integer/float/decimal metadata; exact remainders retain dividend signedness and zero normalization.
- Keep contributor regressions and query-plan updates with the fix. No Beads-specific configuration, schema or private data.
- A consumer must pin the replacement in its own main module; dependency replacements are not transitive.

## Work Guidance

Keep the patch focused on the reported SQL behavior; qualify the complete consumer dependency tuple before installation.

## Verification

Run view and reported SQL queries in normal and prepared mode, query-plan suites, expression/types, catalog and planbuilder tests. Compare broader failures against the unchanged base; pure-Go regex behavior is distinct from ICU expectations.

## Child DOX Index

