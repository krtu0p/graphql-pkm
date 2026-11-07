import type { SearchMode } from '../../hooks/useAiSearch';

export interface SearchFiltersProps {
  activeMode: SearchMode;
  isAiUnavailable: boolean;
  onSelectMode: (mode: SearchMode) => void;
}

interface SearchModeOption {
  id: SearchMode;
  label: string;
  description: string;
  requiresAi: boolean;
}

const SEARCH_MODES: SearchModeOption[] = [
  {
    id: 'smart',
    label: 'Smart Search',
    description: 'AI synthesis with explanation, knowledge gaps, and suggested connections.',
    requiresAi: true,
  },
  {
    id: 'hybrid',
    label: 'Hybrid',
    description: 'Combines exact keyword matches with conceptual semantic similarity.',
    requiresAi: true,
  },
  {
    id: 'semantic',
    label: 'Semantic',
    description: 'Finds conceptually related notes using vector embeddings.',
    requiresAi: true,
  },
  {
    id: 'keyword',
    label: 'Keyword',
    description: 'Fast exact-text matching across note titles, bodies, and tags.',
    requiresAi: false,
  },
];

export function SearchFilters({
  activeMode,
  isAiUnavailable,
  onSelectMode,
}: SearchFiltersProps) {
  const currentMeta =
    SEARCH_MODES.find((m) => m.id === activeMode) ?? SEARCH_MODES[3];

  return (
    <div className="search-filters-panel">
      <div className="search-mode-tabs" role="tablist" aria-label="Search mode">
        {SEARCH_MODES.map((option) => {
          const isSelected = activeMode === option.id;
          const isDisabled = isAiUnavailable && option.requiresAi;

          return (
            <button
              key={option.id}
              type="button"
              role="tab"
              aria-selected={isSelected}
              disabled={isDisabled}
              className={`search-mode-tab ${isSelected ? 'is-active' : ''}`}
              onClick={() => onSelectMode(option.id)}
              title={
                isDisabled
                  ? 'Unavailable while local AI is offline or disabled'
                  : option.description
              }
            >
              <span>{option.label}</span>
            </button>
          );
        })}
      </div>

      <p className="search-mode-caption">{currentMeta.description}</p>
    </div>
  );
}
