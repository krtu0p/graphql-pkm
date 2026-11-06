package database

import (
	"graphql-pkm/internal/models"
	"path/filepath"
	"testing"
	"time"
)

func TestDatabaseContract(t *testing.T) {
	backends := []struct {
		name  string
		newDB func(t *testing.T) Database
	}{
		{
			name: "MemoryDB",
			newDB: func(t *testing.T) Database {
				return NewMemoryDB()
			},
		},
		{
			name: "SQLiteDB",
			newDB: func(t *testing.T) Database {
				dbPath := filepath.Join(t.TempDir(), "nested", "test.db")
				db, err := NewSQLiteDB(dbPath)
				if err != nil {
					t.Fatalf("NewSQLiteDB failed: %v", err)
				}
				t.Cleanup(func() {
					_ = db.Close()
				})
				return db
			},
		},
	}

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			runDatabaseContractSuite(t, b.newDB)
		})
	}
}

func runDatabaseContractSuite(t *testing.T, newDB func(t *testing.T) Database) {
	t.Run("NoteCRUD", func(t *testing.T) {
		db := newDB(t)

		missing, err := db.GetNote("does-not-exist")
		if err != nil {
			t.Fatalf("GetNote missing returned error: %v", err)
		}
		if missing != nil {
			t.Fatalf("expected nil for missing note, got %+v", missing)
		}

		now := time.Now().UTC().Truncate(time.Second)
		summary := "initial summary"
		note := &models.Note{
			ID:          "note-1",
			Title:       "First Note",
			Content:     "Hello world content",
			Tags:        []string{"go", "sqlite"},
			CreatedAt:   now,
			UpdatedAt:   now,
			Summary:     &summary,
			KeyConcepts: []string{"database", "testing"},
		}

		if err := db.CreateNote(note); err != nil {
			t.Fatalf("CreateNote failed: %v", err)
		}

		got, err := db.GetNote("note-1")
		if err != nil {
			t.Fatalf("GetNote failed: %v", err)
		}
		if got == nil {
			t.Fatalf("expected note, got nil")
		}
		if got.Title != note.Title || got.Content != note.Content {
			t.Fatalf("unexpected note fields: %+v", got)
		}
		if got.Summary == nil || *got.Summary != summary {
			t.Fatalf("unexpected summary: %v", got.Summary)
		}
		if len(got.Tags) != 2 || got.Tags[0] != "go" || got.Tags[1] != "sqlite" {
			t.Fatalf("unexpected tags: %v", got.Tags)
		}
		if len(got.KeyConcepts) != 2 {
			t.Fatalf("unexpected key concepts: %v", got.KeyConcepts)
		}

		all, err := db.GetAllNotes()
		if err != nil {
			t.Fatalf("GetAllNotes failed: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("expected 1 note, got %d", len(all))
		}

		updatedSummary := "updated summary"
		updatedNote := &models.Note{
			ID:          "note-1",
			Title:       "Updated Title",
			Content:     "Updated content",
			Tags:        []string{"go", "updated"},
			CreatedAt:   now,
			UpdatedAt:   now.Add(time.Minute),
			Summary:     &updatedSummary,
			KeyConcepts: []string{"updated-concept"},
		}
		if err := db.UpdateNote(updatedNote); err != nil {
			t.Fatalf("UpdateNote failed: %v", err)
		}

		gotAfterUpdate, err := db.GetNote("note-1")
		if err != nil {
			t.Fatalf("GetNote after update failed: %v", err)
		}
		if gotAfterUpdate.Title != "Updated Title" || gotAfterUpdate.Content != "Updated content" {
			t.Fatalf("note was not updated: %+v", gotAfterUpdate)
		}
		if len(gotAfterUpdate.Tags) != 2 || gotAfterUpdate.Tags[1] != "updated" {
			t.Fatalf("tags were not updated: %v", gotAfterUpdate.Tags)
		}

		if err := db.DeleteNote("note-1"); err != nil {
			t.Fatalf("DeleteNote failed: %v", err)
		}

		deleted, err := db.GetNote("note-1")
		if err != nil {
			t.Fatalf("GetNote after delete failed: %v", err)
		}
		if deleted != nil {
			t.Fatalf("expected nil after delete, got %+v", deleted)
		}
	})

	t.Run("TagsAndSearch", func(t *testing.T) {
		db := newDB(t)
		now := time.Now().UTC().Truncate(time.Second)

		n1 := &models.Note{
			ID:          "n1",
			Title:       "Golang Concurrency",
			Content:     "Channels and goroutines in practice",
			Tags:        []string{"golang", "backend"},
			CreatedAt:   now,
			UpdatedAt:   now,
			KeyConcepts: []string{},
		}
		n2 := &models.Note{
			ID:          "n2",
			Title:       "SQLite Local First",
			Content:     "Embedded relational storage with WAL mode",
			Tags:        []string{"sqlite", "backend"},
			CreatedAt:   now.Add(time.Second),
			UpdatedAt:   now.Add(time.Second),
			KeyConcepts: []string{},
		}

		if err := db.CreateNote(n1); err != nil {
			t.Fatalf("CreateNote n1: %v", err)
		}
		if err := db.CreateNote(n2); err != nil {
			t.Fatalf("CreateNote n2: %v", err)
		}

		byBackend, err := db.GetNotesByTag("backend")
		if err != nil {
			t.Fatalf("GetNotesByTag(backend): %v", err)
		}
		if len(byBackend) != 2 {
			t.Fatalf("expected 2 notes with tag backend, got %d", len(byBackend))
		}

		bySQLite, err := db.GetNotesByTag("sqlite")
		if err != nil {
			t.Fatalf("GetNotesByTag(sqlite): %v", err)
		}
		if len(bySQLite) != 1 || bySQLite[0].ID != "n2" {
			t.Fatalf("expected note n2 for tag sqlite, got %+v", bySQLite)
		}

		byMissing, err := db.GetNotesByTag("nonexistent")
		if err != nil {
			t.Fatalf("GetNotesByTag(nonexistent): %v", err)
		}
		if len(byMissing) != 0 {
			t.Fatalf("expected 0 notes for nonexistent tag, got %d", len(byMissing))
		}

		// Case-insensitive search on title
		titleHits, err := db.SearchNotes("gOlAnG")
		if err != nil {
			t.Fatalf("SearchNotes title: %v", err)
		}
		if len(titleHits) != 1 || titleHits[0].ID != "n1" {
			t.Fatalf("expected n1 from title search, got %+v", titleHits)
		}

		// Case-insensitive search on content
		contentHits, err := db.SearchNotes("wAl MoDe")
		if err != nil {
			t.Fatalf("SearchNotes content: %v", err)
		}
		if len(contentHits) != 1 || contentHits[0].ID != "n2" {
			t.Fatalf("expected n2 from content search, got %+v", contentHits)
		}
	})

	t.Run("LinksBacklinksAndSelfLink", func(t *testing.T) {
		db := newDB(t)
		now := time.Now().UTC().Truncate(time.Second)

		for _, id := range []string{"a", "b"} {
			err := db.CreateNote(&models.Note{
				ID:        id,
				Title:     "Note " + id,
				Content:   "Content " + id,
				Tags:      []string{},
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				t.Fatalf("CreateNote %s: %v", id, err)
			}
		}

		desc := "references"
		linkAB := &models.Link{
			ID:           "link-ab",
			SourceNoteID: "a",
			TargetNoteID: "b",
			Description:  &desc,
			CreatedAt:    now,
		}
		if err := db.CreateLink(linkAB); err != nil {
			t.Fatalf("CreateLink link-ab: %v", err)
		}

		// Link with nil description + self-link at DB level
		selfLink := &models.Link{
			ID:           "link-aa",
			SourceNoteID: "a",
			TargetNoteID: "a",
			Description:  nil,
			CreatedAt:    now.Add(time.Second),
		}
		if err := db.CreateLink(selfLink); err != nil {
			t.Fatalf("CreateLink self-link at DB level failed: %v", err)
		}

		aLinks, err := db.GetLinksByNote("a")
		if err != nil {
			t.Fatalf("GetLinksByNote(a): %v", err)
		}
		if len(aLinks) != 2 {
			t.Fatalf("expected 2 outgoing links from a, got %d", len(aLinks))
		}

		bBacklinks, err := db.GetBackLinksByNote("b")
		if err != nil {
			t.Fatalf("GetBackLinksByNote(b): %v", err)
		}
		if len(bBacklinks) != 1 || bBacklinks[0].ID != "link-ab" {
			t.Fatalf("expected 1 backlink for b, got %+v", bBacklinks)
		}
		if bBacklinks[0].Description == nil || *bBacklinks[0].Description != "references" {
			t.Fatalf("expected backlink description 'references', got %v", bBacklinks[0].Description)
		}

		aBacklinks, err := db.GetBackLinksByNote("a")
		if err != nil {
			t.Fatalf("GetBackLinksByNote(a): %v", err)
		}
		if len(aBacklinks) != 1 || aBacklinks[0].ID != "link-aa" {
			t.Fatalf("expected self-link in backlinks of a, got %+v", aBacklinks)
		}

		if err := db.DeleteLink("link-ab"); err != nil {
			t.Fatalf("DeleteLink(link-ab): %v", err)
		}

		// Deleting again should return not-found error
		if err := db.DeleteLink("link-ab"); err == nil {
			t.Fatalf("expected error deleting non-existent link, got nil")
		}
	})
}

func TestSQLiteCascadeAndEmbeddings(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "cascade.db")
	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteDB failed: %v", err)
	}
	defer db.Close()

	cache := NewSQLiteEmbeddingsCache(db)

	now := time.Now().UTC().Truncate(time.Second)
	for _, id := range []string{"n1", "n2"} {
		if err := db.CreateNote(&models.Note{
			ID:        id,
			Title:     "Title " + id,
			Content:   "Content " + id,
			Tags:      []string{"test"},
			CreatedAt: now,
			UpdatedAt: now,
		}); err != nil {
			t.Fatalf("CreateNote %s: %v", id, err)
		}
	}

	if err := db.CreateLink(&models.Link{
		ID:           "l1",
		SourceNoteID: "n1",
		TargetNoteID: "n2",
		CreatedAt:    now,
	}); err != nil {
		t.Fatalf("CreateLink: %v", err)
	}

	if err := cache.Store("n1", []float32{0.1, 0.2, 0.3}); err != nil {
		t.Fatalf("cache.Store initial: %v", err)
	}
	// Upsert test
	if err := cache.Store("n1", []float32{0.4, 0.5, 0.6}); err != nil {
		t.Fatalf("cache.Store upsert: %v", err)
	}
	if err := cache.Store("n2", []float32{0.7, 0.8, 0.9}); err != nil {
		t.Fatalf("cache.Store n2: %v", err)
	}

	vec, ok := cache.Get("n1")
	if !ok || len(vec) != 3 || vec[0] != 0.4 {
		t.Fatalf("unexpected cached embedding for n1: ok=%v vec=%v", ok, vec)
	}

	count, err := cache.Count()
	if err != nil || count != 2 {
		t.Fatalf("expected cache count 2, got %d (err=%v)", count, err)
	}

	stats, err := cache.GetCacheStats()
	if err != nil {
		t.Fatalf("GetCacheStats failed: %v", err)
	}
	if stats["total_embeddings"] != 2 {
		t.Fatalf("expected total_embeddings=2, got %+v", stats)
	}

	// Delete n1 -> should cascade delete link l1 and n1's embedding
	if err := db.DeleteNote("n1"); err != nil {
		t.Fatalf("DeleteNote(n1): %v", err)
	}

	backlinks, err := db.GetBackLinksByNote("n2")
	if err != nil {
		t.Fatalf("GetBackLinksByNote(n2): %v", err)
	}
	if len(backlinks) != 0 {
		t.Fatalf("expected link l1 to be cascade-deleted, got %d backlinks", len(backlinks))
	}

	if _, ok := cache.Get("n1"); ok {
		t.Fatalf("expected n1 embedding to be cascade-deleted")
	}

	countAfter, err := cache.Count()
	if err != nil || countAfter != 1 {
		t.Fatalf("expected 1 embedding remaining after cascade delete, got %d (err=%v)", countAfter, err)
	}

	if err := cache.Delete("n2"); err != nil {
		t.Fatalf("cache.Delete(n2): %v", err)
	}
	if err := cache.Clear(); err != nil {
		t.Fatalf("cache.Clear(): %v", err)
	}

	// Model mismatch test: old model vector must be treated as a cache miss when model changes
	jinaCache := NewSQLiteEmbeddingsCache(db, "jina-embeddings-v3")
	if err := jinaCache.Store("n2", []float32{0.1, 0.2, 0.3, 0.4}); err != nil {
		t.Fatalf("jinaCache.Store: %v", err)
	}
	if _, ok := jinaCache.Get("n2"); !ok {
		t.Fatalf("expected cache hit for same model")
	}

	ollamaCache := NewSQLiteEmbeddingsCache(db, "nomic-embed-text")
	if _, ok := ollamaCache.Get("n2"); ok {
		t.Fatalf("expected cache miss when configured EMBED_MODEL differs from cached model")
	}
}
