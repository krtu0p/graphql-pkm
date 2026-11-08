# CODEBASE.md — GraphQL PKM (`graphql-pkm`)

This document describes the repository so another AI agent (or developer) can answer
questions about it without reading every file. It is descriptive only; it is not a guide for
making changes. For contribution/design rules, read `AGENTS.md` (do not modify it here).

## 1. Project Overview

A Personal Knowledge Management (PKM) application with a Go + GraphQL backend and a React +
TypeScript frontend. Notes support tags, bidirectional links/backlinks, and AI-powered search
and insights.

Current state (as of the local-first migration tracked in `LOCAL_FIRST_MIGRATION.md`, whose
status section shows all three phases complete):

- Backend default storage is **SQLite** (`modernc.org/sqlite`, pure Go, no CGO); **MySQL** is
  still selectable via `DB_DRIVER=mysql`.
- Default AI provider is a **local Ollama** server using an OpenAI-compatible API
  (`AI_BASE_URL=http://localhost:11434/v1`); AI can be disabled via `AI_ENABLED=false`.
- GraphQL endpoint: `POST /query` (gqlgen); GraphQL Playground: `GET /`.
- Frontend: React 19 + TypeScript + Vite, lives in `web/`, calls the backend via relative
  `fetch('/query', ...)`.

## 2. Tech Stack

Backend (`go.mod`):

- Go `1.26.0`
- `github.com/99designs/gqlgen v0.17.84` (GraphQL server/generation)
- `github.com/vektah/gqlparser/v2 v2.5.31`
- `github.com/joho/godotenv v1.5.1` (env loading)
- `github.com/go-sql-driver/mysql v1.9.3`
- `modernc.org/sqlite v1.61.1` (SQLite driver, pure Go)
- `github.com/google/uuid v1.6.0` (indirect)

Frontend (`web/package.json`):

- `react ^19.0.0`, `react-dom ^19.0.0`
- `@types/react ^19.0.0`, `@types/react-dom ^19.0.0`
- `@vitejs/plugin-react ^4.3.4`
- `typescript ~5.7.2`
- `vite ^6.2.0`
- No GraphQL client library, no state library, no CSS framework, no test framework installed.

## 3. Repository Layout

