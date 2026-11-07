import type {
  AiModelList,
  AiSettings,
  Backlink,
  Link,
  ListAiModelsInput,
  Note,
  SaveAiSettingsInput,
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

export function parseAiSettings(raw: unknown): AiSettings | null {
  if (!isRecord(raw)) return null;
  if (
    typeof raw.enabled !== 'boolean' ||
    typeof raw.provider !== 'string' ||
    typeof raw.baseUrl !== 'string' ||
    typeof raw.model !== 'string' ||
    typeof raw.hasApiKey !== 'boolean' ||
    typeof raw.embedBaseUrl !== 'string' ||
    typeof raw.embedModel !== 'string'
  ) {
    return null;
  }

  return {
    enabled: raw.enabled,
    provider: raw.provider,
    baseUrl: raw.baseUrl,
    model: raw.model,
    hasApiKey: raw.hasApiKey,
    embedBaseUrl: raw.embedBaseUrl,
    embedModel: raw.embedModel,
  };
}

export function parseAiModelList(raw: unknown): AiModelList {
  if (!isRecord(raw)) {
    return { models: [], error: 'The server returned an invalid AI model list.' };
  }

  // OpenAI-compatible chat requests send the model name as-is. Gemini's model
  // list prefixes IDs with "models/", so normalize to the bare chat name.
  const models = Array.from(
    new Set(
      asStringArray(raw.models)
        .map((model) => model.trim().replace(/^models\//, ''))
        .filter((model) => model.length > 0),
    ),
  );

  return {
    models,
    error: asNullableString(raw.error),
  };
}

function trimTrailingSlashes(value: string): string {
  return value.trim().replace(/\/+$/, '');
}

export function buildSaveAiSettingsInput(input: SaveAiSettingsInput): SaveAiSettingsInput {
  const provider = input.provider.trim();
  const baseUrl = trimTrailingSlashes(input.baseUrl);
  const model = input.model.trim();
  const embedBaseUrl = trimTrailingSlashes(input.embedBaseUrl);
  const embedModel = input.embedModel.trim();

  if (!provider) {
    throw new Error('Choose an AI provider.');
  }
  if (input.enabled && !baseUrl) {
    throw new Error('Enter a base URL before enabling AI.');
  }
  if (input.enabled && !model) {
    throw new Error('Choose or type a model before enabling AI.');
  }

  const cleanInput: SaveAiSettingsInput = {
    enabled: input.enabled,
    provider,
    baseUrl,
    model,
    embedBaseUrl,
    embedModel,
  };
  if (input.apiKey !== undefined) {
    cleanInput.apiKey = input.apiKey.trim();
  }
  return cleanInput;
}

export function buildListAiModelsInput(input: ListAiModelsInput): ListAiModelsInput {
  const cleanInput: ListAiModelsInput = {
    baseUrl: trimTrailingSlashes(input.baseUrl),
  };

  const provider = input.provider?.trim();
  if (provider) {
    cleanInput.provider = provider;
  }
  if (input.apiKey !== undefined) {
    cleanInput.apiKey = input.apiKey.trim();
  }
  return cleanInput;
}
