import { useEffect, useState, type FormEvent, type ReactNode } from 'react';
import { Button } from '../common/Button';
import { ErrorState } from '../common/ErrorState';
import { Input } from '../common/Input';
import {
  AI_PROVIDER_OPTIONS,
  type AiSettingsController,
} from '../../hooks/useAiSettings';

export interface AiSettingsFormProps {
  controller: AiSettingsController;
  onSettingsSaved: () => void;
}

function KeyIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      aria-hidden="true"
    >
      <circle cx="6.5" cy="13.5" r="2.75" />
      <path strokeLinecap="round" d="m8.5 11.5 7-7M13 4.5l2 2M15 2.5l2 2" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg
      width="15"
      height="15"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      aria-hidden="true"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 10.5 3.5 3.5 7.5-8" />
    </svg>
  );
}

function WarningIcon() {
  return (
    <svg
      width="15"
      height="15"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      aria-hidden="true"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M10 3.25 17 16H3L10 3.25Z" />
      <path strokeLinecap="round" d="M10 8v3.25" />
      <circle cx="10" cy="13.5" r="0.8" fill="currentColor" stroke="none" />
    </svg>
  );
}

interface WarningNoticeProps {
  id: string;
  live?: boolean;
  children: ReactNode;
}

function WarningNotice({ id, live = false, children }: WarningNoticeProps) {
  return (
    <p id={id} className="settings-warning" aria-live={live ? 'polite' : undefined}>
      <WarningIcon />
      <span>{children}</span>
    </p>
  );
}

