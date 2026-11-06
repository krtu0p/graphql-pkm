# Codebase Analysis: GraphQL-PKM

A GraphQL-based Personal Knowledge Management (PKM) server written in Go, with AI-powered
search (Groq LLM + Jina embeddings), MySQL persistence, and a bidirectional note-linking
("knowledge graph") system.

- **Module**: `graphql-pkm` (Go 1.25)
- **Key dependencies**: `99designs/gqlgen` (GraphQL codegen), `go-sql-driver/mysql`, `joho/godotenv`, `vektah/gqlparser`
- **Entry point**: `cmd/server/main.go`
- **Public API**: GraphQL at `POST /query`, Playground at `GET /`

---

## 1. Architecture at a Glance

The project follows a classic **layered architecture**:

```
                         HTTP (net/http)
                              │
        ┌─────────────────────┴─────────────────────┐
        │  cmd/server/main.go (composition root)    │   ← wires everything together
        └─────────────────────┬─────────────────────┘
                              │ builds dependency graph
                              ▼
    ┌────────────────────────────────────────────────────┐
    │  GraphQL layer  (internal/gql)                     │
    │  schema.graphqls → generated.go → resolvers/*      │
    │  resolver.go = root Resolver struct                │
    └──────┬─────────────────────────────────────────────┘
           │ calls (thin, no business logic)
           ▼
    ┌────────────────────────────────────────────────────┐
    │  Service layer  (internal/service)                 │
    │  NoteService, AIService, SearchService             │
    │  ← all business rules live here                    │
    └──────┬──────────────────────┬──────────────────────┘
           │                      │
           ▼                      ▼
    ┌────────────────┐    ┌───────────────────────────┐
    │ database layer │    │  AI layer (internal/ai)   │
    │ Database iface │    │  Client (HTTP client)     │
    │ MySQLDB        │    │  Groq chat / Jina embeds  │
    │ MemoryDB       │    └───────────────────────────┘
    │ EmbeddingsCache│
    └───────┬────────┘
            ▼
        MySQL
```

Cross-cutting types used by every layer live in **`internal/models/models.go`**
(`Note`, `Link`, `Backlink`, `SearchResult`, `SmartSearchResponse`, input structs).

---

## 2. File Inventory

| File | Role |
|---|---|
| `cmd/server/main.go` | Composition root: loads config, opens DB, builds services, mounts HTTP handlers |
| `config/config.go` | Reads environment variables into a `Config` struct |
| `internal/models/models.go` | Domain models shared by all layers; bound to the GraphQL schema |
| `internal/database/interface.go` | `Database` interface (note CRUD + links); compile-time checks |
| `internal/database/mysql.go` | MySQL implementation; also auto-creates tables on startup |
| `internal/database/memory.go` | In-memory fallback implementation (`sync.RWMutex`-guarded maps) — currently unused by main |
| `internal/database/embeddings_interface.go` | `EmbeddingsCache` interface |
| `internal/database/embbeding_cache.go` | MySQL implementation of the embedding cache (note_id → JSON vector) |
| `internal/ai/client.go` | HTTP client for Groq (chat) and Jina (embeddings); prompts for summary/concepts |
| `internal/ai/types.go` | Wire-format structs for OpenAI-compatible API + `SmartSearchResult` |
| `internal/ai/search.go` | AI "smart search": sends notes + query, expects JSON back |
| `internal/service/noteService.go` | Note CRUD + link logic, ID generation, tag normalization, AI metadata on write |
| `internal/service/aiService.go` | Semantic / hybrid / smart search, cosine similarity, dedup, relationship check |
| `internal/service/searchService.go` | Facade combining NoteService + AIService |
| `internal/gql/schema.graphqls` | GraphQL SDL (single source of truth for the API) |
| `internal/gql/generated/generated.go` | gqlgen-generated runtime (do not edit) |
| `internal/gql/resolvers/resolver.go` | Root `Resolver` struct that holds services |
| `internal/gql/resolvers/query.go` | Query resolvers (notes, tags, links, backlinks, relationship) |
| `internal/gql/resolvers/search.go` | Query resolvers (semantic/smart/hybrid search) |
| `internal/gql/resolvers/mutation.go` | Mutation resolvers (CRUD + linking) |
| `internal/gql/resolvers/note.go` | Note **field** resolvers (lazy `links` / `backlinks`) |
| `internal/gql/resolvers/subscription.go` | Stubbed subscriptions (return "not implemented") |
| `gqlgen.yml` | gqlgen config: autobinds GraphQL types to `internal/models` structs |

