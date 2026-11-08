import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  createNote,
  deleteNote,
  fetchNoteBacklinks,
  fetchNoteById,
  fetchNotes,
  linkNotes,
  unlinkNotes,
  updateNote,
  type Backlink,
  type Note,
} from '../lib/graphql';
import { clearDraftBackup } from '../lib/utils/draftStorage';
import type { EditorDraftPayload } from '../pages/EditorPage';

export function useNotesWorkspace() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [activeNoteId, setActiveNoteId] = useState<string | null>(null);
  const [backlinks, setBacklinks] = useState<Backlink[]>([]);
  const [isLoadingNotes, setIsLoadingNotes] = useState(true);
  const [notesError, setNotesError] = useState<string | null>(null);
  const [isEnriching, setIsEnriching] = useState(false);
  const [isLinking, setIsLinking] = useState(false);
  const [linkError, setLinkError] = useState<string | null>(null);
  const [notePendingDelete, setNotePendingDelete] = useState<Note | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const enrichTimersRef = useRef<number[]>([]);

  const clearEnrichTimers = useCallback(() => {
    for (const timer of enrichTimersRef.current) {
      window.clearTimeout(timer);
    }
    enrichTimersRef.current = [];
  }, []);

  useEffect(() => {
    return () => clearEnrichTimers();
  }, [clearEnrichTimers]);

  const loadNotes = useCallback(async (signal?: AbortSignal) => {
    setIsLoadingNotes(true);
    setNotesError(null);
    try {
      const list = await fetchNotes(signal);
      setNotes(list);
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        return;
      }
      setNotesError(
        err instanceof Error ? err.message : 'Unable to load notes from /query.',
      );
    } finally {
      setIsLoadingNotes(false);
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void loadNotes(controller.signal);
    return () => controller.abort();
  }, [loadNotes]);

  const refreshActiveNoteDetails = useCallback(async (noteId: string) => {
    try {
      const [freshNote, freshBacklinks] = await Promise.all([
        fetchNoteById(noteId),
        fetchNoteBacklinks(noteId),
      ]);
      setBacklinks(freshBacklinks);
      if (freshNote) {
        setNotes((prev) => {
          const exists = prev.some((item) => item.id === freshNote.id);
          if (!exists) return [freshNote, ...prev];
          return prev.map((item) => (item.id === freshNote.id ? freshNote : item));
        });
      }
      return freshNote;
    } catch {
      return null;
    }
  }, []);

  useEffect(() => {
    setLinkError(null);
    if (!activeNoteId) {
      setBacklinks([]);
      setIsEnriching(false);
      return;
    }
    const controller = new AbortController();
    fetchNoteBacklinks(activeNoteId, controller.signal)
      .then((items) => setBacklinks(items))
      .catch(() => {
        // Ignore transient backlink fetch errors on switch
      });
    return () => controller.abort();
  }, [activeNoteId]);

  const scheduleEnrichmentPoll = useCallback(
    (noteId: string) => {
      clearEnrichTimers();
      setIsEnriching(true);

      const delays = [3000, 7500];
      delays.forEach((delay, idx) => {
        const timerId = window.setTimeout(async () => {
          const updated = await refreshActiveNoteDetails(noteId);
          const hasMetadata = Boolean(
            updated && (updated.summary || updated.keyConcepts.length > 0),
          );
          if (hasMetadata || idx === delays.length - 1) {
            setIsEnriching(false);
          }
        }, delay);
        enrichTimersRef.current.push(timerId);
      });
    },
    [clearEnrichTimers, refreshActiveNoteDetails],
  );

  const saveNote = useCallback(
    async (noteId: string | null, draft: EditorDraftPayload): Promise<Note> => {
      if (!noteId) {
        const created = await createNote({
          title: draft.title,
          content: draft.content,
          tags: draft.tags,
        });
        setNotes((prev) => [created, ...prev.filter((n) => n.id !== created.id)]);
        setActiveNoteId(created.id);
        if (!created.summary && created.keyConcepts.length === 0) {
          scheduleEnrichmentPoll(created.id);
        }
        return created;
      }

      const updated = await updateNote(noteId, {
        title: draft.title,
        content: draft.content,
        tags: draft.tags,
      });
      setNotes((prev) =>
        prev.map((item) => (item.id === updated.id ? updated : item)),
      );
      if (!updated.summary || updated.keyConcepts.length === 0) {
        scheduleEnrichmentPoll(updated.id);
      }
      return updated;
    },
    [scheduleEnrichmentPoll],
  );

  const confirmDeleteNote = useCallback(async (): Promise<boolean> => {
    if (!notePendingDelete) return false;
    setIsDeleting(true);
    setDeleteError(null);
    const targetId = notePendingDelete.id;
    try {
      const deleted = await deleteNote(targetId);
      if (!deleted) {
        throw new Error('Server did not confirm note deletion.');
      }
      clearDraftBackup(targetId);
      setNotes((prev) => prev.filter((n) => n.id !== targetId));
      if (activeNoteId === targetId) {
        setActiveNoteId(null);
        setBacklinks([]);
      }
      setNotePendingDelete(null);
      return true;
    } catch (err) {
      setDeleteError(
        err instanceof Error ? err.message : 'Failed to delete note.',
      );
      return false;
    } finally {
      setIsDeleting(false);
    }
  }, [notePendingDelete, activeNoteId]);

  const handleCreateLink = useCallback(
    async (targetNoteId: string, description: string) => {
      if (!activeNoteId) return;
      setIsLinking(true);
      setLinkError(null);
      try {
        await linkNotes(activeNoteId, targetNoteId, description);
        await refreshActiveNoteDetails(activeNoteId);
      } catch (err) {
        setLinkError(
          err instanceof Error ? err.message : 'Unable to link notes.',
        );
      } finally {
        setIsLinking(false);
      }
    },
    [activeNoteId, refreshActiveNoteDetails],
  );

  const handleRemoveLink = useCallback(
    async (linkId: string) => {
      if (!activeNoteId) return;
      setIsLinking(true);
      setLinkError(null);
      try {
        await unlinkNotes(linkId);
        await refreshActiveNoteDetails(activeNoteId);
      } catch (err) {
        setLinkError(
          err instanceof Error ? err.message : 'Unable to remove link.',
        );
      } finally {
        setIsLinking(false);
      }
    },
    [activeNoteId, refreshActiveNoteDetails],
  );

  const activeNote = useMemo(
    () => notes.find((n) => n.id === activeNoteId) ?? null,
    [notes, activeNoteId],
  );

  return {
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
  };
}
