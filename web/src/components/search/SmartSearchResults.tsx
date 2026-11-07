import type { SmartSearchResponse } from '../../lib/graphql/types';
import { createExcerpt, formatRelativeDate } from '../../lib/utils/format';
import { EmptyState } from '../common/EmptyState';
import { TagList } from '../notes/TagList';

export interface SmartSearchResultsProps {
  submittedQuery: string;
  response: SmartSearchResponse;
  activeNoteId: string | null;
  onSelectNote: (noteId: string) => void;
  onSelectTag: (tag: string) => void;
}

export function SmartSearchResults({
  submittedQuery,
  response,
  activeNoteId,
  onSelectNote,
  onSelectTag,
}: SmartSearchResultsProps) {
  const { results, explanation, gaps, connections } = response;
  const hasAnyEditorialInsight = Boolean(
    explanation.trim() || gaps.length > 0 || connections.length > 0,
  );

  return (
    <div className="smart-search-layout">
      <div className="ai-disclaimer-bar" role="note">
        <span>
          AI synthesis is generated from your note library and may be incomplete or inaccurate. Verify important details in your source notes.
        </span>
      </div>

      {hasAnyEditorialInsight && (
        <section
          className="smart-editorial-brief"
          aria-labelledby="smart-synthesis-heading"
        >
          <h2 id="smart-synthesis-heading" className="smart-brief-title">
            Synthesis for &ldquo;{submittedQuery}&rdquo;
          </h2>

          {explanation.trim() && (
            <p className="smart-explanation-body">{explanation}</p>
          )}

          {(gaps.length > 0 || connections.length > 0) && (
            <div className="smart-insights-columns">
              <div className="smart-insight-card">
                <h3 className="context-section-heading">Knowledge Gaps</h3>
                {gaps.length > 0 ? (
                  <ul className="smart-bullet-list">
                    {gaps.map((gap, index) => (
                      <li key={`${index}-${gap.slice(0, 24)}`}>{gap}</li>
                    ))}
                  </ul>
                ) : (
                  <p className="insight-placeholder-text">
                    No specific knowledge gaps identified for this query.
                  </p>
                )}
              </div>

              <div className="smart-insight-card">
                <h3 className="context-section-heading">Suggested Connections</h3>
                {connections.length > 0 ? (
                  <ul className="smart-bullet-list">
                    {connections.map((conn, index) => (
                      <li key={`${index}-${conn.slice(0, 24)}`}>{conn}</li>
                    ))}
                  </ul>
                ) : (
                  <p className="insight-placeholder-text">
                    No additional cross-note connections suggested.
                  </p>
                )}
              </div>
            </div>
          )}
        </section>
      )}

      <section
        className="search-results-section"
        aria-labelledby="smart-relevant-notes-heading"
      >
        <div className="section-subheader">
          <h2 id="smart-relevant-notes-heading" className="section-label">
            Relevant Notes ({results.length})
          </h2>
        </div>

        {results.length === 0 ? (
          <EmptyState
            title="No directly matching notes"
            description="Smart search did not find notes matching this query."
          />
        ) : (
          <div className="ranked-results-list" role="list">
            {results.map(({ note, score, reason }) => (
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
                    <span className="search-score-pill">
                      {Math.round(Math.max(0, Math.min(1, score)) * 100)}% match
                    </span>
                  </div>
                </div>

                {reason && (
                  <p className="search-match-reason">
                    <strong>Why it is relevant:</strong> {reason}
                  </p>
                )}

                <div
                  className="note-card-body"
                  onClick={() => onSelectNote(note.id)}
                  role="presentation"
                >
                  <p className="note-card-excerpt">
                    {createExcerpt(note.content, 170)}
                  </p>
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
        )}
      </section>
    </div>
  );
}