export function AiSettingsForm({ controller, onSettingsSaved }: AiSettingsFormProps) {
  const {
    draft,
    apiKey,
    apiKeyMode,
    hasSavedApiKey,
    savedApiKeyApplies,
    isOllama,
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
  } = controller;

  const [isCustomModel, setIsCustomModel] = useState(false);
  const [isCustomEmbeddingModel, setIsCustomEmbeddingModel] = useState(false);
  const [showAllEmbeddingModels, setShowAllEmbeddingModels] = useState(false);
  const hasModelOptions = models.length > 0 && !modelsError;
  const showModelSelect = hasModelOptions && !isCustomModel;
  const modelFieldId = showModelSelect ? 'ai-model-select' : 'ai-model-input';
  const needsApiKeyForProvider =
    !isOllama &&
    hasSavedApiKey &&
    !savedApiKeyApplies &&
    apiKeyMode === 'keep' &&
    apiKey.trim().length === 0;

  useEffect(() => {
    setIsCustomEmbeddingModel(false);
    setShowAllEmbeddingModels(false);
  }, [draft.provider]);

  const embeddingModelsForProvider = showAllEmbeddingModels
    ? embeddingModels
    : embeddingModels.filter((model) => model.toLowerCase().includes('embed'));
  const hasEmbeddingModelOptions =
    embeddingModelsForProvider.length > 0 && !embeddingModelsError;
  const showEmbeddingModelSelect =
    hasEmbeddingModelOptions && !isCustomEmbeddingModel;
  const embeddingModelFieldId = showEmbeddingModelSelect
    ? 'ai-embed-model-select'
    : 'ai-embed-model-input';
  const showNoEmbeddingModels =
    hasLoadedEmbeddingModels &&
    !embeddingModelsError &&
    embeddingModelsForProvider.length === 0;
  const showNotEmbeddingModelWarning =
    draft.embedModel.trim().length > 0 &&
    !draft.embedModel.toLowerCase().includes('embed');
  const showMissingEmbeddingModelWarning =
    !isOllama && draft.embedModel.trim().length === 0;
  const embedModelDescribedBy = [
    'ai-embed-model-help',
    showMissingEmbeddingModelWarning ? 'ai-embed-model-missing-warning' : '',
    showNotEmbeddingModelWarning ? 'ai-embed-model-warning' : '',
    showAllEmbeddingModels ? 'ai-embed-models-show-all-warning' : '',
    showNoEmbeddingModels ? 'ai-embed-models-empty' : '',
    embeddingModelsError ? 'ai-embed-models-error' : '',
  ]
    .filter(Boolean)
    .join(' ');

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const saved = await save();
    if (saved) {
      onSettingsSaved();
    }
  };

  const handleModelSelectChange = (value: string) => {
    if (value === '__custom__') {
      setIsCustomModel(true);
      return;
    }
    setIsCustomModel(false);
    updateDraft({ model: value });
  };

  const handleEmbeddingModelSelectChange = (value: string) => {
    if (value === '__custom__') {
      setIsCustomEmbeddingModel(true);
      return;
    }
    setIsCustomEmbeddingModel(false);
    updateDraft({ embedModel: value });
  };

  return (
    <form className="settings-form" onSubmit={(event) => void handleSubmit(event)}>
      {loadError && (
        <ErrorState
          title="AI settings could not be refreshed"
          message={loadError}
          onRetry={() => void load()}
        />
      )}

      {saveError && (
        <ErrorState title="AI settings were not saved" message={saveError} />
      )}

      {saveMessage && (
        <p className="settings-save-message" role="status">
          <CheckIcon />
          <span>{saveMessage}</span>
        </p>
      )}

      <div className="settings-toggle-row">
        <div className="settings-toggle-copy">
          <span id="ai-enabled-label" className="settings-field-label">
            Enable AI features
          </span>
          <p className="settings-hint">
            Turn AI off to keep your library fully local. Keyword search and editing keep
            working either way.
          </p>
        </div>
        <button
          type="button"
          role="switch"
          aria-checked={draft.enabled}
          aria-labelledby="ai-enabled-label"
          className={`settings-switch ${draft.enabled ? 'is-on' : ''}`}
          onClick={() => updateDraft({ enabled: !draft.enabled })}
          disabled={isSaving || isLoading}
        >
          <span className="settings-switch-thumb" aria-hidden="true" />
        </button>
      </div>

      <div className="settings-grid">
        <div className="settings-field">
          <label htmlFor="ai-provider" className="input-label">
            Provider
          </label>
          <select
            id="ai-provider"
            className="select-input"
            value={draft.provider}
            onChange={(event) => selectProvider(event.target.value)}
            disabled={isSaving || isLoading}
          >
            {AI_PROVIDER_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
          <p className="settings-hint">
            Choosing a provider fills its default endpoint. Choose Custom to enter your
            own.
          </p>
        </div>

        <Input
          id="ai-base-url"
          label="Base URL"
          value={draft.baseUrl}
          onChange={(event) => updateDraft({ baseUrl: event.target.value })}
          placeholder="http://localhost:11434/v1"
          autoComplete="off"
          spellCheck={false}
          disabled={isSaving || isLoading}
        />
      </div>

      {isOllama ? (
        <p className="settings-hint">
          Ollama runs on this computer and does not require an API key.
        </p>
      ) : (
        <div className="settings-field">
          <label htmlFor="ai-api-key" className="input-label">
            API key
          </label>
          <div className="settings-key-row">
            <Input
              id="ai-api-key"
              type="password"
              value={apiKey}
              onChange={(event) => handleApiKeyChange(event.target.value)}
              placeholder={
                needsApiKeyForProvider
                  ? 'Enter an API key for this provider'
                  : hasSavedApiKey
                    ? 'Enter a new key to replace the saved key'
                    : 'Enter an API key'
              }
              autoComplete="off"
              spellCheck={false}
              disabled={isSaving || isLoading}
            />
            {hasSavedApiKey && apiKeyMode !== 'clear' && (
              <Button variant="secondary" onClick={clearApiKey} disabled={isSaving}>
                Clear
              </Button>
            )}
            {apiKeyMode === 'clear' && (
              <Button variant="secondary" onClick={undoClearApiKey} disabled={isSaving}>
                Keep saved key
              </Button>
            )}
          </div>

          {apiKeyMode === 'clear' ? (
            <p className="settings-key-status" role="status">
              The saved key will be removed when you save.
            </p>
          ) : needsApiKeyForProvider ? (
            <p className="settings-key-status">
              <KeyIcon />
              <span>Enter an API key for this provider.</span>
            </p>
          ) : hasSavedApiKey ? (
            <p className="settings-key-status">
              <KeyIcon />
              <span>
                Key saved
                {apiKeyMode === 'replace'
                  ? ' — a new key will replace it when you save.'
                  : '.'}
              </span>
            </p>
          ) : apiKeyMode === 'replace' ? (
            <p className="settings-key-status">
              <KeyIcon />
              <span>A new key will be saved with these settings.</span>
            </p>
          ) : (
            <p className="settings-hint">No key saved yet.</p>
          )}
        </div>
      )}

      <div className="settings-field">
        <div className="settings-label-row">
          <label htmlFor={modelFieldId} className="input-label">
            Model
          </label>
          <Button
            variant="secondary"
            onClick={() => void loadModels()}
            disabled={isSaving || isLoadingModels}
          >
            {isLoadingModels ? 'Loading models…' : 'Load models'}
          </Button>
        </div>

        {modelsError && (
          <ErrorState
            title="Models could not be loaded"
            message={modelsError}
            onRetry={() => void loadModels()}
          />
        )}

        {showModelSelect ? (
          <select
            id="ai-model-select"
            className="select-input"
            value={draft.model}
            onChange={(event) => handleModelSelectChange(event.target.value)}
            disabled={isSaving || isLoading}
          >
            <option value="">Select a model…</option>
            {draft.model && !models.includes(draft.model) && (
              <option value={draft.model}>{draft.model} (current)</option>
            )}
            {models.map((model) => (
              <option key={model} value={model}>
                {model}
              </option>
            ))}
            <option value="__custom__">Custom model…</option>
          </select>
        ) : (
          <Input
            id="ai-model-input"
            value={draft.model}
            onChange={(event) => updateDraft({ model: event.target.value })}
            placeholder="e.g. llama3.2"
            autoComplete="off"
            spellCheck={false}
            disabled={isSaving || isLoading}
          />
        )}

        {!hasModelOptions && (
          <p className="settings-hint">
            {modelsError
              ? 'You can still type the model name manually.'
              : hasLoadedModels
                ? 'No models were returned. Type the model name manually.'
                : 'Load models or type the model name manually.'}
          </p>
        )}

        {isCustomModel && hasModelOptions && (
          <button
            type="button"
            className="settings-text-button"
            onClick={() => setIsCustomModel(false)}
          >
            Choose from the model list instead
          </button>
        )}
      </div>

      <details className="settings-advanced">
        <summary>Advanced</summary>
        <div className="settings-advanced-content">
          <p id="ai-embed-settings-hint" className="settings-hint">
            Optional embedding settings for semantic and hybrid search. Leave the
            embedding base URL blank to reuse the chat endpoint.
          </p>
          <Input
            id="ai-embed-base-url"
            label="Embedding base URL"
            value={draft.embedBaseUrl}
            onChange={(event) => updateDraft({ embedBaseUrl: event.target.value })}
            placeholder="http://localhost:11434/v1"
            autoComplete="off"
            spellCheck={false}
            disabled={isSaving || isLoading}
            aria-describedby="ai-embed-settings-hint"
          />

          <div className="settings-field">
            <div className="settings-label-row">
              <label htmlFor={embeddingModelFieldId} className="input-label">
                Embedding model
              </label>
              <Button
                variant="secondary"
                onClick={() => void loadEmbeddingModels()}
                disabled={isSaving || isLoadingEmbeddingModels}
              >
                {isLoadingEmbeddingModels ? 'Loading models…' : 'Load models'}
              </Button>
            </div>

            <p id="ai-embed-model-help" className="settings-hint">
              Choose a model with &apos;embed&apos; in its name. Chat models like
              gemini-2.5-flash or gpt-4o can&apos;t be used for embeddings. Leaving this
              blank falls back to nomic-embed-text, an Ollama model.
            </p>

            {embeddingModelsError && (
              <div id="ai-embed-models-error" aria-live="polite">
                <ErrorState
                  title="Embedding models could not be loaded"
                  message={embeddingModelsError}
                  onRetry={() => void loadEmbeddingModels()}
                />
              </div>
            )}

            {showEmbeddingModelSelect ? (
              <select
                id="ai-embed-model-select"
                className="select-input"
                value={draft.embedModel}
                onChange={(event) =>
                  handleEmbeddingModelSelectChange(event.target.value)
                }
                disabled={isSaving || isLoading}
                aria-describedby={embedModelDescribedBy || undefined}
              >
                <option value="">Select an embedding model…</option>
                {draft.embedModel &&
                  !embeddingModelsForProvider.includes(draft.embedModel) && (
                    <option value={draft.embedModel}>{draft.embedModel} (current)</option>
                  )}
                {embeddingModelsForProvider.map((model) => (
                  <option key={model} value={model}>
                    {model}
                  </option>
                ))}
                <option value="__custom__">Custom model…</option>
              </select>
            ) : (
              <Input
                id="ai-embed-model-input"
                value={draft.embedModel}
                onChange={(event) => updateDraft({ embedModel: event.target.value })}
                placeholder="nomic-embed-text"
                autoComplete="off"
                spellCheck={false}
                disabled={isSaving || isLoading}
                aria-describedby={embedModelDescribedBy || undefined}
              />
            )}

            <label className="settings-checkbox-row">
              <input
                type="checkbox"
                checked={showAllEmbeddingModels}
                onChange={(event) =>
                  setShowAllEmbeddingModels(event.target.checked)
                }
                disabled={isSaving || isLoading}
                aria-describedby={
                  showAllEmbeddingModels
                    ? 'ai-embed-models-show-all-warning'
                    : undefined
                }
              />
              <span>Show all models</span>
            </label>

            {showAllEmbeddingModels && (
              <WarningNotice id="ai-embed-models-show-all-warning">
                Only pick a model meant for embeddings, or semantic search will fail.
              </WarningNotice>
            )}

            {showNoEmbeddingModels && (
              <WarningNotice id="ai-embed-models-empty" live>
                No embedding models found for this provider. Semantic search will stay
                off. Chat and smart search still work.
              </WarningNotice>
            )}

            {showMissingEmbeddingModelWarning && (
              <WarningNotice id="ai-embed-model-missing-warning">
                Set an embedding model, or semantic search will stay off.
              </WarningNotice>
            )}

            {showNotEmbeddingModelWarning && (
              <WarningNotice id="ai-embed-model-warning">
                This doesn&apos;t look like an embedding model.
              </WarningNotice>
            )}

            {isCustomEmbeddingModel && hasEmbeddingModelOptions && (
              <button
                type="button"
                className="settings-text-button"
                onClick={() => setIsCustomEmbeddingModel(false)}
              >
                Choose from the model list instead
              </button>
            )}
          </div>
        </div>
      </details>

      <div className="settings-actions-row">
        <Button type="submit" disabled={isSaving || isLoading}>
          {isSaving ? 'Saving…' : 'Save AI settings'}
        </Button>
        <Button
          variant="secondary"
          onClick={() => void load()}
          disabled={isSaving || isLoading}
        >
          Reload from server
        </Button>
      </div>
    </form>
  );
}