```
cmd/server/main.go            # Composition root: env -> config, DB, AI client, services, gqlgen server
config/
  config.go                   # Env var loading + defaults (Host, Port, DB_*, AI_*, EMBED_*, SEMANTIC_THRESHOLD)
  config_test.go
internal/
  ai/
    client.go                  # OpenAI-compatible chat + embeddings HTTP client (Ollama/Groq)
    search.go                  # SmartSearch chat completion + JSON response parsing
    types.go                   # OpenAI request/response types, SmartSearchResult
  database/
    interface.go               # Database + EmbeddingsCache interfaces (+ compile-time checks)
    memory.go                  # In-memory DB implementation (NOT used by main; used in tests)
    mysql.go                   # MySQL implementation
    sqlite.go                  # SQLite implementation (default backend)
    embbeding_cache.go         # MySQL embeddings cache (NOTE: file name is misspelled "embbeding")
    sqlite_embeddings.go      # SQLite embeddings cache
    embeddings_interface.go   # EmbeddingsCache interface
  gql/
    schema.graphqls           # GraphQL schema
    generated/generated.go    # gqlgen-generated types + execution (do not edit by hand)
    resolvers/
      resolver.go              # Resolver struct + sub-resolver wiring
      query.go                 # Query resolvers (notes, search, backlinks, links, checks)
      mutation.go              # Mutation resolvers (create/update/delete note, link/unlink)
      note.go                  # Field resolvers for Note.links / Note.backlinks
      search.go                # SemanticSmartHybrid search resolvers
      subscription.go          # NOTE: unimplemented (returns errors)
  models/
    models.go                 # Go domain models bound to GraphQL via gqlgen autobind
  service/
    noteService.go            # Note CRUD, links, background AI enrichment worker
    aiService.go              # Semantic/smart/hybrid search, relationship check, cosine similarity
    searchService.go          # Facade over NoteService/AIService
    service_test.go
web/
  index.html
  package.json / package-lock.json / tsconfig*.json / vite.config.ts
  src/
    main.tsx                   # React entry, imports styles/globals.css
    App.tsx                    # Section-based routing shell + workspace state orchestration
    pages/
      NotesPage.tsx           # All Notes library (search, sort, note list)
      EditorPage.tsx          # Note editor (draft, autosave, draft backup, save states)
      SearchPage.tsx          # AI/keyword search UI + relationship checker
      TagsPage.tsx            # Tag browsing + notesByTag results
      GraphPage.tsx           # Knowledge graph page
      (Settings section rendered inline in App.tsx)
    components/
      common/                  # Button, Dialog, Input, LoadingState, EmptyState, ErrorState, Tooltip
      layout/                  # AppSidebar, AppHeader, WorkspaceLayout, ContextPanel
      notes/                   # NoteList, NoteCard, NoteEditor, NoteTitle, NoteMetadata, TagList, TagInput, NoteConnectionsSection
      search/                  # SearchBar, SearchFilters, SearchResults, SmartSearchResults, RelationshipChecker
      graph/                   # KnowledgeGraph, GraphNode, GraphControls, GraphInspector
    hooks/
      useNotesWorkspace.ts    # Notes CRUD/save/delete/link state + AI enrichment polling
      useAiSearch.ts           # Search modes (smart/hybrid/semantic/keyword) + AI-unavailable fallback
    lib/
      graphql/
        client.ts              # executeGraphQL -> fetch('/query') with error handling
        operations.ts          # GraphQL query/mutation strings + async fetch functions
        types.ts               # TS interfaces mirroring GraphQL models
        validators.ts          # Runtime response shape validation/mapping
        index.ts               # Barrel exports
      utils/
        draftStorage.ts       # localStorage unsaved-draft backup (key prefix `folio_unsaved_draft_`)
        format.ts              # date formatting, excerpts, tag normalization
        graphLayout.ts        # Graphviz-style layout for graph nodes/edges
    styles/
      globals.css             # Global styles (imports tokens.css), layout/component classes
      tokens.css              # Design tokens (brown/white palette, typography, spacing, radii)
    types/
      navigation.ts           # NavigationSection union + section metadata
gqlgen.yml                      # gqlgen config (schema, generated exec, resolver layout, model bindings)
go.mod / go.sum
LOCAL_FIRST_MIGRATION.md        # Task log for local-first migration (SQLite + local AI phases)
README.md                       # User-facing docs; partially stale (still mentions Groq/MySQL as primary)
```

## 4. Backend (`internal/`, `cmd/server`, `config/`)

### 4.1 Wiring

`cmd/server/main.go` is the composition root:

1. `godotenv.Load()`.
2. `config.LoadConfig()` (defaults: host `127.0.0.1`, port `8080`, driver `sqlite`,
   db path `./data/pkm.db`, AI base URL `http://localhost:11434/v1`, AI model `llama3.2`,
   embed model `nomic-embed-text`, semantic threshold `0.6`).
3. Open DB per `DB_DRIVER` (`sqlite` default -> `database.NewSQLiteDB`, embeddings cache
   `NewSQLiteEmbeddingsCache`; `mysql` -> `NewMySQLDB` + `NewMySQLEmbeddingsCache`; other -> fatal).
4. Build `ai.Client` via `NewClientWithConfig` from config; enabled when base URL is set and
   `AI_ENABLED` allows.
5. Services: `NoteService(db, aiClient)`, `AIService(aiClient, embeddingsCache, noteService)`,
   `SearchService(noteService, aiService)`; resolver: `resolvers.NewResolver(noteService, searchService)`.
6. gqlgen handler: `handler.NewDefaultServer(generated.NewExecutableSchema(...))`; mounts
   playground at `/` and GraphQL at `/query`.
7. Starts `http.ListenAndServe` on `cfg.Host:cfg.Port`.

Layering rule (per `LOCAL_FIRST_MIGRATION.md`): calls only go downward —
`resolvers -> services -> (database | ai)`.

### 4.2 GraphQL Schema (`internal/gql/schema.graphqls`)

