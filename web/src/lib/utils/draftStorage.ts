export interface NoteDraftBackup {
  noteId: string | null;
  title: string;
  content: string;
  tags: string[];
  savedAt: number;
}

const DRAFT_KEY_PREFIX = 'folio_unsaved_draft_';

function getStorageKey(noteId: string | null): string {
  return `${DRAFT_KEY_PREFIX}${noteId ?? 'new'}`;
}

export function saveDraftBackup(draft: NoteDraftBackup): void {
  try {
    const key = getStorageKey(draft.noteId);
    window.localStorage.setItem(key, JSON.stringify(draft));
  } catch {
    // Ignore storage quota or private browsing errors
  }
}

export function loadDraftBackup(noteId: string | null): NoteDraftBackup | null {
  try {
    const key = getStorageKey(noteId);
    const raw = window.localStorage.getItem(key);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    if (!parsed || typeof parsed !== 'object') return null;
    const record = parsed as Record<string, unknown>;
    if (
      typeof record.title !== 'string' ||
      typeof record.content !== 'string' ||
      !Array.isArray(record.tags)
    ) {
      return null;
    }
    const tags = record.tags.filter((t): t is string => typeof t === 'string');
    return {
      noteId: typeof record.noteId === 'string' ? record.noteId : null,
      title: record.title,
      content: record.content,
      tags,
      savedAt: typeof record.savedAt === 'number' ? record.savedAt : Date.now(),
    };
  } catch {
    return null;
  }
}

export function clearDraftBackup(noteId: string | null): void {
  try {
    const key = getStorageKey(noteId);
    window.localStorage.removeItem(key);
  } catch {
    // Ignore storage errors
  }
}
