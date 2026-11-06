# AGENTS.md — GraphQL-PKM UI Guidelines

## 1. Project Context

This project is a Personal Knowledge Management (PKM) note-taking application backed by an existing Go 1.25 GraphQL server.

The backend already provides:
- Note CRUD operations
- Tags and note categorization
- Bidirectional note links and backlinks
- Keyword search
- Semantic search
- Hybrid search
- AI-powered smart search
- AI-generated summaries and key concepts
- Relationship checking between notes

Your responsibility is to build a polished, functional frontend that consumes the existing GraphQL API.

IMPORTANT:
- This is a note-taking application, not an administrative dashboard.
- Do not redesign or replace the existing backend.
- Do not invent backend capabilities.
- Inspect the existing codebase before making implementation decisions.
- Reuse existing project conventions and dependencies whenever practical.
- The backend's local-first migration is tracked in `LOCAL_FIRST_MIGRATION.md`. Do not work on it unless asked.

## 2. Design Direction

Design a premium, minimalist, editorial-style note-taking application.

The visual inspiration is a combination of:
- Craft: refined writing experience.
- Notion: practical information organization.
- Obsidian: interconnected knowledge.
- Editorial paper journals: warm, calm, distraction-free aesthetics.

The application should feel like a personal digital library rather than a corporate SaaS dashboard.

### Core Design Principles

1. Content comes first.
2. Writing must feel effortless.
3. Navigation must be predictable.
4. AI features should feel integrated, not intrusive.
5. Whitespace is intentional.
6. Every component should have a clear purpose.
7. Avoid unnecessary decoration and visual noise.

## 3. Strict Color System

ONLY use brown and white as the design color families.

Do not introduce blue, purple, green, red, or other accent colors.

### Color Tokens

Use these CSS variables consistently:

```css
:root {
  --background: #FAF9F6;
  --surface: #FFFFFF;
  --surface-secondary: #F3EFE8;

  --foreground: #33271F;
  --foreground-secondary: #78685B;
  --foreground-muted: #A99A8D;

  --primary: #79563B;
  --primary-hover: #62432E;
  --primary-foreground: #FFFFFF;

  --border: #E8E0D6;
  --border-strong: #D6C9BA;

  --accent: #B08A68;
  --accent-soft: #EDE2D6;

  --selection: #E8D9C8;
}
```

### Color Rules

- Main background: warm off-white.
- Main content surfaces: pure white.
- Primary text: dark espresso brown.
- Secondary text: muted brown.
- Primary buttons: medium brown with white text.
- Borders: subtle warm beige.
- Hover states: slightly darker brown or warmer background.
- Selected states: light brown background.
- Focus states: brown outline.

Never use black as the primary text color.
Never use gradients with unrelated colors.
Never introduce colorful category badges.
Never use colorful AI indicators.

Contrast rules:
- `--foreground-muted` and `--accent` have low contrast on light backgrounds. Use them only for decorative or non-essential elements (icons, dividers, placeholder hints), never for text the user must read.
- Errors and warnings cannot use red. Communicate them with an icon, clear text, and a strong brown border, never color alone.

## 4. Typography

Use a refined typography system.

Recommended fonts:
- Interface: Inter or Geist Sans.
- Editorial headings: Lora or Georgia.
- Note editor: Inter or a readable system sans-serif.

Self-host fonts (npm package or local files). Do not load fonts from a CDN.

Typography hierarchy:

- Page heading: 28–32px.
- Section heading: 18–22px.
- Note title: 32–40px.
- Body text: 15–17px.
- Metadata: 12–13px.

Use generous line height for long-form writing.

Avoid oversized typography and excessive font-weight variations.

## 5. Application Layout

Use a three-part workspace layout.

### Sidebar

Persistent left navigation.

Elements:
- App logo and name.
- New Note button.
- All Notes.
- Favorites (only if supported by the backend or local implementation).
- Tags.
- Knowledge Graph.
- AI Search.
- Settings.

