# SQL analyzer corrections for Dolt

This dependency fork starts at dolthub/go-mysql-server `239e0ad6b2f9f32ba9c5752a8b5930e4d682d2cf`.

It adapts angelamayxie's [PR #4044](https://github.com/dolthub/go-mysql-server/pull/4044), head `8de54b7c6a8955680da461ed2bf7fbf25414f201`, for [Dolt #12141](https://github.com/dolthub/dolt/issues/12141). Every projected alias receives a column ID even when it is unavailable to sibling projections. Preserve the contributor's view and query-plan regressions.

The newer base retains the MIN/MAX index-bounds and decimal MOD corrections missing from the proposal branch. Keep the Apache-2.0 license, module path and supported SQL semantics unchanged. Dolt and Beads pin this source explicitly and qualify their dependency graph separately; no Beads data or runtime is stored here.

## Additional SQL and catalog corrections

- [Dolt #12122](https://github.com/dolthub/dolt/issues/12122): adapt the disappearing-database catalog approach from snemarch's proposal. Skip only typed missing-database errors in INFORMATION_SCHEMA enumerations; preserve unrelated errors, routine visibility and the selected session database. Dolt supplies typed failures and cleans up failed database creation.
- [Dolt #12100](https://github.com/dolthub/dolt/issues/12100): retain the existing fixes for cases 02, 05, 14, 16–18 and 20–21. Adapt [GMS #4035](https://github.com/dolthub/go-mysql-server/pull/4035) for MATCH over derived tables/JOIN conditions and [#4036](https://github.com/dolthub/go-mysql-server/pull/4036) for COALESCE numeric precision; retain contributor regressions.
- The remaining case fixes preserve filtered range-join inputs, require every relation referenced by a moved predicate, retain NULL-extended empty-relation identity, reject row-dependent ESCAPE arguments, deduplicate keyed FULLTEXT documents without discarding keyless physical duplicates, and exclude unordered indexes from sort elimination. Mixed numeric/text IN subqueries use SQL comparison; negated scalar subqueries retain empty/NULL and multi-row cardinality behavior. Explicit incompatible collations and lossy index-bound conversions fall back to filtering. Decimal comparisons preserve both scales; exact lookup keys must fit without rounding.
- Empty relations with a nonempty schema use the existing unsupported-reordering fallback. This keeps correct NULL-extension semantics at the cost of absent optimizer estimates for those plans; two plan fixtures reflect that documented fallback. Ordinary indexed string lookups retain their plans.
- [Dolt #12126](https://github.com/dolthub/dolt/issues/12126): MOD returns integer, floating or decimal results according to its operands. Preserve dividend signedness and normalized zero, retain fractions in numeric strings, and size exact remainder arithmetic for exponent gaps. Integer-only fixture metadata is updated alongside typed normal/prepared regressions.

Qualification uses disposable databases, normal/prepared queries, the 23 supplied #12100 repro scripts, additional NULL/collation/index/cardinality controls and the affected analyzer/engine suites. Pure-Go regex tests have separately reproduced ICU-dependent baseline differences; they are not a regex compatibility claim. The existing non-array JSON-path edge cases are not generalized from the supplied JSON-null repro. This fork does not claim full MySQL compatibility or a general-purpose set of optimizer proofs.

Subquery unnesting preserves ordinary negated predicates when another IN subquery becomes a join. Negated scalar subqueries remain scalar, retaining empty-result NULL semantics and multi-row errors. Consumer regression tests include Graph Preview label, lifecycle and ready-work filters.

## Ordinary predicates on Full-Text columns

A Full-Text index is not an ordered lookup for equality, NULL, range or other ordinary filters. Cost it only for a matching MATCH predicate; keep other conjuncts as residual filters and fall back for OR branches that cannot be represented by a complete lookup. Ordinary indexes also leave MATCH residuals intact. Keyed/keyless fixtures, with and without an ordinary same-column index, preserve NULL and physical-duplicate semantics across twenty predicates in normal and prepared mode. This correction also reproduces on the official Dolt 2.5.1 engine/analyzer pair.
