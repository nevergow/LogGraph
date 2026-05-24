# Changelog

All notable changes to LogGraph will be documented in this file.

## [0.7.0] — 2026-05-24

### Added

- **`&Name` project syntax** — new `&projectName` prefix replaces `#projectName` for project binding; avoids markdown heading/tag ambiguity
- **Drag-and-drop between projects** — cards can be dragged between project groups in ProjectView; content auto-updates project tag
- **Per-project quick-add button** — each project group has a "+" button to quickly add logs bound to that project
- **Card follow-up button** — expanded cards have a "Follow-up" action that pre-fills SmartInput with project + reference; submitting auto-completes the parent card with a shrink/fade animation
- **`^` reference search** — typing `^` in SmartInput shows a searchable popup of existing blocks; filtered by selected project; supports multi-select
- **Related block visual signal** — blocks with `referenced_by` relations get an amber left-border gradient (30% width only) for subtle visual distinction
- **Timeline sticky date label** — scrolling the timeline shows a sticky YYYY-MM-DD label that updates as cards pass the viewport
- **Unfiled pool** — blocks without a project assignment appear in a dashed-border "Unfiled" section with inbox icon and CTA hint
- **Micro knowledge graph (placeholder)** — sidebar prepared for a lightweight SVG-based connection graph

### Changed

- **Design baseline overhaul** — Electric Blue (`#2563EB`) replaces indigo/violet gradients; Geist + Geist Mono fonts replace Inter; three-layer depth system (Canvas/Surface/Chrome) replaces flat glassmorphism; cards use solid white + subtle borders instead of backdrop-blur
- **Logo** — replaced gradient "LG" square with SVG two-dots-and-line mark symbolizing "your work memory, connected"
- **SmartInput** — placeholder updated to `&project @person ^reference`; `#` autocomplete replaced with `&`; project dropdown prepends `&Name` instead of `#Name`; new `prefillProject`/`prefillContent` props for follow-up flow
- **BlockCard** — footer action bar now includes "Connections" and "Follow-up" buttons; amber border gradient for related blocks; supports `draggable` prop

### Fixed

- **Node rename conflict** — renaming a project/person to an existing name now returns HTTP 409 instead of silently merging logs
- **Header menu blur** — three-dot dropdown now uses solid white background instead of glass-strong with backdrop-blur
- **Stale project selector** — `useNodes` refs moved to module-level singletons; all components share the same reactive state

### Removed

- **Graph slide-over panel** — RightGraphPanel and full-screen Vue Flow overlay replaced with planned micro SVG graph in sidebar
- **`#Name` autocomplete** — frontend no longer suggests `#` for project binding; `&` is the only suggestion trigger
- **Mobile Graph button** — removed from header

## [0.6.0] — 2026-05-22

### Added

- **Priority tags** — `!!` or `!high` syntax in SmartInput automatically sets `priority: high` on blocks, strips control chars from stored content; High-priority blocks render a subtle red badge
- **Desktop hover actions** — mouse hover on cards reveals a mini toolbar with Complete, Archive, and Delete buttons using `group-hover` fade-in
- **Mobile swipe gestures** — swipe left on cards reveals red Delete zone, swipe right reveals green Complete + amber Archive zones; 25px dead zone on left edge preserves system back gesture
- **Toast + undo system** — delete triggers a bottom toast with "已删除 1 条日志 [撤销]" for 3 seconds; clicking undo restores the block before the API call fires
- **Project status counters** — project headers show per-status colored bubbles (Active blue, Done emerald, Blocked red), only when count > 0
- **Project auto-dimming** — projects where all blocks are Done auto-dim to gray text; new Active block instantly reactivates
- **Global filter clear** — filter indicator with Clear button appears in the nav bar when any filter is active
- **Dual-view filtering** — filter state is shared between Project and Timeline views

### Changed

- **BlockCard component** — extracted shared card rendering (markdown, expand/collapse, status badge, toolbar) used by both CenterTimeline and ProjectView
- **Filter state sync** — Timeline dropdowns now bind directly to global filter state via props instead of local refs, eliminating view-switch reset bug
- **Archived blocks** — blocks with `metadata.isArchived = true` are hidden from main views via `visibleBlocks` computed
- **API types** — `create` and `update` methods accept `metadata` parameter for priority/archive storage

## [0.5.1] — 2026-05-22

### Changed