- `scalar Time`.
- Types: `Note` (`id`, `title`, `content`, `tags: [String!]!`, `createdAt`, `updatedAt`,
  `summary: String`, `keyConcepts: [String!]!`, `links: [Link!]!`, `backlinks: [Link!]!`),
  `Link` (`id`, `sourceNoteId`, `targetNoteId`, `description: String`, `createdAt`),
  `Backlink` (Query-level; `id`, `sourceNote: Note!`, `description`, `createdAt`),
  `SearchResult` (`note`, `score: Float!`, `reason`, `matchType`),
  `SmartSearchResponse` (`results`, `explanation`, `gaps`, `connections`).
- Inputs: `CreateNoteInput` (title/content/tags), `UpdateNoteInput` (nullable title/content,
  tags).
- Queries: `notes`, `note(id)`, `searchNotes(query)`, `notesByTag(tag)`, `backlinks(noteId)`,
  `links(noteId)`, `semanticSearch(query)`, `smartSearch(query)`, `hybridSearch(query)`,
  `checkRelationship(noteIDA: ID!, noteIDB: ID!)`.
- Mutations: `createNote(input)`, `updateNote(id, input)`, `deleteNote(id)`,
  `linkNotes(sourceId, targetId, description)`, `unlinkNotes(linkId)`.
- Subscription: `noteCreated`, `noteUpdated` — **declared but not implemented** (resolvers return
  errors: "subscriptions not implemented yet").

gqlgen config: `gqlgen.yml` (schema `internal/gql/schema.graphqls`, exec
`internal/gql/generated/generated.go`, resolver layout `internal/gql/resolvers`,
autobind to `graphql-pkm/internal/models`, custom model mapping for Note/Link/inputs/SearchResult/
SmartSearchResponse).

### 4.3 Search

- Keyword: resolver (`search.go`) -> `SearchService.SearchNotes` -> `NoteService` ->
  `Database.SearchNotes`. Backends do case-insensitive substring match (`LIKE '%term%'` on
  lower-cased title/content; in-memory: `strings.Contains`). Ordered by `created_at DESC`.
- Semantic: resolver -> `AIService.SemanticSearch` -> `aiClient.GetEmbedding(query)` (embeddings
  endpoint), then for each note compute/get embedding (cached via `EmbeddingsCache`, cache key is
  note ID; model + dimension mismatch => cache miss), cosine similarity; keep results with
  similarity > threshold (default `0.6`, configurable via `SEMANTIC_THRESHOLD`). Result:
  `score = similarity`, `reason = "Conceptually similar"`, `matchType = "semantic"`. Falls back to
  keyword search if AI/embedding unavailable.
- Hybrid: `AIService.HybridSearch` = keyword results (score `0.8`, reason `"Keyword match"`,
  matchType `"keyword"`) + semantic results; dedupe by note ID; sort by score descending.
- Smart: `AIService.SmartSearch` -> keyword search for context notes (max 10) -> AI chat
  (`ai.Client.SmartSearch`) with prompt containing user query + context notes, expects JSON with
  `relevantNotes: [{noteId, reason, score}]`, `explanation`, `gaps`, `connections`; resolves
  relevant note IDs to notes. Falls back to keyword results with explanation "AI unavailable; ..."`
  when AI is disabled or errors.
- Relationship check: `AIService.CheckRelationship` builds an AI prompt comparing two notes and
  returns the AI-generated relationship string (returns "AI features are disabled." when AI off).

### 4.4 AI / Ollama Integration

- `internal/ai/client.go`: OpenAI-compatible client. Chat: `POST {BaseURL}/chat/completions`
  (default base `http://localhost:11434/v1`, model `llama3.2`, low temperature `0.2`,
  max tokens vary: 150 summary / 100 key-concepts / 4096 smart). Embeddings: `POST
  {EmbedBaseURL}/embeddings` with `{model, input:[text]}` (default embed model `nomic-embed-text`).
  Timeout default 30s. `IsEnabled()` returns true when a base URL is configured and AI is not
  disabled.
- `internal/ai/search.go`: `SmartSearch(query, contextNotes map[string]string)` -> chat, then
  `cleanAIContent` (strips code fences) + `extractJSON` (regex to first `{...}`) + unmarshal into
  `SmartSearchResult`.
- `internal/service/noteService.go`: background enrichment worker (bounded channel, capacity 100,
  drops job when full). On create/update (when AI enabled) it enqueues `{noteID, title, content}`;
  worker calls `aiClient.GenerateSummary` and `aiClient.ExtractKeyConcepts`, then persists summary +
  key concepts back through the DB (note `UpdatedAt` bumped).
