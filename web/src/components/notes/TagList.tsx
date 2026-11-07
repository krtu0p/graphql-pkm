export interface TagListProps {
  tags: string[];
  selectedTag?: string | null;
  onSelectTag?: (tag: string) => void;
  onRemoveTag?: (tag: string) => void;
  emptyLabel?: string;
}

export function TagList({
  tags,
  selectedTag = null,
  onSelectTag,
  onRemoveTag,
  emptyLabel,
}: TagListProps) {
  if (tags.length === 0) {
    return emptyLabel ? <span className="metadata-muted">{emptyLabel}</span> : null;
  }

  return (
    <ul className="tag-list" aria-label="Note tags">
      {tags.map((tag) => {
        const isSelected = selectedTag === tag;
        return (
          <li key={tag} className="tag-list-item">
            {onSelectTag ? (
              <button
                type="button"
                className={`tag-pill tag-pill-interactive ${isSelected ? 'is-selected' : ''}`}
                onClick={(event) => {
                  event.stopPropagation();
                  onSelectTag(tag);
                }}
                aria-pressed={isSelected}
              >
                <span>#{tag}</span>
              </button>
            ) : (
              <span className="tag-pill">
                <span>#{tag}</span>
                {onRemoveTag && (
                  <button
                    type="button"
                    className="tag-remove-btn"
                    onClick={() => onRemoveTag(tag)}
                    aria-label={`Remove tag ${tag}`}
                  >
                    <svg
                      width="11"
                      height="11"
                      viewBox="0 0 14 14"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.7"
                      aria-hidden="true"
                    >
                      <path strokeLinecap="round" d="M3.5 3.5l7 7M10.5 3.5l-7 7" />
                    </svg>
                  </button>
                )}
              </span>
            )}
          </li>
        );
      })}
    </ul>
  );
}