- **Project View** — project names display without `#` prefix; clicking a project header navigates to filtered Timeline View instead of a separate Filter button
- **Timeline axis timestamps** — timestamps moved from card header to the vertical timeline axis next to status dots
- **Timeline project filter** — added project filter dropdown alongside status and date range filters

## [0.5.0] — 2026-05-22

### Added

- **Project View** — new default homepage with blocks grouped by `#project` tag into collapsible accordion cards
- **Segmented control** — Projects | Timeline toggle in the navigation bar for view switching
- **Vertical timeline axis** — 2px gray line with status-colored dots (blue/emerald/red) on the left side of each timeline card
- **Done status de-emphasis** — completed blocks get line-through, gray text, and reduced opacity
- **Touch drag tooltip** — frosted-glass bubble shows the timestamp of the nearest block when dragging on the timeline axis

### Changed

- Default view is now Project View instead of Timeline View
- Left sidebar hidden in Project View (shown only in Timeline View)
- Markdown headings (`##`, `###`) no longer incorrectly parsed as project tags

## [0.4.0] — 2026-05-22

### Added

- **Markdown rendering** — replaced regex-based content rendering with `marked` + `DOMPurify` + `highlight.js` pipeline, supporting GFM (tables, task lists, strikethrough), code syntax highlighting, and XSS sanitization
- **Content folding** — block content exceeding ~200 characters auto-collapses to 300px with an expand button (展开阅读)
- **Progressive input panel** — compact mode (capsule input + Send) and expanded mode (full-height drawer with overlay mask, markdown toolbar, Cmd/Ctrl+Enter to send)
- **Tailwind typography plugin** — prose classes for rendered markdown content in timeline and AI reports
- **Project README** — architecture overview, quick start guide, environment variable reference, and API endpoint documentation
- **CHANGELOG.md** — this file

### Changed

- **Card redesign** — removed 1px borders, added `shadow-sm` with `rounded-xl`, increased padding for cleaner visual hierarchy
- **Status badges** — simplified to colored background pills (emerald/red/blue), removed the colored dot indicator
- **Header polish** — subtler bottom border (`slate-200/60`), icon buttons changed to mid-gray (`slate-400`)
- **Info header** — restructured with status + author on the left, timestamp on the right; timestamps are smaller and lighter
- **AI Panel** — now uses the shared markdown rendering composable

### Security

- XSS sanitization via DOMPurify on all rendered markdown
- External links automatically get `target="_blank" rel="noopener noreferrer"`

## [0.3.0] — 2026-05-21

### Added

- LAN access — Vite dev server binds to `0.0.0.0` for network accessibility
- Mobile responsive layout — three breakpoints (mobile < 768px, tablet 768-1024px, desktop > 1024px) with Teleport overlays for side panels on small screens
- Expandable formatting toolbar — Feishu-style toggle with Bold, Italic, Strikethrough, Heading, Bullet list, Inline code, Blockquote, and quick insert buttons for `#`, `@`, `^`
- Cursor-aware text operations — wrap selection, insert at cursor, insert at line start

### Changed

- Side panels use overlay pattern on mobile/tablet instead of inline layout

## [0.2.0] — 2026-05-21

### Added

- Date display on block cards (`MM-DD HH:mm:ss`)
- Status filter dropdown (All/Active/Completed/Blocked)
- Date range filter (Since/Until)
- Resizable panels (left sidebar and right graph panel) with drag handles
- Project-centric graph panel — shows node graph when a project filter is active
- AI settings panel — configure LLM base URL, API key, and model
- Filtered AI report generation with date range support
- Modern UI redesign with updated color scheme and typography

### Fixed

- Block editing — content now pre-fills correctly in the input area
- Sidebar refresh — project/person lists update after creating new blocks
- Status toggle — cycling through Active → Done → Blocked → Active now works
- Project filter SQL — corrected the query for project-based block filtering

## [0.1.0] — 2026-05-21

### Added

- Go API server with PostgreSQL backend
  - Block CRUD with cursor-based pagination
  - Node extraction and management (#project, @person, ^reference)
  - Relation graph (mentions, blocks, references)
  - Webhook token authentication
  - Lark bot integration
  - Content parsing pipeline
- Vue 3 frontend
  - Three-column layout (left sidebar, center timeline, right graph)
  - SmartInput with `#` `@` `^` autocomplete
  - Vue Flow knowledge graph visualization
  - Status toggle (active/completed/blocked)
  - Filter by project/person
- Python AI microservice
  - LLM-powered project report generation
  - Smart reminder parsing
  - Runtime settings management
