import { useCallback, useEffect, useRef, useState } from 'react';
import {
  fetchAiModels,
  fetchAiSettings,
  saveAiSettings,
  type AiSettings,
} from '../lib/graphql';

export type AiStatus = 'loading' | 'on' | 'off' | 'unavailable';
export type ApiKeyMode = 'keep' | 'replace' | 'clear';

export interface AiProviderOption {
  value: string;
  label: string;
  baseUrl: string;
}

export const AI_PROVIDER_OPTIONS: AiProviderOption[] = [
  {
    value: 'openai',
    label: 'OpenAI',
    baseUrl: 'https://api.openai.com/v1',
  },
  {
    value: 'groq',
    label: 'Groq',
    baseUrl: 'https://api.groq.com/openai/v1',
  },
  {
    value: 'gemini',
    label: 'Google Gemini',
    baseUrl: 'https://generativelanguage.googleapis.com/v1beta/openai',
  },
  {
    value: 'openrouter',
    label: 'OpenRouter',
    baseUrl: 'https://openrouter.ai/api/v1',
  },
  {
    value: 'ollama',
    label: 'Ollama (local)',
    baseUrl: 'http://localhost:11434/v1',
  },
  {
    value: 'custom',
    label: 'Custom',
    baseUrl: '',
  },
];

export interface AiSettingsDraft {
  enabled: boolean;
  provider: string;
  baseUrl: string;
  model: string;
  embedBaseUrl: string;
  embedModel: string;
}

const EMBEDDING_MODEL_DEFAULTS: Record<string, string> = {
  openai: 'text-embedding-3-small',
  ollama: 'nomic-embed-text',
};

const DEFAULT_DRAFT: AiSettingsDraft = {
  enabled: false,
  provider: 'ollama',
  baseUrl: 'http://localhost:11434/v1',
  model: '',
  embedBaseUrl: '',
  embedModel: '',
};

function normalizeProvider(provider: string): string {
  const match = AI_PROVIDER_OPTIONS.find((option) => option.value === provider);
  return match ? match.value : 'custom';
}

function normalizeBaseUrl(value: string): string {
  return value.trim().replace(/\/+$/, '');
}

function draftFromSettings(settings: AiSettings): AiSettingsDraft {
  return {
    enabled: settings.enabled,
    provider: normalizeProvider(settings.provider),
    baseUrl: settings.baseUrl,
    model: settings.model,
    embedBaseUrl: settings.embedBaseUrl,
    embedModel: settings.embedModel,
  };
}

function getErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

