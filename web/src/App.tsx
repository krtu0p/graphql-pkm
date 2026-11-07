import { useState } from 'react';
import { Dialog } from './components/common/Dialog';
import { WorkspaceLayout } from './components/layout/WorkspaceLayout';
import { useNotesWorkspace } from './hooks/useNotesWorkspace';
import { EditorPage } from './pages/EditorPage';
import { GraphPage } from './pages/GraphPage';
import { NotesPage } from './pages/NotesPage';
import { SearchPage } from './pages/SearchPage';
import { TagsPage } from './pages/TagsPage';
import { NAVIGATION_METADATA, type NavigationSection } from './types/navigation';

export function App() {
  const [activeSection, setActiveSection] = useState<NavigationSection>('notes');
  const [selectedTag, setSelectedTag] = useState<string | null>(null);
  const [librarySearchQuery, setLibrarySearchQuery] = useState('');

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

  const sectionMeta = NAVIGATION_METADATA[activeSection];

  return (
    <WorkspaceLayout
      activeSection={activeSection}
      activeNote={activeNote}
      allNotes={notes}
      backlinks={backlinks}
      selectedTag={selectedTag}
      isEnriching={isEnriching}
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
        <section aria-labelledby="workspace-page-title">
          <div className="page-header">
            <div>
              <h1 id="workspace-page-title" className="page-title">
                {sectionMeta.label}
              </h1>
              <p className="page-description">{sectionMeta.subtitle}</p>
            </div>
            <button type="button" className="btn-primary" onClick={handleNewNote}>
              New Note
            </button>
          </div>

          <div className="editorial-panel">
            <h2>Local-First Workspace Status</h2>
            <p>
              Connected to the local GraphQL endpoint at <code>/query</code>. Notes in library:{' '}
              <strong>{notes.length}</strong>.
            </p>
            <div className="settings-actions-row">
              <button
                type="button"
                className="btn-secondary"
                onClick={() => void loadNotes()}
              >
                Reload Library from Server
              </button>
              <button
                type="button"
                className="btn-secondary"
                onClick={() => setActiveSection('notes')}
              >
                Return to All Notes
              </button>
            </div>
          </div>
        </section>
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
