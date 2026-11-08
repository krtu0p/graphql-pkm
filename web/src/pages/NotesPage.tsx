import { useMemo, useState } from 'react';
import { Button } from '../components/common/Button';
import { EmptyState } from '../components/common/EmptyState';
import { ErrorState } from '../components/common/ErrorState';
import { Input } from '../components/common/Input';
import { LoadingState } from '../components/common/LoadingState';
import { NoteList, type NoteLayoutView } from '../components/notes/NoteList';
import type { Note } from '../lib/graphql/types';
import { formatRelativeDate } from '../lib/utils/format';

export type NoteSortOption = 'updated-desc' | 'updated-asc' | 'title-asc' | 'created-desc';

export interface NotesPageProps {
  notes: Note[];
  activeNoteId: string | null;
  selectedTag: string | null;
  searchQuery: string;
  isLoading: boolean;
  errorMessage: string | null;
  onSearchQueryChange: (query: string) => void;
  onSelectTag: (tag: string | null) => void;
  onSelectNote: (noteId: string) => void;
  onCreateNote: () => void;
  onRequestDeleteNote: (note: Note) => void;
  onRetry: () => void;
}

export function NotesPage({
  notes,
  activeNoteId,
  selectedTag,
  searchQuery,
  isLoading,
  errorMessage,
  onSearchQueryChange,
  onSelectTag,
  onSelectNote,
  onCreateNote,
  onRequestDeleteNote,
  onRetry,
}: NotesPageProps) {
  const [sortBy, setSortBy] = useState<NoteSortOption>('updated-desc');
  const [viewMode, setViewMode] = useState<NoteLayoutView>('list');

  const allTags = useMemo(() => {
    const counts = new Map<string, number>();
    for (const note of notes) {
      for (const tag of note.tags) {
        counts.set(tag, (counts.get(tag) ?? 0) + 1);
      }
    }
    return Array.from(counts.entries())
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
      .map(([tag, count]) => ({ tag, count }));
  }, [notes]);

  const recentNotes = useMemo(() => {
    return [...notes]
      .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
      .slice(0, 3);
  }, [notes]);

  const filteredAndSortedNotes = useMemo(() => {
    const trimmedQuery = searchQuery.trim().toLowerCase();
    const filtered = notes.filter((note) => {
      if (selectedTag && !note.tags.includes(selectedTag)) {
        return false;
      }
      if (!trimmedQuery) {
        return true;
      }
      const inTitle = note.title.toLowerCase().includes(trimmedQuery);
      const inContent = note.content.toLowerCase().includes(trimmedQuery);
      const inTags = note.tags.some((t) => t.toLowerCase().includes(trimmedQuery));
      const inSummary = note.summary?.toLowerCase().includes(trimmedQuery) ?? false;
      return inTitle || inContent || inTags || inSummary;
    });

    return filtered.sort((a, b) => {
      if (sortBy === 'title-asc') {
        return a.title.localeCompare(b.title);
      }
      if (sortBy === 'updated-asc') {
        return new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime();
      }
      if (sortBy === 'created-desc') {
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
      }
      return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
    });
  }, [notes, selectedTag, searchQuery, sortBy]);

  const hasActiveFilter = Boolean(selectedTag || searchQuery.trim());

  return (
    <section className="notes-page" aria-labelledby="notes-page-heading">
      <div className="page-header">
        <div>
          <h1 id="notes-page-heading" className="page-title">
            All Notes
          </h1>
          <p className="page-description">
            Browse, filter, and organize your personal knowledge library.
          </p>
        </div>

        <Button variant="primary" onClick={onCreateNote}>
          <svg
            width="15"
            height="15"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            aria-hidden="true"
          >
            <path strokeLinecap="round" d="M10 4v12M4 10h12" />
          </svg>
          <span>New Note</span>
        </Button>
      </div>

      {errorMessage && (
        <ErrorState
          title="Could not load notes"
          message={errorMessage}
          onRetry={onRetry}
        />
      )}

      {!isLoading && !errorMessage && notes.length > 0 && !hasActiveFilter && (
        <section className="recent-notes-section" aria-labelledby="recent-notes-heading">
          <div className="section-subheader">
            <h2 id="recent-notes-heading" className="section-label">
              Recent Notes
            </h2>
          </div>
          <div className="recent-notes-row">
            {recentNotes.map((note) => (
              <button
                key={note.id}
                type="button"
                className="recent-note-chip"
                onClick={() => onSelectNote(note.id)}
              >
                <span className="recent-note-chip-title">
                  {note.title || 'Untitled Note'}
                </span>
                <span className="recent-note-chip-date">
                  Updated {formatRelativeDate(note.updatedAt)}
                </span>
              </button>
            ))}
          </div>
        </section>
      )}

      <div className="library-toolbar" role="region" aria-label="Filter and sort notes">
        <div className="library-search-box">
          <Input
            id="library-search-input"
            label="Filter notes"
            hideLabel
            type="search"
            value={searchQuery}
            placeholder="Filter notes by title, content, or tag…"
            onChange={(event) => onSearchQueryChange(event.target.value)}
            startIcon={
              <svg
                width="15"
                height="15"
                viewBox="0 0 20 20"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.6"
              >
                <circle cx="9" cy="9" r="5" />
                <path strokeLinecap="round" d="m13 13 3.5 3.5" />
              </svg>
            }
            endAction={
              searchQuery ? (
                <button
                  type="button"
                  className="icon-btn"
                  onClick={() => onSearchQueryChange('')}
                  aria-label="Clear search query"
                >
                  Clear
                </button>
              ) : undefined
            }
          />
        </div>

        <div className="library-controls">
          <div className="sort-control">
            <label htmlFor="notes-sort-select" className="sort-label">
              Sort
            </label>
            <select
              id="notes-sort-select"
              className="select-input"
              value={sortBy}
              onChange={(event) => setSortBy(event.target.value as NoteSortOption)}
            >
              <option value="updated-desc">Recently updated</option>
              <option value="updated-asc">Oldest updated</option>
              <option value="created-desc">Recently created</option>
              <option value="title-asc">Title (A–Z)</option>
            </select>
          </div>

          <div className="view-toggle" role="group" aria-label="Note layout view">
            <button
              type="button"
              className={`icon-btn ${viewMode === 'list' ? 'is-active' : ''}`}
              onClick={() => setViewMode('list')}
              aria-pressed={viewMode === 'list'}
            >
              List
            </button>
            <button
              type="button"
              className={`icon-btn ${viewMode === 'grid' ? 'is-active' : ''}`}
              onClick={() => setViewMode('grid')}
              aria-pressed={viewMode === 'grid'}
            >
              Grid
            </button>
          </div>
        </div>
      </div>

      {allTags.length > 0 && (
        <div className="tag-filter-bar" role="region" aria-label="Filter by tag">
          <button
            type="button"
            className={`tag-pill tag-pill-interactive ${selectedTag === null ? 'is-selected' : ''}`}
            onClick={() => onSelectTag(null)}
            aria-pressed={selectedTag === null}
          >
            All ({notes.length})
          </button>
          {allTags.map(({ tag, count }) => {
            const isSelected = selectedTag === tag;
            return (
              <button
                key={tag}
                type="button"
                className={`tag-pill tag-pill-interactive ${isSelected ? 'is-selected' : ''}`}
                onClick={() => onSelectTag(isSelected ? null : tag)}
                aria-pressed={isSelected}
              >
                #{tag} ({count})
              </button>
            );
          })}
        </div>
      )}

      {isLoading ? (
        <LoadingState label="Loading your note library…" count={3} />
      ) : filteredAndSortedNotes.length === 0 ? (
        hasActiveFilter ? (
          <EmptyState
            title="No matching notes"
            description="No notes matched your current search or tag filter."
            action={
              <Button
                variant="secondary"
                onClick={() => {
                  onSelectTag(null);
                  onSearchQueryChange('');
                }}
              >
                Reset Filters
              </Button>
            }
          />
        ) : (
          <EmptyState
            title="Your library is empty"
            description="Begin your personal knowledge base by writing your first note."
            action={
              <Button variant="primary" onClick={onCreateNote}>
                Create Note
              </Button>
            }
          />
        )
      ) : (
        <NoteList
          notes={filteredAndSortedNotes}
          activeNoteId={activeNoteId}
          selectedTag={selectedTag}
          viewMode={viewMode}
          onSelectNote={onSelectNote}
          onSelectTag={(tag) => onSelectTag(selectedTag === tag ? null : tag)}
          onRequestDelete={onRequestDeleteNote}
        />
      )}
    </section>
  );
}