---

## 3. Where Functions Connect — the Call Graph

### 3.1 Startup wiring (`cmd/server/main.go`)

This is the single place where **every component is created and connected**:

```
main()
 ├─ godotenv.Load()                       (.env → env vars)
 ├─ config.LoadConfig()                   (config/config.go) ──► *Config
 ├─ database.NewMySQLDB(cfg.DatabaseUrl)  ──► *MySQLDB  (also creates tables)
 │    └─ createTables(db)                 notes / links / embeddings_cache
 ├─ database.NewMySQLEmbeddingsCache(mysqlDB)  ──► *MySQLEmbeddingsCache
 │    └─ wraps the SAME *MySQLDB (shares its *sql.DB handle)
 ├─ ai.NewClient(apiKey, apiUrl, model, jinaToken)  ──► *ai.Client
 ├─ service.NewNoteService(mysqlDB, aiClient)       ──► *NoteService
 │      NOTE: mysqlDB passed as `database.Database` (interface)
 ├─ service.NewAIService(aiClient, embeddingsCache, noteService) ──► *AIService
 │      NOTE: aiService is built AFTER noteService because it depends on it
 ├─ service.NewSearchService(noteService, aiService) ──► *SearchService
 ├─ resolvers.NewResolver(noteService, searchService) ──► *resolvers.Resolver
 └─ generated.NewExecutableSchema(Config{Resolvers: resolver}) + handler
      HTTP: "/" → Playground, "/query" → GraphQL server
```

Dependency direction is strictly one-way: **resolvers → services → (database | ai)**.
No layer ever calls upward.

### 3.2 Resolver → Service connections

The root `Resolver` (`resolvers/resolver.go`) holds exactly two dependencies, and every
per-operation resolver struct embeds it:

```go
type Resolver struct { noteService, searchService }

type queryResolver      struct{ *Resolver }   // query.go + search.go
type mutationResolver   struct{ *Resolver }   // mutation.go
type subscriptionResolver struct{ *Resolver } // subscription.go
type noteResolver       struct{ *Resolver }   // note.go
```

```
 ═══ QUERY resolvers ═══                       service method called
 queryResolver.Notes()            ──────────►   NoteService.GetAllNotes()
 queryResolver.Note(id)           ──────────►   NoteService.GetNote(id)
 queryResolver.SearchNotes(q)     ──────────►   SearchService.SearchNotes(q) ─► NoteService.SearchNotes
 queryResolver.NotesByTag(tag)    ──────────►   NoteService.GetNotesByTag(tag)
 queryResolver.Backlinks(noteID)  ──────────►   NoteService.GetBackLinks(noteID) + GetNote() per link
 queryResolver.Links(noteID)      ──────────►   NoteService.GetLinks(noteID)
 queryResolver.CheckRelationship ───────────►   SearchService.CheckRelationship(a, b)

 search.go (still queryResolver):
 queryResolver.SemanticSearch(q)  ──────────►   SearchService.SemanticSearch(q)
 queryResolver.SmartSearch(q)     ──────────►   SearchService.SmartSearch(q)
 queryResolver.HybridSearch(q)    ──────────►   SearchService.HybridSearch(q)

 ═══ MUTATION resolvers ═══
 mutationResolver.CreateNote      ──────────►   NoteService.CreateNote(title, content, tags)
 mutationResolver.UpdateNote      ──────────►   NoteService.UpdateNote(id, …)
 mutationResolver.DeleteNote      ──────────►   NoteService.DeleteNote(id)
 mutationResolver.LinkNotes       ──────────►   NoteService.CreateLink(source, target, desc)
 mutationResolver.UnlinkNotes     ──────────►   NoteService.DeleteLink(linkID)
```

