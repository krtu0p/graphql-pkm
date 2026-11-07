import type { Note, SearchResult } from '../../lib/graphql/types';
import { createExcerpt, formatRelativeDate } from '../../lib/utils/format';
import { EmptyState } from '../common/EmptyState';
import { NoteList } from '../notes/NoteList';
import { TagList } from '../notes/TagList';

export interface SearchResultsProps {
  mode: 'keyword' | 'semantic' | 'hybrid';
  submittedQuery: string;
  keywordNotes: Note[];
  rankedResults: SearchResult[];
  activeNoteId: string | null;
  onSelectNote: (noteId: string) => void;
  onSelectTag: (tag: string) => void;
}

function formatScore(score: number): string {
  const clamped = Math.max(0, Math.min(1, score));
  return `${Math.round(clamped * 100)}% relevance`;
}

export function SearchResults({
  mode,
  submittedQuery,
  keywordNotes,
  rankedResults,
  activeNoteId,
  onSelectNote,
  onSelectTag,
}: SearchResultsProps) {
  if (mode === 'keyword') {
    if (keywordNotes.length === 0) {
      return (
        <EmptyState
          title="No matching notes found"
          description={`No notes matched "${submittedQuery}" using keyword search.`}
        />
      );
    }

    return (
      <div className="search-results-section">
        <div className="section-subheader">
          <h2 className="section-label">
            Keyword Matches ({keywordNotes.length})
          </h2>
        </div>
        <NoteList
          notes={keywordNotes}
          activeNoteId={activeNoteId}
          onSelectNote={onSelectNote}
          onSelectTag={onSelectTag}
        />
      </div>
    );
  }

  if (rankedResults.length === 0) {
    return (
      <EmptyState
        title="No related notes found"
        description={`No notes met the similarity threshold for "${submittedQuery}". Try rephrasing or switching to Keyword mode.`}
      />
    );
  }

  return (
    <div className="search-results-section">
      <div className="section-subheader">
        <h2 className="section-label">
          {mode === 'semantic' ? 'Semantic Matches' : 'Hybrid Matches'} (
          {rankedResults.length})
        </h2>
      </div>

      <div className="ranked-results-list" role="list">
        {rankedResults.map(({ note, score, reason, matchType }) => (
          <article
            key={note.id}
            role="listitem"
            className={`note-card ${activeNoteId === note.id ? 'is-active' : ''}`}
          >
            <div className="note-card-header">
              <button
                type="button"
                className="note-card-title-btn"
                onClick={() => onSelectNote(note.id)}
              >
                <h3 className="note-card-title">
                  {note.title || 'Untitled Note'}
                </h3>
              </button>

              <div className="note-card-meta-right">
                <span className="tag-pill">{matchType || mode}</span>
                <span className="search-score-pill">{formatScore(score)}</span>
              </div>
            </div>

            {reason && (
              <p className="search-match-reason">
                <strong>Match insight:</strong> {reason}
              </p>
            )}

            <div
              className="note-card-body"
              onClick={() => onSelectNote(note.id)}
              role="presentation"
            >
              <p className="note-card-excerpt">{createExcerpt(note.content, 170)}</p>
            </div>

            <div className="ranked-card-footer">
              <TagList tags={note.tags} onSelectTag={onSelectTag} />
              <time className="note-card-timestamp" dateTime={note.updatedAt}>
                Updated {formatRelativeDate(note.updatedAt)}
              </time>
            </div>
          </article>
        ))}
      </div>
    </div>
  );
}
