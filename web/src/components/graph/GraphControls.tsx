import { Input } from '../common/Input';

export interface GraphControlsProps {
  searchQuery: string;
  selectedTag: string | null;
  availableTags: string[];
  showConnectedOnly: boolean;
  zoom: number;
  totalNodes: number;
  visibleNodes: number;
  visibleEdges: number;
  onSearchQueryChange: (query: string) => void;
  onSelectTag: (tag: string | null) => void;
  onToggleConnectedOnly: () => void;
  onZoomIn: () => void;
  onZoomOut: () => void;
  onResetView: () => void;
}

export function GraphControls({
  searchQuery,
  selectedTag,
  availableTags,
  showConnectedOnly,
  zoom,
  totalNodes,
  visibleNodes,
  visibleEdges,
  onSearchQueryChange,
  onSelectTag,
  onToggleConnectedOnly,
  onZoomIn,
  onZoomOut,
  onResetView,
}: GraphControlsProps) {
  const zoomPercent = Math.round(zoom * 100);

  return (
    <div
      className="graph-controls-bar"
      role="region"
      aria-label="Knowledge graph search, filter, and zoom controls"
    >
      <div className="graph-filter-row">
        <div className="graph-search-box">
          <Input
            id="graph-node-search"
            label="Filter graph nodes"
            hideLabel
            type="search"
            value={searchQuery}
            placeholder="Filter nodes by title, content, or tag…"
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
                  aria-label="Clear node search"
                >
                  Clear
                </button>
              ) : undefined
            }
          />
        </div>

        <div className="graph-filter-actions">
          {availableTags.length > 0 && (
            <div className="sort-control">
              <label htmlFor="graph-tag-filter" className="sort-label">
                Tag
              </label>
              <select
                id="graph-tag-filter"
                className="select-input"
                value={selectedTag ?? ''}
                onChange={(event) =>
                  onSelectTag(event.target.value ? event.target.value : null)
                }
              >
                <option value="">All tags</option>
                {availableTags.map((tag) => (
                  <option key={tag} value={tag}>
                    #{tag}
                  </option>
                ))}
              </select>
            </div>
          )}

          <button
            type="button"
            className={`tag-pill tag-pill-interactive ${showConnectedOnly ? 'is-selected' : ''}`}
            onClick={onToggleConnectedOnly}
            aria-pressed={showConnectedOnly}
          >
            <span>Connected only</span>
          </button>
        </div>
      </div>

      <div className="graph-status-zoom-row">
        <div className="graph-metrics" aria-live="polite">
          <span>
            Showing <strong>{visibleNodes}</strong> of {totalNodes}{' '}
            {totalNodes === 1 ? 'note' : 'notes'}
          </span>
          <span aria-hidden="true">·</span>
          <span>
            <strong>{visibleEdges}</strong>{' '}
            {visibleEdges === 1 ? 'link' : 'links'}
          </span>
        </div>

        <div className="graph-zoom-controls" role="group" aria-label="Canvas zoom and pan controls">
          <button
            type="button"
            className="icon-btn"
            onClick={onZoomOut}
            disabled={zoom <= 0.55}
            aria-label="Zoom out graph"
            title="Zoom out"
          >
            −
          </button>
          <span className="graph-zoom-readout" aria-label={`Zoom level ${zoomPercent}%`}>
            {zoomPercent}%
          </span>
          <button
            type="button"
            className="icon-btn"
            onClick={onZoomIn}
            disabled={zoom >= 2.2}
            aria-label="Zoom in graph"
            title="Zoom in"
          >
            +
          </button>
          <button
            type="button"
            className="icon-btn"
            onClick={onResetView}
            aria-label="Reset zoom and pan"
          >
            Reset View
          </button>
        </div>
      </div>
    </div>
  );
}
