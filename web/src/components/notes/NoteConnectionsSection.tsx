import { useState } from 'react';
import type { Backlink, Link, Note } from '../../lib/graphql/types';
import { Button } from '../common/Button';

export interface NoteConnectionsSectionProps {
  currentNote: Note;
  allNotes: Note[];
  backlinks: Backlink[];
  isEnriching: boolean;
  isLinking: boolean;
  linkError: string | null;
  onSelectNote: (noteId: string) => void;
  onCreateLink: (targetNoteId: string, description: string) => Promise<void>;
  onRemoveLink: (linkId: string) => Promise<void>;
  onRefreshMetadata: () => void;
}

export function NoteConnectionsSection({
  currentNote,
  allNotes,
  backlinks,
  isEnriching,
  isLinking,
  linkError,
  onSelectNote,
  onCreateLink,
  onRemoveLink,
  onRefreshMetadata,
}: NoteConnectionsSectionProps) {
  const [selectedTargetId, setSelectedTargetId] = useState('');
  const [linkDescription, setLinkDescription] = useState('');

  const noteMap = new Map<string, Note>();
  for (const n of allNotes) {
    noteMap.set(n.id, n);
  }

  const linkedTargetIds = new Set(currentNote.links.map((l) => l.targetNoteId));
  const linkableNotes = allNotes.filter(
    (n) => n.id !== currentNote.id && !linkedTargetIds.has(n.id),
  );

  const handleAddLink = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!selectedTargetId || isLinking) return;
    await onCreateLink(selectedTargetId, linkDescription);
    setSelectedTargetId('');
    setLinkDescription('');
  };

  const renderOutgoingLink = (link: Link) => {
    const target = noteMap.get(link.targetNoteId);
    const targetTitle = target?.title || `Note ${link.targetNoteId.slice(0, 8)}`;

    return (
      <li key={link.id} className="connection-list-item">
        <div className="connection-item-main">
          <button
            type="button"
            className="connection-note-link"
            onClick={() => onSelectNote(link.targetNoteId)}
          >
            {targetTitle}
          </button>
          {link.description && (
            <span className="connection-description">{link.description}</span>
          )}
        </div>
        <button
          type="button"
          className="icon-btn"
          onClick={() => void onRemoveLink(link.id)}
          disabled={isLinking}
          aria-label={`Remove link to ${targetTitle}`}
        >
          Unlink
        </button>
      </li>
    );
  };

  return (
    <aside className="editor-insights-section" aria-label="Note AI insights and links">
      <div className="editor-insights-grid">
        <section className="insight-box" aria-labelledby="editor-ai-summary-heading">
          <div className="insight-box-header">
            <h2 id="editor-ai-summary-heading" className="context-section-heading">
              AI Summary &amp; Key Concepts
            </h2>
            <button
              type="button"
              className="icon-btn"
              onClick={onRefreshMetadata}
              aria-label="Refresh AI summary and concepts"
            >
              Refresh
            </button>
          </div>

          {currentNote.summary ? (
            <p className="insight-summary-text">{currentNote.summary}</p>
          ) : (
            <p className="insight-placeholder-text">
              {isEnriching
                ? 'Generating summary in background…'
                : 'Summary not ready yet (or local AI is disabled).'}
            </p>
          )}

          <div className="insight-concepts-block">
            <span className="insight-sub-label">Key Concepts</span>
            {currentNote.keyConcepts.length > 0 ? (
              <ul className="concept-list" aria-label="Key concepts">
                {currentNote.keyConcepts.map((concept) => (
                  <li key={concept} className="tag-pill">
                    {concept}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="insight-placeholder-text">
                {isEnriching ? 'Extracting key concepts…' : 'No key concepts extracted yet.'}
              </p>
            )}
          </div>
        </section>

        <section className="insight-box" aria-labelledby="editor-links-heading">
          <h2 id="editor-links-heading" className="context-section-heading">
            Linked Notes ({currentNote.links.length})
          </h2>

          {currentNote.links.length > 0 ? (
            <ul className="connection-list">
              {currentNote.links.map(renderOutgoingLink)}
            </ul>
          ) : (
            <p className="insight-placeholder-text">
              No outgoing links from this note yet.
            </p>
          )}

          <form className="link-create-form" onSubmit={(e) => void handleAddLink(e)}>
            <label htmlFor="link-target-select" className="sr-only">
              Select a note to link
            </label>
            <select
              id="link-target-select"
              className="select-input"
              value={selectedTargetId}
              onChange={(event) => setSelectedTargetId(event.target.value)}
              disabled={isLinking || linkableNotes.length === 0}
            >
              <option value="">
                {linkableNotes.length === 0
                  ? 'No additional notes available to link'
                  : 'Link to another note…'}
              </option>
              {linkableNotes.map((candidate) => (
                <option key={candidate.id} value={candidate.id}>
                  {candidate.title || 'Untitled Note'}
                </option>
              ))}
            </select>

            {selectedTargetId && (
              <div className="link-create-details">
                <label htmlFor="link-description-input" className="sr-only">
                  Optional link relationship description
                </label>
                <input
                  id="link-description-input"
                  type="text"
                  className="text-input"
                  placeholder="Optional relationship note (e.g., expands on)…"
                  value={linkDescription}
                  onChange={(event) => setLinkDescription(event.target.value)}
                  disabled={isLinking}
                />
                <Button type="submit" variant="secondary" disabled={isLinking}>
                  {isLinking ? 'Linking…' : 'Add Link'}
                </Button>
              </div>
            )}
          </form>

          {linkError && (
            <p className="inline-error-text" role="alert">
              {linkError}
            </p>
          )}

          <div className="backlinks-subsection">
            <h3 className="context-section-heading">
              Backlinks ({backlinks.length})
            </h3>
            {backlinks.length > 0 ? (
              <ul className="connection-list">
                {backlinks.map((bl) => (
                  <li key={bl.id} className="connection-list-item">
                    <div className="connection-item-main">
                      <button
                        type="button"
                        className="connection-note-link"
                        onClick={() => onSelectNote(bl.sourceNote.id)}
                      >
                        {bl.sourceNote.title || 'Untitled Note'}
                      </button>
                      {bl.description && (
                        <span className="connection-description">{bl.description}</span>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="insight-placeholder-text">
                No other notes reference this note yet.
              </p>
            )}
          </div>
        </section>
      </div>
    </aside>
  );
}