export function useAiSettings() {
  const [settings, setSettings] = useState<AiSettings | null>(null);
  const [draft, setDraft] = useState<AiSettingsDraft>(DEFAULT_DRAFT);
  const [apiKey, setApiKey] = useState('');
  const [apiKeyMode, setApiKeyMode] = useState<ApiKeyMode>('keep');

  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveMessage, setSaveMessage] = useState<string | null>(null);

  const [models, setModels] = useState<string[]>([]);
  const [modelsError, setModelsError] = useState<string | null>(null);
  const [hasLoadedModels, setHasLoadedModels] = useState(false);
  const [isLoadingModels, setIsLoadingModels] = useState(false);
  const modelsRequestIdRef = useRef(0);

  const [embeddingModels, setEmbeddingModels] = useState<string[]>([]);
  const [embeddingModelsError, setEmbeddingModelsError] = useState<string | null>(null);
  const [hasLoadedEmbeddingModels, setHasLoadedEmbeddingModels] = useState(false);
  const [isLoadingEmbeddingModels, setIsLoadingEmbeddingModels] = useState(false);
  const embeddingModelsRequestIdRef = useRef(0);

  const savedBaseUrl = normalizeBaseUrl(settings?.baseUrl ?? '');
  const savedApiKeyApplies =
    Boolean(settings?.hasApiKey) &&
    savedBaseUrl.length > 0 &&
    savedBaseUrl === normalizeBaseUrl(draft.baseUrl);
  const embeddingBaseUrl = draft.embedBaseUrl.trim() ? draft.embedBaseUrl : draft.baseUrl;

  const load = useCallback(async (signal?: AbortSignal) => {
    setIsLoading(true);
    setLoadError(null);
    try {
      const loaded = await fetchAiSettings(signal);
      if (signal?.aborted) return;

      setSettings(loaded);
      setDraft(draftFromSettings(loaded));
      setApiKey('');
      setApiKeyMode('keep');
      modelsRequestIdRef.current += 1;
      setModels([]);
      setModelsError(null);
      setHasLoadedModels(false);
      setIsLoadingModels(false);
      embeddingModelsRequestIdRef.current += 1;
      setEmbeddingModels([]);
      setEmbeddingModelsError(null);
      setHasLoadedEmbeddingModels(false);
      setIsLoadingEmbeddingModels(false);
      setSaveError(null);
      setSaveMessage(null);
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return;
      if (signal?.aborted) return;
      setLoadError(getErrorMessage(error, 'Unable to load AI settings.'));
    } finally {
      if (!signal?.aborted) {
        setIsLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  }, [load]);

  const resetModelSuggestions = useCallback(() => {
    modelsRequestIdRef.current += 1;
    setModels([]);
    setModelsError(null);
    setHasLoadedModels(false);
    setIsLoadingModels(false);
  }, []);

  const resetEmbeddingModelSuggestions = useCallback(() => {
    embeddingModelsRequestIdRef.current += 1;
    setEmbeddingModels([]);
    setEmbeddingModelsError(null);
    setHasLoadedEmbeddingModels(false);
    setIsLoadingEmbeddingModels(false);
  }, []);

  const getModelsApiKeyInput = useCallback(
    (provider: string): string | undefined => {
      if (provider === 'ollama') return undefined;
      if (apiKeyMode === 'clear') return '';
      const trimmed = apiKey.trim();
      if (apiKeyMode === 'replace' && trimmed.length > 0) return trimmed;
      // Keep mode: omit the key. The backend only reuses the saved key when the
      // requested base URL matches the saved one.
      return undefined;
    },
    [apiKey, apiKeyMode],
  );

  const updateDraft = useCallback(
    (patch: Partial<AiSettingsDraft>) => {
      setDraft((current) => ({ ...current, ...patch }));
      if (patch.baseUrl !== undefined || patch.provider !== undefined) {
        resetModelSuggestions();
      }
      if (
        patch.baseUrl !== undefined ||
        patch.provider !== undefined ||
        patch.embedBaseUrl !== undefined
      ) {
        resetEmbeddingModelSuggestions();
      }
      setSaveError(null);
      setSaveMessage(null);
    },
    [resetEmbeddingModelSuggestions, resetModelSuggestions],
  );

  const selectProvider = useCallback(
    (provider: string) => {
      const preset = AI_PROVIDER_OPTIONS.find((option) => option.value === provider);
      setDraft((current) => ({
        ...current,
        provider,
        baseUrl:
          preset && preset.value !== 'custom' ? preset.baseUrl : current.baseUrl,
        embedModel: EMBEDDING_MODEL_DEFAULTS[provider] ?? current.embedModel,
      }));
      resetModelSuggestions();
      resetEmbeddingModelSuggestions();
      setSaveError(null);
      setSaveMessage(null);
    },
    [resetEmbeddingModelSuggestions, resetModelSuggestions],
  );

  const handleApiKeyChange = useCallback((value: string) => {
    setApiKey(value);
    setApiKeyMode((current) => {
      if (value.trim().length > 0) return 'replace';
      return current === 'replace' ? 'keep' : current;
    });
    setSaveError(null);
    setSaveMessage(null);
  }, []);

  const clearApiKey = useCallback(() => {
    setApiKey('');
    setApiKeyMode('clear');
    setSaveError(null);
    setSaveMessage(null);
  }, []);

  const undoClearApiKey = useCallback(() => {
    setApiKey('');
    setApiKeyMode('keep');
    setSaveError(null);
    setSaveMessage(null);
  }, []);

  const save = useCallback(async (): Promise<boolean> => {
    setIsSaving(true);
    setSaveError(null);
    setSaveMessage(null);

    try {
      const apiKeyInput =
        apiKeyMode === 'replace' ? apiKey : apiKeyMode === 'clear' ? '' : undefined;
      const saved = await saveAiSettings({
        ...draft,
        apiKey: apiKeyInput,
      });

      setSettings(saved);
      setDraft(draftFromSettings(saved));
      setApiKey('');
      setApiKeyMode('keep');
      setLoadError(null);
      setSaveMessage('AI settings saved.');
      return true;
    } catch (error) {
      setSaveError(getErrorMessage(error, 'Unable to save AI settings.'));
      return false;
    } finally {
      setIsSaving(false);
    }
  }, [apiKey, apiKeyMode, draft]);

  const loadModels = useCallback(async () => {
    const isOllamaProvider = draft.provider === 'ollama';

    if (
      !isOllamaProvider &&
      apiKeyMode === 'keep' &&
      settings?.hasApiKey &&
      !savedApiKeyApplies
    ) {
      modelsRequestIdRef.current += 1;
      setIsLoadingModels(false);
      setModels([]);
      setModelsError('Enter an API key for this provider.');
      setHasLoadedModels(true);
      return;
    }

    const requestId = modelsRequestIdRef.current + 1;
    modelsRequestIdRef.current = requestId;
    setIsLoadingModels(true);
    setModelsError(null);

    try {
      const result = await fetchAiModels({
        provider: draft.provider,
        baseUrl: draft.baseUrl,
        apiKey: getModelsApiKeyInput(draft.provider),
      });
      if (modelsRequestIdRef.current !== requestId) return;
      setModels(result.models);
      setModelsError(result.error);
      setHasLoadedModels(true);
    } catch (error) {
      if (modelsRequestIdRef.current !== requestId) return;
      setModels([]);
      setModelsError(getErrorMessage(error, 'Unable to load AI models.'));
      setHasLoadedModels(true);
    } finally {
      if (modelsRequestIdRef.current === requestId) {
        setIsLoadingModels(false);
      }
    }
  }, [
    apiKeyMode,
    draft.baseUrl,
    draft.provider,
    getModelsApiKeyInput,
    savedApiKeyApplies,
    settings?.hasApiKey,
  ]);

  const loadEmbeddingModels = useCallback(async () => {
    const requestId = embeddingModelsRequestIdRef.current + 1;
    embeddingModelsRequestIdRef.current = requestId;
    setIsLoadingEmbeddingModels(true);
    setEmbeddingModelsError(null);

    try {
      const result = await fetchAiModels({
        provider: draft.provider,
        baseUrl: embeddingBaseUrl,
        apiKey: getModelsApiKeyInput(draft.provider),
      });
      if (embeddingModelsRequestIdRef.current !== requestId) return;
      setEmbeddingModels(result.models);
      setEmbeddingModelsError(result.error);
      setHasLoadedEmbeddingModels(true);
    } catch (error) {
      if (embeddingModelsRequestIdRef.current !== requestId) return;
      setEmbeddingModels([]);
      setEmbeddingModelsError(
        getErrorMessage(error, 'Unable to load embedding models.'),
      );
      setHasLoadedEmbeddingModels(true);
    } finally {
      if (embeddingModelsRequestIdRef.current === requestId) {
        setIsLoadingEmbeddingModels(false);
      }
    }
  }, [draft.provider, embeddingBaseUrl, getModelsApiKeyInput]);

  return {
    settings,
    draft,
    apiKey,
    apiKeyMode,
    hasSavedApiKey: settings?.hasApiKey ?? false,
    savedApiKeyApplies,
    isOllama: draft.provider === 'ollama',
    isLoading,
    loadError,
    isSaving,
    saveError,
    saveMessage,
    models,
    modelsError,
    hasLoadedModels,
    isLoadingModels,
    embeddingModels,
    embeddingModelsError,
    hasLoadedEmbeddingModels,
    isLoadingEmbeddingModels,
    load,
    updateDraft,
    selectProvider,
    handleApiKeyChange,
    clearApiKey,
    undoClearApiKey,
    save,
    loadModels,
    loadEmbeddingModels,
  };
}

export type AiSettingsController = ReturnType<typeof useAiSettings>;
