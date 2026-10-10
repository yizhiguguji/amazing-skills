disposition: ship

Evidence history / constraints: The first follow-up verdict was recapture because revision-desktop-filter.png repeated the page-top brand/status at its bottom. Both authoritative screenshot paths were reopened after refresh; the desktop duplication is gone, and both captures are valid. The supplied revision-results.json timestamp actually reads 2026-10-10T06:25:40.856Z, not after 06:28 UTC as described in the handoff; its five passing outcomes are reported harness evidence, not independently executed tests. The reviewer TOML was manually loaded earlier; no native auto-loading is claimed. This is a scoped review of the requested board-only date filter, with no comp or QUALITY BAR card.

## persistence
pass — PRODUCT.md and the existing surface contract remain applicable. DESIGN.md now exists and records the actual teal palette, font stack, shared field controls, responsive layout, and baseline review provenance. The addition reuses those controls and does not need a new visual direction or comp round.

## fidelity
| Scoped element | State | Evidence |
| --- | --- | --- |
| TYPE | match | The new label/select use the existing 14px label and 16px field type; Chinese labels remain readable on both widths. |
| MATERIAL | match | Standard flat native select matches the existing field system, with no new assets or fake physicality. |
| GROUND | match | Canvas, rail, action, field surface, and divider colors visibly retain the baseline world and agree with DESIGN.md/baseline-styles.json. |
| Date control | match | Authorized by the follow-up in .product/input.md; placed beside search on desktop and below it on mobile, using the established field treatment. |
| Board-only scope | match | app.js hides due-label outside board and applies dateFilter only to boardProjects. Project list, activity, review counts, and global hook derive from their prior sources. |
| Filter semantics | match | All, unfinished overdue, today, and future predicates compare against the existing local-calendar day; completed overdue projects are excluded from overdue results. Stage counts derive from filtered boardProjects. |
| Search / no-result behavior | match | Date predicates follow name search. A zero-match filtered board exposes an explicit date-range recovery message. |
| Responsive / interaction | match | Refreshed desktop and mobile captures show aligned existing controls and no clipping. The select is natively labeled and inherits keyboard-visible focus. revision-results.json reports AC-01 through AC-05 passing with no errors. |

## ceiling
reached for the scoped addition — the filter follows familiar tool placement and existing spacing/radius/type rather than creating a competing visual treatment. The extra mobile row preserves the original readable stage sequence. No visual regression is visible against the baseline captures.

## material_fixes
none

## keep
Keep the board-only filtering boundary, global truthful counts, unfinished-only overdue predicate, existing design tokens/control styling, and mobile field stacking.
