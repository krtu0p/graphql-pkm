import type {
  Backlink,
  Link,
  Note,
  SearchResult,
  SmartSearchResponse,
} from './types';

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value);
}

function asString(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback;
}

function asNullableString(value: unknown): string | null {
  return typeof value === 'string' && value.trim().length > 0 ? value : null;
}

function asStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.filter((item): item is string => typeof item === 'string');
}

export function parseLink(raw: unknown): Link | null {
  if (!isRecord(raw)) return null;
  const id = asString(raw.id);
  const sourceNoteId = asString(raw.sourceNoteId);
  const targetNoteId = asString(raw.targetNoteId);
  if (!id || !sourceNoteId || !targetNoteId) return null;

  return {
    id,
    sourceNoteId,
    targetNoteId,
    description: asNullableString(raw.description),
    createdAt: asString(raw.createdAt),
  };
}

export function parseLinkList(raw: unknown): Link[] {
  if (!Array.isArray(raw)) return [];
  const links: Link[] = [];
  for (const item of raw) {
    const parsed = parseLink(item);
    if (parsed) links.push(parsed);
  }
  return links;
}

export function parseNote(raw: unknown): Note | null {
  if (!isRecord(raw)) return null;
  const id = asString(raw.id);
  if (!id) return null;

  return {
    id,
    title: asString(raw.title, 'Untitled Note'),
    content: asString(raw.content, ''),
    tags: asStringArray(raw.tags),
    createdAt: asString(raw.createdAt),
    updatedAt: asString(raw.updatedAt),
    summary: asNullableString(raw.summary),
    keyConcepts: asStringArray(raw.keyConcepts),
    links: parseLinkList(raw.links),
    backlinks: parseLinkList(raw.backlinks),
  };
}

export function parseNoteList(raw: unknown): Note[] {
  if (!Array.isArray(raw)) return [];
  const notes: Note[] = [];
  for (const item of raw) {
    const parsed = parseNote(item);
    if (parsed) notes.push(parsed);
  }
  return notes;
}

export function parseBacklink(raw: unknown): Backlink | null {
  if (!isRecord(raw)) return null;
  const id = asString(raw.id);
  if (!id || !isRecord(raw.sourceNote)) return null;
  const sourceId = asString(raw.sourceNote.id);
  if (!sourceId) return null;

  return {
    id,
    sourceNote: {
      id: sourceId,
      title: asString(raw.sourceNote.title, 'Untitled Note'),
      updatedAt: asString(raw.sourceNote.updatedAt),
    },
    description: asNullableString(raw.description),
    createdAt: asString(raw.createdAt),
  };
}

export function parseBacklinkList(raw: unknown): Backlink[] {
  if (!Array.isArray(raw)) return [];
  const list: Backlink[] = [];
  for (const item of raw) {
    const parsed = parseBacklink(item);
    if (parsed) list.push(parsed);
  }
  return list;
}

export function parseSearchResult(raw: unknown): SearchResult | null {
  if (!isRecord(raw)) return null;
  const note = parseNote(raw.note);
  if (!note) return null;

  return {
    note,
    score: typeof raw.score === 'number' && Number.isFinite(raw.score) ? raw.score : 0,
    reason: asString(raw.reason),
    matchType: asString(raw.matchType),
  };
}

export function parseSearchResultList(raw: unknown): SearchResult[] {
  if (!Array.isArray(raw)) return [];
  const results: SearchResult[] = [];
  for (const item of raw) {
    const parsed = parseSearchResult(item);
    if (parsed) results.push(parsed);
  }
  return results;
}

export function parseSmartSearchResponse(raw: unknown): SmartSearchResponse {
  if (!isRecord(raw)) {
    return { results: [], explanation: '', gaps: [], connections: [] };
  }
  return {
    results: parseSearchResultList(raw.results),
    explanation: asString(raw.explanation),
    gaps: asStringArray(raw.gaps),
    connections: asStringArray(raw.connections),
  };
}