Sidebar requirements:
- Width around 240px.
- Warm off-white background.
- Subtle right border.
- Compact navigation items.
- Clear active state using a light brown surface.
- Collapsible on smaller screens.

Do not overcrowd the sidebar.

### Main Workspace

The central area displays the current page.

Depending on the selected section:
- Note list.
- Note editor.
- Search results.
- Knowledge graph.
- Tag-filtered notes.

Maintain consistent horizontal spacing and alignment.

### Context Panel

Optional right-side panel for:
- Note backlinks.
- Linked notes.
- AI-generated summary.
- Key concepts.
- Related notes.

Width around 280px.

Allow the panel to collapse to maximize writing space.

## 6. Core Screens

### A. All Notes

Create a clean note library.

Include:
- Page title.
- Search input.
- New note action.
- Recent notes.
- Note list or grid.
- Tag filters.
- Sort options.

Each note preview should display:
- Note title.
- Short content excerpt.
- Updated timestamp.
- Tags.
- Optional AI-generated summary.

Keep note cards simple.

Do not use excessive shadows or complicated card layouts.

Support empty states with a useful Create Note action.

### B. Note Editor

This is the most important screen.

Prioritize writing experience over everything else.

Include:
- Editable title.
- Main content editor.
- Tag management.
- Save state.
- Updated timestamp.
- AI-generated summary.
- Key concepts.
- Linked notes.
- Backlinks.

Editor requirements:
- Wide readable writing column.
- Comfortable line length.
- Minimal distractions.
- Clear focus state.
- Keyboard-friendly interactions.
- Autosave only if properly implemented (debounced, no overlapping requests, never overwrites newer content with older content).
- Show saving, saved, and error states accurately.

The editor should feel like writing in a premium digital notebook.

Do not surround the editor with unnecessary borders.

### C. AI Search

Integrate the existing search capabilities.

Support:
- Keyword search.
- Semantic search.
- Hybrid search.
- Smart search.

Smart search should display:
- Relevant notes.
- AI explanation.
- Knowledge gaps.
- Suggested connections.

Present AI results in a readable editorial layout.

Do not pretend AI results are guaranteed to be correct.

Show loading and error states.

### D. Knowledge Graph

Visualize the existing note-linking relationships.

Requirements:
- Notes represented as nodes.
- Links represented as edges.
- Clickable nodes.
- Open selected note.
- Display relationship information when available.
- Zoom and pan controls.
- Search or filter nodes.

Use brown nodes and subtle brown edges on a white or warm-white canvas.

Avoid unnecessary visual complexity.

Do not invent graph relationships.

### E. Tags

Display all available tags.

Allow users to:
- Browse tags.
- Filter notes by tag.
- Open a note from tag results.

Use understated tag styling with brown text and beige backgrounds.

## 7. GraphQL Integration

The existing GraphQL API is the source of truth.

Endpoint:
POST /query

Playground:
GET /

Inspect the actual GraphQL schema before implementing queries and mutations.

Expected operations include:

Queries:
- notes
- note
- notesByTag
- links
- backlinks
- searchNotes
- semanticSearch
- hybridSearch
- smartSearch
- checkRelationship

Mutations:
- createNote
- updateNote
- deleteNote
- linkNotes
- unlinkNotes

Do not assume field names or input structures without checking schema.graphqls.

Implementation rules:
- Centralize GraphQL client configuration.
- Handle GraphQL errors.
- Handle loading states.
- Handle empty responses.
- Avoid unnecessary duplicate requests.
- Keep API logic separate from presentation components.
- Do not hardcode sample data in production screens.
- Do not silently swallow API errors.
- Always pass user input as GraphQL variables. Never build queries by string concatenation.

## 8. Local-First Behavior

The backend is becoming local-first: SQLite storage, local AI (Ollama), AI optional.

- AI may be disabled or unreachable. The UI must still work fully. Disable or hide the semantic, hybrid, and smart search modes when AI is unavailable, and fall back to keyword search with a clear message.
- Summaries and key concepts may be generated in the background after a note is saved. Treat empty values as "not ready yet", show a subtle "Generating…" state, and refetch, instead of showing an error.
- Never send note content to third-party services (analytics, telemetry, CDNs, external fonts or icons).
- The API URL comes from config (env var or dev proxy), defaulting to the local server.

