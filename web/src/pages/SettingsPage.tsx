import { AiSettingsForm } from '../components/settings/AiSettingsForm';
import { Button } from '../components/common/Button';
import { ErrorState } from '../components/common/ErrorState';
import { LoadingState } from '../components/common/LoadingState';
import type { AiSettingsController } from '../hooks/useAiSettings';
import { NAVIGATION_METADATA } from '../types/navigation';

export interface SettingsPageProps {
  notesCount: number;
  aiSettings: AiSettingsController;
  onReloadLibrary: () => void;
  onReturnToNotes: () => void;
  onCreateNote: () => void;
  onSettingsSaved: () => void;
}

export function SettingsPage({
  notesCount,
  aiSettings,
  onReloadLibrary,
  onReturnToNotes,
  onCreateNote,
  onSettingsSaved,
}: SettingsPageProps) {
  const sectionMeta = NAVIGATION_METADATA.settings;
  const { settings, isLoading, loadError, load } = aiSettings;

  return (
    <section className="settings-page" aria-labelledby="settings-page-title">
      <div className="page-header">
        <div>
          <h1 id="settings-page-title" className="page-title">
            {sectionMeta.label}
          </h1>
          <p className="page-description">{sectionMeta.subtitle}</p>
        </div>
        <Button variant="primary" onClick={onCreateNote}>
          New Note
        </Button>
      </div>

      <div className="editorial-panel settings-panel">
        <div className="settings-panel-heading">
          <h2>AI Settings</h2>
          <p className="settings-hint">
            Configure the local or hosted AI provider used for summaries, semantic
            search, and smart search. AI is optional; your notes work without it.
          </p>
        </div>

        {isLoading && !settings ? (
          <LoadingState label="Loading AI settings…" count={2} />
        ) : loadError && !settings ? (
          <ErrorState
            title="AI settings could not be loaded"
            message={loadError}
            onRetry={() => void load()}
          />
        ) : (
          <AiSettingsForm
            controller={aiSettings}
            onSettingsSaved={onSettingsSaved}
          />
        )}
      </div>

      <div className="editorial-panel">
        <h2>Local-First Workspace Status</h2>
        <p>
          Connected to the local GraphQL endpoint at <code>/query</code>. Notes in
          library: <strong>{notesCount}</strong>.
        </p>
        <div className="settings-actions-row">
          <Button variant="secondary" onClick={onReloadLibrary}>
            Reload Library from Server
          </Button>
          <Button variant="secondary" onClick={onReturnToNotes}>
            Return to All Notes
          </Button>
        </div>
      </div>
    </section>
  );
}
