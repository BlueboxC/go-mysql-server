# SQL analyzer corrections for Dolt

This dependency fork starts at Dolt v2.6.0's analyzer, dolthub/go-mysql-server `38a1fb50b434dc88df4a6e6dd3a3b50598dc9446`, and requires Go 1.27.3. The upstream base now includes PRs #4044 and #4056; retain each correction once, together with the fork's SQL, catalog and historical-parameter protections.

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

## Complete predicates during join reordering

Adapt Nick Tobey's GMS PR #4056 (https://github.com/dolthub/go-mysql-server/pull/4056), merged in 38a1fb50b434, for Dolt PR #12193. Reject joins that drop required filters unless functional dependencies prove an equality redundant. Preserve the contributor ten-table keyed/keyless fixture. IMDB plans are regenerated with the existing plangen for this fork; previous NULL/scalar safeguards remain intact.

## Mixed comparison types

Adapt James Cor's GMS PR #4047, carried by Dolt PR #12197, retaining exact-decimal direction/scale tests and contributor type/query regressions. Comparison type is distinct from result type for temporal and string combinations. The adaptation propagates comparison errors and shares decimal-width helpers with the retained COALESCE correction. Text-length inference remains upstream issue #12181; this does not claim complete MySQL type-metadata parity.

## Prepared historical table bindings

Dolt issue [#12191](https://github.com/dolthub/dolt/issues/12191) is addressed by distinguishing metadata preparation from value binding. A bare `AS OF ?` uses the current table schema for prepare metadata, then resolves the supplied revision afresh on every execution. Execution with NULL, absent or invalid revisions remains an error; it cannot fall back to HEAD. Tables or columns absent from the current schema still require a literal revision for preparation.

## Lateral filters, session limits and error handlers

Adapt the SQL fixes carried by [Dolt PR #12205](https://github.com/dolthub/dolt/pull/12205), preserving the existing decimal and Full-Text safeguards:

- [GMS #4062](https://github.com/dolthub/go-mysql-server/pull/4062), Dem1241: keep correlated filters inside a non-preserved lateral input, retaining the outer NULL-extended rows.
- [GMS #4059](https://github.com/dolthub/go-mysql-server/pull/4059), Dem1241: apply sql_select_limit to simple SELECT and DISTINCT; explicit LIMIT takes precedence.
- [GMS #4058](https://github.com/dolthub/go-mysql-server/pull/4058), James Cor: execute every handler operation. The additional fork correction follows the returned instruction counter for IF/GOTO/WHILE and closes drained iterators. The proposal's counter=-1 and sequential iteration can panic on conditional handlers; retain the independent conditional/loop regression.

The normal/prepared embedded-driver fixtures cover both branch outcomes and complete loop execution. This does not add multiple procedure result sets ([Dolt #12199](https://github.com/dolthub/dolt/issues/12199)).
