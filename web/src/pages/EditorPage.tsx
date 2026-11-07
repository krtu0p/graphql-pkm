import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Button } from '../components/common/Button';
import { ErrorState } from '../components/common/ErrorState';
import { NoteConnectionsSection } from '../components/notes/NoteConnectionsSection';
import { NoteEditor } from '../components/notes/NoteEditor';
import { NoteMetadata, type EditorSaveState } from '../components/notes/NoteMetadata';
import { NoteTitle } from '../components/notes/NoteTitle';
import { TagInput } from '../components/notes/TagInput';
import type { Backlink, Note } from '../lib/graphql/types';
import {
  clearDraftBackup,
  loadDraftBackup,
  saveDraftBackup,
} from '../lib/utils/draftStorage';

export interface EditorDraftPayload {
  title: string;
  content: string;
  tags: string[];
}

export interface EditorPageProps {
  note: Note | null;
  allNotes: Note[];
  backlinks: Backlink[];
  isEnriching: boolean;
  isLinking: boolean;
  linkError: string | null;
  onSaveNote: (
    noteId: string | null,
    payload: EditorDraftPayload,
  ) => Promise<Note>;
  onRequestDeleteNote: (note: Note) => void;
  onSelectNote: (noteId: string) => void;
  onBackToNotes: () => void;
  onCreateLink: (targetNoteId: string, description: string) => Promise<void>;
  onRemoveLink: (linkId: string) => Promise<void>;
  onRefreshMetadata: () => void;
}

function areTagsEqual(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  return a.every((tag, index) => tag === b[index]);
}

