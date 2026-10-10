---
name: 项目工作台
description: 明亮、克制的本地项目推进工具
colors:
  canvas: "#f0f7f5"
  surface: "#ffffff"
  rail: "#dfeee8"
  ink: "#134e4a"
  muted: "#475569"
  accent: "#134e4a"
  on-accent: "#ffffff"
  line: "#bdcfca"
  danger: "#a02a20"
  focus: "#0d9488"
  accent-hover: "#0f3c38"
  warning-bg: "#fff0e8"
  backdrop: "rgb(19 78 74/.35)"
typography:
  headline:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "28px"
    fontWeight: 700
    lineHeight: 1.3
  headline-mobile:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "24px"
    fontWeight: 700
    lineHeight: 1.3
  title:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "18px"
    fontWeight: 700
    lineHeight: 1.5
  status:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "18px"
    fontWeight: 550
    lineHeight: 1.5
  project-title:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "16px"
    fontWeight: 600
    lineHeight: 1.5
  body:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "16px"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
  support:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.7
  metadata:
    fontFamily: 'system-ui, -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "12px"
    fontWeight: 400
    lineHeight: 1.7
rounded:
  control: "12px"
spacing:
  xs: "2px"
  sm: "4px"
  md: "8px"
  lg: "12px"
  xl: "16px"
  2xl: "24px"
  3xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    rounded: "{rounded.control}"
    padding: "8px 16px"
  button-primary-hover:
    backgroundColor: "{colors.accent-hover}"
  button-secondary:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "8px 16px"
  field:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "10px 12px"
  project-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "16px"
  panel:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.control}"
    padding: "24px"
---

# Design System: 项目工作台

## Overview

**Creative North Star: "明亮项目工作台"**

A restrained office tool: pale green-gray ground, a tinted navigation rail, deep teal text and actions, and familiar rounded controls. Its hierarchy makes project names, delivery dates, and recorded next steps easy to scan. Density comes from aligned stage columns and a small fixed type ramp rather than decorative treatment.

Provenance: this is a post-build scan of `index.html`, `styles.css`, and `app.js`, corroborated by `evidence/baseline-empty.png`, `baseline-desktop.png`, `baseline-mobile.png`, `baseline-styles.json`, and `baseline-results.json`. The selected language and world come from `PRODUCT.md`, `.workbench/spec.md`, `.product/design-proposal.md`, and `.impeccable/surfaces/index.md`; all user context and captured sample records are synthetic, not customer evidence. `.product/finish-review.md` gives the baseline disposition **ship**. No shipping raster assets exist; evidence screenshots are not application assets. The authored geometric SVG brand mark is in the HTML.

**Key Characteristics:**
- Light tonal surfaces with fine structural borders.
- One system Chinese font stack across text and controls.
- Flat, rounded, keyboard-visible working controls.
- Content-driven height and deliberate narrow-screen reflow.

## Colors

The palette uses deep teal for work and selection, with quiet green-gray surfaces and red for exceptional conditions.

### Primary
- **Deep working teal** (`accent`): primary actions; `ink` intentionally uses the same value for text and the SVG brand.
- **Pressed teal** (`accent-hover`): primary hover and active states.
- **Focus teal** (`focus`): visible keyboard outlines.

### Secondary
- **Exception red** (`danger`): overdue wording and storage-error copy; the wording carries meaning alongside color.
- **Warm warning ground** (`warning-bg`): storage warning container only.

### Neutral
- **Office canvas** (`canvas`): page ground and secondary active state.
- **White surface** (`surface`, `on-accent`): controls, cards, panels, dialogs, selected navigation, and primary text.
- **Pale teal rail** (`rail`): navigation ground, secondary hover, and text selection.
- **Slate supporting text** (`muted`): descriptions, labels, placeholders, history, and metadata.
- **Quiet divider** (`line`): one-pixel control borders and structural rules.
- **Teal scrim** (`backdrop`): native modal backdrop.

**The Meaningful Color Rule.** Selection, action, focus, and exceptions receive color; stage identity remains text, without a separate decorative palette.

## Typography

**Display Font:** no distinct display face.
**Body Font:** the fixed system stack in frontmatter, shared by headings, body, buttons, inputs, selects, and textareas.
**Label/Mono Font:** no separate face; derived counts use tabular numerals.

**Character:** familiar Chinese office typography, with modest size and weight changes. The explicit simulated-user system-font choice is a narrow project exception to the generic system-display refusal; it does not authorize a system-font marketing display treatment.

### Hierarchy
- **Headline:** page heading; desktop and mobile roles are in frontmatter.
- **Title:** brand and section/dialog headings; the brand is bold, while the status uses its own medium-weight role.
- **Project title:** semibold card title. List/detail subheadings inherit bold browser heading weight at the body size.
- **Body:** controls and normal text; fields explicitly preserve the body size.
- **Label / Support:** field labels and stage headings use the label size; descriptions, list notes, and empty-stage text use the support leading. Stage heading weight inherits bold heading weight.
- **Metadata:** card dates and latest notes; the rail's local-storage note uses the same size with inherited body leading.

