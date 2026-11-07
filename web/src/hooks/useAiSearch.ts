import { useCallback, useRef, useState } from 'react';
import {
  fetchHybridSearch,
  fetchSemanticSearch,
  fetchSmartSearch,
  searchNotesKeyword,
  type Note,
  type SearchResult,
  type SmartSearchResponse,
} from '../lib/graphql';

export type SearchMode = 'keyword' | 'semantic' | 'hybrid' | 'smart';

function isAiFallbackResultList(results: SearchResult[]): boolean {
  return (
    results.length > 0 &&
    results.every((r) => r.reason.toLowerCase().includes('ai unavailable'))
  );
}

function isAiFallbackSmartResponse(resp: SmartSearchResponse): boolean {
  const lower = resp.explanation.toLowerCase();
  return (
    lower.includes('ai features are disabled') ||
    lower.includes('ai unavailable')
  );
}

export function useAiSearch() {
  const [mode, setMode] = useState<SearchMode>('smart');
  const [query, setQuery] = useState('');
  const [submittedQuery, setSubmittedQuery] = useState('');
  const [hasSearched, setHasSearched] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isAiUnavailable, setIsAiUnavailable] = useState(false);
  const [aiFallbackMessage, setAiFallbackMessage] = useState<string | null>(null);

  const [keywordResults, setKeywordResults] = useState<Note[]>([]);
  const [rankedResults, setRankedResults] = useState<SearchResult[]>([]);
  const [smartResponse, setSmartResponse] = useState<SmartSearchResponse | null>(null);

  const activeControllerRef = useRef<AbortController | null>(null);

  const runSearch = useCallback(
    async (rawQuery: string, targetMode: SearchMode) => {
      const trimmed = rawQuery.trim();
      if (!trimmed) {
        setHasSearched(false);
        setSubmittedQuery('');
        setKeywordResults([]);
        setRankedResults([]);
        setSmartResponse(null);
        setErrorMessage(null);
        return;
      }

      if (activeControllerRef.current) {
        activeControllerRef.current.abort();
      }
      const controller = new AbortController();
      activeControllerRef.current = controller;

      const effectiveMode: SearchMode =
        isAiUnavailable && targetMode !== 'keyword' ? 'keyword' : targetMode;

      setIsLoading(true);
      setErrorMessage(null);
      setSubmittedQuery(trimmed);
      setHasSearched(true);

      try {
        if (effectiveMode === 'keyword') {
          const notes = await searchNotesKeyword(trimmed, controller.signal);
          setKeywordResults(notes);
          setRankedResults([]);
          setSmartResponse(null);
          return;
        }

        if (effectiveMode === 'semantic') {
          const results = await fetchSemanticSearch(trimmed, controller.signal);
          if (isAiFallbackResultList(results)) {
            setIsAiUnavailable(true);
            setMode('keyword');
            setAiFallbackMessage(
              'Local AI embeddings are disabled or unreachable. Switched to keyword search.',
            );
            setKeywordResults(results.map((r) => r.note));
            setRankedResults([]);
            return;
          }
          setRankedResults(results);
          setKeywordResults([]);
          setSmartResponse(null);
          return;
        }

        if (effectiveMode === 'hybrid') {
          const results = await fetchHybridSearch(trimmed, controller.signal);
          if (isAiFallbackResultList(results)) {
            setIsAiUnavailable(true);
            setMode('keyword');
            setAiFallbackMessage(
              'Local AI service is unavailable. Showing keyword search results instead.',
            );
            setKeywordResults(results.map((r) => r.note));
            setRankedResults([]);
            return;
          }
          setRankedResults(results);
          setKeywordResults([]);
          setSmartResponse(null);
          return;
        }

        const smart = await fetchSmartSearch(trimmed, controller.signal);
        if (isAiFallbackSmartResponse(smart)) {
          setIsAiUnavailable(true);
          setMode('keyword');
          setAiFallbackMessage(
            'Local AI service is disabled or unreachable. Automatically fell back to keyword search.',
          );
          setKeywordResults(smart.results.map((r) => r.note));
          setSmartResponse(null);
          setRankedResults([]);
          return;
        }
        setSmartResponse(smart);
        setKeywordResults([]);
        setRankedResults([]);
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          return;
        }
        const message =
          err instanceof Error ? err.message : 'Search request failed.';

        if (
          effectiveMode !== 'keyword' &&
          (message.toLowerCase().includes('ai') ||
            message.toLowerCase().includes('ollama') ||
            message.toLowerCase().includes('embedding'))
        ) {
          setIsAiUnavailable(true);
          setMode('keyword');
          setAiFallbackMessage(
            'AI search is currently unavailable. Falling back to keyword search.',
          );
          try {
            const fallbackNotes = await searchNotesKeyword(
              trimmed,
              controller.signal,
            );
            setKeywordResults(fallbackNotes);
            setRankedResults([]);
            setSmartResponse(null);
            return;
          } catch {
            // If keyword fallback also fails, surface original error
          }
        }
        setErrorMessage(message);
      } finally {
        if (!controller.signal.aborted) {
          setIsLoading(false);
        }
      }
    },
    [isAiUnavailable],
  );

  const handleModeChange = useCallback(
    (nextMode: SearchMode) => {
      if (isAiUnavailable && nextMode !== 'keyword') {
        return;
      }
      setMode(nextMode);
      if (query.trim()) {
        void runSearch(query, nextMode);
      }
    },
    [isAiUnavailable, query, runSearch],
  );

  const retryAiModes = useCallback(() => {
    setIsAiUnavailable(false);
    setAiFallbackMessage(null);
  }, []);

  return {
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
  };
}

export type AiSearchController = ReturnType<typeof useAiSearch>;
