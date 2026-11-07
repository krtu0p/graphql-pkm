import type { Backlink, Note } from '../../lib/graphql/types';

export interface ContextPanelProps {
  isOpen: boolean;
  activeNote: Note | null;
  allNotes: Note[];
  backlinks: Backlink[];
  isEnriching?: boolean;
  onClose: () => void;
  onToggle?: () => void;
  onSelectNote?: (noteId: string) => void;
  onRemoveLink?: (linkId: string) => Promise<void>;
  onRefreshMetadata?: () => void;
}

export function ContextPanel({
  isOpen,
  activeNote,
  allNotes,
  backlinks,
  isEnriching = false,
  onClose,
  onToggle,
  onSelectNote,
  onRemoveLink,
  onRefreshMetadata,
}: ContextPanelProps) {
  const noteById = new Map<string, Note>();
  for (const n of allNotes) {
    noteById.set(n.id, n);
  }

  const connectionCount =
    (activeNote?.links.length ?? 0) + backlinks.length;

  return (
    <>
      <div
        className={`context-panel-backdrop ${isOpen ? 'is-open' : ''}`}
        onClick={onClose}
        aria-hidden="true"
      />
      <aside
        id="note-context-panel"
        className={`context-panel ${isOpen ? 'is-open' : 'is-collapsed'}`}
        aria-label="Note Context Panel"
      >
        <div className="context-panel-header">
          <button
            type="button"
            className="context-mobile-accordion-trigger"
            onClick={onToggle ?? onClose}
            aria-expanded={isOpen}
            aria-controls="note-context-panel-body"
          >
            <span className="context-panel-title">
              Context &amp; Connections
              {activeNote && connectionCount > 0 ? ` (${connectionCount})` : ''}
            </span>
            <span className="context-mobile-chevron" aria-hidden="true">
              {isOpen ? 'Hide' : 'Show'}
            </span>
          </button>

          <span className="context-panel-title context-desktop-title">
            Context &amp; Connections
          </span>

          <div className="context-panel-header-actions">
            {activeNote && onRefreshMetadata && isOpen && (
              <button
                type="button"
                className="icon-btn"
                onClick={onRefreshMetadata}
                aria-label="Refresh note context"
                title="Refresh AI summary and links"
              >
                Refresh
              </button>
            )}
            <button
              type="button"
              className="icon-btn context-close-btn"
              onClick={onClose}
              aria-label="Collapse context panel"
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
                <path strokeLinecap="round" d="m5 5 10 10M15 5 5 15" />
              </svg>
            </button>
          </div>
        </div>

        {isOpen && (
          <div id="note-context-panel-body" className="context-panel-body">
            {!activeNote ? (
              <div className="context-card">
                Select or open a note to inspect its AI-generated summary, key concepts, and bidirectional links.
              </div>
            ) : (
              <>
                <section className="context-section" aria-labelledby="ctx-active-note-heading">
                  <h2 id="ctx-active-note-heading" className="context-section-heading">
                    Active Note
                  </h2>
                  <div className="context-card">
                    <strong className="context-note-title">
                      {activeNote.title || 'Untitled Note'}
                    </strong>
                  </div>
                </section>

                <section className="context-section" aria-labelledby="ctx-summary-heading">
                  <h2 id="ctx-summary-heading" className="context-section-heading">
                    AI Summary
                  </h2>
                  <div className="context-card">
                    {activeNote.summary ? (
                      <p className="context-summary-text">{activeNote.summary}</p>
                    ) : (
                      <span>
                        {isEnriching
                          ? 'Generating summary…'
                          : 'Not ready yet (or local AI is disabled).'}
                      </span>
                    )}
                  </div>
                </section>

                <section className="context-section" aria-labelledby="ctx-concepts-heading">
                  <h2 id="ctx-concepts-heading" className="context-section-heading">
                    Key Concepts
                  </h2>
                  <div className="context-card">
                    {activeNote.keyConcepts.length > 0 ? (
                      <div className="context-tag-wrap">
                        {activeNote.keyConcepts.map((concept) => (
                          <span key={concept} className="tag-pill">
                            {concept}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <span>
                        {isEnriching ? 'Extracting key concepts…' : 'No key concepts yet.'}
                      </span>
                    )}
                  </div>
                </section>

                <section className="context-section" aria-labelledby="ctx-links-heading">
                  <h2 id="ctx-links-heading" className="context-section-heading">
                    Linked Notes ({activeNote.links.length})
                  </h2>
                  <div className="context-card">
                    {activeNote.links.length > 0 ? (
                      <ul className="context-link-list">
                        {activeNote.links.map((link) => {
                          const target = noteById.get(link.targetNoteId);
                          const title =
                            target?.title || `Note ${link.targetNoteId.slice(0, 8)}`;
                          return (
                            <li key={link.id} className="context-link-item">
                              <div className="context-link-main">
                                <button
                                  type="button"
                                  className="connection-note-link"
                                  onClick={() => onSelectNote?.(link.targetNoteId)}
                                >
                                  {title}
                                </button>
                                {link.description && (
                                  <span className="connection-description">
                                    {link.description}
                                  </span>
                                )}
                              </div>
                              {onRemoveLink && (
                                <button
                                  type="button"
                                  className="icon-btn"
                                  onClick={() => void onRemoveLink(link.id)}
                                  aria-label={`Unlink ${title}`}
                                >
                                  ×
                                </button>
                              )}
                            </li>
                          );
                        })}
                      </ul>
                    ) : (
                      <span>No outgoing links from this note.</span>
                    )}
                  </div>
                </section>

                <section className="context-section" aria-labelledby="ctx-backlinks-heading">
                  <h2 id="ctx-backlinks-heading" className="context-section-heading">
                    Backlinks ({backlinks.length})
                  </h2>
                  <div className="context-card">
                    {backlinks.length > 0 ? (
                      <ul className="context-link-list">
                        {backlinks.map((bl) => (
                          <li key={bl.id} className="context-link-item">
                            <div className="context-link-main">
                              <button
                                type="button"
                                className="connection-note-link"
                                onClick={() => onSelectNote?.(bl.sourceNote.id)}
                              >
                                {bl.sourceNote.title || 'Untitled Note'}
                              </button>
                              {bl.description && (
                                <span className="connection-description">
                                  {bl.description}
                                </span>
                              )}
                            </div>
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <span>No notes reference this note yet.</span>
                    )}
                  </div>
                </section>
              </>
            )}
          </div>
        )}
      </aside>
    </>
  );
}
