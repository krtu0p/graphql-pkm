import { useEffect, useMemo, useState } from 'react';
import { Button } from '../components/common/Button';
import { EmptyState } from '../components/common/EmptyState';
import { ErrorState } from '../components/common/ErrorState';
import { Input } from '../components/common/Input';
import { LoadingState } from '../components/common/LoadingState';
import { NoteList } from '../components/notes/NoteList';
import { fetchNotesByTag, type Note } from '../lib/graphql';

export interface TagsPageProps {
  notes: Note[];
  selectedTag: string | null;
  activeNoteId: string | null;
  isLoadingLibrary: boolean;
  onSelectTag: (tag: string | null) => void;
  onSelectNote: (noteId: string) => void;
  onCreateNote: () => void;
  onRequestDeleteNote: (note: Note) => void;
}

export interface TagSummary {
  tag: string;
  count: number;
  latestUpdatedAt: string;
}

export function TagsPage({
  notes,
  selectedTag,
  activeNoteId,
  isLoadingLibrary,
  onSelectTag,
  onSelectNote,
  onCreateNote,
  onRequestDeleteNote,
}: TagsPageProps) {
  const [tagFilterQuery, setTagFilterQuery] = useState('');
  const [taggedNotes, setTaggedNotes] = useState<Note[]>([]);
  const [isLoadingTagNotes, setIsLoadingTagNotes] = useState(false);
  const [tagError, setTagError] = useState<string | null>(null);
  const [reloadCounter, setReloadCounter] = useState(0);

  const tagSummaries = useMemo<TagSummary[]>(() => {
    const map = new Map<string, { count: number; latestUpdatedAt: string }>();
    for (const note of notes) {
      for (const tag of note.tags) {
        const existing = map.get(tag);
        if (!existing) {
          map.set(tag, { count: 1, latestUpdatedAt: note.updatedAt });
        } else {
          const isNewer =
            new Date(note.updatedAt).getTime() >
            new Date(existing.latestUpdatedAt).getTime();
          map.set(tag, {
            count: existing.count + 1,
            latestUpdatedAt: isNewer ? note.updatedAt : existing.latestUpdatedAt,
          });
        }
      }
    }

    return Array.from(map.entries())
      .map(([tag, info]) => ({
        tag,
        count: info.count,
        latestUpdatedAt: info.latestUpdatedAt,
      }))
      .sort((a, b) => b.count - a.count || a.tag.localeCompare(b.tag));
  }, [notes]);

  const filteredTagSummaries = useMemo(() => {
    const q = tagFilterQuery.trim().toLowerCase();
    if (!q) return tagSummaries;
    return tagSummaries.filter((item) => item.tag.toLowerCase().includes(q));
  }, [tagSummaries, tagFilterQuery]);

  const activeTag = useMemo(() => {
    if (selectedTag && tagSummaries.some((t) => t.tag === selectedTag)) {
      return selectedTag;
    }
    return selectedTag;
  }, [selectedTag, tagSummaries]);

  useEffect(() => {
    if (!activeTag) {
      setTaggedNotes([]);
      setTagError(null);
      setIsLoadingTagNotes(false);
      return;
    }

    const controller = new AbortController();
    setIsLoadingTagNotes(true);
    setTagError(null);

    fetchNotesByTag(activeTag, controller.signal)
      .then((results) => {
        setTaggedNotes(results);
      })
      .catch((err) => {
        if (err instanceof DOMException && err.name === 'AbortError') return;
        setTagError(
          err instanceof Error
            ? err.message
            : `Failed to fetch notes for tag #${activeTag}.`,
        );
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoadingTagNotes(false);
        }
      });

    return () => controller.abort();
  }, [activeTag, notes, reloadCounter]);

  return (
    <section className="tags-page" aria-labelledby="tags-page-heading">
      <div className="page-header">
        <div>
          <h1 id="tags-page-heading" className="page-title">
            Tags
          </h1>
          <p className="page-description">
            Explore notes categorized by topic tags across your knowledge base.
          </p>
        </div>

        <Button variant="primary" onClick={onCreateNote}>
          New Note
        </Button>
      </div>

      {isLoadingLibrary ? (
        <LoadingState label="Loading tags…" count={2} />
      ) : tagSummaries.length === 0 ? (
        <EmptyState
          title="No tags created yet"
          description="Add tags while writing or editing a note to organize your library by topic."
          action={
            <Button variant="primary" onClick={onCreateNote}>
              Create Note
            </Button>
          }
        />
      ) : (
        <div className="tags-workspace-layout">
          <div className="tags-directory-panel" aria-label="Tag directory">
            <div className="tags-directory-header">
              <h2 className="section-label">
                All Tags ({tagSummaries.length})
              </h2>
              <Input
                id="tag-directory-search"
                label="Filter tags"
                hideLabel
                type="search"
                value={tagFilterQuery}
                placeholder="Find a tag…"
                onChange={(event) => setTagFilterQuery(event.target.value)}
              />
            </div>

            {filteredTagSummaries.length === 0 ? (
              <p className="insight-placeholder-text">
                No tags match &ldquo;{tagFilterQuery}&rdquo;.
              </p>
            ) : (
              <ul className="tags-catalog-grid" role="list" aria-label="Available tags">
                {filteredTagSummaries.map(({ tag, count }) => {
                  const isSelected = activeTag === tag;
                  return (
                    <li key={tag} className="tags-catalog-item">
                      <button
                        type="button"
                        className={`tag-catalog-card ${isSelected ? 'is-selected' : ''}`}
                        onClick={() => onSelectTag(isSelected ? null : tag)}
                        aria-pressed={isSelected}
                      >
                        <span className="tag-catalog-name">#{tag}</span>
                        <span className="tag-catalog-count">
                          {count} {count === 1 ? 'note' : 'notes'}
                        </span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>

          <div className="tags-results-panel" aria-live="polite">
            {!activeTag ? (
              <EmptyState
                title="Select a tag to view notes"
                description="Choose any tag above to inspect all notes categorized under that topic."
              />
            ) : (
              <div className="tagged-notes-section">
                <div className="tagged-notes-header">
                  <div>
                    <h2 className="tagged-notes-title">
                      Notes tagged <span className="tag-pill">#{activeTag}</span>
                    </h2>
                  </div>
                  <Button
                    variant="secondary"
                    onClick={() => onSelectTag(null)}
                  >
                    Clear Tag Filter
                  </Button>
                </div>

                {tagError && (
                  <ErrorState
                    title={`Could not load notes for #${activeTag}`}
                    message={tagError}
                    onRetry={() => setReloadCounter((c) => c + 1)}
                  />
                )}

                {isLoadingTagNotes ? (
                  <LoadingState label={`Loading notes tagged #${activeTag}…`} count={2} />
                ) : taggedNotes.length === 0 ? (
                  <EmptyState
                    title={`No notes found with #${activeTag}`}
                    description="There are currently no notes assigned to this tag."
                    action={
                      <Button variant="secondary" onClick={() => onSelectTag(null)}>
                        Show All Tags
                      </Button>
                    }
                  />
                ) : (
                  <NoteList
                    notes={taggedNotes}
                    activeNoteId={activeNoteId}
                    selectedTag={activeTag}
                    onSelectNote={onSelectNote}
                    onSelectTag={(clickedTag) =>
                      onSelectTag(clickedTag === activeTag ? null : clickedTag)
                    }
                    onRequestDelete={onRequestDeleteNote}
                  />
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </section>
  );
}
