# Task: Migrate graphql-pkm to local-first

Read this whole file before starting. Do ONE phase at a time. After each phase: run the
checks, update the Status section below, summarize what changed in a few lines, then STOP and
wait for me to say "next".

## Goal

Make the app run with zero setup on one machine, with all data in a local file and AI optional:

- Keep the Go backend and the GraphQL API unchanged (no breaking schema changes).
- Replace MySQL with SQLite as the default storage. MySQL support stays as an option.
- AI is optional. Default AI provider is a local Ollama server, not Groq/Jina.
- App must start with no `.env` file and no running database server.
- Bind to localhost by default.

## Rules (important)

- Minimal diffs. Do not refactor unrelated code or rename existing files/typos.
- Do NOT hand-edit `internal/gql/generated/generated.go`. Only regenerate it if
  `schema.graphqls` or `gqlgen.yml` changes: `go run github.com/99designs/gqlgen generate`.
  Avoid schema changes unless a phase says otherwise.
- Do not change service or resolver code unless the phase says so. The `database.Database` and
  `database.EmbeddingsCache` interfaces are the boundary; new backends implement them.
- Read only the files you need. Do not scan the whole repo.
- After every phase run: `go build ./... && go vet ./... && go test ./...` and fix failures.
- Never commit secrets. Do not touch the user's real `.env`.
- If something is unclear or the code differs from this file, stop and ask.

## Codebase summary (for context)

Layers (calls only go downward): `resolvers -> services -> (database | ai)`.

- `cmd/server/main.go`: composition root. Loads config, opens MySQL, builds `NoteService`,
  `AIService`, `SearchService`, resolvers, mounts `/` (playground) and `/query`.
