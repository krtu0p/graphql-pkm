import { useState } from 'react';
import { Dialog } from './components/common/Dialog';
import { WorkspaceLayout } from './components/layout/WorkspaceLayout';
import { useAiSearch } from './hooks/useAiSearch';
import { useAiSettings, type AiStatus } from './hooks/useAiSettings';
import { useNotesWorkspace } from './hooks/useNotesWorkspace';
import { EditorPage } from './pages/EditorPage';
import { GraphPage } from './pages/GraphPage';
import { NotesPage } from './pages/NotesPage';
import { SearchPage } from './pages/SearchPage';
import { SettingsPage } from './pages/SettingsPage';
import { TagsPage } from './pages/TagsPage';
import type { NavigationSection } from './types/navigation';

export function App() {
  const [activeSection, setActiveSection] = useState<NavigationSection>('notes');
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [librarySearchQuery, setLibrarySearchQuery] = useState('');

  const aiSearch = useAiSearch();
  const aiSettings = useAiSettings();

  const {
    notes,
    activeNote,
    activeNoteId,
    setActiveNoteId,
    backlinks,
    isLoadingNotes,
    notesError,
    isEnriching,
    isLinking,
    linkError,
    notePendingDelete,
    setNotePendingDelete,
    isDeleting,
    deleteError,
    setDeleteError,
    loadNotes,
    saveNote,
    confirmDeleteNote,
    handleCreateLink,
    handleRemoveLink,
    refreshActiveNoteDetails,
  } = useNotesWorkspace();

  const handleNewNote = () => {
    setActiveNoteId(null);
    setActiveSection('editor');
  };

  const handleOpenNote = (noteId: string) => {
    setActiveNoteId(noteId);
    setActiveSection('editor');
  };

  const handleOpenTagSection = (tag: string) => {
    setSelectedTag(tag);
    setActiveSection('tags');
  };

  const handleConfirmDelete = async () => {
    const targetWasActive =
      notePendingDelete && notePendingDelete.id === activeNoteId;
    const deleted = await confirmDeleteNote();
    if (deleted && targetWasActive && activeSection === 'editor') {
      setActiveSection('notes');
    }
  };

  const aiStatus: AiStatus = aiSettings.settings
    ? aiSearch.isAiUnavailable
      ? 'unavailable'
      : aiSettings.settings.enabled
        ? 'on'
        : 'off'
    : aiSettings.isLoading
      ? 'loading'
      : 'unavailable';

  return (
    <WorkspaceLayout
      activeSection={activeSection}
      activeNote={activeNote}
      allNotes={notes}
      backlinks={backlinks}
      selectedTag={selectedTag}
      isEnriching={isEnriching}
      aiStatus={aiStatus}
      onSelectSection={setActiveSection}
      onSelectTagFromSidebar={handleOpenTagSection}
      onNewNote={handleNewNote}
      onSelectNote={handleOpenNote}
      onRemoveLink={handleRemoveLink}
      onRefreshMetadata={
        activeNoteId ? () => void refreshActiveNoteDetails(activeNoteId) : undefined
      }
    >
      {activeSection === 'notes' && (
        <NotesPage
          notes={notes}
          activeNoteId={activeNoteId}
          selectedTag={selectedTag}
          searchQuery={librarySearchQuery}
          isLoading={isLoadingNotes}
          errorMessage={notesError}
          onSearchQueryChange={setLibrarySearchQuery}
          onSelectTag={setSelectedTag}
          onSelectNote={handleOpenNote}
          onCreateNote={handleNewNote}
          onRequestDeleteNote={setNotePendingDelete}
          onRetry={() => void loadNotes()}
        />
      )}

      {activeSection === 'editor' && (
        <EditorPage
          note={activeNote}
          allNotes={notes}
          backlinks={backlinks}
          isEnriching={isEnriching}
          isLinking={isLinking}
          linkError={linkError}
          onSaveNote={saveNote}
          onRequestDeleteNote={setNotePendingDelete}
          onSelectNote={handleOpenNote}
          onBackToNotes={() => setActiveSection('notes')}
          onCreateLink={handleCreateLink}
          onRemoveLink={handleRemoveLink}
          onRefreshMetadata={() => {
            if (activeNoteId) {
              void refreshActiveNoteDetails(activeNoteId);
            }
          }}
        />
      )}

      {activeSection === 'tags' && (
        <TagsPage
          notes={notes}
          selectedTag={selectedTag}
          activeNoteId={activeNoteId}
          isLoadingLibrary={isLoadingNotes}
          onSelectTag={setSelectedTag}
          onSelectNote={handleOpenNote}
          onCreateNote={handleNewNote}
          onRequestDeleteNote={setNotePendingDelete}
        />
      )}

      {activeSection === 'search' && (
        <SearchPage
          notes={notes}
          activeNoteId={activeNoteId}
          aiSearch={aiSearch}
          onSelectNote={handleOpenNote}
          onSelectTag={handleOpenTagSection}
          onCreateNote={handleNewNote}
        />
      )}

      {activeSection === 'graph' && (
        <GraphPage
          notes={notes}
          activeNoteId={activeNoteId}
          isLoading={isLoadingNotes}
          errorMessage={notesError}
          onFocusNote={setActiveNoteId}
          onOpenNote={handleOpenNote}
          onSelectTag={handleOpenTagSection}
          onCreateNote={handleNewNote}
          onRetry={() => void loadNotes()}
        />
      )}

      {activeSection === 'settings' && (
        <SettingsPage
          notesCount={notes.length}
          aiSettings={aiSettings}
          onReloadLibrary={() => void loadNotes()}
          onReturnToNotes={() => setActiveSection('notes')}
          onCreateNote={handleNewNote}
          onSettingsSaved={aiSearch.retryAiModes}
        />
      )}

      <Dialog
        isOpen={Boolean(notePendingDelete)}
        title="Delete Note"
        description={
          notePendingDelete
            ? `Are you sure you want to permanently delete "${notePendingDelete.title || 'Untitled Note'}"? This action cannot be undone.`
            : undefined
        }
        confirmLabel="Delete Note"
        cancelLabel="Cancel"
        isBusy={isDeleting}
        onConfirm={() => void handleConfirmDelete()}
        onClose={() => {
          setDeleteError(null);
          setNotePendingDelete(null);
        }}
      >
        {deleteError && (
          <p className="inline-error-text" role="alert">
            {deleteError}
          </p>
        )}
      </Dialog>
    </WorkspaceLayout>
  );
}

export default App;
