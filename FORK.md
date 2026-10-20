# SQL analyzer corrections for Dolt

This dependency fork starts at dolthub/go-mysql-server `239e0ad6b2f9f32ba9c5752a8b5930e4d682d2cf`.

It adapts angelamayxie's [PR #4044](https://github.com/dolthub/go-mysql-server/pull/4044), head `8de54b7c6a8955680da461ed2bf7fbf25414f201`, for [Dolt #12141](https://github.com/dolthub/dolt/issues/12141). Every projected alias receives a column ID even when it is unavailable to sibling projections. Preserve the contributor's view and query-plan regressions.

The newer base retains the MIN/MAX index-bounds and decimal MOD corrections missing from the proposal branch. Keep the Apache-2.0 license, module path and supported SQL semantics unchanged. Dolt and Beads pin this source explicitly and qualify their dependency graph separately; no Beads data or runtime is stored here.
