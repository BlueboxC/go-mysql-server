# DOX.md — SQL analyzer dependency

## Purpose

Own the minimal SQL analyzer correction used by the Dolt dependency fork.

## Ownership

Root owns the pinned upstream base, adaptation policy and attribution in FORK.md. Upstream remains dolthub/go-mysql-server.

## Local Contracts

- Preserve the Apache-2.0 license, public module path and current upstream MIN/MAX and MOD fixes.
- Projected aliases require unique column IDs independently of sibling-reference visibility.
- Keep contributor regressions and query-plan updates with the fix. No Beads-specific configuration, schema or private data.
- A consumer must pin the replacement in its own main module; dependency replacements are not transitive.

## Work Guidance

Keep the patch focused on the reported SQL behavior; qualify the complete consumer dependency tuple before installation.

## Verification

Run view queries in normal and prepared mode, query-plan suites and planbuilder tests. Compare broader failures against the unchanged base; pure-Go regex behavior is distinct from ICU expectations.

## Child DOX Index