### 3.3 Field resolvers (lazy loading)

`gqlgen.yml` marks `Note.links` and `Note.backlinks` with `resolver: true`, so these
fields are **not** taken from the `models.Note` struct — they are computed on request:

```
noteResolver.Links(obj)     ──► NoteService.GetLinks(obj.ID)     ──► MySQLDB.GetLinksByNote
noteResolver.Backlinks(obj) ──► NoteService.GetBackLinks(obj.ID) ──► MySQLDB.GetBackLinksByNote
```

All other `Note` fields (id, title, tags, summary…) are resolved directly from the
`models.Note` struct via autobinding — no code runs for them.

### 3.4 Service → low-level layer connections

**SearchService is a pure dispatcher/facade:**
```
SearchService.SearchNotes      ─► NoteService.SearchNotes   (keyword path)
SearchService.SemanticSearch   ─► AIService.SemanticSearch
SearchService.SmartSearch      ─► AIService.SmartSearch
SearchService.HybridSearch     ─► AIService.HybridSearch
SearchService.CheckRelationship─► NoteService.GetNote() ×2   (must exist)
                               ─► AIService.CheckRelationship(noteA, noteB)
```

**AIService depends on three things** — noteService (for data), aiClient (raw AI calls),
embeddingsCache (persistence):

```
AIService.SemanticSearch(query)
 ├─ aiClient.GetEmbedding(query)           Jina API (float32 vector)
 ├─ noteService.GetAllNotes()              full table scan
 ├─ for each note:
 │    ├─ embeddingsCache.Get(note.ID)                 cache lookup (MySQL)
 │    ├─ on miss: aiClient.GetEmbedding(title+content)
 │    │           + embeddingsCache.Store(note.ID, …)  cache write
 │    └─ cosineSimilarity(queryVec, noteVec)  → keep if > 0.6
 └─ sort by score desc

AIService.HybridSearch(query)
 ├─ noteService.SearchNotes(query)         keyword hits → Score 0.8
 ├─ AIService.SemanticSearch(query)        (recursive self-call!)
 └─ deduplicateResults()                   keep highest score per note, sort

AIService.SmartSearch(query)
 ├─ noteService.SearchNotes(query)         → top 10 notes as context
 ├─ aiClient.SmartSearch(query, contextMap)  → LLM returns JSON
 │    └─ parse → ai.SmartSearchResult{RelevantNotes, Explanation, Gaps, Connections}
 └─ noteService.GetNote(id) per returned noteId → []SearchResult

AIService.CheckRelationship(a, b) ─► AIService.AskAI(prompt) ─► aiClient.Ask(prompt)
```

**NoteService is the connection between AI and DB on every write:**

```
NoteService.CreateNote(title, content, tags)
 ├─ generateId()                              crypto/rand 32-hex-char ID
 ├─ normalizeTags(tags)                       trim + lowercase, drop empties
 ├─ (if AI enabled):
 │    ├─ aiClient.GenerateSummary(title, content)     → note.Summary
 │    └─ aiClient.ExtractKeyConcepts(title, content)  → note.KeyConcepts
 │         └─ on JSON-parse failure: extractConceptsFromText() fallback
 └─ db.CreateNote(note)                       Database interface

NoteService.UpdateNote(id, …)
 ├─ db.GetNote(id)                            must exist
 ├─ apply optional field changes (pointers = "was it set?")
 ├─ re-run GenerateSummary/ExtractKeyConcepts if title or content changed
 └─ db.UpdateNote(note)

NoteService.CreateLink(sourceID, targetID, desc)
 ├─ db.GetNote(sourceID)                      BOTH must exist
 ├─ db.GetNote(targetID)
 ├─ reject sourceID == targetID               (self-link prevention)
 └─ db.CreateLink(link)                       FK constraints cascade on delete
```

