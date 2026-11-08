import type { EditorDraftPayload } from '../../pages/EditorPage';

export function areTagsEqual(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false;
  return a.every((tag, index) => tag === b[index]);
}

// isDraftUnchanged compares a draft against the last confirmed server state,
// applying exactly the normalization the backend applies on save: title and
// content are whitespace-trimmed (strings.TrimSpace) and compared exactly,
// so case-only edits are always saved; tags are trimmed and lowercased
// (normalizeTags), so case-only tag edits are redundant.
export function isDraftUnchanged(draft: EditorDraftPayload, saved: EditorDraftPayload): boolean {
  const normalizedTags = (tags: string[]) => tags.map((t) => t.trim().toLowerCase());
  return (
    draft.title.trim() === saved.title.trim() &&
    draft.content.trim() === saved.content.trim() &&
    areTagsEqual(normalizedTags(draft.tags), normalizedTags(saved.tags))
  );
}