- `config/config.go`: env vars to `Config`.
- `internal/models/models.go`: shared types, bound to GraphQL by gqlgen autobind.
- `internal/database/`: `interface.go` (Database), `mysql.go` (creates tables on startup),
  `memory.go` (in-memory impl, unused by main), `embeddings_interface.go`,
  `embbeding_cache.go` (MySQL cache: note_id -> JSON vector, wraps MySQLDB's `*sql.DB`).
- `internal/ai/`: `client.go` (Groq chat + Jina embeddings, hard-coded), `types.go`, `search.go`.
  `IsEnabled()` currently means `apiKey != ""`.
- `internal/service/`: `noteService.go` (CRUD, links, tag normalization, calls AI for
  summary/key concepts synchronously on create/update), `aiService.go` (semantic, hybrid,
  smart search, cosine similarity, threshold 0.6), `searchService.go` (facade).
- Deleting a note cascades to `links` and `embeddings_cache` via `ON DELETE CASCADE`
  foreign keys (not Go code). SQLite must have foreign keys enabled for this to work.

## Phase 1: SQLite backend

Add `internal/database/sqlite.go` and `internal/database/sqlite_embeddings.go`.

- Use `modernc.org/sqlite` (pure Go, no CGO). Driver name `"sqlite"`.
- DSN must enable: foreign keys, WAL, busy timeout. Example:
  `file:<path>?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)`
- Create the parent directory of the DB file if missing.
- Implement every method of `database.Database` with the same behavior as `mysql.go`
  (including 404/not-found errors, e.g. `DeleteLink` when 0 rows affected).
- Translate MySQL-specific SQL:
  - `JSON_CONTAINS(tags, '"x"')` -> `EXISTS (SELECT 1 FROM json_each(notes.tags) WHERE value = ?)`
  - `INSERT ... ON DUPLICATE KEY UPDATE` -> `INSERT ... ON CONFLICT(note_id) DO UPDATE SET ... = excluded....`
  - Table DDL: remove `AUTO_INCREMENT`, store JSON columns as `TEXT`, declare time columns as
    `DATETIME`, keep the foreign keys with `ON DELETE CASCADE`.
  - Keep search as case-insensitive `LIKE` on title/content.
- Implement `EmbeddingsCache` for SQLite (Store, Get, Delete, Clear, Count, GetCacheStats),
  sharing the same `*sql.DB` handle like the MySQL version does.
- Add compile-time checks: `var _ Database = (*SQLiteDB)(nil)` etc.
- While in `mysql.go`: make `GetLinksByNote` handle NULL `description` (use `sql.NullString`).
- Tests: write one shared contract test suite for the `Database` interface and run it against
  `MemoryDB` and `SQLiteDB` (use `t.TempDir()` for the SQLite file). Must cover: create/get/
  update/delete note, tags filter, search, links + backlinks, self-link behavior at DB level,
  and cascade delete (deleting a note removes its links and embeddings).

Done when: tests pass for both backends and the cascade test passes on SQLite.

## Phase 2: Config and wiring

- `config/config.go`: add `DB_DRIVER` (`sqlite` default, `mysql` optional), `DB_PATH`
  (default `./data/pkm.db`), `HOST` (default `127.0.0.1`), and fix the port key so `PORT`
  works (currently reads `Port`; accept both). Keep `DATABASE_URL` for MySQL.
- `cmd/server/main.go`: choose backend from `DB_DRIVER`. Everything else stays the same.
- No required env vars. App must start with no `.env`.
- Add `.env.example` documenting all variables. Add `data/` to `.gitignore`.
- Update README with a short "Run locally" section.

Done when: `go run ./cmd/server` with no `.env` starts, creates `./data/pkm.db`, and a
`createNote` + `notes` + `deleteNote` round trip works via the GraphQL playground on localhost.

## Phase 3: Local AI (Ollama) and optional AI

- Make `ai.Client` provider-agnostic, using OpenAI-compatible endpoints:
  - Chat: `{AI_BASE_URL}/chat/completions`, default base `http://localhost:11434/v1`.
  - Embeddings: `{EMBED_BASE_URL}/embeddings` (defaults to `AI_BASE_URL`).
  - New env vars: `AI_BASE_URL`, `AI_API_KEY` (optional, empty is fine for Ollama),
    `AI_MODEL`, `EMBED_MODEL`, `AI_ENABLED` (default true if `AI_BASE_URL` is set).
  - Remove the hard-coded Groq/Jina URLs and the Jina-specific request builder; reuse the
    shared request helper. Keep Groq/Jina working by just setting the env vars.
- `IsEnabled()` must not require an API key. If AI is disabled or the server is unreachable,
  every feature must degrade gracefully (keyword search fallback, note CRUD still works).
- Embedding cache: add a `model` column (and dimension if useful). When the configured
  `EMBED_MODEL` differs from a cached row's model, treat it as a miss and recompute.
  Old Jina vectors (1024 dims) must never be compared with new vectors.
- Make the semantic similarity threshold configurable (`SEMANTIC_THRESHOLD`, default 0.6) and
  note in the README that it needs tuning per embedding model.
- Async enrichment: `CreateNote`/`UpdateNote` must save the note first and return quickly.
  Generate summary and key concepts in a background goroutine, then update the note. Use a
  bounded worker (single goroutine + channel is fine), with timeouts and error logging. Do not
  change the GraphQL schema; the fields stay nullable/empty until enrichment finishes.
- Local models return messy JSON: keep `cleanAIContent`/`extractJSON` and the concept
  fallback parser, use low temperature, and set a request timeout.
- Fix while here: `AIService.CheckRelationship` must nil-check notes before building the prompt.

Done when: with Ollama running, creating a note returns immediately and gets a summary a few
seconds later; with Ollama stopped, all non-AI operations still work and search falls back to
keyword search with no panics.

## Phase 4: Data migration tool (one-off)

- Add `cmd/migrate/main.go`: read all notes and links from MySQL (`DATABASE_URL`) through the
  existing `Database` interface and write them to SQLite (`DB_PATH`), preserving IDs and
  timestamps. Skip embeddings (they get recomputed).
- Print counts (notes, links) read and written, and verify they match.
- Must be idempotent or refuse to run into a non-empty SQLite file unless `--force` is passed.

Done when: counts match on a test dataset.

## Phase 5 (do NOT start unless I ask): desktop wrapper

Embed the UI with `go:embed` and ship a single binary; consider Wails or opening the browser.

## Small fixes to include in the relevant phase (low priority)

- `MemoryDB.GetNotesByTag` uses `RLock` but a write `Unlock` defer; fix to `RUnlock`.
- Unused code (`Client.GenerateEmbedding`, `openrouterEmbbeding*` types): leave alone unless
  they get in the way.

## Status

- [x] Phase 1: SQLite backend
- [x] Phase 2: Config and wiring
- [x] Phase 3: Local AI and optional AI
- [ ] Phase 4: Data migration tool
- [ ] Phase 5: Desktop wrapper (on hold)