export function EditorPage({
  note,
  allNotes,
  backlinks,
  isEnriching,
  isLinking,
  linkError,
  onSaveNote,
  onRequestDeleteNote,
  onSelectNote,
  onBackToNotes,
  onCreateLink,
  onRemoveLink,
  onRefreshMetadata,
}: EditorPageProps) {
  const noteId = note?.id ?? null;

  const [title, setTitle] = useState<string>(() => note?.title ?? '');
  const [content, setContent] = useState<string>(() => note?.content ?? '');
  const [tags, setTags] = useState<string[]>(() => note?.tags ?? []);
  const [saveState, setSaveState] = useState<EditorSaveState>(
    note ? 'saved' : 'new',
  );
  const [saveError, setSaveError] = useState<string | null>(null);
  const [restoredDraftNotice, setRestoredDraftNotice] = useState(false);

  const bodyTextareaRef = useRef<HTMLTextAreaElement | null>(null);
  const currentNoteIdRef = useRef<string | null>(noteId);
  const revisionRef = useRef(0);
  const savedRevisionRef = useRef(0);
  const isSavingRef = useRef(false);
  const queuedSaveRef = useRef(false);

  const latestDraftRef = useRef<EditorDraftPayload>({
    title: note?.title ?? '',
    content: note?.content ?? '',
    tags: note?.tags ?? [],
  });
  latestDraftRef.current = { title, content, tags };

  useEffect(() => {
    currentNoteIdRef.current = noteId;
    revisionRef.current = 0;
    savedRevisionRef.current = 0;
    setSaveError(null);

    const backup = loadDraftBackup(noteId);
    const serverUpdatedMs = note?.updatedAt ? new Date(note.updatedAt).getTime() : 0;
    const hasNewerBackup =
      backup &&
      backup.savedAt > serverUpdatedMs &&
      (backup.title !== (note?.title ?? '') ||
        backup.content !== (note?.content ?? '') ||
        !areTagsEqual(backup.tags, note?.tags ?? []));

    if (hasNewerBackup && backup) {
      setTitle(backup.title);
      setContent(backup.content);
      setTags(backup.tags);
      revisionRef.current = 1;
      setSaveState('unsaved');
      setRestoredDraftNotice(true);
    } else {
      setTitle(note?.title ?? '');
      setContent(note?.content ?? '');
      setTags(note?.tags ?? []);
      setSaveState(note ? 'saved' : 'new');
      setRestoredDraftNotice(false);
    }
    // Reset editor state only when switching to a different note ID
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [noteId]);

  const performSave = useCallback(async () => {
    const draft = latestDraftRef.current;
    const targetRevision = revisionRef.current;

    if (!draft.title.trim()) {
      return;
    }

    if (isSavingRef.current) {
      queuedSaveRef.current = true;
      return;
    }

    isSavingRef.current = true;
    setSaveState('saving');
    setSaveError(null);

    try {
      const savedNote = await onSaveNote(currentNoteIdRef.current, draft);
      clearDraftBackup(currentNoteIdRef.current);
      clearDraftBackup(savedNote.id);
      currentNoteIdRef.current = savedNote.id;
      savedRevisionRef.current = Math.max(savedRevisionRef.current, targetRevision);
      setRestoredDraftNotice(false);

      if (revisionRef.current > targetRevision) {
        setSaveState('unsaved');
        queuedSaveRef.current = true;
      } else {
        setSaveState('saved');
      }
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'Unable to save note changes.';
      setSaveError(message);
      setSaveState('error');
    } finally {
      isSavingRef.current = false;
      if (queuedSaveRef.current) {
        queuedSaveRef.current = false;
        if (revisionRef.current > savedRevisionRef.current) {
          void performSave();
        }
      }
    }
  }, [onSaveNote]);

  useEffect(() => {
    if (revisionRef.current <= savedRevisionRef.current) {
      return;
    }
    if (!title.trim()) {
      return;
    }

    const timer = window.setTimeout(() => {
      void performSave();
    }, 1000);

    return () => window.clearTimeout(timer);
  }, [title, content, tags, performSave]);

  const markDirtyAndBackup = (nextDraft: EditorDraftPayload) => {
    revisionRef.current += 1;
    setSaveState('unsaved');
    saveDraftBackup({
      noteId: currentNoteIdRef.current,
      title: nextDraft.title,
      content: nextDraft.content,
      tags: nextDraft.tags,
      savedAt: Date.now(),
    });
  };

  const handleTitleChange = (nextTitle: string) => {
    setTitle(nextTitle);
    markDirtyAndBackup({ title: nextTitle, content, tags });
  };

  const handleContentChange = (nextContent: string) => {
    setContent(nextContent);
    markDirtyAndBackup({ title, content: nextContent, tags });
  };

  const handleTagsChange = (nextTags: string[]) => {
    setTags(nextTags);
    markDirtyAndBackup({ title, content, tags: nextTags });
  };

  const handleManualSave = () => {
    if (!title.trim()) {
      setSaveError('Please enter a title before saving your note.');
      setSaveState('error');
      return;
    }
    revisionRef.current = Math.max(revisionRef.current, savedRevisionRef.current + 1);
    void performSave();
  };

  const wordCount = useMemo(() => {
    const trimmed = content.trim();
    if (!trimmed) return 0;
    return trimmed.split(/\s+/).length;
  }, [content]);

  return (
    <article className="editor-page" aria-label="Note Editor">
      <div className="editor-top-bar">
        <button
          type="button"
          className="icon-btn"
          onClick={onBackToNotes}
        >
          <svg
            width="15"
            height="15"
            viewBox="0 0 20 20"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.7"
            aria-hidden="true"
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M12.5 4.5 7 10l5.5 5.5" />
          </svg>
          <span>All Notes</span>
        </button>

        <div className="editor-top-actions">
          {note && (
            <Button
              variant="secondary"
              onClick={() => onRequestDeleteNote(note)}
            >
              Delete
            </Button>
          )}
          <Button
            variant="primary"
            onClick={handleManualSave}
            disabled={saveState === 'saving' || !title.trim()}
          >
            {saveState === 'saving' ? 'Saving…' : 'Save Note'}
          </Button>
        </div>
      </div>

      {restoredDraftNotice && (
        <div className="draft-restore-banner" role="status">
          <span>Restored unsaved local draft from your browser.</span>
          {note && (
            <button
              type="button"
              className="icon-btn"
              onClick={() => {
                clearDraftBackup(note.id);
                setTitle(note.title);
                setContent(note.content);
                setTags(note.tags);
                revisionRef.current = savedRevisionRef.current;
                setSaveState('saved');
                setRestoredDraftNotice(false);
              }}
            >
              Discard local draft
            </button>
          )}
        </div>
      )}

      {saveError && (
        <ErrorState
          title="Could not save note"
          message={saveError}
          onRetry={handleManualSave}
        />
      )}

      <div className="editor-writing-column">
        <NoteTitle
          value={title}
          onChange={handleTitleChange}
          onEnterKey={() => bodyTextareaRef.current?.focus()}
        />

        <NoteMetadata
          updatedAt={note?.updatedAt ?? null}
          createdAt={note?.createdAt ?? null}
          saveState={saveState}
          saveErrorMessage={saveError}
          wordCount={wordCount}
        />

        <TagInput tags={tags} onChange={handleTagsChange} />

        <NoteEditor
          content={content}
          onChange={handleContentChange}
          onSaveShortcut={handleManualSave}
          textareaRef={bodyTextareaRef}
        />

        {note && (
          <NoteConnectionsSection
            currentNote={note}
            allNotes={allNotes}
            backlinks={backlinks}
            isEnriching={isEnriching}
            isLinking={isLinking}
            linkError={linkError}
            onSelectNote={onSelectNote}
            onCreateLink={onCreateLink}
            onRemoveLink={onRemoveLink}
            onRefreshMetadata={onRefreshMetadata}
          />
        )}
      </div>
    </article>
  );
}
