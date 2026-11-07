import { Button } from '../components/common/Button';
import { EmptyState } from '../components/common/EmptyState';
import { ErrorState } from '../components/common/ErrorState';
import { LoadingState } from '../components/common/LoadingState';
import { RelationshipChecker } from '../components/search/RelationshipChecker';
import { SearchBar } from '../components/search/SearchBar';
import { SearchFilters } from '../components/search/SearchFilters';
import { SearchResults } from '../components/search/SearchResults';
import { SmartSearchResults } from '../components/search/SmartSearchResults';
import { useAiSearch } from '../hooks/useAiSearch';
import type { Note } from '../lib/graphql/types';

export interface SearchPageProps {
  notes: Note[];
  activeNoteId: string | null;
  onSelectNote: (noteId: string) => void;
  onSelectTag: (tag: string) => void;
  onCreateNote: () => void;
}

export function SearchPage({
  notes,
  activeNoteId,
  onSelectNote,
  onSelectTag,
  onCreateNote,
}: SearchPageProps) {
  const {
    mode,
    query,
    setQuery,
    submittedQuery,
    hasSearched,
    isLoading,
    errorMessage,
    isAiUnavailable,
    aiFallbackMessage,
    keywordResults,
    rankedResults,
    smartResponse,
    runSearch,
    handleModeChange,
    retryAiModes,
  } = useAiSearch();

  return (
    <section className="search-page" aria-labelledby="search-page-heading">
      <div className="page-header">
        <div>
          <h1 id="search-page-heading" className="page-title">
            AI Search
          </h1>
          <p className="page-description">
            Search across your library using smart synthesis, hybrid ranking, semantic similarity, or keyword matching.
          </p>
        </div>

        <Button variant="primary" onClick={onCreateNote}>
          New Note
        </Button>
      </div>

      <div className="search-controls-surface">
        <SearchBar
          query={query}
          isLoading={isLoading}
          onQueryChange={setQuery}
          onSubmit={() => void runSearch(query, mode)}
          onClear={() => {
            setQuery('');
            void runSearch('', mode);
          }}
        />

        <SearchFilters
          activeMode={mode}
          isAiUnavailable={isAiUnavailable}
          onSelectMode={handleModeChange}
        />
      </div>

      {isAiUnavailable && (
        <div className="ai-fallback-banner" role="status">
          <div className="ai-fallback-banner-text">
            <strong>Local AI unavailable — using Keyword Search.</strong>
            <span>
              {aiFallbackMessage ??
                'Semantic, Hybrid, and Smart Search modes are disabled while local AI is offline.'}
            </span>
          </div>
          <Button variant="secondary" onClick={retryAiModes}>
            Re-enable AI Modes
          </Button>
        </div>
      )}

      {errorMessage && (
        <ErrorState
          title="Search could not be completed"
          message={errorMessage}
          onRetry={() => void runSearch(query, mode)}
        />
      )}

      {isLoading ? (
        <LoadingState
          label={
            mode === 'smart'
              ? 'Synthesizing notes, explanations, and connections…'
              : 'Searching your knowledge library…'
          }
          count={3}
        />
      ) : !hasSearched ? (
        <EmptyState
          title="Explore your knowledge base"
          description="Enter a question, topic, or keyword above to search across your notes."
        />
      ) : mode === 'smart' && smartResponse ? (
        <SmartSearchResults
          submittedQuery={submittedQuery}
          response={smartResponse}
          activeNoteId={activeNoteId}
          onSelectNote={onSelectNote}
          onSelectTag={onSelectTag}
        />
      ) : (
        <SearchResults
          mode={mode === 'smart' ? 'keyword' : mode}
          submittedQuery={submittedQuery}
          keywordNotes={keywordResults}
          rankedResults={rankedResults}
          activeNoteId={activeNoteId}
          onSelectNote={onSelectNote}
          onSelectTag={onSelectTag}
        />
      )}

      <RelationshipChecker
        notes={notes}
        isAiUnavailable={isAiUnavailable}
        onDetectedAiUnavailable={() => {
          handleModeChange('keyword');
        }}
      />
    </section>
  );
}