### 3.5 Database layer connections

Both storage backends satisfy the same `Database` interface
(`interface.go` enforces this with `var _ Database = (*MemoryDB)(nil)`):

```
Database interface                          MySQLDB                          MemoryDB
──────────────────────                      ───────                          ────────
CreateNote / GetNote / GetAllNotes      →   SQL INSERT/SELECT (tags &        map lookups under
UpdateNote / DeleteNote                     key_concepts as JSON columns)    sync.RWMutex
GetNotesByTag                           →   JSON_CONTAINS(tags, '"tag"')     scan all notes
SearchNotes                             →   LOWER(title/content) LIKE %q%    strings.Contains
CreateLink / GetLinksByNote             →   links WHERE source_note_id       map scan
GetBackLinksByNote                      →   links WHERE target_note_id       map scan
DeleteLink                              →   DELETE, 404 if 0 rows affected   map delete
```

`MemoryDB` satisfies the interface but is **only used within the package** (compile-time
check); `main.go` always picks MySQL. Delete-cascade behavior:

- `main.go` config lifetimes: notes deletion cascades to `links` and `embeddings_cache`
  rows via foreign keys (handled by MySQL, not Go code).

**Embedding cache** (`MySQLEmbeddingsCache`, wraps `MySQLDB.db.db` — the raw `*sql.DB`):

```
Store(noteID, []float32)   → INSERT … ON DUPLICATE KEY UPDATE  (upsert)
Get(noteID)                → SELECT embedding_json → []float32
Delete / Clear / Count / GetCacheStats
```

### 3.6 AI layer connections

`ai.Client` is the only place that speaks to the outside world besides MySQL:

```
Client.makeRequest(endpoint, payload)     shared: JSON marshal + Bearer auth + HTTP POST
 └─ used by:
     ├─ GenerateSummary(title, content)   → Groq /chat/completions (max_tokens 150)
     ├─ ExtractKeyConcepts(title, content)→ Groq /chat/completions, expects JSON array
     ├─ Ask(prompt)                       → Groq /chat/completions (free-form)
     └─ ai.SmartSearch(query, context)    → Groq /chat/completions, expects JSON object

Client.GetEmbedding(text)                 → Jina https://api.jina.ai/v1/embeddings
                                            (separate code path — its own request builder,
                                             uses jinaToken, model jina-embeddings-v3, 1024 dims)
Client.GenerateEmbedding(text)            → GetEmbedding + float32→float64 conversion
                                            (not called by anything else — dead code)
Client.IsEnabled()                        → apiKey != ""  (gates everything)
```

Helpers that glue AI output into Go data:
```
cleanAIContent()     strips ```json fences
extractJSON()        regex-picks the {...} block out of chat text
formatNotesForAI()   truncates each note to 500 chars, builds prompt context
extractConceptsFromText() fallback parser for ExtractKeyConcepts
```

---

## 4. Full Request Traces (end-to-end examples)

### 4.1 `mutation { createNote(input: {…}) }`

```
HTTP /query
  → generated.go (gqlgen runtime parses/executes)
    → mutationResolver.CreateNote(ctx, input)           [mutation.go]
      → NoteService.CreateNote(title, content, tags)    [noteService.go]
        ├─ generateId() → e.g. "683b9e86…"
        ├─ normalizeTags()
        ├─ aiClient.GenerateSummary()                   → Groq HTTP
        ├─ aiClient.ExtractKeyConcepts()                → Groq HTTP
        └─ MySQLDB.CreateNote(note)                     → SQL INSERT notes
  ← JSON response to client
