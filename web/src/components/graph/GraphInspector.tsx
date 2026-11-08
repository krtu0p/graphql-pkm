import type { Backlink, Link, Note } from '../../lib/graphql/types';
import { createExcerpt, formatRelativeDate } from '../../lib/utils/format';
import { Button } from '../common/Button';
import { TagList } from '../notes/TagList';

export interface GraphInspectorProps {
  selectedNote: Note | null;
  allNotes: Note[];
  outgoingLinks: Link[];
  incomingBacklinks: Backlink[];
  isLoadingConnections: boolean;
  onOpenNote: (noteId: string) => void;
  onFocusNode: (noteId: string) => void;
  onSelectTag?: (tag: string) => void;
}

export function GraphInspector({
  selectedNote,
  allNotes,
  outgoingLinks,
  incomingBacklinks,
  isLoadingConnections,
  onOpenNote,
  onFocusNode,
  onSelectTag,
}: GraphInspectorProps) {
  if (!selectedNote) {
    return (
      <aside className="graph-inspector-card" aria-label="Selected note relationship details">
        <h2 className="context-section-heading">Node Details</h2>
        <p className="insight-placeholder-text">
          Select any note node in the graph to inspect its outgoing links, incoming backlinks, and relationship descriptions, or open it in the editor.
        </p>
      </aside>
    );
  }

  const noteById = new Map<string, Note>();
  for (const note of allNotes) {
    noteById.set(note.id, note);
  }

  return (
    <aside className="graph-inspector-card" aria-label="Selected note relationship details">
      <div className="graph-inspector-header">
        <div>
          <span className="section-label">Selected Note</span>
          <h2 className="graph-inspector-title">
            {selectedNote.title || 'Untitled Note'}
          </h2>
          <span className="note-card-timestamp">
            Updated {formatRelativeDate(selectedNote.updatedAt)}
          </span>
        </div>

        <Button
          variant="primary"
          onClick={() => onOpenNote(selectedNote.id)}
        >
          Open Note
        </Button>
      </div>

      {selectedNote.summary ? (
        <div className="note-card-summary">
          <span className="note-card-summary-label">AI Summary:</span>{' '}
          <span>{selectedNote.summary}</span>
        </div>
      ) : (
        <p className="note-card-excerpt">
          {createExcerpt(selectedNote.content, 160)}
        </p>
      )}

      {selectedNote.tags.length > 0 && (
        <TagList tags={selectedNote.tags} onSelectTag={onSelectTag} />
      )}

      <div className="graph-inspector-connections">
        <div className="graph-inspector-column">
          <h3 className="context-section-heading">
            Outgoing Links ({outgoingLinks.length})
          </h3>
          {isLoadingConnections && outgoingLinks.length === 0 ? (
            <p className="insight-placeholder-text">Loading outgoing links…</p>
          ) : outgoingLinks.length === 0 ? (
            <p className="insight-placeholder-text">
              This note does not link to any other notes yet.
            </p>
          ) : (
            <ul className="connection-list">
              {outgoingLinks.map((link) => {
                const target = noteById.get(link.targetNoteId);
                const targetTitle =
                  target?.title || `Note ${link.targetNoteId.slice(0, 8)}`;
                return (
                  <li key={link.id} className="connection-list-item">
                    <div className="connection-item-main">
                      <button
                        type="button"
                        className="connection-note-link"
                        onClick={() => onFocusNode(link.targetNoteId)}
                      >
                        {targetTitle}
                      </button>
                      <span className="connection-description">
                        {link.description
                          ? `Relationship: ${link.description}`
                          : 'Direct outgoing link'}
                      </span>
                    </div>
                    <button
                      type="button"
                      className="icon-btn"
                      onClick={() => onOpenNote(link.targetNoteId)}
                      aria-label={`Open ${targetTitle} in editor`}
                    >
                      Open
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>

        <div className="graph-inspector-column">
          <h3 className="context-section-heading">
            Incoming Backlinks ({incomingBacklinks.length})
          </h3>
          {isLoadingConnections && incomingBacklinks.length === 0 ? (
            <p className="insight-placeholder-text">Loading backlinks…</p>
          ) : incomingBacklinks.length === 0 ? (
            <p className="insight-placeholder-text">
              No other notes reference this note yet.
            </p>
          ) : (
            <ul className="connection-list">
              {incomingBacklinks.map((bl) => {
                const sourceTitle = bl.sourceNote.title || 'Untitled Note';
                return (
                  <li key={bl.id} className="connection-list-item">
                    <div className="connection-item-main">
                      <button
                        type="button"
                        className="connection-note-link"
                        onClick={() => onFocusNode(bl.sourceNote.id)}
                      >
                        {sourceTitle}
                      </button>
                      <span className="connection-description">
                        {bl.description
                          ? `Relationship: ${bl.description}`
                          : 'Referenced by this note'}
                      </span>
                    </div>
                    <button
                      type="button"
                      className="icon-btn"
                      onClick={() => onOpenNote(bl.sourceNote.id)}
                      aria-label={`Open ${sourceTitle} in editor`}
                    >
                      Open
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </div>
    </aside>
  );
}