- `internal/service/aiService.go`: orchestrates search + relationship; holds embeddings cache,
  semantic threshold; `cosineSimilarity(a, b []float32) float64` implementation.

### 4.5 Storage

- Interface `database.Database`: Create/Get/Update/Delete note, GetAllNotes, GetNotesByTag,
  SearchNotes, CreateLink, GetLinksByNote, DeleteLink, GetBackLinksByNote.
- Interface `database.EmbeddingsCache`: Store/Get/Delete/Clear/Count/GetCacheStats; implemented
  by `MySQLEmbeddingsCache` (`internal/database/embbeding_cache.go`) and
  `SQLiteEmbeddingsCache` (`internal/database/sqlite_embeddings.go`).
- SQLite implementation notes: DSN includes `foreign_keys(1)`, `journal_mode(WAL)`, `busy_timeout`;
  tables `notes`, `links`, `embeddings_cache`; tags/key concepts stored as JSON text; tag filter
  uses `EXISTS (SELECT 1 FROM json_each(notes.tags) WHERE value = ?)`; upsert via
  `ON CONFLICT(note_id) DO UPDATE`; cascades via `ON DELETE CASCADE`; notes ordered by
  `created_at DESC`.
- MySQL implementation: JSON columns, `JSON_CONTAINS` for tags, `ON DUPLICATE KEY UPDATE`.
- `MemoryDB` implements the same interface for tests but is not wired into `main`.
- Tests: `internal/database/database_test.go` (contract suite run against in-memory + SQLite,
  incl. cascade/embeddings), `internal/service/service_test.go` (enrichment + AI-unreachable
  graceful degradation), `config/config_test.go`.

## 5. Frontend (`web/`)

- Entry: `web/src/main.tsx` -> `web/src/App.tsx`. **No router library**; navigation is state-based
  via `NavigationSection` union in `web/src/types/navigation.ts` (`notes | editor | tags | graph |
  search | settings`) with metadata map.
- `App.tsx` owns state: `activeSection`, `selectedTag`, `librarySearchQuery`, plus data from
  `useNotesWorkspace` (notes, activeNote, backlinks, save/delete/link handlers, loading/error,
  AI enrichment flag). Renders the matching page inside `WorkspaceLayout` (which composes
  AppSidebar + AppHeader + ContextPanel + content).
- Pages: `NotesPage` (library: search input client-side filters, sort options, note list,
  empty/error/loading states), `EditorPage` (title/content/tag draft; autosave, draft backup,
  save states, AI summary/key concepts, backlinks/links), `SearchPage` (uses `useAiSearch`,
  `SearchBar`, `SearchFilters`, `SearchResults`/`SmartSearchResults`, `RelationshipChecker`),
  `TagsPage` (uses `fetchNotesByTag`, tag summaries), `GraphPage` (uses `fetchNoteLinks` /
  `fetchNoteBacklinks`, builds layout via `lib/utils/graphLayout`).
- Components organized by domain: `common/`, `layout/`, `notes/`, `search/`, `graph/` —
  roughly mirroring the structure suggested in `AGENTS.md` section 9.
- State management: React hooks only (`useNotesWorkspace`, `useAiSearch`); no Redux/Zustand.
  Server state lives in async functions in `web/src/lib/graphql/operations.ts`; no React Query.
- Editor autosave: debounce ~1000ms via `useEffect` on title/content/tags; save-state machine
  `saved | saving | unsaved | error | new`; manual save button; queued re-save when a newer
  revision was made during save; unsaved drafts persisted to localStorage
  (`folio_unsaved_draft_<noteId|new>`) and cleared after successful save; restore compares
  `backup.savedAt` to server `updatedAt`.
- AI enrichment UX: after save, `useNotesWorkspace` schedules enrichment polling (3s/7.5s) and
  shows "Generating…/Not ready yet" states until summary/key concepts appear.
- AI search fallback UX: `useAiSearch` starts in mode `smart`; when AI is unavailable (backend
  fallback reasons or error text mentioning AI/Ollama/embedding) it switches to keyword mode and
  exposes `retryAiModes`/`isAiUnavailable`/`aiFallbackMessage`.

### 5.1 Where GraphQL calls live