## 9. Component Architecture

Prefer reusable, focused components.

Suggested structure:

```
src/
  components/
    layout/
      AppSidebar
      AppHeader
      WorkspaceLayout
      ContextPanel

    notes/
      NoteList
      NoteCard
      NoteEditor
      NoteTitle
      NoteMetadata
      TagList
      TagInput

    search/
      SearchBar
      SearchFilters
      SearchResults
      SmartSearchResults

    graph/
      KnowledgeGraph
      GraphNode
      GraphControls

    common/
      Button
      Input
      Dialog
      Tooltip
      EmptyState
      LoadingState
      ErrorState

  pages/
    NotesPage
    EditorPage
    SearchPage
    GraphPage
    TagsPage

  lib/
    graphql/
    utils/

  styles/
    globals.css
    tokens.css
```

Adapt this structure to the existing frontend framework instead of forcing a new architecture.

## 10. Interaction and UX Requirements

Every visible interactive element must work.

Required interactions:
- Create note.
- Edit note.
- Delete note with confirmation.
- Search notes.
- Filter by tag.
- Navigate between notes.
- Create note links.
- Remove note links.
- Open backlinks.
- Switch search modes.
- Navigate the knowledge graph.

UX requirements:
- Provide useful keyboard interactions.
- Avoid unnecessary confirmation dialogs.
- Use skeletons or subtle loading indicators.
- Provide clear empty states.
- Show errors with actionable information.
- Maintain responsive layouts.
- Preserve editor content during temporary API failures.

Do not create decorative buttons that do nothing.

## 11. Responsive Design

Desktop:
- Full sidebar.
- Main workspace.
- Optional context panel.

Tablet:
- Collapsible sidebar.
- Main workspace prioritized.
- Context panel becomes optional.

Mobile:
- Single-column layout.
- Sidebar becomes a drawer.
- Editor takes full width.
- Context information becomes a collapsible section.
- Touch-friendly controls.

Writing must remain comfortable on every screen size.

## 12. Accessibility

- Use semantic HTML.
- Provide accessible labels.
- Ensure keyboard navigation.
- Maintain readable contrast (WCAG AA for text).
- Support visible focus states.
- Use proper heading hierarchy.
- Do not rely on color alone to communicate status.

## 13. Design Anti-Patterns

Strictly avoid:

- Generic admin dashboards.
- Excessive rounded cards.
- Glassmorphism.
- Neon colors.
- Gradients.
- Excessive shadows.
- Colorful status badges.
- Unnecessary animations.
- Oversized navigation.
- Dense interfaces.
- Fake analytics.
- Fake note counts.
- Fake AI results.
- Placeholder interactions.
- Excessive modal dialogs.

## 14. Stack and Dependencies

- Inspect the repo first and use the existing frontend stack.
- If no frontend exists yet, stop and ask me which stack to use before scaffolding. (Suggested default: Vite + React + TypeScript.)
- Ask before adding any new dependency. Prefer small, well-maintained packages, and prefer the platform (fetch, CSS) over a library when it is simple enough.
- Pin versions via the lockfile. Do not add packages for things the project already has.
- Run the package manager's audit after installing and report any high-severity issues.

## 15. Secure Coding Rules

- Never use `dangerouslySetInnerHTML`, `innerHTML`, `eval`, or `new Function`. If note content is rendered as Markdown or HTML, use a well-known renderer with a sanitizer, and never render raw HTML from notes.
- Render all AI output (summaries, explanations, gaps, connections) as plain text. It is generated from user notes and must be treated as untrusted.
- Treat all API data as untrusted: validate shapes at the boundary (for example, with generated types or a schema validator), and handle null and missing fields.
- No secrets in frontend code, env files committed to git, or logs. The frontend must not contain API keys. AI provider keys belong to the backend only.
- Do not log note content, titles, or tokens to the console or to any external service. Remove debug logging before finishing.
- Do not store note content in `localStorage` except for a temporary unsaved-draft backup, and clear it after a successful save.
- Links to external URLs use `rel="noopener noreferrer"`. Only allow `http`, `https`, and `mailto` URL schemes for user-provided links.
- Do not disable security features (CSP, CORS checks, TLS verification) to make something work. In development, use a dev-server proxy instead of a wildcard CORS setting.
- Validate and trim user input on the client (lengths, empty titles), but never rely on that as the only protection.
- Deleting data requires an explicit user action, and the UI must reflect the real result of the API call.

