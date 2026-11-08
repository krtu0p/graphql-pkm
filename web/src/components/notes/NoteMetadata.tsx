import { formatDateTime } from '../../lib/utils/format';

export type EditorSaveState = 'saved' | 'saving' | 'unsaved' | 'error' | 'new';

export interface NoteMetadataProps {
  updatedAt: string | null;
  createdAt: string | null;
  saveState: EditorSaveState;
  saveErrorMessage: string | null;
  wordCount: number;
}

export function NoteMetadata({
  updatedAt,
  createdAt,
  saveState,
  saveErrorMessage,
  wordCount,
}: NoteMetadataProps) {
  const renderSaveBadge = () => {
    if (saveState === 'saving') {
      return (
        <span className="save-status-badge is-saving" role="status" aria-live="polite">
          <span className="status-dot" aria-hidden="true" />
          <span>Saving…</span>
        </span>
      );
    }
    if (saveState === 'error') {
      return (
        <span
          className="save-status-badge is-error"
          role="alert"
          title={saveErrorMessage ?? 'Failed to save changes'}
        >
          <svg
            width="13"
            height="13"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            aria-hidden="true"
          >
            <circle cx="10" cy="10" r="7.5" />
            <path strokeLinecap="round" d="M10 6.5v4.25" />
            <circle cx="10" cy="13.5" r="0.85" fill="currentColor" stroke="none" />
          </svg>
          <span>Save failed — draft preserved locally</span>
        </span>
      );
    }
    if (saveState === 'unsaved') {
      return (
        <span className="save-status-badge is-unsaved" role="status" aria-live="polite">
          <span className="status-dot" aria-hidden="true" />
          <span>Unsaved changes</span>
        </span>
      );
    }
    if (saveState === 'saved') {
      return (
        <span className="save-status-badge is-saved" role="status" aria-live="polite">
          <svg
            width="13"
            height="13"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            aria-hidden="true"
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 10.5 3.5 3.5 7.5-7.5" />
          </svg>
          <span>Saved</span>
        </span>
      );
    }
    return (
      <span className="save-status-badge is-new" role="status">
        <span>New draft</span>
      </span>
    );
  };

  return (
    <div className="note-metadata-bar" aria-label="Note status and timestamps">
      <div className="note-metadata-left">
        {renderSaveBadge()}
        {updatedAt && (
          <span className="note-metadata-item">
            Updated <time dateTime={updatedAt}>{formatDateTime(updatedAt)}</time>
          </span>
        )}
        {!updatedAt && createdAt && (
          <span className="note-metadata-item">
            Created <time dateTime={createdAt}>{formatDateTime(createdAt)}</time>
          </span>
        )}
      </div>

      <div className="note-metadata-right">
        <span className="note-metadata-item">
          {wordCount} {wordCount === 1 ? 'word' : 'words'}
        </span>
      </div>
    </div>
  );
}