- `web/src/lib/graphql/client.ts`: `executeGraphQL<TData>(query, variables, signal)` -> plain
  `fetch('/query', {method POST, JSON body {query, variables}})`, throws `GraphQLRequestError`,
  handles network errors and GraphQL `errors`, requires `data`.
- `web/src/lib/graphql/operations.ts`: query/mutation strings (shared `NOTE_FIELDS` fragment,
  `NOTES_QUERY`, `NOTE_QUERY`, `NOTES_BY_TAG_QUERY`, `SEARCH_NOTES_QUERY`, `LINKS_QUERY`,
  `BACKLINKS_QUERY`, `SEMANTIC_SEARCH_QUERY`, `HYBRID_SEARCH_QUERY`, `SMART_SEARCH_QUERY`,
  `CHECK_RELATIONSHIP_QUERY`, `CREATE_NOTE_MUTATION`, `UPDATE_NOTE_MUTATION`, `DELETE_NOTE_MUTATION`,
  `LINK_NOTES_MUTATION`, `UNLINK_NOTES_MUTATION`) plus exported async functions returning parsed
  data. Queries/mutations always use GraphQL variables, never string concatenation.
- `web/src/lib/graphql/types.ts`: TS interfaces (`Note`, `Link`, `Backlink`, `SearchResult`,
  `SmartSearchResponse`, inputs, `GraphQLResponse<T>`).
- `web/src/lib/graphql/validators.ts`: runtime validation/mapping from raw API JSON to TS types.
- Barrel: `web/src/lib/graphql/index.ts`.

### 5.2 Styling

- `web/src/styles/tokens.css`: CSS custom properties — strict brown & white palette
  (`--background #FAF9F6`, `--surface #FFFFFF`, `--surface-secondary #F3EFE8`,
  `--foreground #33271F`, `--primary #79563B`, `--border`, `--accent`, `--selection`, etc.),
  typography tokens (`--font-sans/--font-serif/--font-editor/--font-mono`,
  `--text-*`, `--leading-*`), layout dimensions (`--sidebar-width: 240px`, `--context-panel-width:
  280px`, `--editor-max-width: 720px`, ...), spacing scale, radii, transitions.
- `web/src/styles/globals.css`: imports tokens, resets, layout classes (`.workspace-shell`,
  `.app-sidebar`, `.workspace-main`, `.context-panel`, `.page-header`, `.editorial-panel`, form
  controls, empty/loading/error/skeleton states, editor styles, responsive `@media` blocks at
  `max-width: 1024px` and `max-width: 768px`).
- Design rules enforced in code: brown/white only, no gradients/decorative shadows, fonts via CSS
  variables (note: no `@font-face`/font files are currently loaded — see Known Gaps).

## 6. How to Run, Build, Test, Lint

Backend (from repo root):

```bash
go run ./cmd/server                 # start server (default: http://127.0.0.1:8080, /query, playground /)
go run github.com/99designs/gqlgen generate   # regenerate GraphQL code
go build ./...                      # build check
go build -o pkm-server cmd/server/main.go
go test ./...                       # backend tests
go vet ./...                        # backend vet
```

Frontend (from `web/`):

```bash
npm run dev                          # Vite dev server (proxy /query -> http://127.0.0.1:8080)
npm run build                        # tsc -b && vite build
npm run lint                         # tsc --noEmit
npm run preview                      # vite preview
```

There is no frontend test script or tests in `web/`; backend tests exist under `internal/` and `config/`.

## 7. AGENTS.md Section 18 — Implementation Status vs. Actual Code

The workflow in `AGENTS.md` section 18 prescribes an order of steps (tokens/global styles ->
shell/sidebar -> notes library -> editor -> GraphQL integration -> tags/nav -> AI search ->
knowledge graph -> responsive -> accessibility/polish).

Implemented based on the actual code:

- Step 1 (tokens & global styles): mostly done (`tokens.css`, `globals.css`). Gap: font stacks are
  declared as CSS variables but no font files are self-hosted or loaded (AGENTS.md requires
  self-hosting; not done).
- Step 2 (shell & sidebar): done — `WorkspaceLayout`, `AppSidebar`, `AppHeader`, `ContextPanel`;
  sidebar contains All Notes, Tags, Knowledge Graph, AI Search, Settings + quick tag filters;
  sidebar collapses on tablet and becomes drawer on mobile; context panel collapsible.
