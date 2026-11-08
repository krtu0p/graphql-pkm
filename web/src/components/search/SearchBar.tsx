import type { FormEvent } from 'react';
import { Button } from '../common/Button';
import { Input } from '../common/Input';

export interface SearchBarProps {
  query: string;
  isLoading: boolean;
  onQueryChange: (nextQuery: string) => void;
  onSubmit: () => void;
  onClear: () => void;
}

export function SearchBar({
  query,
  isLoading,
  onQueryChange,
  onSubmit,
  onClear,
}: SearchBarProps) {
  const handleFormSubmit = (event: FormEvent) => {
    event.preventDefault();
    onSubmit();
  };

  return (
    <form className="ai-search-form" onSubmit={handleFormSubmit} role="search">
      <div className="ai-search-input-wrap">
        <Input
          id="ai-search-query-input"
          label="Search across your knowledge base"
          hideLabel
          type="search"
          value={query}
          placeholder="Ask a question, search by concept, or enter keywords…"
          onChange={(event) => onQueryChange(event.target.value)}
          startIcon={
            <svg
              width="16"
              height="16"
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
            query ? (
              <button
                type="button"
                className="icon-btn"
                onClick={onClear}
                aria-label="Clear search query"
              >
                Clear
              </button>
            ) : undefined
          }
        />
      </div>

      <Button
        type="submit"
        variant="primary"
        disabled={isLoading || !query.trim()}
      >
        {isLoading ? 'Searching…' : 'Search'}
      </Button>
    </form>
  );
}