**The Chinese Continuity Rule.** Preserve the exact fallback order and avoid remote font loading. Long project names and notes wrap with `overflow-wrap:anywhere`; do not truncate them into unreadable fragments.

Not canonized: the proposal describes a rem type ramp, but the shipped stylesheet uses px. This document records the actual px values, without treating the proposal wording as implemented behavior.

## Layout

Desktop shell: a fixed rail (208px) and flexible main region. The rail has padding (32px 20px); main has padding (32px), full available width, and a maximum width (1440px). Header and list rows use flex layouts; the search label is at most (280px). The status has vertical padding (16px), a bottom divider, and a following gap (32px).

The board has four equal flexible columns and a gap (16px). At viewport widths at most (1100px), it becomes two columns. At at most (760px), the rail becomes a top block with padding (20px), wrapping horizontal navigation, and a bottom divider; its local-storage note hides. Main padding becomes (20px 16px), the board becomes one column with gaps (24px), search fills the toolbar, list rows and warnings stack, and panels use padding (20px). The mobile heading and status step down to their recorded sizes. Baseline screenshots corroborate the (1440px) desktop and (390px) mobile cases; the intermediate breakpoint is source-evidenced.

Spacing custom properties define the recorded small scale, though the stylesheet mainly uses the matching literal values. Additional observed local measures (20px, 40px) serve rail, list, label, and navigation spacing; they are not invented scale primitives. Panels cap at (760px). Dialogs cap at (500px), keep (16px) viewport margins, limit height to the viewport minus (32px), and scroll internally.

## Elevation & Depth

No box shadows. White surfaces, tinted ground, and one-pixel borders establish containment. Native modal dialogs add a translucent teal backdrop; no fabricated physical material or ornamental depth is used.

**The Flat Working Surface Rule.** Preserve tonal containment and thin borders for ordinary controls, cards, and panels.

## Shapes

Controls, cards, list containers, panels, warnings, and dialogs share gently curved corners through the single radius token. List containers clip their internal row dividers. Stage headers use straight bottom rules. The brand is an inline stroked geometric SVG, without raster imagery or glyph icons.

## Components

### Buttons
Familiar working controls: minimum height (44px), one-pixel border, inherited font, and the recorded padding/radius. Primary hover and active share pressed teal; secondary hover uses the rail and active uses canvas. Buttons transition background color (150ms), using default CSS easing. Disabled buttons use opacity (.5) and a not-allowed cursor. Keyboard focus is an outline (3px) with offset (3px); reduced-motion preference removes the button transition.

### Inputs / Fields
White, bordered, full-width fields with minimum height (44px). Labels sit above fields with a gap (8px); dialog labels have top margins (20px). Placeholder text uses supporting slate; caret uses the primary accent. Search, date, select, and textarea keep native behavior. Focus shares the keyboard outline. Required project name/date use browser validation, including the custom Chinese missing-name message; there is no custom invalid-border token. Native disabled input rendering remains browser-defined.

### Navigation
Transparent left-aligned buttons at rest; selected state is white with a visible border and font weight (650), driven by `aria-pressed`. Desktop navigation uses a gap (8px) and starts after (40px); mobile starts after (20px), wraps horizontally, and uses smaller labels and padding (8px 12px).

### Cards / Containers
Project cards are full-width button surfaces with padding (16px), a following gap (12px), semibold names, small dates, and the latest actual note. Overdue dates include “已超期” and exception red. Lists use one bordered rounded container with row padding (16px 20px), dividers, and a trailing project action. Panels use the recorded desktop padding. Empty stages stay plain supporting copy; empty lists hide rather than rendering ghost cards.

### Dialogs and Feedback
Native modal dialogs contain labeled fields, right-aligned action rows with gaps (12px), and progress history. Storage errors appear as a warm warning block with exception-colored text and a retry action; corrupt storage suppresses retry and disables new-project/owner input. Save status and derived status use live status regions. No inline success toast, custom validation animation, or loading skeleton is shipped.

### Stage Comparison
A three-column row combines stage text, native meter, and right-aligned tabular count; columns are (70px), flexible, and (32px), with gaps (16px), becoming (12px) on mobile. Meter height is (16px), and its accent is deep teal. Native meter and date/select details may vary by browser; the paired labels and numbers remain the stable design.

## Do's and Don'ts

### Do:
- **Do** reuse the recorded Chinese font stack, semantic palette, shared radius, and working-control states.
- **Do** pair exception color with meaningful text and retain visible keyboard outlines.
- **Do** allow long Chinese names and notes to wrap and preserve the one-column mobile sequence.
- **Do** use actual project data or honest empty states, with synthetic preview content identified as such.

### Don't:
- **Don't** add a marketing hero, invented statistic cards, or decorative status kickers to this working-tool world.
- **Don't** introduce a separate font, gradient lettering, glyph icons, or ornamental shadows as inherited system rules.
- **Don't** treat evidence screenshots or synthesized sidecar color ramps as shipping assets or new normative palette tokens.