## 16. Agent Safety Rules (how you work)

- Work only inside this project folder. Do not read or modify files outside it.
- Do not read, print, edit, or commit `.env` files, API keys, or database files (`data/`, `*.db`, `*.sqlite`).
- Never run destructive commands without asking first: `rm -rf`, `git reset --hard`, `git clean`, `git push --force`, dropping or truncating tables, deleting the `data/` folder.
- Do not run the app against my real data when testing writes. Use a copy of the database or a separate test database.
- Do not modify the Go backend, `schema.graphqls`, or `generated.go` unless I ask. If the UI needs a backend change, stop and describe what is needed.
- Do not run network requests to external services. Local development servers only.
- Do not install global packages, change system settings, or edit git config.
- Do not commit or push unless I ask. Never rewrite git history.
- If a command fails repeatedly, stop and report instead of trying risky workarounds.

## 17. Clean Code Standards

- Use TypeScript in strict mode when the project uses TypeScript. No `any` and no `@ts-ignore` without a comment explaining why.
- One responsibility per component or function. Split components that grow past roughly 150 lines or mix data fetching with presentation.
- Keep API calls and data mapping in `lib/graphql/` or hooks. Components receive typed props.
- Use clear, descriptive names. Components are PascalCase, hooks start with `use`, and booleans read like `isLoading` or `hasError`.
- No dead code, commented-out code, unused imports, leftover `console.log`, or TODOs without a short explanation.
- No duplicated logic. Extract shared helpers when the same code appears three times.
- Follow the existing formatter and linter config (Prettier, ESLint). Do not change lint rules to make errors disappear.
- Handle loading, empty, and error states for every data-driven view.
- Prefer simple solutions. Do not add abstractions, state libraries, or patterns that the current features do not need.
- Keep CSS in tokens and shared styles. Use the CSS variables from section 3, never hard-coded hex values in components.
- Write tests for non-trivial logic (utilities, hooks, data mapping) using the project's existing test setup.
- Keep changes small and focused. Do not make unrelated edits or reformat files you did not need to change.

## 18. Implementation Workflow

Before writing code:

1. Inspect the repository structure.
2. Identify the existing frontend framework.
3. Inspect package.json and installed dependencies.
4. Inspect the GraphQL schema.
5. Identify existing API integration.
6. Understand existing routing and state management.

Then implement in this order:

1. Design tokens and global styles.
2. Application shell and sidebar.
3. Notes library.
4. Note editor.
5. GraphQL integration.
6. Tags and navigation.
7. AI search.
8. Knowledge graph.
9. Responsive behavior.
10. Accessibility and polish.

Work in small steps. After each numbered step, make sure the project builds, then continue. If the task is large, stop after a logical checkpoint and report.

After implementation:

- Run the project.
- Run the type check, linter, and tests, and fix all errors.
- Check for build errors.
- Test GraphQL integration.
- Test all CRUD operations (on a copy of the data).
- Test responsive layouts.
- Verify that every visible action works.
- Summarize what you built, what you could not verify, and anything that needs a backend change.

## 19. Final Quality Standard

The finished product should look and feel like a carefully designed premium PKM application.

The experience should be:
- Minimal.
- Warm.
- Elegant.
- Functional.
- Content-focused.
- Consistent.

The brown-and-white color system is mandatory.

Prioritize usability and writing comfort over visual gimmicks.

Do not stop after building a static mockup. Deliver a working application integrated with the existing GraphQL backend.