```

### 4.2 `query { smartSearch(query: "…") }`

```
HTTP /query
  → queryResolver.SmartSearch(q)                        [search.go]
    → SearchService.SmartSearch(q)                      [searchService.go]
      → AIService.SmartSearch(q)
        ├─ NoteService.SearchNotes(q)                   → MySQL LIKE query
        │    (cap 10 notes → contextMap {id: "title: content"})
        ├─ aiClient.SmartSearch(q, contextMap)          [ai/search.go]
        │    ├─ formatNotesForAI()
        │    ├─ makeRequest("/chat/completions")        → Groq HTTP
        │    └─ cleanAIContent + extractJSON + Unmarshal → SmartSearchResult
        └─ NoteService.GetNote(noteID) per hit          → MySQL SELECT
      ← SmartSearchResponse{Results, Explanation, Gaps, Connections}
  ← JSON response
```

### 4.3 `query { note(id) { links backlinks } }`

```
HTTP /query
  → queryResolver.Note(id)            → NoteService.GetNote → MySQL SELECT   (1 query)
  → noteResolver.Links(obj)           → NoteService.GetLinks → MySQL         (per request)
  → noteResolver.Backlinks(obj)       → NoteService.GetBacklinks → MySQL     (per request)
```

The "knowledge graph" is never materialized in Go: forward/backward edges are simply two
directional WHERE clauses over the same `links` table.

---

## 5. Design Decisions Worth Knowing

1. **autobind, not generated models** — gqlgen binds GraphQL types directly to
   `internal/models` structs (`gqlgen.yml`), so the DB layer and GraphQL share one type
   definition. Only `links`/`backlinks` need explicit field resolvers.
2. **Interfaces at the DB boundary** — `database.Database` and `database.EmbeddingsCache`
   allow swapping MySQL for `MemoryDB` (already implements the interface) without touching
   service code.
3. **Graceful AI degradation** — `aiClient.IsEnabled()` gates every AI call:
   - search falls back to keyword search (`AIService.fallbackSearch`),
   - note CRUD still succeeds if summary/concepts generation fails (warn-and-continue).
4. **Cost control via embedding cache** — note embeddings computed once by Jina are stored
   as JSON in MySQL and reused across semantic searches (upsert keyed by `note_id`).
5. **Threshold-based similarity** — semantic matches keep only cosine similarity > 0.6;
   keyword matches in hybrid search get a fixed score 0.8; results deduped by note ID.
6. **Cascade deletion in the schema, not in code** — `ON DELETE CASCADE` FKs mean `DeleteNote`
   is a single statement.
7. **Dead / stub code**: `Client.GenerateEmbedding` (unused), `openrouterEmbbeding*` types
   (unused), subscriptions return errors (declared in SDL but not implemented).
8. **AI tags on write mutations** — summarize/key-concepts happen synchronously inside
   createNote/updateNote, so those requests block on two extra LLM round-trips.

---

## 6. Notable Quirks / Potential Issues

- `config.getEnv("Port", …)` — the env key is `Port` (capitalized), while README/.env
  examples say `PORT`; on case-sensitive systems the default 8080 wins unless `Port` is set.
- `MemoryDB.GetNotesByTag` acquires `RLock` then defers `Unlock` (unlocks write lock):
  mismatched but harmless in practice — still worth fixing.
- `aiClient.ExtractKeyConcepts` parses `content` (response JSON) but `content` was shadowed...
  it unmarshal-falls back to `extractConceptsFromText`, so malformed JSON array is tolerated.
- `AIService.CheckRelationship` does not nil-check the notes returned by `GetNote`
  (nonexistent IDs would send nil-dereference content into the prompt).
- `MySQLDB.GetLinksByNote` scans `link.Description` directly (no NullString), unlike
  `GetBackLinksByNote` which handles NULL — NULL descriptions would error.
- `SemanticSearch` does one `GetAllNotes` full scan + per-note cache lookups per request;
  fine for a personal vault, not for large corpora.
- Package/file typos kept as-is for traceability: `embbeding_cache.go`,
  `aiService.go`, `openrouterEmbbedingRequest` (the code now actually uses Jina + Groq,
  but names still say "openrouter").