- Step 3 (notes library): done — `NotesPage`, `NoteList`, `NoteCard`, search/sort/empty states.
- Step 4 (editor): done — `EditorPage` with autosave, save states, draft backup, AI summary UI.
- Step 5 (GraphQL integration): done — `lib/graphql/*` client/operations/types/validators,
  plain `fetch`, relative `/query`, Vite proxy, variables used, error/loading/empty handling.
- Step 6 (tags & navigation): done — `TagsPage`, `notesByTag`, sidebar tag filters.
- Step 7 (AI search): done — `useAiSearch` with smart/hybrid/semantic/keyword modes and backend
  fallback to keyword when AI is unavailable.
- Step 8 (knowledge graph): done — `GraphPage` + `KnowledgeGraph`/`GraphNode`/`GraphControls`/
  `GraphInspector`, pan/zoom, tag filter, open/focus, from real links.
- Step 9 (responsive): partially/mostly done — `@media` blocks at 1024px and 768px present
  (tablet collapse, mobile drawer, full-width editor, collapsible context panel).
- Step 10 (accessibility & polish): partially done — semantic HTML, aria labels, focus-visible
  styles, sr-only helpers present; full WCAG AA compliance unverified.

Not done / unclear: frontend unit tests (none exist; AGENTS.md asks for tests of non-trivial
logic — backend tests exist for service/database only); ESLint config (frontend `lint` script is
just `tsc --noEmit`); visual verification of every screen; `web/dist` embedding into Go binary not
yet done.

## 8. Known Gaps, TODOs, and Things That Look Broken

1. **GraphQL subscriptions are not implemented.** `internal/gql/resolvers/subscription.go`
   returns errors for `noteCreated` and `noteUpdated`. The schema still declares them.
2. **README.md is stale.** It still describes Groq as the AI provider and MySQL as the primary
   database, while the code (post-local-first migration) defaults to Ollama + SQLite. Keep in mind
   when answering questions about the stack.
3. **AI search fallbacks are coarse.** `HybridSearch` gives keyword results a hardcoded score of
   `0.8`; dedupe keeps the first occurrence (keyword) when a note matches both. Smart search
   fallback swallows AI errors into a generic keyword explanation.
4. **AI enrichment is fire-and-forget with limits.** Bounded channel (cap 100) silently drops
   enrichment jobs when full; enrichment updates the note but does not trigger re-enrichment of
   related notes.
5. **AI client naming/types.** `internal/ai/client.go` and `types.go` keep "openrouter" in type and
   field names even though the client is now provider-agnostic (Ollama/OpenAI-compatible).
6. **No backend GraphQL-level integration tests** (only service/database unit tests); gqlgen
   generated code is large (`internal/gql/generated/generated.go`) and must be regenerated via
   `go run github.com/99designs/gqlgen generate` after schema changes.
7. **Frontend**: no router, no tests, no ESLint (only `tsc --noEmit`); editor autosave relies on
   localStorage draft backup (per AGENTS.md allowed) with restore logic comparing timestamps; AI
   "generating" polling uses fixed delays (3s/7.5s).
8. **Minor**: database file filename typo `embbeding_cache.go`; `data/` directory holds local DB
   files (not read/committed); `.env` is gitignored, only `.env.example` is committed.

## 9. Quick Reference for Common Questions

- "Where is the GraphQL schema?" -> `internal/gql/schema.graphqls`; generated code
  `internal/gql/generated/generated.go`; resolver wiring `internal/gql/resolvers/`.
- "How does semantic search differ from keyword?" -> keyword is substring LIKE over title/content;
  semantic embeds query + notes and uses cosine similarity over threshold (0.6 default).
- "How are embeddings cached?" -> per-note cache table (`embeddings_cache` in SQLite / MySQL),
  keyed by note ID, with model + dimension checks.
- "How does the frontend talk to the backend?" -> `web/src/lib/graphql/client.ts` posts to
  relative `/query`; Vite proxy in `web/vite.config.ts` forwards to `http://127.0.0.1:8080`.
- "How is AI output treated?" -> rendered as plain text; never injected as HTML/markdown; AI
  outputs treated as untrusted and displayed as text (see AGENTS.md security rules).
