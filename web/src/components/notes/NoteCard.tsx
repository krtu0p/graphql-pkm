import type { Note } from '../../lib/graphql/types';
import { createExcerpt, formatRelativeDate } from '../../lib/utils/format';
import { TagList } from './TagList';

export interface NoteCardProps {
  note: Note;
  isActive?: boolean;
  selectedTag?: string | null;
  onSelectNote: (noteId: string) => void;
  onSelectTag?: (tag: string) => void;
  onRequestDelete?: (note: Note) => void;
}

export function NoteCard({
  note,
  isActive = false,
  selectedTag = null,
  onSelectNote,
  onSelectTag,
  onRequestDelete,
}: NoteCardProps) {
  const excerpt = createExcerpt(note.content, 160);

  return (
    <article
      className={`note-card ${isActive ? 'is-active' : ''}`}
      aria-label={note.title || 'Untitled Note'}
    >
      <div className="note-card-header">
        <button
          type="button"
          className="note-card-title-btn"
          onClick={() => onSelectNote(note.id)}
        >
          <h3 className="note-card-title">{note.title || 'Untitled Note'}</h3>
        </button>

        <div className="note-card-meta-right">
          <time className="note-card-timestamp" dateTime={note.updatedAt}>
            {formatRelativeDate(note.updatedAt)}
          </time>
          {onRequestDelete && (
            <button
              type="button"
              className="icon-btn note-card-delete-btn"
              onClick={(event) => {
                event.stopPropagation();
                onRequestDelete(note);
              }}
              aria-label={`Delete note ${note.title}`}
              title="Delete note"
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 20 20"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.6"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M4.5 6h11M8 6V4.5a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1V6m1.5 0v9.5a1.5 1.5 0 0 1-1.5 1.5h-5A1.5 1.5 0 0 1 6 15.5V6"
                />
              </svg>
            </button>
          )}
        </div>
      </div>

      <div
        className="note-card-body"
        onClick={() => onSelectNote(note.id)}
        role="presentation"
      >
        <p className="note-card-excerpt">{excerpt}</p>

        {note.summary && (
          <div className="note-card-summary">
            <span className="note-card-summary-label">AI Summary:</span>{' '}
            <span>{createExcerpt(note.summary, 130)}</span>
          </div>
        )}
      </div>

      {note.tags.length > 0 && (
        <div className="note-card-footer">
          <TagList
            tags={note.tags}
            selectedTag={selectedTag}
            onSelectTag={onSelectTag}
          />
        </div>
      )}
    </article>
  );
}
